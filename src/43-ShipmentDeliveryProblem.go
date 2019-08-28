package main

import "github.com/Shopify/sarama"

func ShipmentDeliveryProblemMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func ShipmentDeliveryProblemAction(message *sarama.ConsumerMessage) error {

	err := ShipmentDeliveryProblemProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentDeliveryProblemProduce(topic string, payload []byte) error {
	return nil
}
