package actor_event

import (
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

type IActorEvent interface {
	events.IEvent
	Order() 		entities.Order
	ItemsId()		[]string
	ActorType() 	actors.ActorType
	ActorAction() 	actors.IActorAction
}