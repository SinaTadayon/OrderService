package return_shipment_success_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Return_Shipment_Success"
	stepIndex int		= 55
)

type returnShipmentSuccessStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentSuccessStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentSuccessStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &returnShipmentSuccessStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (returnShipmentSuccess returnShipmentSuccessStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (returnShipmentSuccess returnShipmentSuccessStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}



//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func ReturnShipmentDeliveredGrpcAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentSuccessRequest) error {
//	err := main.MoveOrderToNewState("buyer", "", main.ReturnShipmentSuccess, "return-shipment-success", ppr)
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
