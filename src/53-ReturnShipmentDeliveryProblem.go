package main

import "github.com/Shopify/sarama"

func ReturnShipmentDeliveryProblemMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ReturnShipmentDeliveryProblem)
	if err != nil {
		return mess, err
	}
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
