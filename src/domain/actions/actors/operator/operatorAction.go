package operator_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
)

const (
	actorType = actors.OperatorActor
	actionType = actions.ActorAction
)

type operatorActorActionImpl struct {
	actionType actions.ActionType
	actorType  actors.ActorType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return operatorActorActionImpl{actionType, actorType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return operatorActorActionImpl{actionType, actorType, iEnumAction}
}

func (operatorAction operatorActorActionImpl) ActorType() actors.ActorType {
	return operatorAction.actorType
}

func (operatorAction operatorActorActionImpl) ActionType() actions.ActionType {
	return operatorAction.actionType
}

func (operatorAction operatorActorActionImpl) ActionEnums() []actions.IEnumAction {
	return operatorAction.enumAction
}