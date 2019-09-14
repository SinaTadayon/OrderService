package main

import "github.com/Shopify/sarama"

func ReturnShipmentDetailDelayedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ReturnShipmentDetailDelayed)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func ReturnShipmentDetailDelayedAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentDetailDelayedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentDetailDelayedProduce(topic string, payload []byte) error {
	return nil
}
