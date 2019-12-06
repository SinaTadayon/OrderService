package state_32

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
)

const (
	stepName  string = "Delivered"
	stepIndex int    = 32
	//ShipmentDeliveredPending             = "ShipmentDeliveredPending"
	//AutoApprovedShipmentDelivered        = "AutoApproved"
)

type shipmentDeliveredState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &shipmentDeliveredState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &shipmentDeliveredState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &shipmentDeliveredState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (shipmentDelivered shipmentDeliveredState) Process(ctx context.Context, iFrame frame.IFrame) {
	panic("implementation required")
}

//func (shipmentDelivered shipmentDeliveredState) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
//
//	if param == nil {
//		logger.Audit("Order Received in %s step, orderId: %d, Actions: %s", shipmentDelivered.Name(), order.OrderId, ShipmentDeliveredPending)
//		shipmentDelivered.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, ShipmentDeliveredPending, true, "", true, states.OrderInProgressStatus)
//		if err := shipmentDelivered.persistOrder(ctx, &order); err != nil {
//			returnChannel := make(chan future.IDataFuture, 1)
//			defer close(returnChannel)
//			returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//			return future.NewFuture(returnChannel, 1, 1)
//		}
//		returnChannel := make(chan future.IDataFuture, 1)
//		defer close(returnChannel)
//		returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
//		return future.NewFuture(returnChannel, 1, 1)
//	} else if param == "actionApproved" {
//		logger.Audit("Order Received in %s step, orderId: %d, Actions: %s", shipmentDelivered.Name(), order.OrderId, AutoApprovedShipmentDelivered)
//		shipmentDelivered.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, AutoApprovedShipmentDelivered, true, "", true, states.OrderInProgressStatus)
//		if err := shipmentDelivered.persistOrder(ctx, &order); err != nil {
//			returnChannel := make(chan future.IDataFuture, 1)
//			defer close(returnChannel)
//			returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//			return future.NewFuture(returnChannel, 1, 1)
//		}
//
//		return shipmentDelivered.Childes()[2].ProcessOrder(ctx, order, itemsId, nil)
//	}
//
//	logger.Audit("invalid action, Order Received in %s step, orderId: %d, Actions: %s", shipmentDelivered.Name(), order.OrderId, param)
//	returnChannel := make(chan future.IDataFuture, 1)
//	defer close(returnChannel)
//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//	return future.NewFuture(returnChannel, 1, 1)
//
//	//} else {
//	//	req, ok := param.(message.RequestSellerOrderAction)
//	//	if ok != true {
//	//		//if len(order.Items) == len(itemsId) {
//	//		//	sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, nil, "CLOSED", false)
//	//		//} else {
//	//		//	sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, nil, "InProgress", false)
//	//		//}
//	//		//sellerApprovalPending.persistOrder(ctx, &order)
//	//
//	//		logger.Err("param not a message.RequestSellerOrderAction type , order: %v", order)
//	//		returnChannel := make(chan future.IDataFuture, 1)
//	//		defer close(returnChannel)
//	//		returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.InternalError, Reason:"Unknown Error"}}
//	//		return future.NewFuture(returnChannel, 1, 1)
//	//	}
//	//
//	//	if req.Actions == "success" {
//	//		shipmentDelivered.UpdateAllOrderStatus(ctx, &order, nil, "InProgress", false)
//	//		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, Approved, true, "", false)
//	//		shipmentDelivered.persistOrder(ctx, &order)
//	//		return shipmentDelivered.Childes()[2].ProcessOrder(ctx, order, itemsId, nil)
//	//	} else if req.Actions == "failed" {
//	//		actionData := req.Get.(*message.RequestSellerOrderAction_Failed)
//	//		if ok != true {
//	//			logger.Err("request data not a message.RequestSellerOrderAction_Failed type , order: %v", order)
//	//			returnChannel := make(chan future.IDataFuture, 1)
//	//			defer close(returnChannel)
//	//			returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.InternalError, Reason:"Unknown Error"}}
//	//			return future.NewFuture(returnChannel, 1, 1)
//	//		}
//	//		shipmentDelivered.UpdateAllOrderStatus(ctx, &order, nil, "InProgress", false)
//	//		shipmentDelivered.updateOrderItemsProgress(ctx, &order, itemsId, Approved, false, actionData.Failed.Reason, false)
//	//		shipmentDelivered.persistOrder(ctx, &order)
//	//		return shipmentDelivered.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
//	//	}
//	//
//	//	logger.Err("%s step received invalid action, order: %v, action: %s", shipmentDelivered.ActionName(), order, req.Actions)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	defer close(returnChannel)
//	//	returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.InternalError, Reason:"Unknown Error"}}
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//}
//
//func (shipmentDelivered shipmentDeliveredState) persistOrder(ctx context.Context, order *entities.Order) error {
//	_, err := global.Singletons.OrderRepository.Save(*order)
//	if err != nil {
//		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentDelivered.Name(), order, err.Error())
//	}
//	return err
//}
//
//func (shipmentDelivered shipmentDeliveredState) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, reason string, isSetExpireTime bool, itemStatus string) {
//
//	findFlag := false
//	if itemsId != nil && len(itemsId) > 0 {
//		for _, id := range itemsId {
//			findFlag = false
//			for i := 0; i < len(order.Items); i++ {
//				if order.Items[i].ItemId == id {
//					shipmentDelivered.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
//					findFlag = true
//					break
//				}
//			}
//			if !findFlag {
//				logger.Err("%s received itemId %d not exist in order, orderId: %d", shipmentDelivered.Name(), id, order.OrderId)
//			}
//		}
//	} else {
//		for i := 0; i < len(order.Items); i++ {
//			shipmentDelivered.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
//		}
//	}
//}
//
//func (shipmentDelivered shipmentDeliveredState) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
//	actionName string, result bool, reason string, isSetExpireTime bool, itemStatus string) {
//
//	order.Items[index].Status = itemStatus
//	order.Items[index].UpdatedAt = time.Now().UTC()
//
//	length := len(order.Items[index].Progress.StepsHistory) - 1
//
//	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
//		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
//	}
//
//	var action entities.Action
//	if isSetExpireTime {
//		// TODO must be checked calculation
//		expiredTime := order.Items[index].UpdatedAt.Add(time.Hour*
//			time.Duration(order.Items[index].ShipmentSpec.ShippingTime) +
//			time.Minute*time.Duration(0) +
//			time.Second*time.Duration(0))
//
//		action = entities.Action{
//			Name:   actionName,
//			Result: result,
//			Reason: reason,
//			Data: map[string]interface{}{
//				"expiredTime": expiredTime,
//			},
//			CreatedAt: order.Items[index].UpdatedAt,
//		}
//	} else {
//		action = entities.Action{
//			Name:      actionName,
//			Result:    result,
//			Reason:    reason,
//			CreatedAt: order.Items[index].UpdatedAt,
//		}
//	}
//
//	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
//}
