package payment_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{
	"Success",
	"Fail",
}

const (
	Success ActionEnums = iota
	Fail
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < Success || action > Fail {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < Success || action > Fail {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "Success":
		return Success, nil
	case "Fail":
		return Fail, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
