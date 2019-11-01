package shipment_rejected_by_seller_step

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
	stepName string 	= "Shipment_Rejected_By_Seller"
	stepIndex int		= 21
)

type shipmentRejectedBySellerStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentRejectedBySellerStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentRejectedBySellerStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentRejectedBySellerStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {
	shipmentRejectedBySeller.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
	shipmentRejectedBySeller.updateOrderItemsProgress(ctx, &order, itemsId, "RejectedBySeller", true)
	shipmentRejectedBySeller.persistOrder(ctx, &order)
	return shipmentRejectedBySeller.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) persistOrder(ctx context.Context, order *entities.Order) {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentRejectedBySeller.Name(), order, err.Error())
	}
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					shipmentRejectedBySeller.doUpdateOrderItemsProgress(ctx, order, i, action, result)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", shipmentRejectedBySeller.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipmentRejectedBySeller.doUpdateOrderItemsProgress(ctx, order, i, action, result)
		}
	}
}

func (shipmentRejectedBySeller shipmentRejectedBySellerStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool) {

	order.Items[index].Status = actionName
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.ActionHistory = append(order.Items[index].Progress.ActionHistory, action)
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
