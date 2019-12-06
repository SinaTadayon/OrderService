package voucher_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Payment
)

type voucherActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return voucherActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (voucher voucherActionImpl) ActionType() actions.ActionType {
	return voucher.actionType
}

func (voucher voucherActionImpl) ActionEnum() actions.IEnumAction {
	return voucher.enumAction
}
