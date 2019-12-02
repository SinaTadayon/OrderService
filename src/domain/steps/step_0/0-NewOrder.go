package new_order_step

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
	pb "gitlab.faza.io/protos/order"
	"time"
)

const (
	stepName      string = "New_Order"
	stepIndex     int    = 0
	StockReserved        = "StockReserved"
	StockReleased        = "StockReleased"
)

type newOrderProcessingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &newOrderProcessingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (newOrderProcessing newOrderProcessingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	var requestNewOrder pb.RequestNewOrder

	logger.Audit("New Order Received . . .")

	if err := ptypes.UnmarshalAny(request.Data, &requestNewOrder); err != nil {
		logger.Err("Could not unmarshal requestNewOrder from anything field, error: %s, request: %v", err, request)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.BadRequest, Reason: "Invalid requestNewOrder"}}
		close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	//timestamp, err := ptypes.Timestamp(request.Time)
	//if err != nil {
	//	logger.Err("timestamp of requestNewOrder invalid, error: %s, requestNewOrder: %v", err, requestNewOrder)
	//	returnChannel := make(chan promise.FutureData, 1)
	//	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.BadRequest, Reason:"Invalid Request Timestamp"}}
	//	defer close(returnChannel)
	//	return promise.NewPromise(returnChannel, 1, 1)
	//}

	value, err := global.Singletons.Converter.Map(requestNewOrder, entities.Order{})
	if err != nil {
		logger.Err("Converter.Map requestNewOrder to order object failed, error: %s, requestNewOrder: %v", err, requestNewOrder)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.BadRequest, Reason: "Received requestNewOrder invalid"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrder := value.(*entities.Order)
	//newOrderEvent := actor_event.NewActorEvent(actors.CheckoutActor, checkout_action.NewOf(checkout_action.NewOrderAction),
	//	newOrder, nil, nil, timestamp)
	//
	//checkoutState, ok := newOrderProcessing.StatesMap()[0].(listener_state.IListenerState)
	//if ok != true || checkoutState.ActorType() != actors.CheckoutActor {
	//	logger.Err("checkout state doesn't exist in index 0 of statesMap, requestNewOrder: %v", requestNewOrder)
	//	returnChannel := make(chan promise.FutureData, 1)
	//	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	//	defer close(returnChannel)
	//	return promise.NewPromise(returnChannel, 1, 1)
	//}

	//for _, item := range newOrder.Items {
	//	userCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	//	ipromise := global.Singletons.UserService.GetSellerProfile(userCtx, strconv.Itoa(int(item.SellerInfo.SellerId)))
	//	futureData := ipromise.Data()
	//	if futureData.Ex != nil {
	//		logger.Err("ProcessMessage() => get NewOrder sellerProfile failed, order: %v, sellerId: %d",
	//			newOrder, item.SellerInfo.SellerId)
	//		returnChannel := make(chan promise.FutureData, 1)
	//		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.BadRequest, Reason: "Received requestNewOrder invalid"}}
	//		defer close(returnChannel)
	//		return promise.NewPromise(returnChannel, 1, 1)
	//	}
	//
	//	sellerInfo, ok := futureData.Data.(entities.SellerProfile)
	//	if ok != true {
	//		logger.Err("ProcessMessage() => get NewOrder sellerProfile invalid, order: %v, sellerId: %d",
	//			newOrder, item.SellerInfo.SellerId)
	//		returnChannel := make(chan promise.FutureData, 1)
	//		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.BadRequest, Reason: "Received requestNewOrder invalid"}}
	//		defer close(returnChannel)
	//		return promise.NewPromise(returnChannel, 1, 1)
	//	}
	//	item.SellerInfo.Profile = &sellerInfo
	//}

	newOrderProcessing.UpdateAllOrderStatus(ctx, newOrder, nil, steps.NewStatus, false)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	if err != nil {
		logger.Err("Save NewOrder Step Failed, error: %s, order: %v", err, newOrder)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	iPromise := global.Singletons.StockService.BatchStockActions(ctx, *order, nil, StockReserved)
	futureData := iPromise.Data()
	if futureData == nil {
		newOrderProcessing.UpdateAllOrderStatus(ctx, order, nil, steps.ClosedStatus, true)
		newOrderProcessing.updateOrderItemsProgress(ctx, order, nil, StockReserved, false, steps.ClosedStatus)
		if err := newOrderProcessing.persistOrder(ctx, order); err != nil {
		}
		logger.Err("StockService promise channel has been closed, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		go func() {
			newOrderProcessing.Childes()[0].ProcessOrder(ctx, *order, nil, nil)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if futureData.Ex != nil {
		newOrderProcessing.UpdateAllOrderStatus(ctx, order, nil, steps.ClosedStatus, true)
		newOrderProcessing.updateOrderItemsProgress(ctx, order, nil, StockReserved, false, steps.ClosedStatus)
		if err := newOrderProcessing.persistOrder(ctx, order); err != nil {
		}
		logger.Err("Reserved stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		go func() {
			newOrderProcessing.Childes()[0].ProcessOrder(ctx, *order, nil, nil)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrderProcessing.updateOrderItemsProgress(ctx, order, nil, StockReserved, true, steps.NewStatus)
	if err := newOrderProcessing.persistOrder(ctx, order); err != nil {
		newOrderProcessing.releasedStock(ctx, order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}

		go func() {
			newOrderProcessing.Childes()[0].ProcessOrder(ctx, *order, nil, nil)
		}()

		return promise.NewPromise(returnChannel, 1, 1)
	}

	return newOrderProcessing.Childes()[1].ProcessOrder(ctx, *order, nil, "PaymentCallbackUrlRequest")
	//return checkoutState.ActionListener(ctx, newOrderEvent, nil)
}

func (newOrderProcessing newOrderProcessingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {
	panic("implementation required")
}
func (newOrderProcessing newOrderProcessingStep) releasedStock(ctx context.Context, order *entities.Order) {
	iPromise := global.Singletons.StockService.BatchStockActions(ctx, *order, nil, StockReleased)
	futureData := iPromise.Data()
	if futureData == nil {
		newOrderProcessing.updateOrderItemsProgress(ctx, order, nil, StockReleased, false, steps.ClosedStatus)
		logger.Err("StockService promise channel has been closed, step: %s, orderId: %d", newOrderProcessing.Name(), order.OrderId)
		return
	}

	if futureData.Ex != nil {
		newOrderProcessing.updateOrderItemsProgress(ctx, order, nil, StockReleased, false, steps.ClosedStatus)
		logger.Err("Reserved stock from stockService failed, step: %s, orderId: %d, error: %s", newOrderProcessing.Name(), order.OrderId, futureData.Ex.Error())
		return
	}

	newOrderProcessing.updateOrderItemsProgress(ctx, order, nil, StockReleased, true, steps.ClosedStatus)
	logger.Audit("Reserved stock from stockService success, step: %s, orderId: %d", newOrderProcessing.Name(), order.OrderId)
}

func (newOrderProcessing newOrderProcessingStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", newOrderProcessing.Name(), order, err.Error())
	}
	return err
}

func (newOrderProcessing newOrderProcessingStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					newOrderProcessing.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
					findFlag = true
				}
			}

			if !findFlag {
				logger.Err("%s received itemId %d not exist in order, orderId: %d", newOrderProcessing.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			newOrderProcessing.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
		}
	}
}

func (newOrderProcessing newOrderProcessingStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
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
