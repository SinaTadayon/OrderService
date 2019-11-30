package operator_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Operator
)

type operatorActorActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return operatorActorActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (operatorAction operatorActorActionImpl) ActionType() actions.ActionType {
	return operatorAction.actionType
}

func (operatorAction operatorActorActionImpl) ActionEnum() actions.IEnumAction {
	return operatorAction.enumAction
}
