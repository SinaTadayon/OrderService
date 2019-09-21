package main

import pb "gitlab.faza.io/protos/payment"

func ShipmentSuccessAction(ppr PaymentPendingRequest, req *pb.ShipmentSuccessRequest) error {
	err := MoveOrderToNewState("buyer", "", ShipmentSuccess, "shipment-success", ppr)
	if err != nil {
		return err
	}
	return nil
}
