package payment_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	payment_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/payment"
	"gitlab.faza.io/order-project/order-service/domain/events"
	active_event "gitlab.faza.io/order-project/order-service/domain/events/active"
	actor_event "gitlab.faza.io/order-project/order-service/domain/events/actor"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	launcher_state "gitlab.faza.io/order-project/order-service/domain/states/launcher"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"time"
)

const (
	actorType        = actors.PaymentActor
	stateName string = "Payment_Action_State"
)

type paymentActionListener struct {
	*listener_state.BaseListenerImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) listener_state.IListenerState {
	return &paymentActionListener{listener_state.NewBaseListener(stateName, index, childes, parents,
		actions, actorType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction) listener_state.IListenerState {
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
func (paymentAction paymentActionListener) ActionListener(ctx context.Context, event events.IEvent, param interface{}) promise.IPromise {

	if event == nil {
		logger.Err("Received Event is nil")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	nextToStepState, ok := paymentAction.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("nextToStepState isn't child of paymentAction, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if event.EventType() == events.ActiveEvent {
		activeEvent := event.(active_event.IActiveEvent)
		order := activeEvent.Order()
		paymentAction.persistOrderState(ctx, &order, activeEvent.ItemsId(), activeEvent.ActiveAction().ActionEnums()[0], true, "", nil)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: activeEvent.Data(), Ex: nil}
		return promise.NewPromise(returnChannel, 1, 1)
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
			paymentAction.persistOrderState(ctx, &order, actorEvent.ItemsId(), payment_action.SuccessAction, true, "", &paymentResult)
			go func() {
				nextToStepState.ActionLauncher(ctx, order, actorEvent.ItemsId(), payment_action.SuccessAction)
			}()
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: actorEvent.Data(), Ex: nil}
			return promise.NewPromise(returnChannel, 1, 1)
		} else {
			paymentAction.persistOrderState(ctx, &order, actorEvent.ItemsId(), payment_action.FailedAction, false, "", &paymentResult)
			return nextToStepState.ActionLauncher(ctx, order, actorEvent.ItemsId(), payment_action.FailedAction)
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
		logger.Err("Save orderPayment State Failed, error: %s, order: %v", err, order)
	}
}

func (paymentAction paymentActionListener) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string, paymentResult *payment_service.PaymentResult) {
	//order.Items[index].Tracking.CurrentState.Name = paymentAction.Name()
	//order.Items[index].Tracking.CurrentState.Index = paymentAction.Index()
	//order.Items[index].Tracking.CurrentState.Type = paymentAction.Actions().ActionType().Name()
	//order.Items[index].Tracking.CurrentState.CreatedAt = time.Now().UTC()
	//order.Items[index].Tracking.CurrentState.Result = result
	//order.Items[index].Tracking.CurrentState.Reason = reason
	//
	//if acceptedAction != nil {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Name = acceptedAction.Name()
	//} else {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Name = ""
	//}
	//
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Type = actors.PaymentActor.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Base = actions.ActorAction.String()
	//// TODO implement stringfy paymentResult
	//if paymentResult != nil {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Data = nil
	//} else {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Data = nil
	//}
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Time = &order.Items[index].Tracking.CurrentState.CreatedAt
	//
	//order.Items[index].Tracking.CurrentState.Actions = []entities.Action{order.Items[index].Tracking.CurrentState.AcceptedAction}
	//
	//stateHistory := entities.StateHistory {
	//	Name: order.Items[index].Tracking.CurrentState.Name,
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
