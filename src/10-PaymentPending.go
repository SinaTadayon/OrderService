package main

import "github.com/Shopify/sarama"

func PaymentPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PaymentPendingAction(message *sarama.ConsumerMessage) error {

	err := PaymentPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PaymentPendingProduce(topic string, payload []byte) error {
	return nil
}
