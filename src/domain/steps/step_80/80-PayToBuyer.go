package pay_to_buyer_step

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
	stepName string 	= "Pay_To_Buyer"
	stepIndex int		= 80
)

type payToBuyerStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToBuyerStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToBuyerStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &payToBuyerStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (payToBuyer payToBuyerStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (payToBuyer payToBuyerStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	if len(order.Items) == len(itemsId) {
		payToBuyer.UpdateOrderStep(ctx, &order, itemsId, "Closed", false)
	} else {
		payToBuyer.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
	}

	payToBuyer.updateOrderItemsProgress(ctx, &order, itemsId, "PayToBuyer", true)
	payToBuyer.persistOrder(ctx, &order)
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data:nil, Ex:nil}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (payToBuyer payToBuyerStep) persistOrder(ctx context.Context, order *entities.Order) {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", payToBuyer.Name(), order, err.Error())
	}
}

func (payToBuyer payToBuyerStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					payToBuyer.doUpdateOrderItemsProgress(ctx, order, i, action, result)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", payToBuyer.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			payToBuyer.doUpdateOrderItemsProgress(ctx, order, i, action, result)
		}
	}
}

func (payToBuyer payToBuyerStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
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
//// TODO: must be implemented
//func PayToBuyerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PayToBuyer)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PayToBuyerAction(message *sarama.ConsumerMessage) error {
//
//	err := PayToBuyerProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PayToBuyerProduce(topic string, payload []byte) error {
//	return nil
//}
