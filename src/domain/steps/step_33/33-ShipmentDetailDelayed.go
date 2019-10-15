package shipment_detail_delayed_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Shipment_Detail_Delayed"
	stepIndex int		= 33
)

type shipmentDetailDelayedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDetailDelayedStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDetailDelayedStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentDetailDelayedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentDetailDelayed shipmentDetailDelayedStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (shipmentDetailDelayed shipmentDetailDelayedStep) ProcessOrder(ctx context.Context, order entities.Order) error {
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
