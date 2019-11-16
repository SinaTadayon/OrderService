package retry_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.RetryAction
	actionType = actions.ActiveAction
)

type retryActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return retryActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return retryActiveActionImpl{actionType, activeType, iEnumAction}
}

func (retryAction retryActiveActionImpl) ActiveType() actives.ActiveType {
	return retryAction.activeType
}

func (retryAction retryActiveActionImpl) ActionType() actions.ActionType {
	return retryAction.actionType
}

func (retryAction retryActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return retryAction.enumAction
}
