package main

import "github.com/Shopify/sarama"

func ShipmentDeliveryDelayedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ShipmentDeliveryDelayedAction(message *sarama.ConsumerMessage) error {

	err := ShipmentDeliveryDelayedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentDeliveryDelayedProduce(topic string, payload []byte) error {
	return nil
}
