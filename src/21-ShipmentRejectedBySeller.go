package main

import "github.com/Shopify/sarama"

func ShipmentRejectedBySellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, ShipmentRejectedBySeller)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func ShipmentRejectedBySellerAction(message *sarama.ConsumerMessage) error {

	err := ShipmentRejectedBySellerProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func ShipmentRejectedBySellerProduce(topic string, payload []byte) error {
	return nil
}
