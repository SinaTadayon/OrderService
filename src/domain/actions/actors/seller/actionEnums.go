package seller_action

import (
	"errors"
)

type ActionEnums int

var actionStrings = []string{"ApprovedAction", "RejectAction", "DeliveredAction",
	"NeedSupportAction", "EnterShipmentDetailAction"}

const (
	ApprovedAction ActionEnums = iota
	RejectAction
	DeliveredAction
	NeedSupportAction
	EnterShipmentDetailAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < ApprovedAction || action > EnterShipmentDetailAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < ApprovedAction || action > EnterShipmentDetailAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "ApprovedAction":
		return ApprovedAction, nil
	case "RejectAction":
		return RejectAction, nil
	case "DeliveredAction":
		return DeliveredAction, nil
	case "NeedSupportAction":
		return NeedSupportAction, nil
	case "EnterShipmentDetailAction":
		return EnterShipmentDetailAction, nil
	default:
		return -1, errors.New("invalid actorActionImpl string")
	}
}
