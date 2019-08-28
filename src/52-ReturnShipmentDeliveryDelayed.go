package main

import "github.com/Shopify/sarama"

func ReturnShipmentDeliveryDelayedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ReturnShipmentDeliveryDelayedAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentDeliveryDelayedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentDeliveryDelayedProduce(topic string, payload []byte) error {
	return nil
}
