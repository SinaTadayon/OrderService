package actors

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type IActorAction interface {
	actions.IAction
	ActionEnums() 		[]actions.IEnumAction
	ActorType() 		ActorType
}