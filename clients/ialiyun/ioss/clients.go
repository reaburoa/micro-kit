package ioss

import (
	"fmt"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/protos"
)

type AliyunOSSManger struct {
	clients sync.Map
}

func NewAliyunOSSManger() *AliyunOSSManger {
	manger := &AliyunOSSManger{
		clients: sync.Map{},
	}

	return manger
}

func (o *AliyunOSSManger) GetOssClient(cfg *protos.OssConfig) (*oss.Client, error) {
	ossClient, err := oss.New(cfg.Endpoint, cfg.AccessKeyId, cfg.AccessKeySecret, oss.Timeout(10, 20))
	if err != nil {
		return nil, err
	}

	return ossClient, nil
}

func (o *AliyunOSSManger) RegisterAllClient() error {
	ossCfgs := map[string]*protos.OssConfig{}
	err := config.Get("oss").Scan(&ossCfgs)
	if err != nil {
		return err
	}

	for key, cfg := range ossCfgs {
		err = o.registerClient(key, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *AliyunOSSManger) registerClient(key string, cfg *protos.OssConfig) error {
	client, err := o.GetOssClient(cfg)
	if err != nil {
		return err
	}
	o.clients.Store(key, &OssBucket{
		client:   client,
		endpoint: cfg.Endpoint,
	})

	return nil
}

func (o *AliyunOSSManger) RegisterClient(key string) error {
	cfg := protos.OssConfig{}
	err := config.Get(fmt.Sprintf("oss.%s", key)).Scan(&cfg)
	if err != nil {
		return err
	}

	return o.registerClient(key, &cfg)
}

func (o *AliyunOSSManger) GetClient(key string) (*OssBucket, error) {
	if obj, ok := o.clients.Load(key); ok {
		return obj.(*OssBucket), nil
	}

	return nil, fmt.Errorf("no %s oss client", key)
}
