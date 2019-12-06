package actions

import (
	"github.com/pkg/errors"
)

type ActionType int

var actionTypeStrings = []string{
	"Payment",
	"Operator",
	"Seller",
	"Buyer",
	"Scheduler",
	"Stock",
	"Notification",
	"Voucher",
	"System",
}

const (
	Payment ActionType = iota
	Operator
	Seller
	Buyer
	Scheduler
	Stock
	Notification
	Voucher
	System
)

func (actorType ActionType) ActionName() string {
	return actorType.String()
}

func (actorType ActionType) ActionOrdinal() int {
	if actorType < Payment || actorType > System {
		return -1
	}
	return int(actorType)
}

func (actorType ActionType) Values() []string {
	return actionTypeStrings
}

func (actorType ActionType) String() string {
	if actorType < Payment || actorType > System {
		return ""
	}

	return actionTypeStrings[actorType]
}

func FromString(actionType string) (ActionType, error) {
	switch actionType {
	case "Payment":
		return Payment, nil
	case "Operator":
		return Operator, nil
	case "Seller":
		return Seller, nil
	case "Buyer":
		return Buyer, nil
	case "Scheduler":
		return Scheduler, nil
	case "Stock":
		return Stock, nil
	case "Notification":
		return Notification, nil
	case "Voucher":
		return Voucher, nil
	case "System":
		return System, nil
	default:
		return -1, errors.New("invalid actionType string")
	}
}
