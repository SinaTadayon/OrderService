package main

import "github.com/Shopify/sarama"

func ShipmentSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ShipmentSuccess)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func ShipmentSuccessAction(message *sarama.ConsumerMessage) error {

	err := ShipmentSuccessProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentSuccessProduce(topic string, payload []byte) error {
	return nil
}
