package rocketmq

import (
	"context"
	"fmt"
	"strings"
	"time"

	rmq "github.com/apache/rocketmq-client-go/v2"
	rmqPrimitive "github.com/apache/rocketmq-client-go/v2/primitive"
	rmqProducer "github.com/apache/rocketmq-client-go/v2/producer"
	prod "github.com/reaburoa/micro-kit/clients/irocketmq/producer"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/cloud/tracer"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const defaultTopic = "DEFAULT"

type MessageOption func(*rmqPrimitive.Message)

func WithMsgTag(tag string) MessageOption {
	return func(msg *rmqPrimitive.Message) {
		if tag != "" {
			msg.WithTag(tag)
		}
	}
}

func WithKeys(keys ...string) MessageOption {
	return func(msg *rmqPrimitive.Message) {
		filtered := make([]string, 0, len(keys))
		for _, key := range keys {
			if key != "" {
				filtered = append(filtered, key)
			}
		}
		if len(filtered) > 0 {
			msg.WithKeys(filtered)
			msg.WithShardingKey(filtered[0])
		}
	}
}

func WithDelay(delay time.Duration) MessageOption {
	return func(msg *rmqPrimitive.Message) {
		if delay > 0 {
			if level := delayLevel(delay); level > 0 {
				msg.WithDelayTimeLevel(level)
			}
		}
	}
}

type Options struct {
	Topic       string
	GroupName   string
	NamesrvAddr []string
	Tag         string
	Retry       int
}

type Option func(opts *Options)

func WithGroupName(group string) Option {
	return func(opts *Options) { opts.GroupName = group }
}

func WithNamesrvAddrs(addrs ...string) Option {
	return func(opts *Options) { opts.NamesrvAddr = append(opts.NamesrvAddr, addrs...) }
}

func WithTag(tag string) Option {
	return func(opts *Options) { opts.Tag = tag }
}

func WithRetry(retry int) Option {
	return func(opts *Options) { opts.Retry = retry }
}

func NewClient(topic string, opts ...Option) (prod.Client, error) {
	var cfg protos.RocketMQ
	if err := config.Get(fmt.Sprintf("rocketmq.%s", topic)).Scan(&cfg); err != nil {
		return nil, err
	}
	defaultOpts := Options{Topic: topic, Tag: cfg.Tag, Retry: int(cfg.Retry), GroupName: cfg.ProducerGroup}
	for _, o := range opts {
		o(&defaultOpts)
	}
	if defaultOpts.GroupName == "" {
		defaultOpts.GroupName = "micro-kit-producer"
	}
	if defaultOpts.Retry <= 0 {
		defaultOpts.Retry = 3
	}
	if len(defaultOpts.NamesrvAddr) == 0 {
		if len(cfg.NamesrvAddrs) > 0 {
			defaultOpts.NamesrvAddr = cfg.NamesrvAddrs
		} else if cfg.NamesrvAddr != "" {
			defaultOpts.NamesrvAddr = strings.Split(cfg.NamesrvAddr, ",")
		}
	}
	if len(defaultOpts.NamesrvAddr) == 0 {
		return nil, fmt.Errorf("rocketmq producer config for topic %q is missing namesrv addrs", topic)
	}
	addr, err := rmqPrimitive.NewNamesrvAddr(defaultOpts.NamesrvAddr...)
	if err != nil {
		return nil, fmt.Errorf("invalid rocketmq namesrv addrs: %w", err)
	}
	p, err := rmq.NewProducer(
		rmqProducer.WithNameServer(addr),
		rmqProducer.WithGroupName(defaultOpts.GroupName),
		rmqProducer.WithRetry(defaultOpts.Retry),
	)
	if err != nil {
		return nil, fmt.Errorf("new rocketmq producer: %w", err)
	}
	if err = p.Start(); err != nil {
		return nil, fmt.Errorf("start rocketmq producer: %w", err)
	}
	return &client{topic: defaultOpts.Topic, group: defaultOpts.GroupName, producer: p, tag: defaultOpts.Tag}, nil
}

type client struct {
	topic    string
	group    string
	tag      string
	producer rmq.Producer
}

func (c *client) Topic() string { return c.topic }

func (c *client) Publish(ctx context.Context, value []byte) error {
	return c.PublishWithTag(ctx, value, c.tag)
}

func (c *client) PublishWithTag(ctx context.Context, value []byte, tag string) error {
	return c.sendMessage(ctx, value, WithMsgTag(tag))
}

func (c *client) PublishWithKey(ctx context.Context, value []byte, key string) error {
	return c.sendMessage(ctx, value, WithKeys(key))
}

func (c *client) PublishWithDelay(ctx context.Context, value []byte, delay time.Duration, tag string) error {
	return c.sendMessage(ctx, value, WithMsgTag(tag), WithDelay(delay))
}

func (c *client) sendMessage(ctx context.Context, value []byte, opts ...MessageOption) error {
	if len(value) == 0 {
		return nil
	}
	msg := newMessage(c.topic, value, opts...)
	ctx, span := startRocketMQProducerSpan(ctx, "rocketmq-publish", c.topic, msg.GetTags(), msg.GetKeys())
	defer span.End()

	_, err := c.producer.SendSync(ctx, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Errorf("rocketmq producer send sync failed: %v", err)
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

func startRocketMQProducerSpan(ctx context.Context, operation, topic, tag, keys string) (context.Context, trace.Span) {
	if tracer.TraceProvider == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	ctx, span := tracer.TraceProvider.Start(ctx, operation, trace.WithSpanKind(trace.SpanKindProducer))
	span.SetAttributes(
		attribute.String("messaging.system", "rocketmq"),
		attribute.String("messaging.destination", topic),
		attribute.String("messaging.operation", "publish"),
		attribute.String("rocketmq.tag", tag),
		attribute.String("rocketmq.keys", keys),
	)
	return ctx, span
}

func (c *client) Close() error { return c.producer.Shutdown() }

func newMessage(topic string, body []byte, opts ...MessageOption) *rmqPrimitive.Message {
	if topic == "" {
		topic = defaultTopic
	}
	msg := rmqPrimitive.NewMessage(topic, body)
	for _, opt := range opts {
		if opt != nil {
			opt(msg)
		}
	}
	return msg
}

func delayLevel(delay time.Duration) int {
	if delay <= 0 {
		return 0
	}

	for _, level := range []struct {
		threshold time.Duration
		value     int
	}{
		{time.Second, 1},
		{5 * time.Second, 2},
		{10 * time.Second, 3},
		{30 * time.Second, 4},
		{time.Minute, 5},
		{2 * time.Minute, 6},
		{3 * time.Minute, 7},
		{4 * time.Minute, 8},
		{5 * time.Minute, 9},
		{6 * time.Minute, 10},
		{7 * time.Minute, 11},
		{8 * time.Minute, 12},
		{9 * time.Minute, 13},
		{10 * time.Minute, 14},
		{20 * time.Minute, 15},
		{30 * time.Minute, 16},
		{time.Hour, 17},
		{2 * time.Hour, 18},
	} {
		if delay <= level.threshold {
			return level.value
		}
	}
	return 18
}
