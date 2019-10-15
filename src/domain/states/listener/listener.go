package listener_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/states"
)

type IListenerState interface {
	states.IState
	ActorType() actors.ActorType
	ActionListener(ctx context.Context, event events.IEvent, param interface{})
}