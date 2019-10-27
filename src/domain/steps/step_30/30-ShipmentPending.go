package shipment_pending_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Shipment_Pending"
	stepIndex int		= 30
)

type shipmentPendingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentPendingStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentPendingStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentPendingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentPending shipmentPendingStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (shipmentPending shipmentPendingStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}


//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	OrderService "gitlab.faza.io/protos/order"
//)
//
//func ShipmentPendingEnteredDetail(ppr PaymentPendingRequest, req *OrderService.ShipmentDetailRequest) error {
//	ppr.ShipmentDetail.ShipmentDetail.ShipmentProvider = req.ShipmentProvider
//	ppr.ShipmentDetail.ShipmentDetail.ShipmentTrackingNumber = req.ShipmentTrackingNumber
//	ppr.ShipmentDetail.ShipmentDetail.Description = req.GetDescription()
//	err := main.MoveOrderToNewState("seller", "", main.Shipped, "shipped", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
