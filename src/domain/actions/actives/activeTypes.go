package actives

import (
	"errors"
)

type ActiveType int

var actionStrings = [] string {"NotificationAction", "NextToStepAction", "PayToSellerAction",
	"PayToBuyerAction", "PayToMarketAction",
	"StockAction", "ManualPaymentAction", "OrderPaymentAction",
	"RetryAction", "NewOrderAction", "FinalizeAction"}

const (
	NotificationAction ActiveType = iota
	NextToStepAction
	PayToSellerAction
	PayToBuyerAction
	PayToMarketAction
	StockAction
	ManualPaymentAction
	OrderPaymentAction
	RetryAction
	NewOrderAction
	FinalizeAction
)

func (action ActiveType) Name() string {
	return action.String()
}

func (action ActiveType) Ordinal() int {
	if action  < NotificationAction || action  > FinalizeAction {
		return -1
	}
	return int(action)
}

func (action ActiveType) Values() []string {
	return actionStrings
}

func (action ActiveType) String() string {
	if action < NotificationAction || action > FinalizeAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActiveType, error) {
	switch action {
	case "NotificationAction":
		return NotificationAction, nil
	case "NextToStepAction":
		return NextToStepAction, nil
	case "PayToSellerAction":
		return PayToSellerAction, nil
	case "PayToBuyerAction":
		return PayToBuyerAction, nil
	case "PayToMarketAction":
		return PayToMarketAction, nil
	case "StockAction":
		return StockAction, nil
	case "ManualPaymentAction":
		return ManualPaymentAction, nil
	case "OrderPaymentAction":
		return OrderPaymentAction, nil
	case "RetryAction":
		return RetryAction, nil
	case "NewOrderAction":
		return NewOrderAction, nil
	case "FinalizeAction":
		return FinalizeAction, nil
	default:
		return -1, errors.New("invalid activeType string")
	}
}

