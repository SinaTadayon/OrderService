package scheduler_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"Cancel",
	"Close",
	"DeliveryDelay",
	"Deliver",
	"DeliveryPending",
	"RejectReturn",
	"AcceptReturn",
}

const (
	Cancel ActionEnums = iota
	Close
	DeliveryDelay
	Deliver
	DeliveryPending
	RejectReturn
	AcceptReturn
)

func (actionEnum ActionEnums) ActionName() string {
	return actionEnum.String()
}

func (actionEnum ActionEnums) ActionOrdinal() int {
	if actionEnum < Cancel || actionEnum > AcceptReturn {
		return -1
	}

	return int(actionEnum)
}

func (actionEnum ActionEnums) Values() []string {
	return actionStrings
}

func (actionEnum ActionEnums) String() string {
	if actionEnum < Cancel || actionEnum > AcceptReturn {
		return ""
	}

	return actionStrings[actionEnum]
}

func (actionEnum ActionEnums) FromString(action string) actions.IEnumAction {
	switch action {
	case "Cancel":
		return Cancel
	case "Close":
		return Close
	case "DeliveryDelay":
		return DeliveryDelay
	case "DeliveryPending":
		return DeliveryPending
	case "Deliver":
		return Deliver
	case "RejectReturn":
		return RejectReturn
	case "AcceptReturn":
		return AcceptReturn
	default:
		return nil
	}
}
