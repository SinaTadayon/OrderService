package stock_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Stock
)

type stockActiveActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return stockActiveActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (stock stockActiveActionImpl) ActionType() actions.ActionType {
	return stock.actionType
}

func (stock stockActiveActionImpl) ActionEnum() actions.IEnumAction {
	return stock.enumAction
}
