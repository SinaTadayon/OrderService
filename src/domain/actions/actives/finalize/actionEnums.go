package finalize_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string{"MarketFinalizeAction", "PaymentFailedFinalizeAction", "OrderFinalizeAction", "BuyerFinalizeAction"}

const (
	MarketFinalizeAction ActionEnums = iota
	PaymentFailedFinalizeAction
	OrderFailedFinalizeAction
	BuyerFinalizeAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < MarketFinalizeAction || action > BuyerFinalizeAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < MarketFinalizeAction || action > BuyerFinalizeAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "MarketFinalizeAction":
		return MarketFinalizeAction, nil
	case "PaymentFailedFinalizeAction":
		return PaymentFailedFinalizeAction, nil
	case "OrderFailedFinalizeAction":
		return OrderFailedFinalizeAction, nil
	case "BuyerFinalizeAction":
		return BuyerFinalizeAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
