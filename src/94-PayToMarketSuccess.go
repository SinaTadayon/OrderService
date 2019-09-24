package main

import pb "gitlab.faza.io/protos/payment"

func PayToMarketSuccessAction(ppr PaymentPendingRequest, req *pb.PayToMarketSuccessRequest) error {
	err := MoveOrderToNewState("operator", req.GetDescription(), PayToMarketSuccess, "pay-to-market-success", ppr)
	if err != nil {
		return err
	}
	return nil
}
