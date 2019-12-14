package events

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"time"
)

type IEvent interface {
	EventType() EventType
	OrderId() uint64
	PackageId() uint64
	UserId() uint64
	StateIndex() int32
	Action() actions.IAction
	Data() interface{}
	Timestamp() time.Time
}
