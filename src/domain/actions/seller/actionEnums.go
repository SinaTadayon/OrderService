package seller_action

import (
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"Approve",
	"Reject",
	"Cancel",
	"AcceptReturn",
	"CancelReturn",
	"RejectReturn",
	"Deliver",
	"DeliveryFail",
	"EnterShipmentDetails",
}

const (
	Approve ActionEnums = iota
	Reject
	Cancel
	AcceptReturn
	CancelReturn
	RejectReturn
	Deliver
	DeliveryFail
	EnterShipmentDetails
)

func (actionEnum ActionEnums) ActionName() string {
	return actionEnum.String()
}

func (actionEnum ActionEnums) ActionOrdinal() int {
	if actionEnum < Approve || actionEnum > EnterShipmentDetails {
		return -1
	}

	return int(actionEnum)
}

func (actionEnum ActionEnums) Values() []string {
	return actionStrings
}

func (actionEnum ActionEnums) String() string {
	if actionEnum < Approve || actionEnum > EnterShipmentDetails {
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
	case "AcceptReturn":
		return AcceptReturn
	case "CancelReturn":
		return CancelReturn
	case "RejectReturn":
		return RejectReturn
	case "EnterShipmentDetails":
		return EnterShipmentDetails
	default:
		return nil
	}
}
