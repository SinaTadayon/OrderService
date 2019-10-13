package main

import pb "gitlab.faza.io/protos/order"

func ReturnShipmentDeliveryProblemAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentDeliveryProblemRequest) error {
	err := MoveOrderToNewState("buyer", "", ReturnShipmentDeliveryProblem, "return-shipment-delivery-problem", ppr)
	if err != nil {
		return err
	}
	return nil
}
