package main

import "github.com/Shopify/sarama"

func ShipmentDeliveredMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ShipmentDelivered)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func ShipmentDeliveredAction(message *sarama.ConsumerMessage) error {

	err := ShipmentDeliveredProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentDeliveredProduce(topic string, payload []byte) error {
	return nil
}
