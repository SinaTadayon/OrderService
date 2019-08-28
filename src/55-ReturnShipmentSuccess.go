package main

import "github.com/Shopify/sarama"

func ReturnShipmentSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ReturnShipmentSuccessAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentSuccessProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentSuccessProduce(topic string, payload []byte) error {
	return nil
}
