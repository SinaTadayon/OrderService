package pay_to_seller_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.PayToSellerAction
	actionType = actions.ActiveAction
)

type payToSellerActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return payToSellerActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return payToSellerActiveActionImpl{actionType, activeType, iEnumAction}
}

func (payToSeller payToSellerActiveActionImpl) ActiveType() actives.ActiveType {
	return payToSeller.activeType
}

func (payToSeller payToSellerActiveActionImpl) ActionType() actions.ActionType {
	return payToSeller.actionType
}

func (payToSeller payToSellerActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return payToSeller.enumAction
}
