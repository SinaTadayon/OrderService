package payment_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
)

const (
	actorType  = actors.PaymentActor
	actionType = actions.ActorAction
)

type paymentActorActionImpl struct {
	actionType actions.ActionType
	actorType  actors.ActorType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return paymentActorActionImpl{actionType, actorType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actors.IActorAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return paymentActorActionImpl{actionType, actorType, iEnumAction}
}

func (paymentAction paymentActorActionImpl) ActorType() actors.ActorType {
	return paymentAction.actorType
}

func (paymentAction paymentActorActionImpl) ActionType() actions.ActionType {
	return paymentAction.actionType
}

func (paymentAction paymentActorActionImpl) ActionEnums() []actions.IEnumAction {
	return paymentAction.enumAction
}
