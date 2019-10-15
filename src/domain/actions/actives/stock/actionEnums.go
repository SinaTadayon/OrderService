package stock_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string{"ReservedAction", "ReleasedAction", "SettlementAction", "FailedAction"}

const (
	ReservedAction ActionEnums = iota
	ReleasedAction
	SettlementAction
	FailedAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < ReservedAction || action > FailedAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < ReservedAction || action > FailedAction {
		return ""
	}

	return  actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "ReservedAction":
		return ReservedAction, nil
	case "ReleasedAction":
		return ReleasedAction, nil
	case "SettlementAction":
		return SettlementAction, nil
	case "FailedAction":
		return FailedAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}

