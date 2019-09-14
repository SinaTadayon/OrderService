package main

import "github.com/Shopify/sarama"

func ReturnShipmentDeliveredMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ReturnShipmentDelivered)
	if err != nil {
		return mess, err
	}
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
