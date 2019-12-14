package buyer_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Buyer
)

type buyerActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return buyerActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (buyerAction buyerActionImpl) ActionType() actions.ActionType {
	return buyerAction.actionType
}

func (buyerAction buyerActionImpl) ActionEnum() actions.IEnumAction {
	return buyerAction.enumAction
}
