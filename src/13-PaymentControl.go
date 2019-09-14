package main

import "github.com/Shopify/sarama"

func PaymentControlMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PaymentControl)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PaymentControlAction(message *sarama.ConsumerMessage) error {

	err := PaymentControlProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PaymentControlProduce(topic string, payload []byte) error {
	return nil
}
