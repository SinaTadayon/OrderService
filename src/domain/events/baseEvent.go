package events

import (
	"time"
)

type BaseEventImpl struct {
	eventType EventType
	data      interface{}
	timestamp time.Time
}

func NewBaseEventImpl(eventType EventType, data interface{}, timestamp time.Time) *BaseEventImpl {
	return &BaseEventImpl{eventType, data, timestamp}
}

func (baseEvent BaseEventImpl) SetTimestamp(time time.Time) {
	baseEvent.timestamp = time
}

func (baseEvent BaseEventImpl) SetEventType(event EventType) {
	baseEvent.eventType = event
}

func (baseEvent BaseEventImpl) SetData(data interface{}) {
	baseEvent.data = data
}

func (baseEvent BaseEventImpl) Timestamp() time.Time {
	return baseEvent.timestamp
}

func (baseEvent BaseEventImpl) EventType() EventType {
	return baseEvent.eventType
}

func (baseEvent BaseEventImpl) Data() interface{} {
	return baseEvent.data
}
