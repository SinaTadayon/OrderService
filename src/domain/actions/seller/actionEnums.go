package seller_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{
	"Approve",
	"Reject",
	"Cancel",
	"AcceptReturn",
	"CancelReturn",
	"RejectReturn",
	"Deliver",
	"EnterShipmentDetails",
}

const (
	Approve ActionEnums = iota
	Reject
	Cancel
	AcceptReturn
	CancelReturn
	RejectReturn
	Deliver
	EnterShipmentDetails
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < Approve || action > EnterShipmentDetails {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < Approve || action > EnterShipmentDetails {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "Approve":
		return Approve, nil
	case "Reject":
		return Reject, nil
	case "Cancel":
		return Cancel, nil
	case "Deliver":
		return Deliver, nil
	case "AcceptReturn":
		return AcceptReturn, nil
	case "CancelReturn":
		return CancelReturn, nil
	case "RejectReturn":
		return RejectReturn, nil
	case "EnterShipmentDetails":
		return EnterShipmentDetails, nil
	default:
		return -1, errors.New("invalid actorActionImpl string")
	}
}
