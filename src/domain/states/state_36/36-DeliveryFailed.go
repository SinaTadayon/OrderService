package state_36

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Delivery_Failed"
	stepIndex int    = 36
)

type shipmentCanceledStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentCanceledStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentCanceledStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &shipmentCanceledStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (shipmentCanceled shipmentCanceledStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (shipmentCanceled shipmentCanceledStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
	panic("implementation required")
}

//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//// TODO Improvement ShipmentCanceledActoin
//func ShipmentCanceledActoin(ppr PaymentPendingRequest, req *pb.ShipmentCanceledRequest) error {
//	err := main.MoveOrderToNewState(req.GetOperator(), req.GetReason(), main.ShipmentCanceled, "shipment-canceled", ppr)
//	if err != nil {
//		return err
//	}
//	newPpr, err := main.GetOrder(ppr.OrderNumber)
//	if err != nil {
//		return err
//	}
//	err = main.MoveOrderToNewState("system", "", main.PayToBuyer, "pay-to-buyer", newPpr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
