package actor_event

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"time"
)

type iActorEventImpl struct {
	*events.BaseEventImpl
	actor   actions.ActionType
	action  actors.IActorAction
	order   entities.Order
	itemsId []uint64
}

func NewActorEvent(actor actions.ActionType, action actors.IActorAction, order entities.Order, itemsId []uint64, data interface{}, timestamp time.Time) IActorEvent {
	return &iActorEventImpl{events.NewBaseEventImpl(events.ActorEvent, data, timestamp), actor, action, order, itemsId}
}

func (actorEvent iActorEventImpl) ActorType() actions.ActionType {
	return actorEvent.actor
}

func (actorEvent iActorEventImpl) ActorAction() actors.IActorAction {
	return actorEvent.action
}

func (actorEvent iActorEventImpl) Order() entities.Order {
	return actorEvent.order
}

func (actorEvent iActorEventImpl) ItemsId() []uint64 {
	return actorEvent.itemsId
}
