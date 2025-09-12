package iconfluentbroker

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/panjf2000/ants/v2"
	"github.com/reaburoa/micro-kit/utils/log"
	"go.uber.org/zap"
)

// SubscribeTopics 简化消费者创建消费业务场景，只需要提供topic配置以及业务方法，异步方式消费Kafka消息
func SubscribeTopics(topic string, syncPoolNum int, handler func(ctx context.Context, msg *kafka.Message) error) {
	p, err := ants.NewPoolWithFunc(syncPoolNum, func(msg interface{}) {
		kafkaMsg := msg.(*kafka.Message)
		err := handler(context.Background(), kafkaMsg)
		if err != nil {
			log.Error("consumer kafka message failed", zap.Any("msg", string(kafkaMsg.Value)), zap.Error(err))
		}
	})
	if err != nil {
		log.Error("consumer ants poll error", zap.Error(err))
		return
	}
	defer p.Release()

	bk := NewBroker(topic, Handler(func(message *kafka.Message) {
		log.Info("async consumer", zap.String("message on", fmt.Sprintf("%s: %s\n", message.TopicPartition, string(message.Value))))

		err := p.Invoke(message)
		if err != nil {
			log.Error("syncUpdate", zap.Error(err))
		}
	}))

	err = bk.Start(context.Background())
	if err != nil {
		log.Fatalf("start consumer failed with %+v", err)
	}
}

// SubscribeTopicsSync 简化消费者创建消费业务场景，只需要提供topic配置以及业务方法，同步方式消费Kafka消息
func SubscribeTopicsSync(topic string, handler func(ctx context.Context, msg *kafka.Message) error) {
	bk := NewBroker(topic, Handler(func(message *kafka.Message) {
		log.Info("sync consumer", zap.String("message on", fmt.Sprintf("%s: %s\n", message.TopicPartition, string(message.Value))))
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic error, %v", err)
			}
		}()
		err := handler(context.Background(), message)
		if err != nil {
			log.Error("consumer kafka message failed", zap.Any("msg", string(message.Value)), zap.Error(err))
		}
	}))

	err := bk.Start(context.Background())
	if err != nil {
		log.Fatalf("start consumer failed with %+v", err)
	}
}
