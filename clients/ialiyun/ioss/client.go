package ioss

import (
	"context"
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.opentelemetry.io/otel/codes"
)

type Client struct {
	client   *oss.Client
	endpoint string
}

func NewOssBucket(client *oss.Client) *Client {
	return &Client{client: client}
}

func NewClient(client *oss.Client) *Client {
	return &Client{client: client}
}

func (c *Client) Bucket(name string) (*BucketClient, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("oss client is not initialized")
	}
	ctx, span := startOSSSpan(context.Background(), "oss-bucket-open", name, "")
	defer span.End()
	bk, err := c.client.Bucket(name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	_ = ctx
	return &BucketClient{bucket: bk, bucketName: name}, nil
}
