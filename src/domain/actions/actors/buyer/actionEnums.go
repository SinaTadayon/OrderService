package buyer_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{"ApprovedAction", "RejectAction", "DeliveredAction",
	"NeedSupportAction", "CanceledAction",
	"ReturnIfPossibleAction", "EnterReturnShipmentDetailAction"}

const (
	ApprovedAction ActionEnums = iota
	RejectAction
	DeliveredAction
	NeedSupportAction
	CanceledAction
	ReturnIfPossibleAction
	EnterReturnShipmentDetailAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < ApprovedAction || action > EnterReturnShipmentDetailAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < ApprovedAction || action > EnterReturnShipmentDetailAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(actionEnums string) (ActionEnums, error) {
	switch actionEnums {
	case "ApprovedAction":
		return ApprovedAction, nil
	case "RejectAction":
		return RejectAction, nil
	case "DeliveredAction":
		return DeliveredAction, nil
	case "NeedSupportAction":
		return NeedSupportAction, nil
	case "CanceledAction":
		return CanceledAction, nil
	case "ReturnIfPossibleAction":
		return ReturnIfPossibleAction, nil
	case "EnterReturnShipmentDetailAction":
		return EnterReturnShipmentDetailAction, nil
	default:
		return -1, errors.New("invalid actionEnums string")
	}
}
