package main

import "github.com/Shopify/sarama"

func PaymentSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PaymentSuccessAction(message *sarama.ConsumerMessage) error {

	err := PaymentSuccessProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PaymentSuccessProduce(topic string, payload []byte) error {
	return nil
}
