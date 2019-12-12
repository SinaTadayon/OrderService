package buyer_action

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type ActionEnums int

var actionStrings = []string{
	"DeliveryDelay",
	"Cancel",
	"SubmitReturnRequest",
	"EnterShipmentDetail",
}

const (
	DeliveryDelay ActionEnums = iota
	Cancel
	SubmitReturnRequest
	EnterShipmentDetail
)

func (action ActionEnums) ActionName() string {
	return action.String()
}

func (action ActionEnums) ActionOrdinal() int {
	if action < DeliveryDelay || action > EnterShipmentDetail {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < DeliveryDelay || action > EnterShipmentDetail {
		return ""
	}

	return actionStrings[action]
}

func (action ActionEnums) FromString(actionEnums string) actions.IEnumAction {
	switch actionEnums {
	case "DeliveryDelay":
		return DeliveryDelay
	case "Cancel":
		return Cancel
	case "SubmitReturnRequest":
		return SubmitReturnRequest
	case "EnterShipmentDetail":
		return EnterShipmentDetail

	default:
		return nil
	}
}
