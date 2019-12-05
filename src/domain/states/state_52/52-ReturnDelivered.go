package state_52

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Return_Delivery"
	stepIndex int    = 52
)

type returnShipmentDeliveryDelayedStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &returnShipmentDeliveryDelayedStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &returnShipmentDeliveryDelayedStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &returnShipmentDeliveryDelayedStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (returnShipmentDeliveryDelayed returnShipmentDeliveryDelayedStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (returnShipmentDeliveryDelayed returnShipmentDeliveryDelayedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
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
