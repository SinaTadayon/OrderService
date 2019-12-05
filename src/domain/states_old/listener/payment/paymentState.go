package payment_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	payment_action "gitlab.faza.io/order-project/order-service/domain/actions/payment"
	"gitlab.faza.io/order-project/order-service/domain/events"
	active_event "gitlab.faza.io/order-project/order-service/domain/events/active"
	actor_event "gitlab.faza.io/order-project/order-service/domain/events/actor"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	launcher_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"time"
)

const (
	actorType        = actions.Payment
	stateName string = "Payment_Action_State"
)

type paymentActionListener struct {
	*listener_state.BaseListenerImpl
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction) listener_state.IListenerState {
	return &paymentActionListener{listener_state.NewBaseListener(stateName, index, childes, parents,
		actions, actorType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction) listener_state.IListenerState {
	return &paymentActionListener{listener_state.NewBaseListener(name, index, childes, parents,
		actions, actorType)}
}

func NewFrom(base *listener_state.BaseListenerImpl) listener_state.IListenerState {
	return &paymentActionListener{base}
}

func NewValueOf(base *listener_state.BaseListenerImpl, params ...interface{}) listener_state.IListenerState {
	panic("implementation required")
}

// TODO must be complete implement
func (paymentAction paymentActionListener) ActionListener(ctx context.Context, event events.IEvent, param interface{}) future.IFuture {

	if event == nil {
		logger.Err("Received Event is nil")
		returnChannel := make(chan future.IDataFuture, 1)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewFuture(returnChannel, 1, 1)
	}

	nextToStepState, ok := paymentAction.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("nextToStepState isn't child of paymentAction, event: %v", event)
		returnChannel := make(chan future.IDataFuture, 1)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewFuture(returnChannel, 1, 1)
	}

	if event.EventType() == events.ActiveEvent {
		activeEvent := event.(active_event.IActiveEvent)
		order := activeEvent.Order()
		paymentAction.persistOrderState(ctx, &order, activeEvent.ItemsId(), activeEvent.ActiveAction().ActionEnums()[0], true, "", nil)
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: activeEvent.Data(), Ex: nil}
		return future.NewFuture(returnChannel, 1, 1)
	} else {
		actorEvent := event.(actor_event.IActorEvent)
		order := actorEvent.Order()
		paymentResult := actorEvent.Data().(payment_service.PaymentResult)
		order.PaymentService[0].PaymentResult = &entities.PaymentResult{
			Result:      paymentResult.Result,
			Reason:      "",
			PaymentId:   paymentResult.PaymentId,
			InvoiceId:   paymentResult.InvoiceId,
			Amount:      uint64(paymentResult.Amount),
			ReqBody:     paymentResult.ReqBody,
			ResBody:     paymentResult.ResBody,
			CardNumMask: paymentResult.CardMask,
			CreatedAt:   time.Now(),
		}

		if paymentResult.Result == true {
			paymentAction.persistOrderState(ctx, &order, actorEvent.ItemsId(), payment_action.Success, true, "", &paymentResult)
			go func() {
				nextToStepState.ActionLauncher(ctx, order, actorEvent.ItemsId(), payment_action.Success)
			}()
			returnChannel := make(chan future.IDataFuture, 1)
			defer close(returnChannel)
			returnChannel <- future.IDataFuture{Data: actorEvent.Data(), Ex: nil}
			return future.NewFuture(returnChannel, 1, 1)
		} else {
			paymentAction.persistOrderState(ctx, &order, actorEvent.ItemsId(), payment_action.Fail, false, "", &paymentResult)
			return nextToStepState.ActionLauncher(ctx, order, actorEvent.ItemsId(), payment_action.Fail)
		}
	}
}

func (paymentAction paymentActionListener) persistOrderState(ctx context.Context, order *entities.Order, itemsId []uint64,
	acceptedAction actions.IEnumAction, result bool, reason string, paymentResult *payment_service.PaymentResult) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					paymentAction.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason, paymentResult)
				} else {
					logger.Err("orderPayment received itemId %d not exist in order, orderId: %d", id, order.OrderId)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			paymentAction.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason, paymentResult)
		}
	}

	if _, err := global.Singletons.OrderRepository.Save(*order); err != nil {
		logger.Err("Save orderPayment Status Failed, error: %s, order: %v", err, order)
	}
}

func (paymentAction paymentActionListener) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string, paymentResult *payment_service.PaymentResult) {
	//order.Items[index].Tracking.CurrentState.ActionName = paymentAction.ActionName()
	//order.Items[index].Tracking.CurrentState.Index = paymentAction.Index()
	//order.Items[index].Tracking.CurrentState.Type = paymentAction.Actions().ActionType().ActionName()
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
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Type = actors.Payment.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Base = actions.ActorAction.String()
	//// TODO implement stringfy paymentResult
	//if paymentResult != nil {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Get = nil
	//} else {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Get = nil
	//}
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Time = &order.Items[index].Tracking.CurrentState.CreatedAt
	//
	//order.Items[index].Tracking.CurrentState.Actions = []entities.Actions{order.Items[index].Tracking.CurrentState.AcceptedAction}
	//
	//stateHistory := entities.StateHistory {
	//	ActionName: order.Items[index].Tracking.CurrentState.ActionName,
	//	Index: order.Items[index].Tracking.CurrentState.Index,
	//	Type: order.Items[index].Tracking.CurrentState.Type,
	//	Actions: order.Items[index].Tracking.CurrentState.AcceptedAction,
	//	Result: order.Items[index].Tracking.CurrentState.Result,
	//	Reason: order.Items[index].Tracking.CurrentState.Reason,
	//	CreatedAt:order.Items[index].Tracking.CurrentState.CreatedAt,
	//}
	//
	//order.Items[index].Tracking.States[len(order.Items[index].Tracking.States)].States =
	//	append(order.Items[index].Tracking.States[len(order.Items[index].Tracking.States)].States, stateHistory)
}
