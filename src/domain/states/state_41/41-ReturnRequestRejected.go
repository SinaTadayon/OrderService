package state_41

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Return_Request_Rejected"
	stepIndex int    = 41
)

type returnShipmentPendingStep struct {
	*states.BaseStepImpl
}

func New(childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &returnShipmentPendingStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &returnShipmentPendingStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStepImpl) states.IStep {
	return &returnShipmentPendingStep{base}
}

func NewValueOf(base *states.BaseStepImpl, params ...interface{}) states.IStep {
	panic("implementation required")
}

func (returnShipmentPending returnShipmentPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (returnShipmentPending returnShipmentPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {
	panic("implementation required")
}

//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func ReturnShipmentPendingAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentPendingRequest) error {
//	err := main.MoveOrderToNewState(req.GetOperator(), req.GetReason(), main.ReturnShipmentPending, "return-shipment-pending", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func ReturnShipmentDetailAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentDetailRequest) error {
//	ppr.ShippingDetail.ReturnShipmentDetail.ShipmentProvider = req.GetShipmentProvider()
//	ppr.ShippingDetail.ReturnShipmentDetail.Description = req.GetDescription()
//	ppr.ShippingDetail.ReturnShipmentDetail.ShipmentTrackingNumber = req.GetShipmentTrackingNumber()
//
//	err := main.MoveOrderToNewState("buyer", "", main.ReturnShipped, "return-shipped", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
