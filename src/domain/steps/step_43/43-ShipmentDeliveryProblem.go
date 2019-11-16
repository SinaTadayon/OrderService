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
	stepName      string = "Shipment_Delivery_Problem"
	stepIndex     int    = 43
	StockReleased        = "StockReleased"
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
	//shipmentDeliveryProblem.UpdateAllOrderStatus(ctx, &order, itemsId, "InProgress", false)
	//shipmentDeliveryProblem.updateOrderItemsProgress(ctx, &order, itemsId, "BuyerShipmentDeliveryProblem", true)
	//shipmentDeliveryProblem.persistOrder(ctx, &order)
	//returnChannel := make(chan promise.FutureData, 1)
	//defer close(returnChannel)
	//returnChannel <- promise.FutureData{Data:nil, Ex:nil}
	//return promise.NewPromise(returnChannel, 1, 1)

	logger.Audit("Order Received in %s step, orderId: %s, Action: %s", shipmentDeliveryProblem.Name(), order.OrderId, "shipmentDelivered")
	req, ok := param.(*message.RequestBackOfficeOrderAction)
	if ok != true {
		logger.Err("param not a message.RequestBackOfficeOrderAction type , order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if !shipmentDeliveryProblem.validateAction(ctx, &order, itemsId) {
		logger.Err("%s step received invalid action, order: %v, action: %s", shipmentDeliveryProblem.Name(), order, req.Action)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotAccepted, Reason: "Action Expired"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if req.Action == "success" {
		if len(order.Items) == len(itemsId) {
			shipmentDeliveryProblem.UpdateAllOrderStatus(ctx, &order, itemsId, steps.ClosedStatus, false)
		} else {
			shipmentDeliveryProblem.UpdateAllOrderStatus(ctx, &order, itemsId, steps.InProgressStatus, false)
		}

		shipmentDeliveryProblem.updateOrderItemsProgress(ctx, &order, itemsId, req, steps.ClosedStatus)
		if err := shipmentDeliveryProblem.persistOrder(ctx, &order); err != nil {
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}
		return shipmentDeliveryProblem.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
	} else if req.Action == "cancel" {

		iPromise := global.Singletons.StockService.BatchStockActions(ctx, order, itemsId, StockReleased)
		futureData := iPromise.Data()
		if futureData == nil {
			if err := shipmentDeliveryProblem.persistOrder(ctx, &order); err != nil {
			}
			logger.Err("StockService promise channel has been closed, order: %s", order.OrderId)
		} else if futureData.Ex != nil {
			if err := shipmentDeliveryProblem.persistOrder(ctx, &order); err != nil {
			}
			logger.Err("released stock from stockService failed, error: %s, orderId: %s", futureData.Ex.Error(), order.OrderId)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		if len(order.Items) == len(itemsId) {
			shipmentDeliveryProblem.UpdateAllOrderStatus(ctx, &order, itemsId, steps.ClosedStatus, false)
		} else {
			shipmentDeliveryProblem.UpdateAllOrderStatus(ctx, &order, itemsId, steps.InProgressStatus, false)
		}
		shipmentDeliveryProblem.updateOrderItemsProgress(ctx, &order, itemsId, req, steps.ClosedStatus)
		if err := shipmentDeliveryProblem.persistOrder(ctx, &order); err != nil {
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		return shipmentDeliveryProblem.Childes()[1].ProcessOrder(ctx, order, itemsId, nil)
	}

	logger.Err("%s step received invalid action, order: %v, action: %s", shipmentDeliveryProblem.Name(), order, req.Action)
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) validateAction(ctx context.Context, order *entities.Order,
	itemsId []string) bool {
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				length := len(order.Items[i].Progress.StepsHistory) - 1
				if order.Items[i].ItemId == id && order.Items[i].Progress.StepsHistory[length].Name != "32.Shipment_Delivered" {
					return false
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			length := len(order.Items[i].Progress.StepsHistory) - 1
			if order.Items[i].Progress.StepsHistory[length].Name != "32.Shipment_Delivered" {
				return false
			}
		}
	}

	return true
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentDeliveryProblem.Name(), order, err.Error())
	}

	return err
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	req *message.RequestBackOfficeOrderAction, itemStatus string) *entities.Order {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentDeliveryProblem.doUpdateOrderItemsProgress(ctx, order, i, req, itemStatus)
					findFlag = true
					break
				}
			}

			if findFlag == false {
				logger.Err("%s received itemId %s not exist in order, orderId: %v", shipmentDeliveryProblem.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentDeliveryProblem.doUpdateOrderItemsProgress(ctx, order, i, req, itemStatus)
		}
	}

	return order
}

func (shipmentDeliveryProblem shipmentDeliveryProblemStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	req *message.RequestBackOfficeOrderAction, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      req.Action,
		Result:    true,
		Reason:    req.Description,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
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
