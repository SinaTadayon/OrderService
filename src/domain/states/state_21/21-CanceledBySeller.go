package state_21

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
	"time"
)

const (
	stepName         string = "Canceled_By_Seller"
	stepIndex        int    = 21
	RejectedBySeller        = "RejectedBySeller"
)

type shipmentRejectedBySellerStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentRejectedBySellerStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &shipmentRejectedBySellerStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &shipmentRejectedBySellerStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {

	logger.Audit("shipmentRejectedBySeller step, orderId: %d", order.OrderId)

	if len(order.Items) == len(itemsId) {
		shipmentRejectedBySeller.UpdateAllOrderStatus(ctx, &order, itemsId, states.ClosedStatus, false)
	} else {
		shipmentRejectedBySeller.UpdateAllOrderStatus(ctx, &order, itemsId, states.InProgressStatus, false)
	}

	shipmentRejectedBySeller.updateOrderItemsProgress(ctx, &order, itemsId, RejectedBySeller, true, states.ClosedStatus)
	if err := shipmentRejectedBySeller.persistOrder(ctx, &order); err != nil {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}
	return shipmentRejectedBySeller.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentRejectedBySeller.Name(), order, err.Error())
	}

	return err
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentRejectedBySeller.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
					findFlag = true
					break
				}
			}

			if findFlag == false {
				logger.Err("%s received itemId %d not exist in order, orderId: %d", shipmentRejectedBySeller.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentRejectedBySeller.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
		}
	}
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
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
//// TODO Must be implement
//func ShipmentRejectedBySellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.ShipmentRejectedBySeller)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func ShipmentRejectedBySellerAction(message *sarama.ConsumerMessage) error {
//
//	err := ShipmentRejectedBySellerProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func ShipmentRejectedBySellerProduce(topic string, payload []byte) error {
//	return nil
//}
