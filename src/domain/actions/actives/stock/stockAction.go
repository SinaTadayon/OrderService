package stock_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.StockAction
	actionType = actions.ActiveAction
)

type stockActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return stockActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return stockActiveActionImpl{actionType, activeType, iEnumAction}
}

func (stock stockActiveActionImpl) ActiveType() actives.ActiveType {
	return stock.activeType
}

func (stock stockActiveActionImpl) ActionType() actions.ActionType {
	return stock.actionType
}

func (stock stockActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return stock.enumAction
}
