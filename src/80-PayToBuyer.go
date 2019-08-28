package main

import "github.com/Shopify/sarama"

func PayToBuyerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PayToBuyerAction(message *sarama.ConsumerMessage) error {

	err := PayToBuyerProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToBuyerProduce(topic string, payload []byte) error {
	return nil
}
