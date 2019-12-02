package system_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{"ComposeActorsAction", "CombineActorsAction"}

const (
	ComposeActorsAction ActionEnums = iota
	CombineActorsAction
)

func (action ActionEnums) ActionName() string {
	return action.String()
}

func (action ActionEnums) ActionOrdinal() int {
	if action < ComposeActorsAction || action > CombineActorsAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < ComposeActorsAction || action > CombineActorsAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "ComposeActorsAction":
		return ComposeActorsAction, nil
	case "CombineActorsAction":
		return CombineActorsAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
