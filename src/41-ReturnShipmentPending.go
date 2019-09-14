package main

import "github.com/Shopify/sarama"

func ReturnShipmentPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ReturnShipmentPending)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func ReturnShipmentPendingAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentPendingProduce(topic string, payload []byte) error {
	return nil
}
