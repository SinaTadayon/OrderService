package seller_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
)

const (
	actorType  = actors.SellerActor
	actionType = actions.ActorAction
)

type sellerActorActionImpl struct {
	actionType actions.ActionType
	actorType  actors.ActorType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return sellerActorActionImpl{actionType, actorType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return sellerActorActionImpl{actionType, actorType, iEnumAction}
}

func (sellerAction sellerActorActionImpl) ActorType() actors.ActorType {
	return sellerAction.actorType
}

func (sellerAction sellerActorActionImpl) ActionType() actions.ActionType {
	return sellerAction.actionType
}

func (sellerAction sellerActorActionImpl) ActionEnums() []actions.IEnumAction {
	return sellerAction.enumAction
}
