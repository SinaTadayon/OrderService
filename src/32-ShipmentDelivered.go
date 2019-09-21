package main

import pb "gitlab.faza.io/protos/payment"

func ShipmentDeliveredAction(ppr PaymentPendingRequest, req *pb.ShipmentDeliveredRequest) error {
	err := MoveOrderToNewState("buyer", "", ShipmentDelivered, "shipment-delivered", ppr)
	if err != nil {
		return err
	}
	return nil
}
