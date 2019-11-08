package seller_approval_pending_step

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
	stepName string 	= "Seller_Approval_Pending"
	stepIndex int		= 20
	Approved			= "Approved"
	ApprovalPending = "ApprovalPending"
)

type sellerApprovalPendingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &sellerApprovalPendingStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &sellerApprovalPendingStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &sellerApprovalPendingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (sellerApprovalPending sellerApprovalPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (sellerApprovalPending sellerApprovalPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	if param == nil {
		logger.Audit("Order Received in %s step, orderId: %s, Action: %s", sellerApprovalPending.Name(), order.OrderId, ApprovalPending)
		sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, steps.InProgressStatus, false)
		sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, ApprovalPending, true, "", true, steps.InProgressStatus)
		if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {}
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: nil}
		return promise.NewPromise(returnChannel, 1, 1)
	} else {
		logger.Audit("Order Received in %s step, orderId: %s, Action: %s", sellerApprovalPending.Name(), order.OrderId, Approved)
		req, ok := param.(*message.RequestSellerOrderAction)
		if ok != true {
			//if len(order.Items) == len(itemsId) {
			//	sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, nil, "CLOSED", false)
			//} else {
			//	sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, nil, "InProgress", false)
			//}
			//sellerApprovalPending.persistOrder(ctx, &order)

			logger.Err("param not a message.RequestSellerOrderAction type , order: %v", order)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		if !sellerApprovalPending.validateAction(ctx, &order, itemsId) {
			logger.Err("%s step received invalid action, order: %v, action: %s", sellerApprovalPending.Name(), order, req.Action)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: nil, Ex:promise.FutureError{Code:promise.NotAccepted, Reason:"Action Expired"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		if req.Action == "success" {
			sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, steps.InProgressStatus, true)
			sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, Approved, true, "", false, steps.InProgressStatus)
			if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
				return promise.NewPromise(returnChannel, 1, 1)
			}
			return sellerApprovalPending.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
		} else if req.Action == "failed" {
			if req.Data == nil {
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.BadRequest, Reason:"Reason Data Required"}}
				return promise.NewPromise(returnChannel, 1, 1)
			}

			actionData := req.Data.(*message.RequestSellerOrderAction_Failed)
			if ok != true {
				logger.Err("request data not a message.RequestSellerOrderAction_Failed type , order: %v", order)
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
				return promise.NewPromise(returnChannel, 1, 1)
			}
			if len(order.Items) == len(itemsId) {
				sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, steps.ClosedStatus, true)
			} else {
				sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, steps.InProgressStatus, true)
			}
			sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, Approved, false, actionData.Failed.Reason, false, steps.ClosedStatus)
			if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
				return promise.NewPromise(returnChannel, 1, 1)
			}

			return sellerApprovalPending.Childes()[1].ProcessOrder(ctx, order, itemsId, nil)
		}

		logger.Err("%s step received invalid action, order: %v, action: %s", sellerApprovalPending.Name(), order, req.Action)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}
}

func (sellerApprovalPending sellerApprovalPendingStep) validateAction(ctx context.Context, order *entities.Order,
	itemsId []string) bool {
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				length := len(order.Items[i].Progress.StepsHistory)-1
				if order.Items[i].ItemId == id && order.Items[i].Progress.StepsHistory[length].Name != sellerApprovalPending.Name() {
						return false
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			length := len(order.Items[i].Progress.StepsHistory) - 1
			if order.Items[i].Progress.StepsHistory[length].Name != sellerApprovalPending.Name() {
				return false
			}
		}
	}

	return true
}

func (sellerApprovalPending sellerApprovalPendingStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", sellerApprovalPending.Name(), order, err.Error())
	}

	return err
}

func (sellerApprovalPending sellerApprovalPendingStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order,
	itemsId []string, action string, result bool, reason string, isSetExpireTime bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					sellerApprovalPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
					findFlag = true
					break
				}
			}
			if findFlag == false {
				logger.Err("%s received itemId %s not exist in order, orderId: %v", sellerApprovalPending.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			sellerApprovalPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
		}
	}
}

// TODO set time from redis config
func (sellerApprovalPending sellerApprovalPendingStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, reason string, isSetExpireTime bool, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	var action entities.Action
	if isSetExpireTime {
		expiredTime := order.Items[index].UpdatedAt.Add(time.Hour *
			time.Duration(24) +
			time.Minute * time.Duration(0) +
			time.Second * time.Duration(0))

		action = entities.Action{
			Name:      actionName,
			Result:    result,
			Reason:    reason,
			Data:		map[string]interface{}{
						"expiredTime": expiredTime,
						},
			CreatedAt: order.Items[index].UpdatedAt,
		}
	} else {
		action = entities.Action{
			Name:      actionName,
			Result:    result,
			Reason:    reason,
			CreatedAt: order.Items[index].UpdatedAt,
		}
	}

	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
}


//
//import "gitlab.faza.io/order-project/order-service"
//
//func ApprovalPendingApproved(ppr PaymentPendingRequest) error {
//	err := main.MoveOrderToNewState("seller", "", main.ShipmentPending, "shipment-pending", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//// TODO: Improvement ApprovalPendingRejected
//func ApprovalPendingRejected(ppr PaymentPendingRequest, reason string) error {
//	err := main.MoveOrderToNewState("seller", reason, main.ShipmentRejectedBySeller, "shipment-rejected-by-seller", ppr)
//	if err != nil {
//		return err
//	}
//	newPpr, err := main.GetOrder(ppr.OrderNumber)
//	if err != nil {
//		return err
//	}
//	err = main.MoveOrderToNewState("system", reason, main.PayToBuyer, "pay-to-buyer", newPpr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
