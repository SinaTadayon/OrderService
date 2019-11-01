package shipment_delivery_problem_step

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
	"time"
)

const (
	stepName string 	= "Shipment_Delivery_Problem"
	stepIndex int		= 43
)

type shipmentDeliveryProblemStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDeliveryProblemStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDeliveryProblemStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentDeliveryProblemStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

// TODO operator action required handled
func (shipmentDeliveryProblem shipmentDeliveryProblemStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {
	shipmentDeliveryProblem.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
	shipmentDeliveryProblem.updateOrderItemsProgress(ctx, &order, itemsId, "BuyerShipmentDeliveryProblem", true)
	shipmentDeliveryProblem.persistOrder(ctx, &order)
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data:nil, Ex:nil}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) persistOrder(ctx context.Context, order *entities.Order) {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentDeliveryProblem.Name(), order, err.Error())
	}
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentDeliveryProblem.doUpdateOrderItemsProgress(ctx, order, i, action, result)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", shipmentDeliveryProblem.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentDeliveryProblem.doUpdateOrderItemsProgress(ctx, order, i, action, result)
		}
	}
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool) {

	order.Items[index].Status = actionName
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.ActionHistory = append(order.Items[index].Progress.ActionHistory, action)
}


//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func ShipmentDeliveryProblemAction(ppr PaymentPendingRequest, req *pb.ShipmentDeliveryProblemRequest) error {
//	err := main.MoveOrderToNewState("buyer", "", main.ShipmentDeliveryProblem, "shipment-delivery-problem", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
