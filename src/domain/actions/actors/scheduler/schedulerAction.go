package scheduler_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
)

const (
	actorType = actors.SchedulerActor
	actionType = actions.ActorAction
)

type schedulerActorActionImpl struct {
	actionType actions.ActionType
	actorType  actors.ActorType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return schedulerActorActionImpl{actionType, actorType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return schedulerActorActionImpl{actionType, actorType, iEnumAction}
}

func (schedulerAction schedulerActorActionImpl) ActorType() actors.ActorType {
	return schedulerAction.actorType
}

func (schedulerAction schedulerActorActionImpl) ActionType() actions.ActionType {
	return schedulerAction.actionType
}

func (schedulerAction schedulerActorActionImpl) ActionEnums() []actions.IEnumAction {
	return schedulerAction.enumAction
}