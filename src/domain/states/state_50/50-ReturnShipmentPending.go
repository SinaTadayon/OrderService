package state_50

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Return_Shipment_Pending"
	stepIndex int    = 50
)

type returnShipmentDeliveredStep struct {
	*states.BaseStepImpl
}

func New(childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &returnShipmentDeliveredStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &returnShipmentDeliveredStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStepImpl) states.IStep {
	return &returnShipmentDeliveredStep{base}
}

func NewValueOf(base *states.BaseStepImpl, params ...interface{}) states.IStep {
	panic("implementation required")
}

func (returnShipmentDelivered returnShipmentDeliveredStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (returnShipmentDelivered returnShipmentDeliveredStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {
	panic("implementation required")
}

//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func ReturnShipmentDeliveredAction(ppr PaymentPendingRequest, req *pb.ReturnShipmentDeliveredRequest) error {
//	err := main.MoveOrderToNewState("buyer", "", main.ReturnShipmentDelivered, "return-shipment-delivered", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
