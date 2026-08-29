package server

import (
	"fmt"
	"net"

	"github.com/reaburoa/micro-kit/cloud/register/nacos"
	"github.com/reaburoa/micro-kit/utils/log"
)

func EnableNacosRegistry(conf *Server) error {
	if conf == nil || !conf.EnableNacos {
		return nil
	}
	if conf.Name == "" {
		return fmt.Errorf("server name is empty")
	}
	if conf.Port <= 0 {
		return fmt.Errorf("server port is empty")
	}

	client, err := nacos.NewServiceRegistryClient()
	if err != nil {
		return err
	}
	if client == nil || client.Config() == nil {
		return fmt.Errorf("nacos service registry config is nil")
	}

	ip, err := localIP()
	if err != nil {
		return err
	}
	groupName := conf.GroupName
	if groupName == "" && client.Config().GroupName != "" {
		groupName = client.Config().GroupName
	}
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}
	cfg := client.Config()
	if cfg.ServiceName == "" {
		cfg.ServiceName = conf.Name
	}
	if cfg.GroupName == "" {
		cfg.GroupName = groupName
	}

	return client.RegisterInstance(ip, uint64(conf.Port), conf.Name, map[string]string{
		"server_name": conf.Name,
		"network":     conf.Network,
	})
}

func DiscoverNacosService(serviceName string) (*nacos.ServiceRegistryClient, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name is empty")
	}
	client, err := nacos.NewServiceRegistryClient()
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("nacos service registry client is nil")
	}
	_, err = client.SelectOneHealthyInstance(serviceName)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func localIP() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, a := range addr {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}
		return ipNet.IP.String(), nil
	}
	return "", fmt.Errorf("get local ip failed")
}

func AutoRegisterWithNacos(conf *Server) error {
	if conf == nil {
		return nil
	}
	if !conf.EnableNacos {
		log.Info("nacos registration is disabled by default")
		return nil
	}
	return EnableNacosRegistry(conf)
}
