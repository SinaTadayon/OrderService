package main

import "github.com/Shopify/sarama"

func SellerApprovalPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, SellerApprovalPending)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func SellerApprovalPendingAction(message *sarama.ConsumerMessage) error {

	err := SellerApprovalPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func SellerApprovalPendingProduce(topic string, payload []byte) error {
	return nil
}
