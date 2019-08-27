package kafkaadapter

import (
	"github.com/Shopify/sarama"
)

type Kafka struct {
	brokers     []string
	topic       string
	Config      *sarama.Config
	bulkMessage []*sarama.ProducerMessage
}

func NewKafka(broker []string, topic string) *Kafka {
	return &Kafka{
		brokers: broker,
		topic:   topic,
		Config:  sarama.NewConfig(),
	}
}
