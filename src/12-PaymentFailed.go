package main

import "github.com/Shopify/sarama"

func PaymentFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PaymentFailed)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PaymentFailedAction(message *sarama.ConsumerMessage) error {

	err := PaymentFailedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PaymentFailedProduce(topic string, payload []byte) error {
	return nil
}
