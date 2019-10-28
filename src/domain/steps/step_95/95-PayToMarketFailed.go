package pay_to_market_failed_step

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
	stepName string 	= "Pay_To_Market_Failed"
	stepIndex int		= 95
)

type payToMarketFailedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, states ...states.IState) steps.IStep {
	return &payToMarketFailedStep{steps.NewBaseStep(stepName, stepIndex, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewOf(name string, index int, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToMarketFailedStep{steps.NewBaseStep(name, index, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &payToMarketFailedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (payToMarketFailed payToMarketFailedStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (payToMarketFailed payToMarketFailedStep) ProcessOrder(ctx context.Context, order entities.Order) promise.IPromise {
	panic("implementation required")
}


//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//func PayToMarketFailedMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PayToMarketFailed)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PayToMarketFailedAction(message *sarama.ConsumerMessage) error {
//
//	err := PayToMarketFailedProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func PayToMarketFailedProduce(topic string, payload []byte) error {
//	return nil
//}
