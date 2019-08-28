package main

import "github.com/Shopify/sarama"

func ShipmentRejectedBySellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
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
