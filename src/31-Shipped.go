package main

import "github.com/Shopify/sarama"

func ShippedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ShippedAction(message *sarama.ConsumerMessage) error {

	err := ShippedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShippedProduce(topic string, payload []byte) error {
	return nil
}
