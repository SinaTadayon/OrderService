package state_33

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Shipment_Delayed"
	stepIndex int    = 33
)

type shipmentDetailDelayedStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentDetailDelayedStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentDetailDelayedStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &shipmentDetailDelayedStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (shipmentDetailDelayed shipmentDetailDelayedStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (shipmentDetailDelayed shipmentDetailDelayedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
	panic("implementation required")
}

//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	OrderService "gitlab.faza.io/protos/order"
//)
//
//func BuyerCancel(ppr PaymentPendingRequest, req *OrderService.BuyerCancelRequest) error {
//	err := main.MoveOrderToNewState("buyer", req.GetReason(), main.ShipmentCanceled, "shipment-canceled", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
