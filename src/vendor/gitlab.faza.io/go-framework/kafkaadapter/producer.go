package kafkaadapter

import (
	"errors"

	"github.com/Shopify/sarama"
)

func (k *Kafka) AddToBulk(key string, val []byte) *Kafka {
	if key == "" {
		k.bulkMessage = append(k.bulkMessage, &sarama.ProducerMessage{
			Topic: k.topic,
			Value: sarama.ByteEncoder(val),
		})
	} else {
		k.bulkMessage = append(k.bulkMessage, &sarama.ProducerMessage{
			Topic: k.topic,
			Key:   sarama.StringEncoder(key),
			Value: sarama.ByteEncoder(val),
		})
	}

	return k
}

func (k *Kafka) SendBulk() error {
	syncProducer, err := sarama.NewSyncProducer(k.brokers, k.Config)
	if err != nil {
		return errors.New("failed to create producer: " + err.Error())
	}

	err = syncProducer.SendMessages(k.bulkMessage)

	if err != nil {
		return errors.New("failed to send message to " + k.topic + err.Error())
	}

	_ = syncProducer.Close()
	return nil
}

// SendOne send single message to kafka brokers.
// Key[OPTIONAL] is the partition key, if provided messages will consume in the order
// Value is the body of the message
func (k *Kafka) SendOne(key string, val []byte) (int32, int64, error) {
	syncProducer, err := sarama.NewSyncProducer(k.brokers, k.Config)
	if err != nil {
		return 0, 0, errors.New("failed to create producer: " + err.Error())
	}

	var body *sarama.ProducerMessage
	if key != "" {
		body = &sarama.ProducerMessage{
			Topic: k.topic,
			Key:   sarama.StringEncoder(key),
			Value: sarama.ByteEncoder(val),
		}
	} else {
		body = &sarama.ProducerMessage{
			Topic: k.topic,
			Value: sarama.StringEncoder(val),
		}
	}

	partition, offset, err := syncProducer.SendMessage(body)

	if err != nil {
		return 0, 0, errors.New("failed to send message to " + k.topic + err.Error())
	}

	_ = syncProducer.Close()
	return partition, offset, nil
}
