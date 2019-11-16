package order_payment_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string{"OrderPaymentAction", "OrderPaymentFailedAction"}

const (
	OrderPaymentAction ActionEnums = iota
	OrderPaymentFailedAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < OrderPaymentAction || action > OrderPaymentFailedAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < OrderPaymentAction || action > OrderPaymentFailedAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "OrderPaymentAction":
		return OrderPaymentAction, nil
	case "OrderPaymentFailedAction":
		return OrderPaymentFailedAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
