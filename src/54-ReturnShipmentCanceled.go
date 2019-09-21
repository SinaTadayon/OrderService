package main

import pb "gitlab.faza.io/protos/payment"

func ReturnShipmentCanceledActoin(ppr PaymentPendingRequest, req *pb.ReturnShipmentCanceledRequest) error {
	err := MoveOrderToNewState("operator", req.GetReason(), ReturnShipmentCanceled, "return-shipment-canceled", ppr)
	if err != nil {
		return err
	}
	newPpr, err := GetOrder(ppr.OrderNumber)
	if err != nil {
		return err
	}
	err = MoveOrderToNewState("system", "", PayToSeller, "pay-to-seller", newPpr)
	if err != nil {
		return err
	}
	return nil
}
