package return_shipment_delivery_problem_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Return_Shipment_Delivery_Problem"
	stepIndex int		= 53
)

type returnShipmentDeliveryProblemStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentDeliveryProblemStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentDeliveryProblemStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &returnShipmentDeliveryProblemStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (returnShipmentDeliveryProblem returnShipmentDeliveryProblemStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (returnShipmentDeliveryProblem returnShipmentDeliveryProblemStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}


//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func ReturnShipmentDeliveryProblemAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentDeliveryProblemRequest) error {
//	err := main.MoveOrderToNewState("buyer", "", main.ReturnShipmentDeliveryProblem, "return-shipment-delivery-problem", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
