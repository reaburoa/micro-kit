package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/log"
)

func LoadServiceRegistry() (*protos.NacosServiceRegistry, error) {
	var cfg protos.NacosServiceRegistry
	if err := config.Get("nacos.service_registry").Scan(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type ServiceRegistryClient struct {
	client naming_client.INamingClient
	cfg    *protos.NacosServiceRegistry
}

func NewServiceRegistryClient() (*ServiceRegistryClient, error) {
	cfg, err := LoadServiceRegistry()
	if err != nil {
		return nil, err
	}
	return NewServiceRegistryClientWithConfig(cfg)
}

func NewServiceRegistryClientWithConfig(cfg *protos.NacosServiceRegistry) (*ServiceRegistryClient, error) {
	if cfg == nil {
		return nil, errors.New("nacos service registry config is nil")
	}
	if len(cfg.Servers) == 0 {
		return nil, errors.New("nacos service registry server list is empty")
	}
	cc := buildClientConfig(cfg.Client)
	sc := buildServerConfigs(cfg.Servers)
	if len(sc) == 0 {
		return nil, errors.New("nacos server config is empty")
	}
	clientParam := vo.NacosClientParam{ClientConfig: cc, ServerConfigs: sc}
	client, err := clients.NewNamingClient(clientParam)
	log.Info("nacos service registry config: ", cfg, " client ", client, " error ", err)
	if err != nil {
		return nil, err
	}
	return &ServiceRegistryClient{client: client, cfg: cfg}, nil
}

// RegisterFactory loads service discovery config from the shared proto schema and creates a naming client.
func RegisterFactory() (naming_client.INamingClient, error) {
	cfg, err := LoadServiceRegistry()
	if err != nil {
		return nil, err
	}
	return RegisterFactoryWithConfig(cfg)
}

// RegisterFactoryWithConfig creates the naming client from the shared proto schema.
func RegisterFactoryWithConfig(cfg *protos.NacosServiceRegistry) (naming_client.INamingClient, error) {
	client, err := NewServiceRegistryClientWithConfig(cfg)
	if err != nil {
		return nil, err
	}
	return client.client, nil
}

func (r *ServiceRegistryClient) RegisterInstance(ip string, port uint64, serviceName string, metadata map[string]string) error {
	if r == nil || r.client == nil {
		return errors.New("nacos service registry client is nil")
	}
	groupName := "DEFAULT_GROUP"
	if r.cfg != nil && r.cfg.GroupName != "" {
		groupName = r.cfg.GroupName
	}
	param, err := buildRegisterInstanceParam(ip, port, serviceName, groupName, metadata)
	if err != nil {
		return err
	}
	if r.cfg != nil && r.cfg.ServiceName != "" {
		param.ServiceName = r.cfg.ServiceName
	}
	if r.cfg != nil && r.cfg.Weight > 0 {
		param.Weight = r.cfg.Weight
	}
	if r.cfg != nil {
		param.Enable = r.cfg.Enable
		param.Healthy = r.cfg.Healthy
		param.Ephemeral = r.cfg.Ephemeral
	}
	_, err = r.client.RegisterInstance(param)
	return err
}

func (r *ServiceRegistryClient) DeregisterInstance(ip string, port uint64, serviceName string) error {
	if r == nil || r.client == nil {
		return errors.New("nacos service registry client is nil")
	}
	if serviceName == "" {
		serviceName = r.cfg.ServiceName
	}
	groupName := "DEFAULT_GROUP"
	if r.cfg != nil && r.cfg.GroupName != "" {
		groupName = r.cfg.GroupName
	}
	_, err := r.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   groupName,
	})
	return err
}

func (r *ServiceRegistryClient) SelectOneHealthyInstance(serviceName string) (*model.Instance, error) {
	if r == nil || r.client == nil {
		return nil, errors.New("nacos service registry client is nil")
	}
	if serviceName == "" && r.cfg != nil {
		serviceName = r.cfg.ServiceName
	}
	groupName := "DEFAULT_GROUP"
	if r.cfg != nil && r.cfg.GroupName != "" {
		groupName = r.cfg.GroupName
	}
	return r.client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{ServiceName: serviceName, GroupName: groupName})
}

func (r *ServiceRegistryClient) Config() *protos.NacosServiceRegistry {
	if r == nil {
		return nil
	}
	return r.cfg
}
