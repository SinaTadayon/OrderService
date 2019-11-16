package actors

import (
	"errors"
)

type ActorType int

var actorTypeStrings = []string{"PaymentActor",
	"OperatorActor", "SellerActor",
	"BuyerActor", "SchedulerActor", "CheckoutActor", "SystemActor"}

const (
	PaymentActor ActorType = iota
	OperatorActor
	SellerActor
	BuyerActor
	SchedulerActor
	CheckoutActor
	SystemActor
)

func (actorType ActorType) Name() string {
	return actorType.String()
}

func (actorType ActorType) Ordinal() int {
	if actorType < PaymentActor || actorType > SystemActor {
		return -1
	}
	return int(actorType)
}

func (actorType ActorType) Values() []string {
	return actorTypeStrings
}

func (actorType ActorType) String() string {
	if actorType < PaymentActor || actorType > SystemActor {
		return ""
	}

	return actorTypeStrings[actorType]
}

func FromString(actorType string) (ActorType, error) {
	switch actorType {
	case "PaymentActor":
		return PaymentActor, nil
	case "OperatorActor":
		return OperatorActor, nil
	case "SellerActor":
		return SellerActor, nil
	case "BuyerActor":
		return BuyerActor, nil
	case "SchedulerActor":
		return SchedulerActor, nil
	case "CheckoutActor":
		return CheckoutActor, nil
	case "SystemActor":
		return SystemActor, nil
	default:
		return -1, errors.New("invalid actorType string")
	}
}
