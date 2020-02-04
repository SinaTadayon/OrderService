package scheduler_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"Cancel",
	"Close",
	"PaymentFail",
	"DeliveryDelay",
	"Deliver",
	"DeliveryPending",
	"Notification",
	"Reject",
	"Accept",
}

const (
	Cancel ActionEnums = iota
	Close
	PaymentFail
	DeliveryDelay
	Deliver
	DeliveryPending
	Notification
	Reject
	Accept
)

func (actionEnum ActionEnums) ActionName() string {
	return actionEnum.String()
}

func (actionEnum ActionEnums) ActionOrdinal() int {
	if actionEnum < Cancel || actionEnum > Accept {
		return -1
	}

	return int(actionEnum)
}

func (actionEnum ActionEnums) Values() []string {
	return actionStrings
}

func (actionEnum ActionEnums) String() string {
	if actionEnum < Cancel || actionEnum > Accept {
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
	case "PaymentFail":
		return PaymentFail
	case "DeliveryDelay":
		return DeliveryDelay
	case "DeliveryPending":
		return DeliveryPending
	case "Deliver":
		return Deliver
	case "Notification":
		return Notification
	case "Reject":
		return Reject
	case "Accept":
		return Accept
	default:
		return nil
	}
}
