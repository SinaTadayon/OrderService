package checkout_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
)

const (
	actorType  = actors.CheckoutActor
	actionType = actions.ActorAction
)

type checkoutActorActionImpl struct {
	actionType actions.ActionType
	actorType  actors.ActorType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return checkoutActorActionImpl{actionType, actorType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return checkoutActorActionImpl{actionType, actorType, iEnumAction}
}

func (checkoutAction checkoutActorActionImpl) ActorType() actors.ActorType {
	return checkoutAction.actorType
}

func (checkoutAction checkoutActorActionImpl) ActionType() actions.ActionType {
	return checkoutAction.actionType
}

func (checkoutAction checkoutActorActionImpl) ActionEnums() []actions.IEnumAction {
	return checkoutAction.enumAction
}
