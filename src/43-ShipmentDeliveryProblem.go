package main

import pb "gitlab.faza.io/protos/order"

func ShipmentDeliveryProblemAction(ppr PaymentPendingRequest, req *pb.ShipmentDeliveryProblemRequest) error {
	err := MoveOrderToNewState("buyer", "", ShipmentDeliveryProblem, "shipment-delivery-problem", ppr)
	if err != nil {
		return err
	}
	return nil
}
