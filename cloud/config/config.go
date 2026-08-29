package config

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/env"
)

var (
	defaultConfig          config.Config
	configMu               sync.RWMutex
	nacosConfigOverride    *protos.NacosConfigCenter
	forceLocalConfigSource bool
)

type nacosConfigClient interface {
	GetConfig(param vo.ConfigParam) (string, error)
}

func newNacosConfigClient(cfg *protos.NacosConfigCenter) (nacosConfigClient, error) {
	if cfg == nil || len(cfg.Servers) == 0 {
		return nil, fmt.Errorf("nacos config center is not configured")
	}
	if cfg.Group == "" {
		cfg.Group = "DEFAULT_GROUP"
	}
	if cfg.DataId == "" {
		cfg.DataId = env.ServiceName()
	}
	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  buildNacosClientConfig(cfg.Client),
		ServerConfigs: buildNacosServerConfigs(cfg.Servers),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func buildNacosClientConfig(cfg *protos.NacosClientConfig) *constant.ClientConfig {
	if cfg == nil {
		cfg = &protos.NacosClientConfig{}
	}
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = 5000
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "/tmp/nacos/log"
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = "/tmp/nacos/cache"
	}
	if !cfg.NotLoadCacheAtStart {
		cfg.NotLoadCacheAtStart = true
	}
	if !cfg.UpdateCacheWhenEmpty {
		cfg.UpdateCacheWhenEmpty = true
	}
	return constant.NewClientConfig(
		constant.WithNamespaceId(cfg.NamespaceId),
		constant.WithTimeoutMs(cfg.TimeoutMs),
		constant.WithNotLoadCacheAtStart(cfg.NotLoadCacheAtStart),
		constant.WithUpdateCacheWhenEmpty(cfg.UpdateCacheWhenEmpty),
		constant.WithAccessKey(cfg.AccessKey),
		constant.WithSecretKey(cfg.SecretKey),
		constant.WithLogDir(cfg.LogDir),
		constant.WithCacheDir(cfg.CacheDir),
		constant.WithLogLevel(cfg.LogLevel),
	)
}

func buildNacosServerConfigs(serverCfgs []*protos.NacosServerConfig) []constant.ServerConfig {
	servers := make([]constant.ServerConfig, 0, len(serverCfgs))
	for _, serverCfg := range serverCfgs {
		if serverCfg == nil || serverCfg.IpAddr == "" {
			continue
		}
		servers = append(servers, *constant.NewServerConfig(serverCfg.IpAddr, serverCfg.Port))
	}
	return servers
}

type nacosRemoteSource struct {
	client *nacosConfigClient
	dataID string
	group  string
}

func (n *nacosRemoteSource) Load() ([]*config.KeyValue, error) {
	if n == nil || n.client == nil {
		return nil, nil
	}
	client := *n.client
	content, err := getNacosConfig(client, n.dataID, n.group)
	if err != nil || content == "" {
		return nil, nil
	}
	return []*config.KeyValue{{
		Key:    "nacos-config",
		Value:  []byte(content),
		Format: "yaml",
	}}, nil
}

func getNacosConfig(client nacosConfigClient, dataID, group string) (string, error) {
	var (
		content string
		err     error
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("nacos config read panic: %v", r)
			content = ""
		}
	}()
	content, err = client.GetConfig(vo.ConfigParam{DataId: dataID, Group: group})
	return content, err
}

func (n *nacosRemoteSource) Watch() (config.Watcher, error) {
	return &noopWatcher{}, nil
}

type noopWatcher struct{}

func (n *noopWatcher) Next() ([]*config.KeyValue, error) {
	return nil, context.Canceled
}

func (n *noopWatcher) Stop() error { return nil }

func SetNacosConfigOverride(cfg *protos.NacosConfigCenter) {
	configMu.Lock()
	defer configMu.Unlock()
	nacosConfigOverride = cfg
	forceLocalConfigSource = false
}

func SetLocalConfigOnly() {
	configMu.Lock()
	defer configMu.Unlock()
	forceLocalConfigSource = true
	nacosConfigOverride = nil
}

func buildConfigSources(confPath string) ([]config.Source, error) {
	configPath := path.Join(confPath, fmt.Sprintf("configs/%s", env.GetRuntimeEnv()))
	localSource := file.NewSource(fmt.Sprintf("%s/config.yaml", configPath))
	sources := []config.Source{localSource}
	if !forceLocalConfigSource {
		remoteSource, err := loadNacosConfigSource(configPath, nacosConfigOverride)
		if err != nil {
			return nil, err
		}
		if remoteSource != nil {
			sources = append(sources, remoteSource)
		}
	}
	return sources, nil
}

func loadLocalConfig(confPath string) (config.Config, error) {
	sources, err := buildConfigSources(confPath)
	if err != nil {
		return nil, err
	}
	c := config.New(config.WithSource(sources...))
	if err := c.Load(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return c, nil
}

func loadNacosConfigSource(configPath string, override *protos.NacosConfigCenter) (config.Source, error) {
	cfg := override
	if cfg == nil {
		base := config.New(config.WithSource(file.NewSource(path.Join(configPath, "config.yaml"))))
		if err := base.Load(); err != nil {
			return nil, nil
		}
		var localCfg protos.NacosConfigCenter
		if v := base.Value("nacos.config_center"); v != nil {
			if err := v.Scan(&localCfg); err != nil {
				return nil, nil
			}
		}
		cfg = &localCfg
	}
	if cfg == nil {
		return nil, nil
	}
	if cfg.DataId == "" && cfg.Group == "" && len(cfg.Servers) == 0 {
		return nil, nil
	}
	if cfg.DataId == "" {
		cfg.DataId = env.ServiceName()
	}
	if cfg.Group == "" {
		cfg.Group = "DEFAULT_GROUP"
	}
	if len(cfg.Servers) == 0 {
		return nil, nil
	}
	client, err := newNacosConfigClient(cfg)
	if err != nil {
		return nil, nil
	}
	return &nacosRemoteSource{client: &client, dataID: cfg.DataId, group: cfg.Group}, nil
}

func setConfig(c config.Config) {
	configMu.Lock()
	defer configMu.Unlock()
	defaultConfig = c
}

func InitConfig() error {
	var (
		confPath string
		err      error
	)
	if env.IsDebug() {
		confPath, err = env.GetProjectPath()
		if err != nil {
			return fmt.Errorf("get root path: %w", err)
		}
	}
	conf, err := loadLocalConfig(confPath)
	if err != nil {
		return err
	}
	setConfig(conf)
	return nil
}

func Get(key string) config.Value {
	configMu.RLock()
	defer configMu.RUnlock()
	if defaultConfig == nil {
		return nil
	}
	return defaultConfig.Value(key)
}
