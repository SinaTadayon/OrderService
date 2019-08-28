package main

import "github.com/Shopify/sarama"

func ShipmentCanceledMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ShipmentCanceledAction(message *sarama.ConsumerMessage) error {

	err := ShipmentCanceledProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentCanceledProduce(topic string, payload []byte) error {
	return nil
}
