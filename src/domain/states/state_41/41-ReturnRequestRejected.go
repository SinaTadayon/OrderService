package state_41

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
)

const (
	stepName  string = "Return_Request_Rejected"
	stepIndex int    = 41
)

type returnRequestRejectedState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnRequestRejectedState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnRequestRejectedState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &returnRequestRejectedState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state returnRequestRejectedState) Process(ctx context.Context, iFrame frame.IFrame) {
	panic("implementation required")
}

//func (returnShipmentPending returnRequestRejectedState) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
//	panic("implementation required")
//}

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
//	ppr.ShippingDetail.ReturnShipmentDetail.ShipmentProvider = req.GetShipmentProvider()
//	ppr.ShippingDetail.ReturnShipmentDetail.Description = req.GetDescription()
//	ppr.ShippingDetail.ReturnShipmentDetail.ShipmentTrackingNumber = req.GetShipmentTrackingNumber()
//
//	err := main.MoveOrderToNewState("buyer", "", main.ReturnShipped, "return-shipped", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
