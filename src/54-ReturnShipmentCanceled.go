package main

import "github.com/Shopify/sarama"

func ReturnShipmentCanceledMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ReturnShipmentCanceledAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentCanceledProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentCanceledProduce(topic string, payload []byte) error {
	return nil
}
