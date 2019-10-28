package payment_success_step

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
	stepName string 	= "Payment_Success"
	stepIndex int		= 11
)

type paymentSuccessStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, states ...states.IState) steps.IStep {
	return &paymentSuccessStep{steps.NewBaseStep(stepName, stepIndex, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewOf(name string, index int, orderRepository repository.IOrderRepository,
	itemRepository repository.IItemRepository, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentSuccessStep{steps.NewBaseStep(name, index, orderRepository,
		itemRepository, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &paymentSuccessStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentSuccess paymentSuccessStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (paymentSuccess paymentSuccessStep) ProcessOrder(ctx context.Context, order entities.Order) promise.IPromise {
	panic("implementation required")
}

//import (
//	"encoding/json"
//	"gitlab.faza.io/order-project/order-service"
//
//	"gitlab.faza.io/go-framework/logger"
//
//	"github.com/Shopify/sarama"
//)
//
//func PaymentSuccessMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.PaymentSuccess)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func PaymentSuccessAction(message *sarama.ConsumerMessage) error {
//	ppr := PaymentPendingRequest{}
//	err := json.Unmarshal(message.Value, &ppr)
//	if err != nil {
//		return err
//	}
//
//	// @TODO: remove automatic move status auto fraud detection
//	err = main.MoveOrderToNewState("system", "auto approval", main.SellerApprovalPending, "seller-approval-pending", ppr)
//	if err != nil {
//		return err
//	}
//
//	err = main.NotifySellerForNewOrder(ppr)
//	if err != nil {
//		logger.Err("cant notify seller, %v", err)
//	}
//	return nil
//}
