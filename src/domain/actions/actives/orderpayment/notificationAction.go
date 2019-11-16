package order_payment_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
)

const (
	activeType = actives.NotificationAction
	actionType = actions.ActiveAction
)

type notificationActiveActionImpl struct {
	actionType actions.ActionType
	activeType actives.ActiveType
	enumAction []actions.IEnumAction
}

func New(actionEnum []ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnum))
	for i, action := range actionEnum {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return notificationActiveActionImpl{actionType, activeType, iEnumAction}
}

func NewOf(actionEnums ...ActionEnums) actives.IActiveAction {
	iEnumAction := make([]actions.IEnumAction, len(actionEnums))
	for i, action := range actionEnums {
		iEnumAction[i] = actions.IEnumAction(action)
	}
	return notificationActiveActionImpl{actionType, activeType, iEnumAction}
}

func (notification notificationActiveActionImpl) ActiveType() actives.ActiveType {
	return notification.activeType
}

func (notification notificationActiveActionImpl) ActionType() actions.ActionType {
	return notification.actionType
}

func (notification notificationActiveActionImpl) ActionEnums() []actions.IEnumAction {
	return notification.enumAction
}
