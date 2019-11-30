package events

import "github.com/pkg/errors"

type EventType int

var eventTypeStrings = []string{
	"Payment",
	"Operator",
	"Seller",
	"Buyer",
	"Scheduler",
	"Stock",
	"Notification",
	"System",
}

const (
	Payment EventType = iota
	Operator
	Seller
	Buyer
	Scheduler
	Stock
	Notification
	System
)

func (eventType EventType) Name() string {
	return eventType.String()
}

func (eventType EventType) Ordinal() int {
	if eventType < Payment || eventType > Payment {
		return -1
	}
	return int(eventType)
}

func (eventType EventType) Values() []string {
	return eventTypeStrings
}

func (eventType EventType) String() string {
	if eventType < Payment || eventType > Payment {
		return ""
	}

	return eventTypeStrings[eventType]
}

func FromString(eventType string) (EventType, error) {
	switch eventType {
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
	case "System":
		return System, nil
	default:
		return -1, errors.New("invalid eventType string")
	}
}
