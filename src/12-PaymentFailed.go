package main

import "github.com/Shopify/sarama"

func PaymentFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PaymentFailedAction(message *sarama.ConsumerMessage) error {

	err := PaymentFailedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PaymentFailedProduce(topic string, payload []byte) error {
	return nil
}
