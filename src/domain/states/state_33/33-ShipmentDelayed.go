package state_33

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Shipment_Delayed"
	stepIndex int    = 33
)

type shipmentDetailDelayedStep struct {
	*states.BaseStepImpl
}

func New(childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &shipmentDetailDelayedStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &shipmentDetailDelayedStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStepImpl) states.IStep {
	return &shipmentDetailDelayedStep{base}
}

func NewValueOf(base *states.BaseStepImpl, params ...interface{}) states.IStep {
	panic("implementation required")
}

func (shipmentDetailDelayed shipmentDetailDelayedStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipmentDetailDelayed shipmentDetailDelayedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {
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
