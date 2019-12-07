package events

import "github.com/pkg/errors"

type EventType int

var eventTypeStrings = []string{
	"Action",
}

const (
	Action EventType = iota
)

func (eventType EventType) ActionName() string {
	return eventType.String()
}

func (eventType EventType) ActionOrdinal() int {
	if eventType != Action {
		return -1
	}
	return int(eventType)
}

func (eventType EventType) Values() []string {
	return eventTypeStrings
}

func (eventType EventType) String() string {
	if eventType != Action {
		return ""
	}

	return eventTypeStrings[eventType]
}

func FromString(eventType string) (EventType, error) {
	switch eventType {
	case "Action":
		return Action, nil
	default:
		return -1, errors.New("invalid eventType string")
	}
}
