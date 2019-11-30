package events

import (
	"time"
)

type IEvent interface {
	EventType() EventType
	Timestamp() time.Time
}
