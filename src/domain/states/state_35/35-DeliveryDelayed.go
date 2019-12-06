package state_35

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
)

const (
	stepName  string = "Delivery_Delayed"
	stepIndex int    = 35
)

type DeliveryDelayedState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &DeliveryDelayedState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &DeliveryDelayedState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &DeliveryDelayedState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (shipmentDeliveryDelayed DeliveryDelayedState) Process(ctx context.Context, iFrame frame.IFrame) {
	panic("implementation required")
}

//func (shipmentDeliveryDelayed DeliveryDelayedState) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
//	panic("implementation required")
//}

//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func ShipmentDeliveryDelay(ppr PaymentPendingRequest, req *pb.ShipmentDeliveryDelayedRequest) error {
//	err := main.MoveOrderToNewState("buyer", "", main.ShipmentDeliveryDelayed, "shipment-delivered-delayed", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
