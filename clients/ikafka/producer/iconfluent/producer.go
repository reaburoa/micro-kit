package iconfluent

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/reaburoa/micro-kit/utils/log"
)

func NewProducer(topic string) (Client, error) {
	return NewClient(topic, EventHandler(func(e kafka.Event) {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				log.Error("发送失败", m.TopicPartition.Error)
			}
			return
		default:
			log.Error("Ignored event", ev)
		}
	}))
}
