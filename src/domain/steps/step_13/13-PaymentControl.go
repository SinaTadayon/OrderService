package payment_control_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Payment_Control"
	stepIndex int		= 13
)

type paymentControlStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentControlStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentControlStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &paymentControlStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentControl paymentControlStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (paymentControl paymentControlStep) ProcessOrder(ctx context.Context, order entities.Order) error {
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
