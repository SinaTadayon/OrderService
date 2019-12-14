package events

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"time"
)

type BaseEventImpl struct {
	eventType  EventType
	orderId    uint64
	packageId  uint64
	userId     uint64
	stateIndex int32
	action     actions.IAction
	timestamp  time.Time
	data       interface{}
}

func New(eventType EventType, orderId, packageId, userId uint64, stateIndex int32, action actions.IAction, timestamp time.Time, data interface{}) IEvent {
	return &BaseEventImpl{eventType, orderId, packageId, userId, stateIndex,
		action, timestamp, data}
}

func (baseEvent BaseEventImpl) Timestamp() time.Time {
	return baseEvent.timestamp
}

func (baseEvent BaseEventImpl) EventType() EventType {
	return baseEvent.eventType
}

func (baseEvent BaseEventImpl) OrderId() uint64 {
	return baseEvent.orderId
}

func (baseEvent BaseEventImpl) UserId() uint64 {
	return baseEvent.userId
}

func (baseEvent BaseEventImpl) Data() interface{} {
	return baseEvent.data
}

func (baseEvent BaseEventImpl) Action() actions.IAction {
	return baseEvent.action
}

func (baseEvent BaseEventImpl) PackageId() uint64 {
	return baseEvent.packageId
}

func (baseEvent BaseEventImpl) StateIndex() int32 {
	return baseEvent.stateIndex
}
