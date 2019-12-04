package notification_action

import (
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"SellerNotification",
	"BuyerNotification",
}

const (
	SellerNotification ActionEnums = iota
	BuyerNotification
)

func (action ActionEnums) ActionName() string {
	return action.String()
}

func (action ActionEnums) ActionOrdinal() int {
	if action < SellerNotification || action > BuyerNotification {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < SellerNotification || action > BuyerNotification {
		return ""
	}

	return actionStrings[action]
}

func (action ActionEnums) FromString(actionEnums string) actions.IEnumAction {
	switch actionEnums {
	case "SellerNotification":
		return SellerNotification
	case "BuyerNotification":
		return BuyerNotification
	default:
		return nil
	}
}
