package scheduler_action_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

const (
	actorType        = actions.Scheduler
	stateName string = "Scheduler_Action_State"
)

type schedulerActionListener struct {
	*listener_state.BaseListenerImpl
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction) listener_state.IListenerState {
	return &schedulerActionListener{listener_state.NewBaseListener(stateName, index, childes, parents,
		actions, actorType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction) listener_state.IListenerState {
	return &schedulerActionListener{listener_state.NewBaseListener(name, index, childes, parents,
		actions, actorType)}
}

func NewFrom(base *listener_state.BaseListenerImpl) listener_state.IListenerState {
	return &schedulerActionListener{base}
}

func NewValueOf(base *listener_state.BaseListenerImpl, params ...interface{}) listener_state.IListenerState {
	panic("implementation required")
}

func (paymentAction schedulerActionListener) ActionListener(ctx context.Context, event events.IEvent, param interface{}) future.IFuture {
	panic("implementation required")
}
