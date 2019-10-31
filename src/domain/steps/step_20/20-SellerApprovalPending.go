package seller_approval_pending_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName string 	= "Seller_Approval_Pending"
	stepIndex int		= 20
)

type sellerApprovalPendingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &sellerApprovalPendingStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &sellerApprovalPendingStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &sellerApprovalPendingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (sellerApprovalPending sellerApprovalPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (sellerApprovalPending sellerApprovalPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string) promise.IPromise {
	panic("implementation required")
}


//
//import "gitlab.faza.io/order-project/order-service"
//
//func SellerApprovalPendingApproved(ppr PaymentPendingRequest) error {
//	err := main.MoveOrderToNewState("seller", "", main.ShipmentPending, "shipment-pending", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//// TODO: Improvement SellerApprovalPendingRejected
//func SellerApprovalPendingRejected(ppr PaymentPendingRequest, reason string) error {
//	err := main.MoveOrderToNewState("seller", reason, main.ShipmentRejectedBySeller, "shipment-rejected-by-seller", ppr)
//	if err != nil {
//		return err
//	}
//	newPpr, err := main.GetOrder(ppr.OrderNumber)
//	if err != nil {
//		return err
//	}
//	err = main.MoveOrderToNewState("system", reason, main.PayToBuyer, "pay-to-buyer", newPpr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
