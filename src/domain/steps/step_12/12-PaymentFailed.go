package payment_failed_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Payment_Failed"
	stepIndex int		= 12
)

type paymentFailedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentFailedStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentFailedStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &paymentFailedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentFailed paymentFailedStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (paymentFailed paymentFailedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string) promise.IPromise {
	panic("implementation required")
}


//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func PaymentFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PaymentFailed)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PaymentFailedAction(message *sarama.ConsumerMessage) error {
//
//	err := PaymentFailedProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PaymentFailedProduce(topic string, payload []byte) error {
//	return nil
//}
