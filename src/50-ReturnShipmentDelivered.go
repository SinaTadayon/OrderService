package main

import pb "gitlab.faza.io/protos/order"

func ReturnShipmentDeliveredAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentDeliveredRequest) error {
	err := MoveOrderToNewState("buyer", "", ReturnShipmentDelivered, "return-shipment-delivered", ppr)
	if err != nil {
		return err
	}
	return nil
}
