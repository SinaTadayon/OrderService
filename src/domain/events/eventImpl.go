package events

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"time"
)

type eventImpl struct {
	actor     actors.ActorType
	action    actors.IActorAction
	data      actions.IActionData
	timestamp time.Time
}

func New(actor actors.ActorType, action actors.IActorAction,
		data actions.IActionData, timestamp time.Time) IEvent {
	return eventImpl{actor, action, data, timestamp}
}

func (event eventImpl) ActorType() actors.ActorType {
	return event.actor
}

func (event eventImpl) ActorAction() actors.IActorAction {
	return event.action
}

func (event eventImpl) Data() actions.IActionData {
	return event.data
}

func (event eventImpl) Timestamp() time.Time {
	return event.timestamp
}
