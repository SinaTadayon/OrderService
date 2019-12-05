package state_40

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	message "gitlab.faza.io/protos/order"
	"time"
)

const (
	stepName        string = "Return_Request_Pending"
	stepIndex       int    = 40
	ShipmentSuccess        = "ShipmentSuccess"
	StockSettlement        = "StockSettlement"
)

type shipmentSuccessStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentSuccessStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentSuccessStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &shipmentSuccessStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (shipmentSuccess shipmentSuccessStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (shipmentSuccess shipmentSuccessStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {

	logger.Audit("shipmentSuccess step, orderId: %d", order.OrderId)
	shipmentSuccess.UpdateAllOrderStatus(ctx, &order, itemsId, states.InProgressStatus, false)

	iPromise := global.Singletons.StockService.BatchStockActions(ctx, order, itemsId, StockSettlement)
	futureData := iPromise.Get()
	if futureData == nil {
		shipmentSuccess.updateOrderItemsProgress(ctx, &order, itemsId, StockSettlement, false, states.ClosedStatus)
		if err := shipmentSuccess.persistOrder(ctx, &order); err != nil {
		}
		logger.Err("StockService future channel has been closed, order: %d", order.OrderId)
	} else if futureData.Ex != nil {
		shipmentSuccess.updateOrderItemsProgress(ctx, &order, itemsId, StockSettlement, false, states.ClosedStatus)
		if err := shipmentSuccess.persistOrder(ctx, &order); err != nil {
		}
		logger.Err("Settlement stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
	}

	shipmentSuccess.updateOrderItemsProgress(ctx, &order, itemsId, ShipmentSuccess, true, states.InProgressStatus)
	if err := shipmentSuccess.persistOrder(ctx, &order); err != nil {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}
	return shipmentSuccess.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
}

func (shipmentSuccess shipmentSuccessStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentSuccess.Name(), order, err.Error())
	}
	return err
}

func (shipmentSuccess shipmentSuccessStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {

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
				logger.Err("%s received itemId %d not exist in order, orderId: %d", shipmentSuccess.Name(), id, order.OrderId)
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
