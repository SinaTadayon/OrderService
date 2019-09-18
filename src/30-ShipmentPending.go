package main

import (
	OrderService "gitlab.faza.io/protos/payment"
)

func ShipmentPendingEnteredDetail(ppr PaymentPendingRequest, req *OrderService.ShipmentDetailRequest) error {
	ppr.ShipmentInfo.ShipmentDetail.ShipmentProvider = req.ShipmentProvider
	ppr.ShipmentInfo.ShipmentDetail.ShipmentTrackingNumber = req.ShipmentTrackingNumber

	err := MoveOrderToNewState("seller", "", Shipped, "shipped", ppr)
	if err != nil {
		return err
	}
	return nil
}
