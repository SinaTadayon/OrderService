package return_shipment_delivered_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Return_Shipment_Delivered"
	stepIndex int		= 50
)

type returnShipmentDeliveredStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, states ...states.IState) steps.IStep {
	return &returnShipmentDeliveredStep{steps.NewBaseStep(stepName, stepIndex, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewOf(name string, index int, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentDeliveredStep{steps.NewBaseStep(name, index, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &returnShipmentDeliveredStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (returnShipmentDelivered returnShipmentDeliveredStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (returnShipmentDelivered returnShipmentDeliveredStep) ProcessOrder(ctx context.Context, order entities.Order) promise.IPromise {
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
