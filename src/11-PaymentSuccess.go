package main

import (
	"encoding/json"
	"errors"
	"time"

	"gitlab.faza.io/go-framework/kafkaadapter"
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
	pprOld := ppr

	statusHistory := StatusHistory{
		Status:    ppr.Status.Current,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.Current = SellerApprovalPending
	ppr.Status.History = append(ppr.Status.History, statusHistory)

	err = UpdateOrderMongo(ppr)
	if err != nil {
		return err
	}

	newPpr, err := json.Marshal(ppr)
	if err != nil {
		return errors.New("cant convert ppr struct to json: " + err.Error())
	}

	err = PaymentSuccessProduce("seller-approval-pending", newPpr)
	if err != nil {
		err = UpdateOrderMongo(pprOld)
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func PaymentSuccessProduce(topic string, payload []byte) error {
	App.kafka = kafkaadapter.NewKafka(brokers, topic)
	App.kafka.Config.Producer.Return.Successes = true

	_, _, err := App.kafka.SendOne("", payload)
	if err != nil {
		logger.Err("cant insert to kafka: %v", err)
	}

	return nil
}
