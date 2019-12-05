package state_54

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Return_Delivery_Delayed"
	stepIndex int    = 54
)

type returnShipmentCanceledStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &returnShipmentCanceledStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &returnShipmentCanceledStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &returnShipmentCanceledStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (returnShipmentCanceled returnShipmentCanceledStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (returnShipmentCanceled returnShipmentCanceledStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
	panic("implementation required")
}

//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func ReturnShipmentCanceledActoin(ppr PaymentPendingRequest, req *pb.ReturnShipmentCanceledRequest) error {
//	err := main.MoveOrderToNewState("operator", req.GetReason(), main.ReturnShipmentCanceled, "return-shipment-canceled", ppr)
//	if err != nil {
//		return err
//	}
//	newPpr, err := main.GetOrder(ppr.OrderNumber)
//	if err != nil {
//		return err
//	}
//	err = main.MoveOrderToNewState("system", "", main.PayToSeller, "pay-to-seller", newPpr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
