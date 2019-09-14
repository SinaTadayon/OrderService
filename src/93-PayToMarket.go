package main

import "github.com/Shopify/sarama"

func PayToMarketMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PayToMarket)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PayToMarketAction(message *sarama.ConsumerMessage) error {

	err := PayToMarketProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToMarketProduce(topic string, payload []byte) error {
	return nil
}
