package shipment_success_step

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
	stepName string 	= "Shipment_Success"
	stepIndex int		= 40
	ShipmentSuccess		= "ShipmentSuccess"
	StockSettlement		= "StockSettlement"
)

type shipmentSuccessStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentSuccessStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentSuccessStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentSuccessStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentSuccess shipmentSuccessStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipmentSuccess shipmentSuccessStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	logger.Audit("shipmentSuccess step, orderId: %s", order.OrderId)
	shipmentSuccess.UpdateAllOrderStatus(ctx, &order, itemsId, steps.InProgressStatus, false)

	iPromise := global.Singletons.StockService.BatchStockActions(ctx, order, itemsId, StockSettlement)
	futureData := iPromise.Data()
	if futureData == nil {
		shipmentSuccess.updateOrderItemsProgress(ctx, &order, itemsId, StockSettlement, false, steps.ClosedStatus)
		if err := shipmentSuccess.persistOrder(ctx, &order); err != nil {}
		logger.Err("StockService promise channel has been closed, order: %s", order.OrderId)
	} else if futureData.Ex != nil {
		shipmentSuccess.updateOrderItemsProgress(ctx, &order, itemsId, StockSettlement, false, steps.ClosedStatus)
		if err := shipmentSuccess.persistOrder(ctx, &order); err != nil {}
		logger.Err("Settlement stock from stockService failed, error: %s, orderId: %s", futureData.Ex.Error(), order.OrderId)
	}

	shipmentSuccess.updateOrderItemsProgress(ctx, &order, itemsId, ShipmentSuccess, true, steps.InProgressStatus)
	if err := shipmentSuccess.persistOrder(ctx, &order); err != nil {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}
	return shipmentSuccess.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
}

func (shipmentSuccess shipmentSuccessStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentSuccess.Name(), order, err.Error())
	}
	return err
}

func (shipmentSuccess shipmentSuccessStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentSuccess.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
					findFlag = true
					break
				}
			}
			if !findFlag {
				logger.Err("%s received itemId %s not exist in order, order: %v", shipmentSuccess.Name(), id, order)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentSuccess.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
		}
	}
}

func (shipmentSuccess shipmentSuccessStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)

}


//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//// TODO: Transition state to 90.Pay_TO_Seller state
//func ShipmentSuccessAction(ppr PaymentPendingRequest, req *pb.ShipmentSuccessRequest) error {
//	err := main.MoveOrderToNewState("buyer", "", main.ShipmentSuccess, "shipment-success", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
