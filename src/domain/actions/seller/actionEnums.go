package seller_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"Approve",
	"Reject",
	"Cancel",
	"Accept",
	"CancelReturn",
	"RejectReturn",
	"Deliver",
	"DeliveryFail",
	"EnterShipmentDetail",
}

const (
	Approve ActionEnums = iota
	Reject
	Cancel
	Accept
	CancelReturn
	RejectReturn
	Deliver
	DeliveryFail
	EnterShipmentDetail
)

func (actionEnum ActionEnums) ActionName() string {
	return actionEnum.String()
}

func (actionEnum ActionEnums) ActionOrdinal() int {
	if actionEnum < Approve || actionEnum > EnterShipmentDetail {
		return -1
	}

	return int(actionEnum)
}

func (actionEnum ActionEnums) Values() []string {
	return actionStrings
}

func (actionEnum ActionEnums) String() string {
	if actionEnum < Approve || actionEnum > EnterShipmentDetail {
		return ""
	}

	return actionStrings[actionEnum]
}

func (actionEnum ActionEnums) FromString(action string) actions.IEnumAction {
	switch action {
	case "Approve":
		return Approve
	case "Reject":
		return Reject
	case "Cancel":
		return Cancel
	case "Deliver":
		return Deliver
	case "DeliveryFail":
		return DeliveryFail
	case "Accept":
		return Accept
	case "CancelReturn":
		return CancelReturn
	case "RejectReturn":
		return RejectReturn
	case "EnterShipmentDetail":
		return EnterShipmentDetail
	default:
		return nil
	}
}
