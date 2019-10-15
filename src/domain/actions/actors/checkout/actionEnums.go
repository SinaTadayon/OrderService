package checkout_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string {"NewOrderAction"}

const (
	NewOrderAction ActionEnums = iota
)

func (checkoutAction ActionEnums) Name() string {
	return checkoutAction.String()
}

func (checkoutAction ActionEnums) Ordinal() int {
	if checkoutAction != NewOrderAction {
		return -1
	}

	return int(checkoutAction)
}

func (checkoutAction ActionEnums) Values() []string {
	return actionStrings
}

func (checkoutAction ActionEnums) String() string {
	if checkoutAction != NewOrderAction {
		return ""
	}

	return actionStrings[checkoutAction]
}

func FromString(checkoutAction string) (ActionEnums, error) {
	switch checkoutAction {
	case "TimeoutActionScheduler":
		return NewOrderAction, nil
	default:
		return -1, errors.New("invalid checkoutAction string")
	}
}
