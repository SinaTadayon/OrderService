package state_90

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
	stepName  string = "Pay_To_Seller"
	stepIndex int    = 90
	Delivered        = "DELIVERED"
)

type payToSellerStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &payToSellerStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &payToSellerStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &payToSellerStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (payToSeller payToSellerStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (payToSeller payToSellerStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {

	logger.Audit("Pay to Seller step, orderId: %d", order.OrderId)

	if len(order.Items) == len(itemsId) {
		payToSeller.UpdateAllOrderStatus(ctx, &order, itemsId, states.ClosedStatus, false)
	} else {
		payToSeller.UpdateAllOrderStatus(ctx, &order, itemsId, states.InProgressStatus, false)
	}

	payToSeller.updateOrderItemsProgress(ctx, &order, itemsId, Delivered, true, states.ClosedStatus)
	if err := payToSeller.persistOrder(ctx, &order); err != nil {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}
	returnChannel := make(chan future.IDataFuture, 1)
	defer close(returnChannel)
	returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
	return future.NewFuture(returnChannel, 1, 1)
}

func (payToSeller payToSellerStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", payToSeller.Name(), order, err.Error())
	}
	return err
}

func (payToSeller payToSellerStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					payToSeller.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
					findFlag = true
					break
				}
			}

			if findFlag == false {
				logger.Err("%s received itemId %d not exist in order, orderId: %d", payToSeller.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			payToSeller.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
		}
	}
}

func (payToSeller payToSellerStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
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
//func PayToSellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PayToSeller)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PayToSellerAction(message *sarama.ConsumerMessage) error {
//
//	err := PayToSellerProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PayToSellerProduce(topic string, payload []byte) error {
//	return nil
//}
