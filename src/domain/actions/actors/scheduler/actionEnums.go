package scheduler_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string { "TimeoutAction", "WaitForShippingDaysTimeoutAction",
	"NoActionForXDaysTimeoutAction", "WaitXDaysTimeoutAction",
	"AutoApprovedAction"}

const (
	TimeoutAction ActionEnums = iota
	WaitForShippingDaysTimeoutAction
	NoActionForXDaysTimeoutAction
	WaitXDaysTimeoutAction
	AutoApprovedAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < TimeoutAction || action > AutoApprovedAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < TimeoutAction || action > AutoApprovedAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "TimeoutAction":
		return TimeoutAction, nil
	case "WaitForShippingDaysTimeoutAction":
		return WaitForShippingDaysTimeoutAction, nil
	case "NoActionForXDaysTimeoutAction":
		return NoActionForXDaysTimeoutAction, nil
	case "WaitXDaysTimeoutAction":
		return WaitXDaysTimeoutAction, nil
	case "AutoApprovedAction":
		return AutoApprovedAction, nil

	default:
		return -1, errors.New("invalid actionEnums string")
	}
}

