package pay_to_seller_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Pay_To_Seller"
	stepIndex int		= 90
)

type payToSellerStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToSellerStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToSellerStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &payToSellerStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (payToSeller payToSellerStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (payToSeller payToSellerStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string) promise.IPromise {
	panic("implementation required")
}



//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func PayToSellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PayToSeller)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PayToSellerAction(message *sarama.ConsumerMessage) error {
//
//	err := PayToSellerProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PayToSellerProduce(topic string, payload []byte) error {
//	return nil
//}
