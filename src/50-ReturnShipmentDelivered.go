package main

import "github.com/Shopify/sarama"

func ReturnShipmentDeliveredMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ReturnShipmentDeliveredAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentDeliveredProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentDeliveredProduce(topic string, payload []byte) error {
	return nil
}
