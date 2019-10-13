package main

import pb "gitlab.faza.io/protos/order"

func PayToSellerSuccessAction(ppr PaymentPendingRequest, req *pb.PayToSellerSuccessRequest) error {
	err := MoveOrderToNewState("operator", req.GetDescription(), PayToSellerSuccess, "pay-to-seller-success", ppr)
	if err != nil {
		return err
	}
	return nil
}
