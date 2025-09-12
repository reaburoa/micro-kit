package ioss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OssBucket struct {
	client   *oss.Client
	endpoint string      `json:"endpoint"`
	bucket   *oss.Bucket `json:"bucket"`
}

func NewOssBucket(client *oss.Client) *OssBucket {
	return &OssBucket{
		client: client,
	}
}

func (o *OssBucket) GetBucket(bucket string) (*oss.Bucket, error) {
	return o.client.Bucket(bucket)
}
