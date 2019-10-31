package payment_success_step

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/buyer"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	launcher_state "gitlab.faza.io/order-project/order-service/domain/states/launcher"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName string 	= "Payment_Success"
	stepIndex int		= 11
)

type paymentSuccessStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentSuccessStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentSuccessStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &paymentSuccessStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentSuccess paymentSuccessStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}


// TODO PaymentApprovalAction must be handled and implement
// TODO notification must be handled and implement
func (paymentSuccess paymentSuccessStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string) promise.IPromise {
	nextToStepState, ok := paymentSuccess.Childes()[2].(launcher_state.ILauncherState)
	if ok != true || nextToStepState.ActiveType() != actives.StockAction {
		logger.Err("nextToStepState doesn't exist in index 2 of %s Childes() , order: %v", paymentSuccess.Name(), order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	paymentSuccess.UpdateOrderStep(ctx, &order, itemsId)
	return nextToStepState.ActionLauncher(ctx, order, itemsId, buyer_action.ApprovedAction)
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
