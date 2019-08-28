package main

import "github.com/Shopify/sarama"

func PayToBuyerFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PayToBuyerFailedAction(message *sarama.ConsumerMessage) error {

	err := PayToBuyerFailedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToBuyerFailedProduce(topic string, payload []byte) error {
	return nil
}
