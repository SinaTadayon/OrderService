package main

import "github.com/Shopify/sarama"

func PayToSellerSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PayToSellerSuccess)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PayToSellerSuccessAction(message *sarama.ConsumerMessage) error {

	err := PayToSellerSuccessProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToSellerSuccessProduce(topic string, payload []byte) error {
	return nil
}
