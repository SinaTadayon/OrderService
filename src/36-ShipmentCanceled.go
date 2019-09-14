package main

import "github.com/Shopify/sarama"

func ShipmentCanceledMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ShipmentCanceled)
	if err != nil {
		return mess, err
	}
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
