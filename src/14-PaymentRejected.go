package main

import "github.com/Shopify/sarama"

func PaymentRejectedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PaymentRejectedAction(message *sarama.ConsumerMessage) error {

	err := PaymentRejectedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PaymentRejectedProduce(topic string, payload []byte) error {
	return nil
}
