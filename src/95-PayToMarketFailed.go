package main

import "github.com/Shopify/sarama"

func PayToMarketFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PayToMarketFailedAction(message *sarama.ConsumerMessage) error {

	err := PayToMarketFailedProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PayToMarketFailedProduce(topic string, payload []byte) error {
	return nil
}
