package checkout_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	checkout_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/checkout"
	"gitlab.faza.io/order-project/order-service/domain/events"
	actor_event "gitlab.faza.io/order-project/order-service/domain/events/actor"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	launcher_state "gitlab.faza.io/order-project/order-service/domain/states/launcher"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	"time"
	//message "gitlab.faza.io/protos/order"
)

const (
	actorType        = actors.CheckoutActor
	stateName string = "Checkout_Action_State"
)

type checkoutActionListener struct {
	*listener_state.BaseListenerImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) listener_state.IListenerState {
	return &checkoutActionListener{listener_state.NewBaseListener(stateName, index, childes, parents,
		actions, actorType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction) listener_state.IListenerState {
	return &checkoutActionListener{listener_state.NewBaseListener(name, index, childes, parents,
		actions, actorType)}
}

func NewFrom(base *listener_state.BaseListenerImpl) listener_state.IListenerState {
	return &checkoutActionListener{base}
}

func NewValueOf(base *listener_state.BaseListenerImpl, params ...interface{}) listener_state.IListenerState {
	panic("implementation required")
}

// TODO context handling
func (checkoutActionState checkoutActionListener) ActionListener(ctx context.Context, event events.IEvent, param interface{}) promise.IPromise {

	if event == nil {
		logger.Err("Received Event is nil")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	stockState, ok := checkoutActionState.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("StockState isn't child of CheckoutState, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	// TODO checking type and result cast
	actorEvent := event.(actor_event.IActorEvent)

	if actorEvent.ActorType() != actors.CheckoutActor {
		logger.Err("Received actorType of event is not CheckoutActor, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if actorEvent.ActorAction().ActionEnums()[0] != checkout_action.NewOrderAction {
		logger.Err("Received actorAction of event is not NewOrderAction, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrder := actorEvent.Order()
	checkoutActionState.updateOrderStates(ctx, &newOrder, event.Timestamp())

	order, err := global.Singletons.OrderRepository.Save(newOrder)
	if err != nil {
		logger.Err("Save NewOrder Failed, error: %s, newOrder: %v", err, newOrder)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	return stockState.ActionLauncher(ctx, *order, nil, nil)
}

func (checkoutActionState checkoutActionListener) updateOrderStates(ctx context.Context, newOrder *entities.Order, timestamp time.Time) {
	//for i := 0; i < len(newOrder.Items); i++ {
	//	newOrder.Items[i].Progress.CurrentState.Name = checkoutActionState.Name()
	//	newOrder.Items[i].Progress.CurrentState.Index = checkoutActionState.Index()
	//	newOrder.Items[i].Progress.CurrentState.Type = checkoutActionState.Actions().ActionType().Name()
	//	newOrder.Items[i].Progress.CurrentState.CreatedAt = time.Now().UTC()
	//	newOrder.Items[i].Progress.CurrentState.Result = true
	//	newOrder.Items[i].Progress.CurrentState.Reason = ""
	//
	//	newOrder.Items[i].Progress.CurrentState.AcceptedAction.Name = checkout_action.NewOrderAction.String()
	//	newOrder.Items[i].Progress.CurrentState.AcceptedAction.Type = actors.CheckoutActor.String()
	//	newOrder.Items[i].Progress.CurrentState.AcceptedAction.Base = actions.ActorAction.String()
	//	newOrder.Items[i].Progress.CurrentState.AcceptedAction.Data = nil
	//	newOrder.Items[i].Progress.CurrentState.AcceptedAction.Time = &timestamp
	//
	//	newOrder.Items[i].Progress.CurrentState.Actions = []entities.Action{newOrder.Items[i].Progress.CurrentState.AcceptedAction}
	//
	//	stateHistory := entities.StateHistory {
	//		Name: newOrder.Items[i].Progress.CurrentState.Name,
	//		Index: newOrder.Items[i].Progress.CurrentState.Index,
	//		Type: newOrder.Items[i].Progress.CurrentState.Type,
	//		Action: newOrder.Items[i].Progress.CurrentState.AcceptedAction,
	//		Result: newOrder.Items[i].Progress.CurrentState.Result,
	//		Reason: newOrder.Items[i].Progress.CurrentState.Reason,
	//		CreatedAt:newOrder.Items[i].Progress.CurrentState.CreatedAt,
	//	}
	//
	//	newOrder.Items[i].Progress.StepsHistory[len(newOrder.Items[i].Progress.StepsHistory)].StatesHistory = append(newOrder.Items[i].Progress.StepsHistory[len(newOrder.Items[i].Progress.StepsHistory)].StatesHistory, stateHistory)
	//}
}
