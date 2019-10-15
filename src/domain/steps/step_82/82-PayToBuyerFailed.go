package pay_to_buyer_failed_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Pay_To_Buyer_Failed"
	stepIndex int		= 82
)

type payToBuyerFailedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToBuyerFailedStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToBuyerFailedStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &payToBuyerFailedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (payToBuyerFailed payToBuyerFailedStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (payToBuyerFailed payToBuyerFailedStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}


//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func PayToBuyerFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PayToBuyerFailed)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PayToBuyerFailedAction(message *sarama.ConsumerMessage) error {
//
//	err := PayToBuyerFailedProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PayToBuyerFailedProduce(topic string, payload []byte) error {
//	return nil
//}
