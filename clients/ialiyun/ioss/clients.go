package ioss

import (
	"fmt"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/protos"
)

type AliyunOSSManager struct {
	clients sync.Map
}

// AliyunOSSManger is kept as a backward-compatible alias for older callers.
type AliyunOSSManger = AliyunOSSManager

func NewAliyunOSSManger() *AliyunOSSManager {
	return NewAliyunOSSManager()
}

func NewAliyunOSSManager() *AliyunOSSManager {
	return &AliyunOSSManager{clients: sync.Map{}}
}

func (o *AliyunOSSManager) GetOssClient(cfg *protos.OssConfig) (*Client, error) {
	ossClient, err := oss.New(cfg.Endpoint, cfg.AccessKeyId, cfg.AccessKeySecret, oss.Timeout(10, 20))
	if err != nil {
		return nil, err
	}
	return &Client{client: ossClient, endpoint: cfg.Endpoint}, nil
}

func (o *AliyunOSSManager) RegisterAllClient() error {
	ossCfgs := map[string]*protos.OssConfig{}
	err := config.Get("oss").Scan(&ossCfgs)
	if err != nil {
		return err
	}

	for key, cfg := range ossCfgs {
		if err = o.registerClient(key, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (o *AliyunOSSManager) registerClient(key string, cfg *protos.OssConfig) error {
	client, err := o.GetOssClient(cfg)
	if err != nil {
		return err
	}
	o.clients.Store(key, client)
	return nil
}

func (o *AliyunOSSManager) RegisterClient(key string) error {
	cfg := protos.OssConfig{}
	err := config.Get(fmt.Sprintf("oss.%s", key)).Scan(&cfg)
	if err != nil {
		return err
	}
	return o.registerClient(key, &cfg)
}

func (o *AliyunOSSManager) GetClient(key string) (*Client, error) {
	if obj, ok := o.clients.Load(key); ok {
		return obj.(*Client), nil
	}
	return nil, fmt.Errorf("no %s oss client", key)
}

func (o *AliyunOSSManager) GetBucket(key, bucketName string) (*BucketClient, error) {
	client, err := o.GetClient(key)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketName)
}
