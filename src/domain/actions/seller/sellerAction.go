package seller_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Seller
)

type sellerActorActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return sellerActorActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (sellerAction sellerActorActionImpl) ActionType() actions.ActionType {
	return sellerAction.actionType
}

func (sellerAction sellerActorActionImpl) ActionEnum() actions.IEnumAction {
	return sellerAction.enumAction
}
