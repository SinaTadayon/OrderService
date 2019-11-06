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
	ShipmentDeliveredPending			= "ShipmentDeliveredPending"
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

	//if param == nil {
		shipmentDelivered.UpdateAllOrderStatus(ctx, &order, itemsId, steps.InProgressStatus, false)
		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, ShipmentDeliveredPending, true, "", true, steps.InProgressStatus)
		if err:= shipmentDelivered.persistOrder(ctx, &order); err != nil {
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: nil}
		return promise.NewPromise(returnChannel, 1, 1)
	//}
	//} else {
	//	req, ok := param.(message.RequestSellerOrderAction)
	//	if ok != true {
	//		//if len(order.Items) == len(itemsId) {
	//		//	sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, nil, "CLOSED", false)
	//		//} else {
	//		//	sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, nil, "InProgress", false)
	//		//}
	//		//sellerApprovalPending.persistOrder(ctx, &order)
	//
	//		logger.Err("param not a message.RequestSellerOrderAction type , order: %v", order)
	//		returnChannel := make(chan promise.FutureData, 1)
	//		defer close(returnChannel)
	//		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	//		return promise.NewPromise(returnChannel, 1, 1)
	//	}
	//
	//	if req.Action == "success" {
	//		shipmentDelivered.UpdateAllOrderStatus(ctx, &order, nil, "InProgress", false)
	//		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, Approved, true, "", false)
	//		shipmentDelivered.persistOrder(ctx, &order)
	//		return shipmentDelivered.Childes()[2].ProcessOrder(ctx, order, itemsId, nil)
	//	} else if req.Action == "failed" {
	//		actionData := req.Data.(*message.RequestSellerOrderAction_Failed)
	//		if ok != true {
	//			logger.Err("request data not a message.RequestSellerOrderAction_Failed type , order: %v", order)
	//			returnChannel := make(chan promise.FutureData, 1)
	//			defer close(returnChannel)
	//			returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	//			return promise.NewPromise(returnChannel, 1, 1)
	//		}
	//		shipmentDelivered.UpdateAllOrderStatus(ctx, &order, nil, "InProgress", false)
	//		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, Approved, false, actionData.Failed.Reason, false)
	//		shipmentDelivered.persistOrder(ctx, &order)
	//		return shipmentDelivered.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
	//	}
	//
	//	logger.Err("%s step received invalid action, order: %v, action: %s", shipmentDelivered.Name(), order, req.Action)
	//	returnChannel := make(chan promise.FutureData, 1)
	//	defer close(returnChannel)
	//	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	//	return promise.NewPromise(returnChannel, 1, 1)
	//}
}

func (shipmentDelivered shipmentDeliveredStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentDelivered.Name(), order, err.Error())
	}
	return err
}

func (shipmentDelivered shipmentDeliveredStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool, reason string, isSetExpireTime bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentDelivered.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
					findFlag = true
					break
				}
			}
			if !findFlag {
				logger.Err("%s received itemId %s not exist in order, orderId: %v", shipmentDelivered.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentDelivered.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
		}
	}
}

func (shipmentDelivered shipmentDeliveredStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, reason string, isSetExpireTime bool, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	var action entities.Action
	if isSetExpireTime {
		// TODO must be checked calculation
		expiredTime := order.Items[index].UpdatedAt.Add(time.Hour *
			time.Duration(order.Items[index].ShipmentSpec.ShippingTime) +
			time.Minute * time.Duration(0) +
			time.Second * time.Duration(0))

		action = entities.Action{
			Name:      actionName,
			Result:    result,
			Reason:		reason,
			Data: map[string]interface{}{
				"expiredTime": expiredTime,
			},
			CreatedAt: order.Items[index].UpdatedAt,
		}
	} else {
		action = entities.Action {
			Name:      actionName,
			Result:    result,
			Reason:	   reason,
			CreatedAt: order.Items[index].UpdatedAt,
		}
	}

	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
}