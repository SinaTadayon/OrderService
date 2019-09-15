package main

import (
	"encoding/json"
	"errors"
	"time"

	"gitlab.faza.io/go-framework/kafkaadapter"
	"gitlab.faza.io/go-framework/logger"
)

func SellerApprovalPendingApproved(ppr PaymentPendingRequest) error {
	pprOld := ppr

	statusHistory := StatusHistory{
		Status:    ppr.Status.Current,
		CreatedAt: time.Now().UTC(),
		Agent:     "seller",
		Reason:    "",
	}
	ppr.Status.Current = ShipmentPending
	ppr.Status.History = append(ppr.Status.History, statusHistory)

	logger.Audit("updating mongo...")
	err := UpdateOrderMongo(ppr)
	if err != nil {
		return err
	}

	newPpr, err := json.Marshal(ppr)
	if err != nil {
		return errors.New("cant convert ppr struct to json: " + err.Error())
	}

	err = SellerApprovalPendingProduce("shipment-pending", newPpr)
	if err != nil {
		logger.Audit("rollbacking...")
		err = UpdateOrderMongo(pprOld)
		if err != nil {
			return err
		}
		return err
	}
	return nil
}
func SellerApprovalPendingRejected(ppr PaymentPendingRequest, reason string) error {
	pprOld := ppr

	statusHistory := StatusHistory{
		Status:    ppr.Status.Current,
		CreatedAt: time.Now().UTC(),
		Agent:     "seller",
		Reason:    reason,
	}
	ppr.Status.Current = ShipmentRejectedBySeller
	ppr.Status.History = append(ppr.Status.History, statusHistory)

	err := UpdateOrderMongo(ppr)
	if err != nil {
		return err
	}

	newPpr, err := json.Marshal(ppr)
	if err != nil {
		return errors.New("cant convert ppr struct to json: " + err.Error())
	}

	err = SellerApprovalPendingProduce("shipment-pending", newPpr)
	if err != nil {
		err = UpdateOrderMongo(pprOld)
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func SellerApprovalPendingProduce(topic string, payload []byte) error {
	App.kafka = kafkaadapter.NewKafka(brokers, topic)
	App.kafka.Config.Producer.Return.Successes = true

	_, _, err := App.kafka.SendOne("", payload)
	if err != nil {
		logger.Err("cant insert to kafka: %v", err)
	}

	return nil
}
