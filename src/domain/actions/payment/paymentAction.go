package payment_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Payment
)

type paymentActorActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return paymentActorActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (paymentAction paymentActorActionImpl) ActionType() actions.ActionType {
	return paymentAction.actionType
}

func (paymentAction paymentActorActionImpl) ActionEnum() actions.IEnumAction {
	return paymentAction.enumAction
}
