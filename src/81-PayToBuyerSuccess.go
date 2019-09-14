package main

import "github.com/Shopify/sarama"

func PayToBuyerSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PayToBuyerSuccess)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PayToBuyerSuccessAction(message *sarama.ConsumerMessage) error {

	err := PayToBuyerSuccessProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToBuyerSuccessProduce(topic string, payload []byte) error {
	return nil
}
