package main

import (
	"strconv"
	"time"

	"gitlab.faza.io/go-framework/logger"

	"github.com/Shopify/sarama"

	"gitlab.faza.io/go-framework/kafkaadapter"
)

func initTopics() error {
	partition, err := strconv.Atoi(App.config.Kafka.Partition)
	if err != nil {
		return err
	}
	replica, err := strconv.Atoi(App.config.Kafka.Replica)
	if err != nil {
		return err
	}

	err = kafkaadapter.CreateTopic(brokers[0],
		App.config.Kafka.ConsumerTopic,
		int32(partition), int16(replica))
	if err != nil {
		if err.Error() != sarama.ErrTopicAlreadyExists.Error() {
			time.Sleep(1 * time.Second)
			_ = initTopics()
		} else if err.Error() == sarama.ErrTopicAlreadyExists.Error() {
			logger.Audit(err.Error())
			return nil
		} else {
			return err
		}
	}
	return nil
}
