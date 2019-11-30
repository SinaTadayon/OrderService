package state_51

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Return_Shipped"
	stepIndex int    = 51
)

type returnShipmentDeliveryPendingStep struct {
	*states.BaseStepImpl
}

func New(childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &returnShipmentDeliveryPendingStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &returnShipmentDeliveryPendingStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStepImpl) states.IStep {
	return &returnShipmentDeliveryPendingStep{base}
}

func NewValueOf(base *states.BaseStepImpl, params ...interface{}) states.IStep {
	panic("implementation required")
}

func (returnShipmentDeliveryPending returnShipmentDeliveryPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (returnShipmentDeliveryPending returnShipmentDeliveryPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {
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
