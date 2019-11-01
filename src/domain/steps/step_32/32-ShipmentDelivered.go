package shipment_delivered_step

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
	stepName string 	= "Shipment_Delivered"
	stepIndex int		= 32
	Approved			= "Approved"
)

type shipmentDeliveredStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDeliveredStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDeliveredStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentDeliveredStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentDelivered shipmentDeliveredStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipmentDelivered shipmentDeliveredStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	if param == nil {
		shipmentDelivered.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, "BuyerShipmentDeliveredPending", true)
		shipmentDelivered.persistOrder(ctx, &order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: nil}
		return promise.NewPromise(returnChannel, 1, 1)
	} else {
		req, ok := param.(message.RequestSellerOrderAction)
		if ok != true {
			//if len(order.Items) == len(itemsId) {
			//	sellerApprovalPending.UpdateOrderStep(ctx, &order, nil, "CLOSED", false)
			//} else {
			//	sellerApprovalPending.UpdateOrderStep(ctx, &order, nil, "InProgress", false)
			//}
			//sellerApprovalPending.persistOrder(ctx, &order)

			logger.Err("param not a message.RequestSellerOrderAction type , order: %v", order)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		if req.Action == "success" {
			shipmentDelivered.UpdateOrderStep(ctx, &order, nil, "InProgress", false)
			shipmentDelivered.updateApprovedOrderItemsProgress(ctx, &order, itemsId, Approved, true, "")
			shipmentDelivered.persistOrder(ctx, &order)
			return shipmentDelivered.Childes()[2].ProcessOrder(ctx, order, itemsId, nil)
		} else if req.Action == "failed" {
			actionData := req.Data.(*message.RequestSellerOrderAction_Failed)
			if ok != true {
				logger.Err("request data not a message.RequestSellerOrderAction_Failed type , order: %v", order)
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
				return promise.NewPromise(returnChannel, 1, 1)
			}
			shipmentDelivered.UpdateOrderStep(ctx, &order, nil, "InProgress", false)
			shipmentDelivered.updateApprovedOrderItemsProgress(ctx, &order, itemsId, Approved, false, actionData.Failed.Reason)
			shipmentDelivered.persistOrder(ctx, &order)
			return shipmentDelivered.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
		}

		logger.Err("%s step received invalid action, order: %v, action: %s", shipmentDelivered.Name(), order, req.Action)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}
}

func (shipmentDelivered shipmentDeliveredStep) persistOrder(ctx context.Context, order *entities.Order) {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentDelivered.Name(), order, err.Error())
	}
}

func (shipmentDelivered shipmentDeliveredStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentDelivered.doUpdateOrderItemsProgress(ctx, order, i, action, result)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", shipmentDelivered.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentDelivered.doUpdateOrderItemsProgress(ctx, order, i, action, result)
		}
	}
}

func (shipmentDelivered shipmentDeliveredStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool) {

	order.Items[index].Status = actionName
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	// TODO must be checked calculation
	expiredTime := order.Items[index].UpdatedAt.Add(time.Hour *
		time.Duration(order.Items[index].ShipmentSpec.ShippingTime) +
		time.Minute * time.Duration(0) +
		time.Second * time.Duration(0))

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		Data: map[string]interface{}{
			"expiredTime": expiredTime,
		},
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.ActionHistory = append(order.Items[index].Progress.ActionHistory, action)
}

func (shipmentDelivered shipmentDeliveredStep) updateApprovedOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string, action string, result bool, reason string) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentDelivered.doApprovedUpdateOrderItemsProgress(ctx, order, i, action, result, reason)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", shipmentDelivered.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentDelivered.doApprovedUpdateOrderItemsProgress(ctx, order, i, action, result, reason)
		}
	}
}

func (shipmentDelivered shipmentDeliveredStep) doApprovedUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, reason string) {

	order.Items[index].Status = actionName
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		Reason:    reason,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.ActionHistory = append(order.Items[index].Progress.ActionHistory, action)
}
