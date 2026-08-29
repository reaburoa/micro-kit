package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/log"
)

func LoadConfigCenter() (*protos.NacosConfigCenter, error) {
	var cfg protos.NacosConfigCenter
	if err := config.Get("nacos.config_center").Scan(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type ConfigCenterClient struct {
	client config_client.IConfigClient
	cfg    *protos.NacosConfigCenter
}

func NewConfigCenterClient() (*ConfigCenterClient, error) {
	cfg, err := LoadConfigCenter()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, errors.New("nacos config center config is nil")
	}
	if len(cfg.Servers) == 0 {
		return nil, errors.New("nacos config center server list is empty")
	}

	cc := buildClientConfig(cfg.Client)
	servers := buildServerConfigs(cfg.Servers)
	if len(servers) == 0 {
		return nil, errors.New("nacos config center server config is empty")
	}

	client, err := clients.NewConfigClient(vo.NacosClientParam{ClientConfig: cc, ServerConfigs: servers})
	log.Info("nacos config center config: ", cfg, " client ", client, " error ", err)
	if err != nil {
		return nil, err
	}
	return &ConfigCenterClient{client: client, cfg: cfg}, nil
}

func NewConfigClient() (config_client.IConfigClient, error) {
	client, err := NewConfigCenterClient()
	if err != nil {
		return nil, err
	}
	return client.client, nil
}

func (c *ConfigCenterClient) GetConfig(dataID, group string) (string, error) {
	if c == nil || c.client == nil {
		return "", errors.New("nacos config client is nil")
	}
	return c.client.GetConfig(vo.ConfigParam{DataId: dataID, Group: group})
}

func (c *ConfigCenterClient) PublishConfig(dataID, group, content string) (bool, error) {
	if c == nil || c.client == nil {
		return false, errors.New("nacos config client is nil")
	}
	return c.client.PublishConfig(vo.ConfigParam{DataId: dataID, Group: group, Content: content})
}

func (c *ConfigCenterClient) DeleteConfig(dataID, group string) (bool, error) {
	if c == nil || c.client == nil {
		return false, errors.New("nacos config client is nil")
	}
	return c.client.DeleteConfig(vo.ConfigParam{DataId: dataID, Group: group})
}

func (c *ConfigCenterClient) Config() *protos.NacosConfigCenter {
	if c == nil {
		return nil
	}
	return c.cfg
}
