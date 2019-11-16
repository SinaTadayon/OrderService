package events

import "errors"

type EventType int

var actorTypeStrings = []string{"ActorEvent", "ActiveEvent"}

const (
	ActorEvent EventType = iota
	ActiveEvent
)

func (activeType EventType) Name() string {
	return activeType.String()
}

func (activeType EventType) Ordinal() int {
	if activeType < ActorEvent || activeType > ActiveEvent {
		return -1
	}
	return int(activeType)
}

func (activeType EventType) Values() []string {
	return actorTypeStrings
}

func (activeType EventType) String() string {
	if activeType < ActorEvent || activeType > ActiveEvent {
		return ""
	}

	return actorTypeStrings[activeType]
}

func FromString(actorType string) (EventType, error) {
	switch actorType {
	case "ActorEvent":
		return ActorEvent, nil
	case "ActiveEvent":
		return ActiveEvent, nil
	default:
		return -1, errors.New("invalid actorType string")
	}
}
