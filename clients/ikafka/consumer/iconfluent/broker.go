package iconfluentbroker

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	broker "github.com/reaburoa/micro-kit/clients/ikafka/consumer"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/async"
	"github.com/reaburoa/micro-kit/utils/log"
)

type HandlerFunc func(message *kafka.Message)

type Config struct {
	Topics           []string                     `json:"topics"`
	GroupId          string                       `json:"group.id"`
	BootstrapServers string                       `json:"bootstrap.servers"`
	SecurityProtocol string                       `json:"security.protocol"`
	SaslMechanism    string                       `json:"sasl.mechanism"`
	SaslUsername     string                       `json:"sasl.username"`
	SaslPassword     string                       `json:"sasl.password"`
	SslCaLocation    string                       `json:"ssl.ca.location"`
	ConfigMap        map[string]kafka.ConfigValue `json:"config.map"`
}

func NewBroker(topic string, opts ...Option) broker.Broker {
	var cfg protos.Kafka
	err := config.Get(fmt.Sprintf("kafka.%s", topic)).Scan(&cfg)
	if err != nil {
		log.Fatalf("get topic %s consumer config failed", topic)
	}
	// common arguments
	kafkaConf := &kafka.ConfigMap{
		"api.version.request":       "true",
		"auto.offset.reset":         "earliest", // 默认从最早上次消费
		"heartbeat.interval.ms":     3000,
		"session.timeout.ms":        30000,
		"max.poll.interval.ms":      120000,
		"fetch.max.bytes":           1024000,
		"max.partition.fetch.bytes": 256000,
	}
	if cfg.AutoOffsetReset != "" {
		_ = kafkaConf.SetKey("auto.offset.reset", cfg.AutoOffsetReset)
	}
	if cfg.ConfigMap != nil {
		for k, v := range cfg.ConfigMap {
			_ = kafkaConf.SetKey(k, v)
		}
	}
	_ = kafkaConf.SetKey("bootstrap.servers", cfg.Servers)
	_ = kafkaConf.SetKey("group.id", cfg.GroupId)
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
		panic(kafka.NewError(kafka.ErrUnknownProtocol, "unknown protocol", true))
	}
	consumer, err := kafka.NewConsumer(kafkaConf)
	if err != nil {
		log.Fatalf("topic %s to init consumer failed with error %#v", topic, err.Error())
	}
	log.Info("init kafka consumer success")
	bk := &kafkaBK{consumer: consumer}
	bk.opts = Options{
		Topics: cfg.Topics,
	}
	for _, o := range opts {
		o(&bk.opts)
	}
	bk.notifyClose = make(chan error)
	return bk
}

type kafkaBK struct {
	opts        Options
	consumer    *kafka.Consumer
	notifyClose chan error
}

func (bk *kafkaBK) Start(ctx context.Context) error {
	return async.RunWithContext(ctx, bk.start)
}

func (bk *kafkaBK) start() error {
	err := bk.consumer.SubscribeTopics(bk.opts.Topics, bk.opts.RebalanceCb)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	for {
		ev := bk.consumer.Poll(-1)
		switch e := ev.(type) {
		case *kafka.Message:
			if e.TopicPartition.Error != nil {
				log.Errorf("Consumer error: %+v (%+v)\n", e.TopicPartition.Error, e)
			}
			bk.opts.Handler(e)
		case kafka.Error:
			log.Errorf("Consumer error: %+v\n", e)
		default:
			// Ignore other event types
		}
		select {
		case err = <-bk.notifyClose:
			return err
		default:
			// Ignore
		}
	}
}

func (bk *kafkaBK) Stop() error {
	close(bk.notifyClose)
	// bk.KafkaPool.
	return bk.consumer.Close()
}
