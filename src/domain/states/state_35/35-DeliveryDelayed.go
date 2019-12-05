package state_35

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Delivery_Delayed"
	stepIndex int    = 35
)

type shipmentDeliveryDelayedStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentDeliveryDelayedStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentDeliveryDelayedStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &shipmentDeliveryDelayedStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (shipmentDeliveryDelayed shipmentDeliveryDelayedStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (shipmentDeliveryDelayed shipmentDeliveryDelayedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
	panic("implementation required")
}

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
