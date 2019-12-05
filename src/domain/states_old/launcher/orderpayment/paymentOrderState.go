package order_payment_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	order_payment_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/orderpayment"
	active_event "gitlab.faza.io/order-project/order-service/domain/events/active"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"time"
)

const (
	stateName  string = "Payment_Order_Action_State"
	activeType        = actives.OrderPaymentAction
)

type orderPaymentActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &orderPaymentActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &orderPaymentActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &orderPaymentActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

func (orderPayment orderPaymentActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {

	paymentState, ok := orderPayment.Childes()[0].(listener_state.IListenerState)
	if ok != true {
		logger.Err("paymentState isn't child of orderPaymentState, order: %v", order)
		returnChannel := make(chan future.IDataFuture, 1)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewFuture(returnChannel, 1, 1)
	}

	nextToStepState, ok := orderPayment.Childes()[1].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("nextToStep isn't child of orderPaymentState, order: %v", order)
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}

	paymentRequest := payment_service.PaymentRequest{
		Amount:   int64(order.Invoice.Total),
		Gateway:  order.Invoice.PaymentOption,
		Currency: order.Invoice.Currency,
		OrderId:  order.OrderId,
	}

	order.PaymentService = []entities.PaymentService{
		{
			PaymentRequest: &entities.PaymentRequest{
				Amount:    uint64(paymentRequest.Amount),
				Currency:  paymentRequest.Currency,
				Gateway:   paymentRequest.Gateway,
				CreatedAt: time.Time{},
			},
		},
	}

	iPromise := global.Singletons.PaymentService.OrderPayment(ctx, paymentRequest)
	futureData := iPromise.Get()
	if futureData == nil {
		order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
			Result:      false,
			Reason:      "PaymentService.OrderPayment in orderPaymentState failed",
			Description: "",
			CallBackUrl: "",
			InvoiceId:   0,
			PaymentId:   "",
			CreatedAt:   time.Now(),
		}

		orderPayment.persistOrderState(ctx, &order, itemsId, order_payment_action.OrderPaymentAction, false,
			"PaymentService.OrderPayment in orderPaymentState failed", nil)
		logger.Err("PaymentService.OrderPayment in orderPaymentState failed, order: %v", order)
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}

	if futureData.Ex != nil {
		order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
			Result:      false,
			Reason:      futureData.Ex.Error(),
			Description: "",
			CallBackUrl: "",
			InvoiceId:   0,
			PaymentId:   "",
			CreatedAt:   time.Now(),
		}

		orderPayment.persistOrderState(ctx, &order, itemsId, order_payment_action.OrderPaymentAction, false,
			futureData.Ex.Error(), nil)
		logger.Err("PaymentService.OrderPayment in orderPaymentState failed, order: %v, error: %s", order, futureData.Ex.Error())
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: futureData.Ex}
		go func() {
			nextToStepState.ActionLauncher(ctx, order, nil, order_payment_action.OrderPaymentFailedAction)
		}()
		return future.NewFuture(returnChannel, 1, 1)
	}

	paymentResponse := futureData.Data.(payment_service.PaymentResponse)
	activeEvent := active_event.NewActiveEvent(order, itemsId, actives.OrderPaymentAction, order_payment_action.NewOf(order_payment_action.OrderPaymentAction),
		paymentResponse, time.Now())

	order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
		Result:      true,
		Reason:      "",
		Description: "",
		CallBackUrl: paymentResponse.CallbackUrl,
		InvoiceId:   paymentResponse.InvoiceId,
		PaymentId:   paymentResponse.PaymentId,
		CreatedAt:   time.Now(),
	}

	orderPayment.persistOrderState(ctx, &order, itemsId, order_payment_action.OrderPaymentAction,
		true, "", &paymentResponse)
	return paymentState.ActionListener(ctx, activeEvent, nil)
}

func (orderPayment orderPaymentActionLauncher) persistOrderState(ctx context.Context, order *entities.Order, itemsId []uint64,
	acceptedAction actions.IEnumAction, result bool, reason string, paymentResponse *payment_service.PaymentResponse) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					orderPayment.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason, paymentResponse)
				} else {
					logger.Err("orderPayment received itemId %d not exist in order, order: %v", id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			orderPayment.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason, paymentResponse)
		}
	}

	if _, err := global.Singletons.OrderRepository.Save(*order); err != nil {
		logger.Err("Save orderPayment State Failed, error: %s, order: %v", err, order)
	}
}

func (orderPayment orderPaymentActionLauncher) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string, paymentResponse *payment_service.PaymentResponse) {

	//order.Items[index].Tracking.CurrentState.ActionName = orderPayment.ActionName()
	//order.Items[index].Tracking.CurrentState.Index = orderPayment.Index()
	//order.Items[index].Tracking.CurrentState.Type = orderPayment.Actions().ActionType().ActionName()
	//order.Items[index].Tracking.CurrentState.CreatedAt = time.Now().UTC()
	//order.Items[index].Tracking.CurrentState.Result = result
	//order.Items[index].Tracking.CurrentState.Reason = reason
	//
	//if acceptedAction != nil {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.ActionName = acceptedAction.ActionName()
	//} else {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.ActionName = ""
	//}
	//
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Type = actives.OrderPaymentAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Get = nil
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Time = &order.Items[index].Tracking.CurrentState.CreatedAt
	//
	//order.Items[index].Tracking.CurrentState.Actions = []entities.Action{order.Items[index].Tracking.CurrentState.AcceptedAction}
	//
	//stateHistory := entities.StateHistory {
	//	ActionName: order.Items[index].Tracking.CurrentState.ActionName,
	//	Index: order.Items[index].Tracking.CurrentState.Index,
	//	Type: order.Items[index].Tracking.CurrentState.Type,
	//	Action: order.Items[index].Tracking.CurrentState.AcceptedAction,
	//	Result: order.Items[index].Tracking.CurrentState.Result,
	//	Reason: order.Items[index].Tracking.CurrentState.Reason,
	//	CreatedAt:order.Items[index].Tracking.CurrentState.CreatedAt,
	//}
	//
	//order.Items[index].Tracking.StatesHistory[len(order.Items[index].Tracking.StatesHistory)].StatesHistory =
	//	append(order.Items[index].Tracking.StatesHistory[len(order.Items[index].Tracking.StatesHistory)].StatesHistory, stateHistory)
}
