package system_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"ComposeActorsAction",
	"Success",
	"Fail",
	"NextToState",
	"Close",
	"CombineActorsAction",
}

const (
	ComposeActorsAction ActionEnums = iota
	Success
	Fail
	NextToState
	Close
	CombineActorsAction
)

func (actionEnum ActionEnums) ActionName() string {
	return actionEnum.String()
}

func (actionEnum ActionEnums) ActionOrdinal() int {
	if actionEnum < ComposeActorsAction || actionEnum > CombineActorsAction {
		return -1
	}

	return int(actionEnum)
}

func (actionEnum ActionEnums) Values() []string {
	return actionStrings
}

func (actionEnum ActionEnums) String() string {
	if actionEnum < ComposeActorsAction || actionEnum > CombineActorsAction {
		return ""
	}

	return actionStrings[actionEnum]
}

func (actionEnum ActionEnums) FromString(action string) actions.IEnumAction {
	switch action {
	case "ComposeActorsAction":
		return ComposeActorsAction
	case "Success":
		return Success
	case "Fail":
		return Fail
	case "NextToState":
		return NextToState
	case "Close":
		return Close
	case "CombineActorsAction":
		return CombineActorsAction
	default:
		return nil
	}
}
