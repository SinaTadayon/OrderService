package notification_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{"SellerNotificationAction", "BuyerNotificationAction",
	"MarketNotificationAction", "OperatorNotificationAction"}

const (
	SellerNotificationAction ActionEnums = iota
	BuyerNotificationAction
	MarketNotificationAction
	OperatorNotificationAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < SellerNotificationAction || action > OperatorNotificationAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < SellerNotificationAction || action > OperatorNotificationAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "SellerNotificationAction":
		return SellerNotificationAction, nil
	case "BuyerNotificationAction":
		return BuyerNotificationAction, nil
	case "MarketNotificationAction":
		return MarketNotificationAction, nil
	case "OperatorNotificationAction":
		return OperatorNotificationAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
