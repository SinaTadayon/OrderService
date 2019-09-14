package main

import "github.com/Shopify/sarama"

func PayToMarketFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PayToMarketFailed)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PayToMarketFailedAction(message *sarama.ConsumerMessage) error {

	err := PayToMarketFailedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToMarketFailedProduce(topic string, payload []byte) error {
	return nil
}
