package state_51

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Return_Shipped"
	stepIndex int    = 51
)

type returnShipmentDeliveryPendingStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &returnShipmentDeliveryPendingStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &returnShipmentDeliveryPendingStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &returnShipmentDeliveryPendingStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (returnShipmentDeliveryPending returnShipmentDeliveryPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (returnShipmentDeliveryPending returnShipmentDeliveryPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
	panic("implementation required")
}

//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func ReturnShipmentDeliveryPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.ReturnShipmentDeliveryPending)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func ReturnShipmentDeliveryPendingAction(message *sarama.ConsumerMessage) error {
//
//	err := ReturnShipmentDeliveryPendingProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func ReturnShipmentDeliveryPendingProduce(topic string, payload []byte) error {
//	return nil
//}
