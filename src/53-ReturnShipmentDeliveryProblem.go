package main

import "github.com/Shopify/sarama"

func ReturnShipmentDeliveryProblemMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ReturnShipmentDeliveryProblemAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentDeliveryProblemProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentDeliveryProblemProduce(topic string, payload []byte) error {
	return nil
}
