package next_to_step_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string{"NextToStepAction"}

const (
	NextToStepAction ActionEnums = iota
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action != NextToStepAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action != NextToStepAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "NextToStepAction":
		return NextToStepAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
