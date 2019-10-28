package events

import (
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"time"
)

type IEvent interface {
	ActorType() 	actors.ActorType
	ActorAction() 	actors.IActorAction
	Data() 			interface{}
	Timestamp()		time.Time
}
