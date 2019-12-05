package state_12

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
	stepName      string = "Payment_Failed"
	stepIndex     int    = 12
	PaymentFailed        = "PaymentFailed"
)

type paymentFailedStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &paymentFailedStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &paymentFailedStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &paymentFailedStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (paymentFailed paymentFailedStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

// TODO states must be append step history and changes to order object
func (paymentFailed paymentFailedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
	//stockState, ok := paymentFailed.Childes()[0].(launcher_state.ILauncherState)
	//if ok != true || stockState.ActiveType() != actives.StockAction {
	//	logger.Err("stock state doesn't exist in index 0 of %s statesMap , order: %v", paymentFailed.ActionName(), order)
	//	returnChannel := make(chan future.IDataFuture, 1)
	//	defer close(returnChannel)
	//	returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.InternalError, Reason:"Unknown Error"}}
	//	return future.NewFuture(returnChannel, 1, 1)
	//}
	//
	//paymentFailed.UpdateAllOrderStatus(ctx, &order, itemsId, "CLOSED", true)
	//return stockState.ActionLauncher(ctx, order, itemsId, nil)

	paymentFailed.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderClosedStatus, false)
	paymentFailed.updateOrderItemsProgress(ctx, &order, itemsId, PaymentFailed, true, states.OrderClosedStatus)
	if err := paymentFailed.persistOrder(ctx, &order); err != nil {
	}
	returnChannel := make(chan future.IDataFuture, 1)
	defer close(returnChannel)
	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.NotAccepted, Reason: "Order Payment Failed"}}
	return future.NewFuture(returnChannel, 1, 1)
}

func (paymentFailed paymentFailedStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", paymentFailed.Name(), order, err.Error())
	}

	return err
}

func (paymentFailed paymentFailedStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64,
	action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					paymentFailed.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
					findFlag = true
					break
				}
			}

			if findFlag == false {
				logger.Err("%s received itemId %d not exist in order, orderId: %d", paymentFailed.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			paymentFailed.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
		}
	}
}

func (paymentFailed paymentFailedStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
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

//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func PaymentFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PaymentFailed)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PaymentFailedAction(message *sarama.ConsumerMessage) error {
//
//	err := PaymentFailedProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PaymentFailedProduce(topic string, payload []byte) error {
//	return nil
//}
