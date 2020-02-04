package actions

import (
	"github.com/pkg/errors"
)

type ActionType int

var actionTypeStrings = []string{
	"Operator",
	"Seller",
	"Buyer",
	"Schedulers",
	"System",
}

const (
	Operator ActionType = iota
	Seller
	Buyer
	Scheduler
	System
)

func (actorType ActionType) ActionName() string {
	return actorType.String()
}

func (actorType ActionType) ActionOrdinal() int {
	if actorType < Operator || actorType > System {
		return -1
	}
	return int(actorType)
}

func (actorType ActionType) Values() []string {
	return actionTypeStrings
}

func (actorType ActionType) String() string {
	if actorType < Operator || actorType > System {
		return ""
	}

	return actionTypeStrings[actorType]
}

func FromString(actionType string) (ActionType, error) {
	switch actionType {
	case "Operator":
		return Operator, nil
	case "Seller":
		return Seller, nil
	case "Buyer":
		return Buyer, nil
	case "Schedulers":
		return Scheduler, nil
	case "System":
		return System, nil
	default:
		return -1, errors.New("invalid actionType string")
	}
}
