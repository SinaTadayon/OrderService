package actions

import "errors"

type ActionType int

var actionTypeStrings = []string{"ActorAction", "ActiveAction"}

const (
	ActorAction ActionType = iota
	ActiveAction
)

func (actionType ActionType) Name() string {
	return actionType.String()
}

func (actionType ActionType) Ordinal() int {
	if actionType < ActorAction || actionType > ActiveAction {
		return -1
	}
	return int(actionType)
}

func (actionType ActionType) Values() []string {
	return actionTypeStrings
}

func (actionType ActionType) String() string {
	if actionType < ActorAction || actionType > ActiveAction {
		return ""
	}

	return  actionTypeStrings[actionType]
}

func actionFromString(actionType string) (ActionType, error) {
	switch actionType {
	case "ActorAction":
		return ActorAction, nil
	case "ActiveAction":
		return ActiveAction, nil
	default:
		return -1, errors.New("invalid actionType string")
	}
}