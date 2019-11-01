package shipped_step

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
	stepName string 	= "Shipped"
	stepIndex int		= 31
	Shipped				= "Shipped"
)

type shippedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shippedStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shippedStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shippedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipped shippedStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipped shippedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {
	req, ok := param.(message.RequestSellerOrderAction)
	if ok != true {
		//if len(order.Items) == len(itemsId) {
		//	shipped.UpdateOrderStep(ctx, &order, nil, "CLOSED", false)
		//} else {
		//	shipped.UpdateOrderStep(ctx, &order, nil, "InProgress", false)
		//}
		//shipped.persistOrder(ctx, &order)

		logger.Err("param not a message.RequestSellerOrderAction type , order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if req.Action == "success" {
		actionData, ok := req.Data.(*message.RequestSellerOrderAction_Success)
		if ok != true {
			logger.Err("request data not a message.RequestSellerOrderAction_Success type , order: %v", order)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		shipped.UpdateOrderStep(ctx, &order, nil, "InProgress", false)
		shipped.updateOrderItemsProgress(ctx, &order, itemsId, Shipped, true, "", actionData)
		shipped.persistOrder(ctx, &order)
		return shipped.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
	} else if req.Action == "failed" {
		actionData, ok := req.Data.(*message.RequestSellerOrderAction_Failed)
		if ok != true {
			logger.Err("request data not a message.RequestSellerOrderAction_Failed type , order: %v", order)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}
		shipped.UpdateOrderStep(ctx, &order, nil, "InProgress", false)
		shipped.updateOrderItemsProgress(ctx, &order, itemsId, Shipped, false, actionData.Failed.Reason, nil)
		shipped.persistOrder(ctx, &order)
		return shipped.Childes()[1].ProcessOrder(ctx, order, itemsId, nil)
	}

	logger.Err("%s step received invalid action, order: %v, action: %s", shipped.Name(), order, req.Action)
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	return promise.NewPromise(returnChannel, 1, 1)}

func (shipped shippedStep) persistOrder(ctx context.Context, order *entities.Order) {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipped.Name(), order, err.Error())
	}
}

func (shipped shippedStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool, reason string, req *message.RequestSellerOrderAction_Success) {

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					if req != nil {
						order.Items[i].ShipmentDetails.SellerShipmentDetail = entities.ShipmentDetail{
							TrackingNumber: req.Success.TrackingId,
							ShippingMethod: req.Success.ShipmentMethod,
						}
					}
					shipped.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason)
				} else {
					logger.Err("%s received itemId %s not exist in order, order: %v", shipped.Name(), id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			shipped.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason)
		}
	}
}

func (shipped shippedStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, reason string) {

	order.Items[index].Status = actionName
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		Reason: 	reason,
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
//func ShippedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.Shipped)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func ShippedAction(message *sarama.ConsumerMessage) error {
//
//	err := ShippedProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func ShippedProduce(topic string, payload []byte) error {
//	return nil
//}
