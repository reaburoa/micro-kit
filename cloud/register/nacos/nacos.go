package kratos

import (
	"os"
	"path"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
	"github.com/reaburoa/micro-kit/utils/env"
	"github.com/reaburoa/micro-kit/utils/log"
	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	IpAddr string `json:"ip_addr"`
	Port   uint64 `json:"port"`
}

type ClientConfig struct {
	NamespaceId string `json:"namespace_id"`
	TimeoutMs   uint64 `json:"timeout_ms"`
	LogDir      string `json:"log_dir"`
	CacheDir    string `json:"cache_dir"`
	RotateTime  string `json:"rotate_time"`
	MaxAge      int    `json:"max_age"`
	LogLevel    string `json:"log_level"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
}

type AddrConfig struct {
	ServerConfig []ServerConfig `json:"server_configs"`
	ClientConfig ClientConfig   `json:"client_config"`
}

func ParsePluginConfig() (*AddrConfig, error) {
	var pluginConfig *AddrConfig
	configPath := "configs/" + string(env.GetRuntimeEnv())
	if env.IsDebug() {
		rootPath, err := env.GetProjectPath()
		if err != nil {
			panic("get root path " + err.Error())
		}
		configPath = path.Join(rootPath, configPath)
	}
	addrsConfigPath := path.Join(configPath, "plugin.yaml")
	pluginContent, err := os.ReadFile(addrsConfigPath)
	if err != nil {
		return nil, errors.WithMessage(err, "read plugin.yaml failed")
	}
	err = yaml.Unmarshal(pluginContent, &pluginConfig)
	if err != nil {
		return nil, errors.WithMessagef(err, "plugin.yaml content error, content is %s", string(pluginContent))
	}

	return pluginConfig, nil
}

// RegisterFactory 注册
func RegisterFactory() (naming_client.INamingClient, error) {
	pluginConfig, err := ParsePluginConfig()
	if err != nil {
		return nil, err
	}
	if pluginConfig == nil {
		return nil, errors.New("plugin.yaml config is nil,please use initPlugin to init pluginConfig")
	}
	if len(pluginConfig.ServerConfig) <= 0 {
		return nil, errors.New("plugin.yaml config nacos server list is empty")
	}

	return RegisterFactoryWithConfig(pluginConfig.ServerConfig[0], pluginConfig.ClientConfig)
}

// registerFactory 注册
func RegisterFactoryWithConfig(serverCfg ServerConfig, clientCfg ClientConfig) (naming_client.INamingClient, error) {
	cc := constant.NewClientConfig(
		constant.WithNamespaceId(clientCfg.NamespaceId),
		constant.WithTimeoutMs(clientCfg.TimeoutMs),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithUpdateCacheWhenEmpty(true),
		constant.WithAccessKey(clientCfg.AccessKey),
		constant.WithSecretKey(clientCfg.SecretKey),
	)
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(serverCfg.IpAddr, serverCfg.Port),
	}
	clientParam := vo.NacosClientParam{ClientConfig: cc, ServerConfigs: sc}
	client, err := clients.NewNamingClient(clientParam)
	log.Info("register server info ", clientCfg, " client ", client, " error ", err)
	return client, err
}
