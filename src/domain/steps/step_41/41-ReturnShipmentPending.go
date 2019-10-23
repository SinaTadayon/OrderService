package return_shipment_pending_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Return_Shipment_Pending"
	stepIndex int		= 41
)

type returnShipmentPendingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentPendingStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentPendingStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &returnShipmentPendingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (returnShipmentPending returnShipmentPendingStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (returnShipmentPending returnShipmentPendingStep) ProcessOrder(ctx context.Context, order entities.Order) error {
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
//	ppr.ShipmentInfo.ReturnShipmentDetail.ShipmentProvider = req.GetShipmentProvider()
//	ppr.ShipmentInfo.ReturnShipmentDetail.Description = req.GetDescription()
//	ppr.ShipmentInfo.ReturnShipmentDetail.ShipmentTrackingNumber = req.GetShipmentTrackingNumber()
//
//	err := main.MoveOrderToNewState("buyer", "", main.ReturnShipped, "return-shipped", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}