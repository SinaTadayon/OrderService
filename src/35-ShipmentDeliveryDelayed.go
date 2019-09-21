package main

import pb "gitlab.faza.io/protos/payment"

func ShipmentDeliveryDelay(ppr PaymentPendingRequest, req *pb.ShipmentDeliveryDelayedRequest) error {
	err := MoveOrderToNewState("buyer", "", ShipmentDeliveryDelayed, "shipment-delivered-delayed", ppr)
	if err != nil {
		return err
	}
	return nil
}
