package main

import pb "gitlab.faza.io/protos/order"

func PayToBuyerSuccessAction(ppr PaymentPendingRequest, req *pb.PayToBuyerSuccessRequest) error {
	err := MoveOrderToNewState("operator", req.GetDescription(), PayToBuyerSuccess, "pay-to-buyer-success", ppr)
	if err != nil {
		return err
	}
	return nil
}
