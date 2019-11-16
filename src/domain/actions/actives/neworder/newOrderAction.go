package new_order_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.NewOrderAction
	actionType = actions.ActiveAction
)

type manualPaymentActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return manualPaymentActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return manualPaymentActiveActionImpl{actionType, activeType, iEnumAction}
}

func (manualPayment manualPaymentActiveActionImpl) ActiveType() actives.ActiveType {
	return manualPayment.activeType
}

func (manualPayment manualPaymentActiveActionImpl) ActionType() actions.ActionType {
	return manualPayment.actionType
}

func (manualPayment manualPaymentActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return manualPayment.enumAction
}
