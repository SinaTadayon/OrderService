package main

import pb "gitlab.faza.io/protos/order"

func ReturnShipmentDeliveryDelay(ppr PaymentPendingRequest, req *pb.ReturnShipmentDeliveryDelayedRequest) error {
	err := MoveOrderToNewState("buyer", "", ReturnShipmentDeliveryDelayed, "return-shipment-delivered-delayed", ppr)
	if err != nil {
		return err
	}
	return nil
}
