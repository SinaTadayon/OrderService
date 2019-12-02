package scheduler_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{
	"Cancel",
	"Close",
	"Delay",
	"Deliver",
	"Pending",
	"RejectReturn",
	"AcceptReturn",
}

const (
	Cancel ActionEnums = iota
	Close
	Delay
	Deliver
	Pending
	RejectReturn
	AcceptReturn
)

func (action ActionEnums) ActionName() string {
	return action.String()
}

func (action ActionEnums) ActionOrdinal() int {
	if action < Cancel || action > AcceptReturn {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < Cancel || action > AcceptReturn {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "Cancel":
		return Cancel, nil
	case "Close":
		return Close, nil
	case "Delay":
		return Delay, nil
	case "Pending":
		return Pending, nil
	case "Deliver":
		return Deliver, nil
	case "RejectReturn":
		return RejectReturn, nil
	case "AcceptReturn":
		return AcceptReturn, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
