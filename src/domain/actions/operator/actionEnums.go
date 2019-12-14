package operator_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"DeliveryDelay",
	"Deliver",
	"DeliveryFail",
	"Accept",
	"Reject",
}

const (
	DeliveryDelay ActionEnums = iota
	Deliver
	DeliveryFail
	Accept
	Reject
)

func (actionEnum ActionEnums) ActionName() string {
	return actionEnum.String()
}

func (actionEnum ActionEnums) ActionOrdinal() int {
	if actionEnum < DeliveryDelay || actionEnum > Reject {
		return -1
	}

	return int(actionEnum)
}

func (actionEnum ActionEnums) Values() []string {
	return actionStrings
}

func (actionEnum ActionEnums) String() string {
	if actionEnum < DeliveryDelay || actionEnum > Reject {
		return ""
	}

	return actionStrings[actionEnum]
}

func (actionEnum ActionEnums) FromString(action string) actions.IEnumAction {
	switch action {
	case "DeliveryDelay":
		return DeliveryDelay
	case "Deliver":
		return Deliver
	case "DeliveryFail":
		return DeliveryFail
	case "Accept":
		return Accept
	case "Reject":
		return Reject
	default:
		return nil
	}
}
