package pay_to_buyer_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.PayToBuyerAction
	actionType = actions.ActiveAction
)

type payToBuyerActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return payToBuyerActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return payToBuyerActiveActionImpl{actionType, activeType, iEnumAction}
}

func (payToBuyer payToBuyerActiveActionImpl) ActiveType() actives.ActiveType {
	return payToBuyer.activeType
}

func (payToBuyer payToBuyerActiveActionImpl) ActionType() actions.ActionType {
	return payToBuyer.actionType
}

func (payToBuyer payToBuyerActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return payToBuyer.enumAction
}
