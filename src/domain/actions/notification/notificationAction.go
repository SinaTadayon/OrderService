package notification_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	actionType = actions.Notification
)

type notificationActiveActionImpl struct {
	actionType actions.ActionType
	enumAction actions.IEnumAction
}

func New(actionEnum ActionEnums) actions.IAction {
	return notificationActiveActionImpl{actionType, actions.IEnumAction(actionEnum)}
}

func (notification notificationActiveActionImpl) ActionType() actions.ActionType {
	return notification.actionType
}

func (notification notificationActiveActionImpl) ActionEnum() actions.IEnumAction {
	return notification.enumAction
}
