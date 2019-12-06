package state_50

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
)

const (
	stepName  string = "Return_Shipment_Pending"
	stepIndex int    = 50
)

type returnShipmentPendingState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnShipmentPendingState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnShipmentPendingState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &returnShipmentPendingState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state returnShipmentPendingState) Process(ctx context.Context, iFrame frame.IFrame) {
	panic("implementation required")
}

//func (returnShipmentDelivered returnShipmentPendingState) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
//	panic("implementation required")
//}

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
