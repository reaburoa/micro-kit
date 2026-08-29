package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
	"github.com/reaburoa/micro-kit/protos"
)

func normalizeClientConfig(cfg *protos.NacosClientConfig) {
	if cfg == nil {
		return
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
}

func buildClientConfig(cfg *protos.NacosClientConfig) *constant.ClientConfig {
	if cfg == nil {
		cfg = &protos.NacosClientConfig{}
	}
	normalizeClientConfig(cfg)
	cc := constant.NewClientConfig(
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
	return cc
}

func buildServerConfigs(serverCfgs []*protos.NacosServerConfig) []constant.ServerConfig {
	servers := make([]constant.ServerConfig, 0, len(serverCfgs))
	for _, serverCfg := range serverCfgs {
		if serverCfg == nil || serverCfg.IpAddr == "" {
			continue
		}
		servers = append(servers, *constant.NewServerConfig(serverCfg.IpAddr, serverCfg.Port))
	}
	return servers
}

func buildRegisterInstanceParam(ip string, port uint64, serviceName, groupName string, metadata map[string]string) (vo.RegisterInstanceParam, error) {
	if ip == "" {
		return vo.RegisterInstanceParam{}, errors.New("nacos ip is empty")
	}
	if serviceName == "" {
		return vo.RegisterInstanceParam{}, errors.New("nacos service name is empty")
	}
	if port == 0 {
		port = 80
	}
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}
	if metadata == nil {
		metadata = map[string]string{}
	}
	return vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Metadata:    metadata,
		ServiceName: serviceName,
		GroupName:   groupName,
		Ephemeral:   true,
	}, nil
}

func buildSelectOneHealthInstanceParam(serviceName string) vo.SelectOneHealthInstanceParam {
	return vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
	}
}

func buildServiceInstance(serviceName string, instance *model.Instance) string {
	if instance == nil {
		return fmt.Sprintf("service=%s instance=<nil>", serviceName)
	}
	return fmt.Sprintf("service=%s ip=%s port=%d healthy=%v enabled=%v", serviceName, instance.Ip, instance.Port, instance.Healthy, instance.Enable)
}
