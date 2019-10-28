package return_shipment_delivery_delayed_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Return_Shipment_Delivery_Delayed"
	stepIndex int		= 52
)

type returnShipmentDeliveryDelayedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, states ...states.IState) steps.IStep {
	return &returnShipmentDeliveryDelayedStep{steps.NewBaseStep(stepName, stepIndex, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewOf(name string, index int, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &returnShipmentDeliveryDelayedStep{steps.NewBaseStep(name, index, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &returnShipmentDeliveryDelayedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (returnShipmentDeliveryDelayed returnShipmentDeliveryDelayedStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (returnShipmentDeliveryDelayed returnShipmentDeliveryDelayedStep) ProcessOrder(ctx context.Context, order entities.Order) promise.IPromise {
	panic("implementation required")
}



//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func ReturnShipmentDeliveryDelayedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.ReturnShipmentDeliveryDelayed)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func ReturnShipmentDeliveryDelayedAction(message *sarama.ConsumerMessage) error {
//
//	err := ReturnShipmentDeliveryDelayedProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func ReturnShipmentDeliveryDelayedProduce(topic string, payload []byte) error {
//	return nil
//}
