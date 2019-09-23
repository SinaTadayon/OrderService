package main

import (
	pb "gitlab.faza.io/protos/payment"
)

func ReturnShipmentDeliveredGrpcAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentSuccessRequest) error {
	err := MoveOrderToNewState("buyer", "", ReturnShipmentSuccess, "return-shipment-success", ppr)
	if err != nil {
		return err
	}
	newPpr, err := GetOrder(ppr.OrderNumber)
	if err != nil {
		return err
	}
	err = MoveOrderToNewState("system", "", PayToBuyer, "pay-to-buyer", newPpr)
	if err != nil {
		return err
	}
	return nil
}
