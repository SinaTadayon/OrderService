package notification_action

import (
	"github.com/pkg/errors"
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

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "SellerNotification":
		return SellerNotification, nil
	case "BuyerNotification":
		return BuyerNotification, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
