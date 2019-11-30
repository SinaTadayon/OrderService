package events

import (
	"time"
)

type BaseEventImpl struct {
	eventType EventType
	timestamp time.Time
}

func NewBaseEventImpl(eventType EventType, timestamp time.Time) *BaseEventImpl {
	return &BaseEventImpl{eventType, timestamp}
}

func (baseEvent BaseEventImpl) SetTimestamp(time time.Time) {
	baseEvent.timestamp = time
}

func (baseEvent BaseEventImpl) SetEventType(event EventType) {
	baseEvent.eventType = event
}

func (baseEvent BaseEventImpl) Timestamp() time.Time {
	return baseEvent.timestamp
}

func (baseEvent BaseEventImpl) EventType() EventType {
	return baseEvent.eventType
}
