package manual_payment_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string{"ManualPaymentToMarketAction",
	"ManualPaymentToSellerAction",
	"ManualPaymentToBuyerAction"}

const (
	ManualPaymentToMarketAction ActionEnums = iota
	ManualPaymentToSellerAction
	ManualPaymentToBuyerAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < ManualPaymentToMarketAction || action > ManualPaymentToBuyerAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < ManualPaymentToMarketAction || action > ManualPaymentToBuyerAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "ManualPaymentToMarketAction":
		return ManualPaymentToMarketAction, nil
	case "ManualPaymentToSellerAction":
		return ManualPaymentToSellerAction, nil
	case "ManualPaymentToBuyerAction":
		return ManualPaymentToBuyerAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
