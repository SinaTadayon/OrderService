package seller_action_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

const (
	actorType = actions.Seller
)

type sellerActionListener struct {
	*listener_state.BaseListenerImpl
}

func New(name string, index int, childes, parents []states_old.IState, actions actions.IAction) listener_state.IListenerState {
	return &sellerActionListener{listener_state.NewBaseListener(name, index, childes, parents,
		actions, actorType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction) listener_state.IListenerState {
	return &sellerActionListener{listener_state.NewBaseListener(name, index, childes, parents,
		actions, actorType)}
}

func NewFrom(base *listener_state.BaseListenerImpl) listener_state.IListenerState {
	return &sellerActionListener{base}
}

func NewValueOf(base *listener_state.BaseListenerImpl, params ...interface{}) listener_state.IListenerState {
	panic("implementation required")
}

func (sellerAction sellerActionListener) ActionListener(ctx context.Context, event events.IEvent, param interface{}) promise.IPromise {
	panic("implementation required")
}
