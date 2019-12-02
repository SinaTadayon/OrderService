package stock_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{
	"Reserve",
	"Release",
	"Settlement",
}

const (
	Reserve ActionEnums = iota
	Release
	Settlement
)

func (action ActionEnums) ActionName() string {
	return action.String()
}

func (action ActionEnums) ActionOrdinal() int {
	if action < Reserve || action > Settlement {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < Reserve || action > Settlement {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "Reserve":
		return Reserve, nil
	case "Release":
		return Release, nil
	case "Settlement":
		return Settlement, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
