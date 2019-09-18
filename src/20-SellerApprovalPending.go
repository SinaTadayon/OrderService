package main

func SellerApprovalPendingApproved(ppr PaymentPendingRequest) error {
	err := MoveOrderToNewState("seller", "", ShipmentPending, "shipment-pending", ppr)
	if err != nil {
		return err
	}
	return nil
}
func SellerApprovalPendingRejected(ppr PaymentPendingRequest, reason string) error {
	err := MoveOrderToNewState("seller", reason, ShipmentRejectedBySeller, "shipment-rejected-by-seller", ppr)
	if err != nil {
		return err
	}
	newPpr, err := GetOrder(ppr.OrderNumber)
	if err != nil {
		return err
	}
	err = MoveOrderToNewState("system", reason, PayToBuyer, "pay-to-buyer", newPpr)
	if err != nil {
		return err
	}
	return nil
}
