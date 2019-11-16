package pay_to_market_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.PayToMarketAction
	actionType = actions.ActiveAction
)

type payToMarketActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return payToMarketActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return payToMarketActiveActionImpl{actionType, activeType, iEnumAction}
}

func (payToMarket payToMarketActiveActionImpl) ActiveType() actives.ActiveType {
	return payToMarket.activeType
}

func (payToMarket payToMarketActiveActionImpl) ActionType() actions.ActionType {
	return payToMarket.actionType
}

func (payToMarket payToMarketActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return payToMarket.enumAction
}
