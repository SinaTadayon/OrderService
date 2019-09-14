package main

import "github.com/Shopify/sarama"

func PayToSellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PayToSeller)
	if err != nil {
		return mess, err
	}
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
