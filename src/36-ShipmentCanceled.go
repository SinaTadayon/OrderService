package main

import pb "gitlab.faza.io/protos/payment"

func ShipmentCanceledActoin(ppr PaymentPendingRequest, req *pb.ShipmentCanceledRequest) error {
	err := MoveOrderToNewState(req.GetOperator(), req.GetReason(), ShipmentCanceled, "shipment-canceled", ppr)
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
