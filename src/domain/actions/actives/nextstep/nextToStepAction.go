package next_to_step_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.NextToStepAction
	actionType = actions.ActiveAction
)

type nextToStepActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []actions.IEnumAction) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return nextToStepActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return nextToStepActiveActionImpl{actionType, activeType, iEnumAction}
}

func (nextToStep nextToStepActiveActionImpl) ActiveType() actives.ActiveType {
	return nextToStep.activeType
}

func (nextToStep nextToStepActiveActionImpl) ActionType() actions.ActionType {
	return nextToStep.actionType
}

func (nextToStep nextToStepActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return nextToStep.enumAction
}



