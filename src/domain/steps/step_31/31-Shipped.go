package shipped_step

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
	stepName string 	= "Shipped"
	stepIndex int		= 31
)

type shippedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, states ...states.IState) steps.IStep {
	return &shippedStep{steps.NewBaseStep(stepName, stepIndex, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewOf(name string, index int, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shippedStep{steps.NewBaseStep(name, index, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shippedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipped shippedStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (shipped shippedStep) ProcessOrder(ctx context.Context, order entities.Order) promise.IPromise {
	panic("implementation required")
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
