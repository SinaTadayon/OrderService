package payment_control_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/item"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName string 	= "Payment_Control"
	stepIndex int		= 13
)

type paymentControlStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, orderRepository order.IOrderRepository,
	itemRepository item.IItemRepository, states ...states.IState) steps.IStep {
	return &paymentControlStep{steps.NewBaseStep(stepName, stepIndex, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewOf(name string, index int, orderRepository order.IOrderRepository,
	itemRepository item.IItemRepository, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentControlStep{steps.NewBaseStep(name, index, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &paymentControlStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentControl paymentControlStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (paymentControl paymentControlStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string) promise.IPromise {
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
