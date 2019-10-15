package payment_action

import (
	"errors"
)

type ActionEnums int
var actionStrings = []string { "SuccessAction", "FailedAction" }

const (
	SuccessAction ActionEnums = iota
	FailedAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < SuccessAction || action > FailedAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < SuccessAction || action > FailedAction {
		return ""
	}

	return  actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "SuccessAction":
		return SuccessAction, nil
	case "FailedAction":
		return FailedAction, nil
	default:
		return -1, errors.New("invalid actorActionImpl string")
	}
}

