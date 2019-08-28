package main

import "github.com/Shopify/sarama"

func PayToBuyerSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PayToBuyerSuccessAction(message *sarama.ConsumerMessage) error {

	err := PayToBuyerSuccessProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToBuyerSuccessProduce(topic string, payload []byte) error {
	return nil
}
