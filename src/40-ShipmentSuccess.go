package main

import pb "gitlab.faza.io/protos/order"

// TODO: Transition state to 90.Pay_TO_Seller state
func ShipmentSuccessAction(ppr PaymentPendingRequest, req *pb.ShipmentSuccessRequest) error {
	err := MoveOrderToNewState("buyer", "", ShipmentSuccess, "shipment-success", ppr)
	if err != nil {
		return err
	}
	return nil
}
