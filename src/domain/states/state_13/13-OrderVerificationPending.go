package state_13

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Order_Verification_Pending"
	stepIndex int    = 13
)

type paymentControlStep struct {
	*states.BaseStepImpl
}

func New(childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &paymentControlStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IStep, states ...states_old.IState) states.IStep {
	return &paymentControlStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStepImpl) states.IStep {
	return &paymentControlStep{base}
}

func NewValueOf(base *states.BaseStepImpl, params ...interface{}) states.IStep {
	panic("implementation required")
}

func (paymentControl paymentControlStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (paymentControl paymentControlStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {
	panic("implementation required")
}

//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func PaymentControlMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PaymentControl)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PaymentControlAction(message *sarama.ConsumerMessage) error {
//
//	err := PaymentControlProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PaymentControlProduce(topic string, payload []byte) error {
//	return nil
//}
