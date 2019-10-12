package main

import "github.com/Shopify/sarama"

// TODO: must be implemented
func PayToBuyerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PayToBuyer)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PayToBuyerAction(message *sarama.ConsumerMessage) error {

	err := PayToBuyerProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToBuyerProduce(topic string, payload []byte) error {
	return nil
}
