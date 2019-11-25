package operator_action

import (
	"github.com/pkg/errors"
)

type ActionEnums int

var actionStrings = []string{"AcceptAction", "DeliveredAction", "RejectAction",
	"CanceledAction", "ReturnedAction",
	"ReturnCanceledAction", "ReturnDeliveredAction"}

const (
	AcceptAction ActionEnums = iota
	DeliveredAction
	RejectAction
	CanceledAction
	ReturnedAction
	ReturnCanceledAction
	ReturnDeliveredAction
)

func (action ActionEnums) Name() string {
	return action.String()
}

func (action ActionEnums) Ordinal() int {
	if action < AcceptAction || action > ReturnDeliveredAction {
		return -1
	}

	return int(action)
}

func (action ActionEnums) Values() []string {
	return actionStrings
}

func (action ActionEnums) String() string {
	if action < AcceptAction || action > ReturnDeliveredAction {
		return ""
	}

	return actionStrings[action]
}

func FromString(action string) (ActionEnums, error) {
	switch action {
	case "AcceptAction":
		return AcceptAction, nil
	case "DeliveredAction":
		return DeliveredAction, nil
	case "RejectAction":
		return RejectAction, nil
	case "CanceledAction":
		return CanceledAction, nil
	case "ReturnedAction":
		return ReturnedAction, nil
	case "ReturnCanceledAction":
		return ReturnCanceledAction, nil
	case "ReturnDeliveredAction":
		return ReturnDeliveredAction, nil
	default:
		return -1, errors.New("invalid actorActionImpl string")
	}
}
