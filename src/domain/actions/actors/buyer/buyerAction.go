package buyer_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
)

const (
	actorType  = actors.BuyerActor
	actionType = actions.ActorAction
)

type buyerActorActionImpl struct {
	actionType actions.ActionType
	actorType  actors.ActorType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return buyerActorActionImpl{actionType, actorType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return buyerActorActionImpl{actionType, actorType, iEnumAction}
}

func (buyerAction buyerActorActionImpl) ActorType() actors.ActorType {
	return buyerAction.actorType
}

func (buyerAction buyerActorActionImpl) ActionType() actions.ActionType {
	return buyerAction.actionType
}

func (buyerAction buyerActorActionImpl) ActionEnums() []actions.IEnumAction {
	return buyerAction.enumAction
}
