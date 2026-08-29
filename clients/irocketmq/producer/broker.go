package producer

import "context"

type Client interface {
	Topic() string
	Publish(ctx context.Context, value []byte) error
	PublishWithTag(ctx context.Context, value []byte, tag string) error
	Close() error
}
