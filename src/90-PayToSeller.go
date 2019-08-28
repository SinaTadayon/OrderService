package main

import "github.com/Shopify/sarama"

func PayToSellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PayToSellerAction(message *sarama.ConsumerMessage) error {

	err := PayToSellerProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToSellerProduce(topic string, payload []byte) error {
	return nil
}
