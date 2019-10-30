package events

import (
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"time"
)

type eventImpl struct {
	actor    	actors.ActorType
	action    	actors.IActorAction
	orderId		string
	itemsId		[]string
	data      	interface{}
	timestamp	time.Time
}

func New(actor actors.ActorType, action actors.IActorAction, orderId string, itemsId[] string,
		data interface{}, timestamp time.Time) IEvent {
	return eventImpl{actor, action, orderId, itemsId, data, timestamp}
}

func (event eventImpl) ActorType() actors.ActorType {
	return event.actor
}

func (event eventImpl) ActorAction() actors.IActorAction {
	return event.action
}

func (event eventImpl) Data() interface{} {
	return event.data
}

func (event eventImpl) OrderId() string {
	return event.orderId
}

func (event eventImpl) ItemsId() []string {
	return event.itemsId
}

func (event eventImpl) Timestamp() time.Time {
	return event.timestamp
}
