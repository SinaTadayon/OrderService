package stock_action

import "gitlab.faza.io/order-project/order-service/domain/actions"

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

func (actionEnum ActionEnums) ActionName() string {
	return actionEnum.String()
}

func (actionEnum ActionEnums) ActionOrdinal() int {
	if actionEnum < Reserve || actionEnum > Settlement {
		return -1
	}

	return int(actionEnum)
}

func (actionEnum ActionEnums) Values() []string {
	return actionStrings
}

func (actionEnum ActionEnums) String() string {
	if actionEnum < Reserve || actionEnum > Settlement {
		return ""
	}

	return actionStrings[actionEnum]
}

func (actionEnum ActionEnums) FromString(action string) actions.IEnumAction {
	switch action {
	case "Reserve":
		return Reserve
	case "Release":
		return Release
	case "Settlement":
		return Settlement
	default:
		return nil
	}
}
