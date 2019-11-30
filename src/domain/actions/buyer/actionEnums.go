package buyer_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{
	"Delay",
	"Cancel",
	"SubmitReturnRequest",
	"CancelReturn",
	"EnterShipmentDetails",
}

const (
	Delay ActionEnums = iota
	Cancel
	SubmitReturnRequest
	CancelReturn
	EnterShipmentDetails
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < Delay || action > EnterShipmentDetails {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < Delay || action > EnterShipmentDetails {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "Delay":
		return Delay, nil
	case "Cancel":
		return Cancel, nil
	case "SubmitReturnRequest":
		return SubmitReturnRequest, nil
	case "CancelReturn":
		return CancelReturn, nil
	case "EnterShipmentDetails":
		return EnterShipmentDetails, nil

	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
