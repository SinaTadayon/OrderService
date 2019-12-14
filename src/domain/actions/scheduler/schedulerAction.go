package scheduler_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Scheduler
)

type schedulerActorActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return schedulerActorActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (schedulerAction schedulerActorActionImpl) ActionType() actions.ActionType {
	return schedulerAction.actionType
}

func (schedulerAction schedulerActorActionImpl) ActionEnum() actions.IEnumAction {
	return schedulerAction.enumAction
}
