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
	SellerApprovalPending = "SellerApprovalPending"
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
		sellerApprovalPending.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
		sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, SellerApprovalPending, true, "", true)
		sellerApprovalPending.persistOrder(ctx, &order)
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
			sellerApprovalPending.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
			sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, Approved, true, "", false)
			sellerApprovalPending.persistOrder(ctx, &order)
			return sellerApprovalPending.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
		} else if req.Action == "failed" {
			actionData := req.Data.(*message.RequestSellerOrderAction_Failed)
			if ok != true {
				logger.Err("request data not a message.RequestSellerOrderAction_Failed type , order: %v", order)
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
				return promise.NewPromise(returnChannel, 1, 1)
			}
			sellerApprovalPending.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
			sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, Approved, false, actionData.Failed.Reason, false)
			sellerApprovalPending.persistOrder(ctx, &order)
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
				if order.Items[i].ItemId == id && order.Items[i].Status != SellerApprovalPending {
						return false
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			if order.Items[i].Status != SellerApprovalPending {
				return false
			}
		}
	}

	return true
}

func (sellerApprovalPending sellerApprovalPendingStep) persistOrder(ctx context.Context, order *entities.Order) {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", sellerApprovalPending.Name(), order, err.Error())
	}
}

func (sellerApprovalPending sellerApprovalPendingStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order,
	itemsId []string, action string, result bool, reason string, isSetExpireTime bool) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					sellerApprovalPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", sellerApprovalPending.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			sellerApprovalPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime)
		}
	}
}

// TODO set time from redis config
func (sellerApprovalPending sellerApprovalPendingStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, reason string, isSetExpireTime bool) {

	order.Items[index].Status = actionName
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
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

	order.Items[index].Progress.ActionHistory = append(order.Items[index].Progress.ActionHistory, action)
}


//
//import "gitlab.faza.io/order-project/order-service"
//
//func SellerApprovalPendingApproved(ppr PaymentPendingRequest) error {
//	err := main.MoveOrderToNewState("seller", "", main.ShipmentPending, "shipment-pending", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//// TODO: Improvement SellerApprovalPendingRejected
//func SellerApprovalPendingRejected(ppr PaymentPendingRequest, reason string) error {
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
