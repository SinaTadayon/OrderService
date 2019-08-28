package main

import "github.com/Shopify/sarama"

func ShipmentDetailDelayedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ShipmentDetailDelayedAction(message *sarama.ConsumerMessage) error {

	err := ShipmentDetailDelayedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentDetailDelayedProduce(topic string, payload []byte) error {
	return nil
}
