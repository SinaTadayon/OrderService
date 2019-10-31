package events

import (
	"time"
)

type IEvent interface {
	EventType()		EventType
	Data() 			interface{}
	Timestamp()		time.Time
}
