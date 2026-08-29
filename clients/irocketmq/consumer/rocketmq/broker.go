package rocketmq

import (
	"context"
	"fmt"
	"strings"
	"time"

	rmq "github.com/apache/rocketmq-client-go/v2"
	rmqConsumer "github.com/apache/rocketmq-client-go/v2/consumer"
	rmqPrimitive "github.com/apache/rocketmq-client-go/v2/primitive"
	broker "github.com/reaburoa/micro-kit/clients/irocketmq/consumer"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/cloud/tracer"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type HandlerFunc func(message *rmqPrimitive.MessageExt) error

type Options struct {
	Topic          string
	ConsumerGroup  string
	ConsumeTimeout time.Duration
	Handler        HandlerFunc
	Tags           []string
	IsDebug        bool
}

type Option func(opts *Options)

func Handler(handler func(message *rmqPrimitive.MessageExt) error) Option {
	return func(opts *Options) {
		opts.Handler = handler
	}
}

func WithConsumerGroup(group string) Option {
	return func(opts *Options) {
		opts.ConsumerGroup = group
	}
}

func WithTags(tags ...string) Option {
	return func(opts *Options) {
		opts.Tags = append(opts.Tags, tags...)
	}
}

func IsDebug(v bool) Option {
	return func(opts *Options) {
		opts.IsDebug = v
	}
}

func NewBroker(topic string, opts ...Option) broker.Broker {
	var cfg protos.RocketMQ
	if err := config.Get(fmt.Sprintf("rocketmq.%s", topic)).Scan(&cfg); err != nil {
		log.Fatalf("get topic %s rocketmq config failed: %v", topic, err)
	}
	defaultOpts := Options{
		Topic:         topic,
		ConsumerGroup: cfg.ConsumerGroup,
		Tags:          []string{"*"},
	}
	for _, o := range opts {
		o(&defaultOpts)
	}
	if defaultOpts.ConsumerGroup == "" {
		defaultOpts.ConsumerGroup = cfg.ConsumerGroup
	}
	if defaultOpts.ConsumerGroup == "" {
		defaultOpts.ConsumerGroup = "micro-kit-consumer"
	}
	if len(defaultOpts.Tags) == 0 {
		defaultOpts.Tags = []string{"*"}
	}
	if defaultOpts.ConsumeTimeout <= 0 {
		if cfg.ConsumeTimeout != "" {
			d, err := time.ParseDuration(cfg.ConsumeTimeout)
			if err == nil {
				defaultOpts.ConsumeTimeout = d
			}
		}
	}
	if defaultOpts.ConsumeTimeout <= 0 {
		defaultOpts.ConsumeTimeout = 15 * time.Minute
	}
	addrList := cfg.NamesrvAddrs
	if len(addrList) == 0 && cfg.NamesrvAddr != "" {
		addrList = strings.Split(cfg.NamesrvAddr, ",")
	}
	if len(addrList) == 0 {
		log.Fatalf("topic %s rocketmq config missing namesrv addrs", topic)
	}
	nameServer, err := rmqPrimitive.NewNamesrvAddr(addrList...)
	if err != nil {
		log.Fatalf("parse rocketmq namesrv addrs for topic %s failed: %v", topic, err)
	}
	consumerOptions := []rmqConsumer.Option{
		rmqConsumer.WithNameServer(nameServer),
		rmqConsumer.WithGroupName(defaultOpts.ConsumerGroup),
		rmqConsumer.WithConsumeTimeout(defaultOpts.ConsumeTimeout),
	}
	if cfg.ConsumeMessageBatchMaxSize > 0 {
		consumerOptions = append(consumerOptions, rmqConsumer.WithConsumeMessageBatchMaxSize(int(cfg.ConsumeMessageBatchMaxSize)))
	}
	if cfg.PullBatchSize > 0 {
		consumerOptions = append(consumerOptions, rmqConsumer.WithPullBatchSize(cfg.PullBatchSize))
	}
	consumer, err := rmq.NewPushConsumer(consumerOptions...)
	if err != nil {
		log.Fatalf("init rocketmq consumer failed for topic %s: %v", topic, err)
	}
	bk := &rocketmqBK{consumer: consumer, opts: defaultOpts}
	return bk
}

type rocketmqBK struct {
	opts     Options
	consumer rmq.PushConsumer
}

func (bk *rocketmqBK) Start(ctx context.Context) error {
	if bk.opts.Handler == nil {
		return fmt.Errorf("rocketmq consumer handler is nil")
	}
	if err := bk.consumer.Start(); err != nil {
		return err
	}
	return bk.consumer.Subscribe(bk.opts.Topic, rmqConsumer.MessageSelector{Type: rmqConsumer.TAG, Expression: "*"}, func(ctx context.Context, msgs ...*rmqPrimitive.MessageExt) (rmqConsumer.ConsumeResult, error) {
		if tracer.TraceProvider == nil {
			for _, msg := range msgs {
				if err := bk.opts.Handler(msg); err != nil {
					return rmqConsumer.ConsumeRetryLater, err
				}
			}
			return rmqConsumer.ConsumeSuccess, nil
		}
		ctx, span := tracer.TraceProvider.Start(ctx, "rocketmq-consume", trace.WithSpanKind(trace.SpanKindConsumer))
		defer span.End()
		span.SetAttributes(
			attribute.String("messaging.system", "rocketmq"),
			attribute.String("messaging.destination", bk.opts.Topic),
			attribute.String("messaging.operation", "receive"),
			attribute.Int("rocketmq.message_count", len(msgs)),
		)
		for _, msg := range msgs {
			if msg == nil {
				continue
			}
			span.SetAttributes(
				attribute.String("rocketmq.topic", msg.Topic),
				attribute.String("rocketmq.tag", msg.GetTags()),
				attribute.String("rocketmq.keys", msg.GetKeys()),
			)
			if err := bk.opts.Handler(msg); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return rmqConsumer.ConsumeRetryLater, err
			}
		}
		span.SetStatus(codes.Ok, "")
		_ = ctx
		return rmqConsumer.ConsumeSuccess, nil
	})
}

func (bk *rocketmqBK) Stop() error {
	return bk.consumer.Shutdown()
}

func SubscribeTopics(topic string, handler func(ctx context.Context, msg *rmqPrimitive.MessageExt) error) {
	bk := NewBroker(topic, Handler(func(msg *rmqPrimitive.MessageExt) error {
		return handler(context.Background(), msg)
	}))
	if err := bk.Start(context.Background()); err != nil {
		log.Fatalf("start rocketmq consumer failed with %+v", err)
	}
}

func SubscribeTopicsSync(topic string, handler func(ctx context.Context, msg *rmqPrimitive.MessageExt) error) {
	SubscribeTopics(topic, handler)
}
