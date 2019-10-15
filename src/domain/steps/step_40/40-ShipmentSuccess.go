package shipment_success_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Shipment_Success"
	stepIndex int		= 40
)

type shipmentSuccessStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentSuccessStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentSuccessStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentSuccessStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentSuccess shipmentSuccessStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (shipmentSuccess shipmentSuccessStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}


//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//// TODO: Transition state to 90.Pay_TO_Seller state
//func ShipmentSuccessAction(ppr PaymentPendingRequest, req *pb.ShipmentSuccessRequest) error {
//	err := main.MoveOrderToNewState("buyer", "", main.ShipmentSuccess, "shipment-success", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
