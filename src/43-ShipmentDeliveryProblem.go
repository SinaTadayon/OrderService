package main

import "github.com/Shopify/sarama"

func ShipmentDeliveryProblemMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ShipmentDeliveryProblem)
	if err != nil {
		return mess, err
	}
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
