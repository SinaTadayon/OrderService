package main

import "github.com/Shopify/sarama"

func ShipmentDeliveryPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ShipmentDeliveryPendingAction(message *sarama.ConsumerMessage) error {

	err := ShipmentDeliveryPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentDeliveryPendingProduce(topic string, payload []byte) error {
	return nil
}
