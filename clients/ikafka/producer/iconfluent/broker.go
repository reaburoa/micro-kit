package iconfluent

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/cloud/tracer"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/async"
	"github.com/reaburoa/micro-kit/utils/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Client interface {
	Topic() string
	Publish(ctx context.Context, value []byte, key []byte) error
	PublishWithoutKey(ctx context.Context, value []byte) error
	PublishWithEvent(ctx context.Context, value []byte, key []byte, event chan kafka.Event) error
	PublishRaw(ctx context.Context, msg *kafka.Message, event chan kafka.Event) error
	Close()
}

type EventHandlerFunc func(event kafka.Event)

func NewClient(topic string, opts ...Option) (Client, error) {
	var cfg protos.Kafka
	err := config.Get(fmt.Sprintf("kafka.%s", topic)).Scan(&cfg)
	if err != nil {
		return nil, err
	}

	return ConnKafka(&cfg, opts...)
}

func ConnKafka(cfg *protos.Kafka, opts ...Option) (Client, error) {
	log.Info("init kafka producer, it may take a few seconds to init the connection\n")
	// common arguments
	kafkaConf := &kafka.ConfigMap{
		"api.version.request": "true",
		"message.max.bytes":   10000000, // 10MB
		"linger.ms":           10,
		"retries":             3,
		"retry.backoff.ms":    1000,
		"acks":                "1",
		"go.batch.producer":   true,
		// "go.delivery.reports":          withDr,
		// "queue.buffering.max.messages": msgcnt, // 单个broker发送最大消息数
	}
	if cfg.ConfigMap != nil {
		for k, v := range cfg.ConfigMap {
			_ = kafkaConf.SetKey(k, v)
		}
	}
	_ = kafkaConf.SetKey("bootstrap.servers", cfg.Servers)
	switch cfg.Protocol {
	case "plaintext":
		_ = kafkaConf.SetKey("security.protocol", cfg.Protocol)
	case "sasl_ssl":
		_ = kafkaConf.SetKey("security.protocol", cfg.Protocol)
		_ = kafkaConf.SetKey("ssl.ca.location", cfg.CaLocation)
		_ = kafkaConf.SetKey("sasl.username", cfg.Username)
		_ = kafkaConf.SetKey("sasl.password", cfg.Password)
		_ = kafkaConf.SetKey("sasl.mechanism", cfg.Mechanism)
	case "sasl_plaintext":
		_ = kafkaConf.SetKey("security.protocol", cfg.Protocol)
		_ = kafkaConf.SetKey("sasl.username", cfg.Username)
		_ = kafkaConf.SetKey("sasl.password", cfg.Password)
		_ = kafkaConf.SetKey("sasl.mechanism", cfg.Mechanism)
	default:
		return nil, kafka.NewError(kafka.ErrUnknownProtocol, "unknown protocol", true)
	}
	cli := &client{conf: kafkaConf}
	cli.opts = Options{
		Topic: cfg.Topics[0],
	}
	for _, o := range opts {
		o(&cli.opts)
	}
	if err := cli.initProducer(); err != nil {
		return nil, err
	}

	return cli, nil
}

type Options struct {
	Topic        string
	EventHandler EventHandlerFunc
	IsDebug      bool
}

type client struct {
	opts     Options
	conf     *kafka.ConfigMap
	producer *kafka.Producer
}

func (cli *client) initProducer() error {
	producer, err := kafka.NewProducer(cli.conf)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if cli.opts.EventHandler == nil {
		cli.opts.EventHandler = func(e kafka.Event) {}
	}
	async.RunWithRecover(func() {
		for e := range producer.Events() {
			cli.opts.EventHandler(e)
		}
	})
	cli.producer = producer
	log.Info("init kafka producer success\n")
	return nil
}

func (cli *client) Topic() string {
	return cli.opts.Topic
}

func (cli *client) Publish(ctx context.Context, value []byte, key []byte) error {
	return cli.PublishRaw(ctx, &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &cli.opts.Topic, Partition: kafka.PartitionAny},
		Value:          value,
		Key:            key,
	}, nil)
}

func (cli *client) PublishWithoutKey(ctx context.Context, value []byte) error {
	return cli.PublishRaw(ctx, &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &cli.opts.Topic, Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)
}

func (cli *client) PublishWithEvent(ctx context.Context, value []byte, key []byte, event chan kafka.Event) error {
	return cli.PublishRaw(ctx, &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &cli.opts.Topic, Partition: kafka.PartitionAny},
		Value:          value,
		Key:            key,
	}, event)
}

func (cli *client) PublishRaw(ctx context.Context, msg *kafka.Message, event chan kafka.Event) error {
	if tracer.TraceProvider == nil {
		return cli.producer.Produce(msg, event)
	}

	_, span := tracer.TraceProvider.Start(ctx, "kafka-publish", trace.WithSpanKind(trace.SpanKindClient))
	conf := *cli.conf
	span.SetAttributes(attribute.String("Server", conf["bootstrap.servers"].(string)))
	span.SetAttributes(attribute.String("Protocol", conf["security.protocol"].(string)))
	span.SetAttributes(attribute.String("Topic", cli.opts.Topic))
	span.SetAttributes(attribute.String("Key", string(msg.Key)))
	span.SetAttributes(attribute.String("Body", string(msg.Value)))
	span.SetAttributes(attribute.Int("Partition", int(msg.TopicPartition.Partition)))

	defer span.End()

	return cli.producer.Produce(msg, event)
}

func (cli *client) Close() {
	cli.producer.Flush(3 * 1000)
	time.Sleep(1 * time.Second)
	cli.producer.Close()
}

type Option func(opts *Options)

func EventHandler(handler EventHandlerFunc) Option {
	return func(opts *Options) {
		opts.EventHandler = handler
	}
}

func IsDebug(isDebug bool) Option {
	return func(opts *Options) {
		opts.IsDebug = isDebug
	}
}
