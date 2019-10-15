package payment_action_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/states"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states/listener"
	"gitlab.faza.io/order-project/order-service/domain/steps"
)

const (
	actorType = actors.PaymentActor
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

func NewValueOf(base *listener_state.BaseListenerImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentAction paymentActionListener) ActionListener(ctx context.Context, event events.IEvent, param interface{}) {
	panic("implementation required")
}
