package main

import "github.com/Shopify/sarama"

func ReturnShippedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ReturnShippedAction(message *sarama.ConsumerMessage) error {

	err := ReturnShippedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShippedProduce(topic string, payload []byte) error {
	return nil
}
