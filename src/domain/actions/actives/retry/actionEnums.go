package retry_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string{"RetryAction"}

const (
	RetryAction ActionEnums = iota
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action != RetryAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action != RetryAction {
		return ""
	}

	return  actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "RetryAction":
		return RetryAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}

