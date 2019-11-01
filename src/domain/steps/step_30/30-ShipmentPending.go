package shipment_pending_step

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
	stepName string 	= "Shipment_Pending"
	stepIndex int		= 30
	Shipped				= "Shipped"
)

type shipmentPendingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentPendingStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentPendingStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentPendingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentPending shipmentPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipmentPending shipmentPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {
	shipmentPending.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
	shipmentPending.updateOrderItemsProgress(ctx, &order, itemsId, "SellerShipmentPending", true)
	shipmentPending.persistOrder(ctx, &order)
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data:nil, Ex:nil}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (shipmentPending shipmentPendingStep) persistOrder(ctx context.Context, order *entities.Order) {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentPending.Name(), order, err.Error())
	}
}

func (shipmentPending shipmentPendingStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", shipmentPending.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result)
		}
	}
}

func (shipmentPending shipmentPendingStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool) {

	order.Items[index].Status = actionName
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	expiredTime := order.Items[index].UpdatedAt.Add(
		time.Hour * time.Duration(order.Items[index].ShipmentSpec.ReactionTime) +
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


//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	OrderService "gitlab.faza.io/protos/order"
//)
//
//func ShipmentPendingEnteredDetail(ppr PaymentPendingRequest, req *OrderService.ShipmentDetailRequest) error {
//	ppr.ShipmentDetail.ShipmentDetail.ShipmentProvider = req.ShipmentProvider
//	ppr.ShipmentDetail.ShipmentDetail.ShipmentTrackingNumber = req.ShipmentTrackingNumber
//	ppr.ShipmentDetail.ShipmentDetail.Description = req.GetDescription()
//	err := main.MoveOrderToNewState("seller", "", main.Shipped, "shipped", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
