package events

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"time"
)

type eventImpl struct {
	actorType   actors.ActorType
	actorAction actors.IActorAction
	actionData  actions.IActionData
	timestamp   time.Time
}

func New(actorType actors.ActorType, actorAction actors.IActorAction,
		actionData actions.IActionData, timestamp time.Time) IEvent {
	return eventImpl{actorType, actorAction, actionData, timestamp}
}

func (event eventImpl) ActorType() actors.ActorType {
	return event.actorType
}

func (event eventImpl) ActorAction() actors.IActorAction {
	return event.actorAction
}

func (event eventImpl) Data() actions.IActionData {
	return event.actionData
}

func (event eventImpl) Timestamp() time.Time {
	return event.timestamp
}
