package main

import "github.com/Shopify/sarama"

func ShipmentDeliveryPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ShipmentDeliveryPending)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func ShipmentDeliveryPendingAction(message *sarama.ConsumerMessage) error {

	err := ShipmentDeliveryPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentDeliveryPendingProduce(topic string, payload []byte) error {
	return nil
}
