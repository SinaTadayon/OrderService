package main

import (
	"encoding/json"

	"gitlab.faza.io/go-framework/logger"

	"github.com/Shopify/sarama"
)

func PaymentSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	mess, err := CheckOrderKafkaAndMongoStatus(message, PaymentSuccess)
	if err != nil {
		return mess, err
	}
	return message, nil
}

func PaymentSuccessAction(message *sarama.ConsumerMessage) error {
	ppr := PaymentPendingRequest{}
	err := json.Unmarshal(message.Value, &ppr)
	if err != nil {
		return err
	}

	// @TODO: remove automatic move status
	err = MoveOrderToNewState("system", "auto approval", SellerApprovalPending, "seller-approval-pending", ppr)
	if err != nil {
		return err
	}

	err = NotifySellerForNewOrder(ppr)
	if err != nil {
		logger.Err("cant notify seller, %v", err)
	}
	return nil
}
