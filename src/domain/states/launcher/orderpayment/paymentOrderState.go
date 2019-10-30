package order_payment_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"time"
)

const (
	stateName string = "Payment_Order_Action_State"
	activeType = actives.OrderPaymentAction
)

type orderPaymentActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &orderPaymentActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &orderPaymentActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &orderPaymentActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

func (orderPayment orderPaymentActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	paymentState, ok := orderPayment.Childes()[0].(listener_state.IListenerState)
	if ok != true {
		logger.Err("paymentState isn't child of orderPaymentState, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	nextToStepState, ok := orderPayment.Childes()[1].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("nextToStep isn't child of orderPaymentState, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	paymentRequest := payment_service.PaymentRequest{
		Amount:   int64(order.Amount.Total),
		Gateway:  order.Amount.PaymentOption,
		Currency: order.Amount.Currency,
		OrderId:  order.OrderId,
	}

	iPromise := global.Singletons.PaymentService.OrderPayment(ctx, paymentRequest)
	futureData := iPromise.Data()
	if futureData == nil {
		logger.Err("PaymentService.OrderPayment in orderPaymentState failed, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if futureData.Ex != nil {
		logger.Err("PaymentService.OrderPayment in orderPaymentState failed, order: %v, error", order, futureData.Ex.Error())
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:futureData.Ex}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	paymentResponse := futureData.Data.(entities.PaymentResponse)

}

func (orderPayment orderPaymentActionLauncher) persistOrderState(ctx context.Context, order *entities.Order, itemsId []string,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					orderPayment.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
				} else {
					logger.Err("orderPayment received itemId %s not exist in order, order: %v", id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			orderPayment.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
		}
	}

	if _, err := global.Singletons.OrderRepository.Save(*order); err != nil {
		logger.Err("Save orderPayment State Failed, error: %s, order: %v", err, order)
	}
}

func (orderPayment orderPaymentActionLauncher) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	order.Items[index].OrderStep.CreatedAt = ctx.Value(global.CtxStepTimestamp).(time.Time)
	order.Items[index].OrderStep.CurrentName = ctx.Value(global.CtxStepName).(string)
	order.Items[index].OrderStep.CurrentIndex = ctx.Value(global.CtxStepIndex).(int)

	order.Items[index].OrderStep.CurrentState.Name = orderPayment.Name()
	order.Items[index].OrderStep.CurrentState.Index = orderPayment.Index()
	order.Items[index].OrderStep.CurrentState.Type = orderPayment.Actions().ActionType().Name()
	order.Items[index].OrderStep.CurrentState.CreatedAt = time.Now().UTC()
	order.Items[index].OrderStep.CurrentState.Result = result
	order.Items[index].OrderStep.CurrentState.Reason = reason

	if acceptedAction != nil {
		order.Items[index].OrderStep.CurrentState.AcceptedAction.Name = acceptedAction.Name()
	} else {
		order.Items[index].OrderStep.CurrentState.AcceptedAction.Name = ""
	}

	order.Items[index].OrderStep.CurrentState.AcceptedAction.Type = actives.OrderPaymentAction.String()
	order.Items[index].OrderStep.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	order.Items[index].OrderStep.CurrentState.AcceptedAction.Data = ""
	order.Items[index].OrderStep.CurrentState.AcceptedAction.Time = &order.Items[index].OrderStep.CurrentState.CreatedAt

	order.Items[index].OrderStep.CurrentState.Actions = []entities.Action{order.Items[index].OrderStep.CurrentState.AcceptedAction}

	order.Items[index].OrderStep.StepsHistory = []entities.StepHistory{{
		Name: order.Items[index].OrderStep.CurrentState.Name,
		Index: order.Items[index].OrderStep.CurrentState.Index,
		CreatedAt: order.Items[index].OrderStep.CurrentState.CreatedAt,
		StatesHistory: make([]entities.StateHistory, 0, 5),
	}}

	stateHistory := entities.StateHistory {
		Name: order.Items[index].OrderStep.CurrentState.Name,
		Index: order.Items[index].OrderStep.CurrentState.Index,
		Type: order.Items[index].OrderStep.CurrentState.Type,
		Action: order.Items[index].OrderStep.CurrentState.AcceptedAction,
		Result: order.Items[index].OrderStep.CurrentState.Result,
		Reason: order.Items[index].OrderStep.CurrentState.Reason,
		CreatedAt:order.Items[index].OrderStep.CurrentState.CreatedAt,
	}

	order.Items[index].OrderStep.StepsHistory[len(order.Items[index].OrderStep.StepsHistory)].StatesHistory =
		append(order.Items[index].OrderStep.StepsHistory[len(order.Items[index].OrderStep.StepsHistory)].StatesHistory, stateHistory)
}

