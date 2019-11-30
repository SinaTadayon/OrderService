package listener_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type IListenerState interface {
	states_old.IState
	ActorType() actors.ActorType
	ActionListener(ctx context.Context, event events.IEvent, param interface{}) promise.IPromise
}
