package main

import "github.com/Shopify/sarama"

func PayToSellerFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PayToSellerFailed)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PayToSellerFailedAction(message *sarama.ConsumerMessage) error {

	err := PayToSellerFailedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToSellerFailedProduce(topic string, payload []byte) error {
	return nil
}
