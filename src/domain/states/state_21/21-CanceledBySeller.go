package state_21

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
)

const (
	stepName  string = "Canceled_By_Seller"
	stepIndex int    = 21
)

type canceledBySellerState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledBySellerState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledBySellerState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &canceledBySellerState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state canceledBySellerState) Process(ctx context.Context, iFrame frame.IFrame) {

}

//func (shipmentRejectedBySeller canceledBySellerState) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
//
//	logger.Audit("shipmentRejectedBySeller step, orderId: %d", order.OrderId)
//
//	if len(order.Items) == len(itemsId) {
//		shipmentRejectedBySeller.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderClosedStatus, false)
//	} else {
//		shipmentRejectedBySeller.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//	}
//
//	shipmentRejectedBySeller.updateOrderItemsProgress(ctx, &order, itemsId, RejectedBySeller, true, states.OrderClosedStatus)
//	if err := shipmentRejectedBySeller.persistOrder(ctx, &order); err != nil {
//		returnChannel := make(chan future.IDataFuture, 1)
//		defer close(returnChannel)
//		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		return future.NewFuture(returnChannel, 1, 1)
//	}
//	return shipmentRejectedBySeller.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
//}
//
//func (shipmentRejectedBySeller canceledBySellerState) persistOrder(ctx context.Context, order *entities.Order) error {
//	_, err := global.Singletons.OrderRepository.Save(*order)
//	if err != nil {
//		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentRejectedBySeller.Name(), order, err.Error())
//	}
//
//	return err
//}
//
//func (shipmentRejectedBySeller canceledBySellerState) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {
//
//	findFlag := false
//	if itemsId != nil && len(itemsId) > 0 {
//		for _, id := range itemsId {
//			findFlag = false
//			for i := 0; i < len(order.Items); i++ {
//				if order.Items[i].ItemId == id {
//					shipmentRejectedBySeller.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
//					findFlag = true
//					break
//				}
//			}
//
//			if findFlag == false {
//				logger.Err("%s received itemId %d not exist in order, orderId: %d", shipmentRejectedBySeller.Name(), id, order.OrderId)
//			}
//		}
//	} else {
//		for i := 0; i < len(order.Items); i++ {
//			shipmentRejectedBySeller.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
//		}
//	}
//}
//
//func (shipmentRejectedBySeller canceledBySellerState) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
//	actionName string, result bool, itemStatus string) {
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
//	action := entities.Action{
//		Name:      actionName,
//		Result:    result,
//		CreatedAt: order.Items[index].UpdatedAt,
//	}
//
//	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
//
//}
//
////
////import (
////	"github.com/Shopify/sarama"
////	"gitlab.faza.io/order-project/order-service"
////)
////
////// TODO Must be implement
////func ShipmentRejectedBySellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
////	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.ShipmentRejectedBySeller)
////	if err != nil {
////		return mess, err
////	}
////	return message, nil
////}
////
////func ShipmentRejectedBySellerAction(message *sarama.ConsumerMessage) error {
////
////	err := ShipmentRejectedBySellerProduce("", []byte{})
////	if err != nil {
////		return err
////	}
////	return nil
////}
////
////func ShipmentRejectedBySellerProduce(topic string, payload []byte) error {
////	return nil
////}
