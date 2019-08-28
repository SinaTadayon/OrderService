package main

import "github.com/Shopify/sarama"

func ReturnShipmentPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ReturnShipmentPendingAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentPendingProduce(topic string, payload []byte) error {
	return nil
}
