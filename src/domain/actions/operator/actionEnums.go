package operator_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{
	"Delay",
	"Deliver",
	"DeliveryFail",
	"AcceptReturn",
	"RejectReturn",
	"CancelReturn",
}

const (
	Delay ActionEnums = iota
	Deliver
	DeliveryFail
	AcceptReturn
	RejectReturn
	CancelReturn
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < Delay || action > CancelReturn {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < Delay || action > CancelReturn {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "Delay":
		return Delay, nil
	case "Deliver":
		return Deliver, nil
	case "DeliveryFail":
		return DeliveryFail, nil
	case "AcceptReturn":
		return AcceptReturn, nil
	case "RejectReturn":
		return RejectReturn, nil
	case "CancelReturn":
		return CancelReturn, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
