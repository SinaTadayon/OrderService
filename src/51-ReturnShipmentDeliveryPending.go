package main

import "github.com/Shopify/sarama"

func ReturnShipmentDeliveryPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ReturnShipmentDeliveryPending)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func ReturnShipmentDeliveryPendingAction(message *sarama.ConsumerMessage) error {

	err := ReturnShipmentDeliveryPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentDeliveryPendingProduce(topic string, payload []byte) error {
	return nil
}
