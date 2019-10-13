package main

import OrderService "gitlab.faza.io/protos/order"

func BuyerCancel(ppr PaymentPendingRequest, req *OrderService.BuyerCancelRequest) error {
	err := MoveOrderToNewState("buyer", req.GetReason(), ShipmentCanceled, "shipment-canceled", ppr)
	if err != nil {
		return err
	}
	return nil
}
