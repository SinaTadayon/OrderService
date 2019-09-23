package main

import pb "gitlab.faza.io/protos/payment"

func ReturnShipmentPendingAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentPendingRequest) error {
	err := MoveOrderToNewState(req.GetOperator(), req.GetReason(), ReturnShipmentPending, "return-shipment-pending", ppr)
	if err != nil {
		return err
	}
	return nil
}

func ReturnShipmentDetailAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentDetailRequest) error {
	ppr.ShipmentInfo.ReturnShipmentDetail.ShipmentProvider = req.GetShipmentProvider()
	ppr.ShipmentInfo.ReturnShipmentDetail.Description = req.GetDescription()
	ppr.ShipmentInfo.ReturnShipmentDetail.ShipmentTrackingNumber = req.GetShipmentTrackingNumber()

	err := MoveOrderToNewState("buyer", "", ReturnShipped, "return-shipped", ppr)
	if err != nil {
		return err
	}
	return nil
}
