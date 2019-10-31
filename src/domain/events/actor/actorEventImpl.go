package actor_event

import (
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"time"
)

type iActorEventImpl struct {
	*events.BaseEventImpl
	actor   actors.ActorType
	action  actors.IActorAction
	order   entities.Order
	itemsId []string
}

func NewActorEvent(actor actors.ActorType, action actors.IActorAction, order entities.Order, itemsId[] string, data interface{}, timestamp time.Time) IActorEvent {
	return &iActorEventImpl{events.NewBaseEventImpl(events.ActorEvent, data, timestamp), actor, action, order, itemsId}
}

func (actorEvent iActorEventImpl) ActorType() actors.ActorType {
	return actorEvent.actor
}

func (actorEvent iActorEventImpl) ActorAction() actors.IActorAction {
	return actorEvent.action
}

func (actorEvent iActorEventImpl) Order() entities.Order {
	return actorEvent.order
}

func (actorEvent iActorEventImpl) ItemsId() []string {
	return actorEvent.itemsId
}
