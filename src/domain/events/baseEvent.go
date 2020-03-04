package events

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"time"
)

type BaseEventImpl struct {
	EType      EventType
	OId        uint64
	PId        uint64
	UId        uint64
	SIdx       int32
	EAction    actions.IAction
	ETimestamp time.Time
	EData      interface{}
}

func New(eventType EventType, OId, PId, UId uint64, SIdx int32, EAction actions.IAction, timestamp time.Time, EData interface{}) IEvent {
	return &BaseEventImpl{eventType, OId, PId, UId, SIdx,
		EAction, timestamp, EData}
}

func (baseEvent BaseEventImpl) Timestamp() time.Time {
	return baseEvent.ETimestamp
}

func (baseEvent BaseEventImpl) EventType() EventType {
	return baseEvent.EType
}

func (baseEvent BaseEventImpl) OrderId() uint64 {
	return baseEvent.OId
}

func (baseEvent BaseEventImpl) UserId() uint64 {
	return baseEvent.UId
}

func (baseEvent BaseEventImpl) Data() interface{} {
	return baseEvent.EData
}

func (baseEvent BaseEventImpl) Action() actions.IAction {
	return baseEvent.EAction
}

func (baseEvent BaseEventImpl) PackageId() uint64 {
	return baseEvent.PId
}

func (baseEvent BaseEventImpl) StateIndex() int32 {
	return baseEvent.SIdx
}
