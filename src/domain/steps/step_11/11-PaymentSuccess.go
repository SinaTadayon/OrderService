package payment_success_step

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
	"time"
)

const (
	stepName string 	= "Payment_Success"
	stepIndex int		= 11
	PaymentSuccess 		= "PaymentSuccess"
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
func (paymentSuccess paymentSuccessStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {
	//nextToStepState, ok := paymentSuccess.Childes()[2].(launcher_state.ILauncherState)
	//if ok != true || nextToStepState.ActiveType() != actives.StockAction {
	//	logger.Err("nextToStepState doesn't exist in index 2 of %s Childes() , order: %v", paymentSuccess.Name(), order)
	//	returnChannel := make(chan promise.FutureData, 1)
	//	defer close(returnChannel)
	//	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	//	return promise.NewPromise(returnChannel, 1, 1)
	//}
	logger.Audit("Order Received in %s step, orderId: %s, Action: %s", paymentSuccess.Name(), order.OrderId, PaymentSuccess)
	paymentSuccess.UpdateOrderStep(ctx, &order, itemsId, "InProgress", false)
	paymentSuccess.updateOrderItemsProgress(ctx, &order, nil, PaymentSuccess, true)
	if err := paymentSuccess.persistOrder(ctx, &order); err != nil {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	return paymentSuccess.Childes()[1].ProcessOrder(ctx, order, nil, nil)
	//return nextToStepState.ActionLauncher(ctx, order, itemsId, buyer_action.ApprovedAction)
}

func (paymentSuccess paymentSuccessStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", paymentSuccess.Name(), order, err.Error())
	}

	return err
}

func (paymentSuccess paymentSuccessStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string, action string, result bool) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					paymentSuccess.doUpdateOrderItemsProgress(ctx, order, i, action, result)
					findFlag = true
				}
			}
			if !findFlag {
				logger.Err("%s received itemId %s not exist in order, orderId: %v", paymentSuccess.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			paymentSuccess.doUpdateOrderItemsProgress(ctx, order, i, action, result)
		}
	}
}

func (paymentSuccess paymentSuccessStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool) {

	order.Items[index].Status = paymentSuccess.Name()
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.ActionHistory = append(order.Items[index].Progress.ActionHistory, action)
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
