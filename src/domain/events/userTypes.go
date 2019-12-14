package events

import "github.com/pkg/errors"

type UserType int

var userTypeStrings = []string{
	"Operator",
	"Seller",
	"Buyer",
	"Scheduler",
}

const (
	Operator UserType = iota
	Seller
	Buyer
	Scheduler
)

func (userType UserType) ActionName() string {
	return userType.String()
}

func (userType UserType) ActionOrdinal() int {
	if userType < Operator || Operator > userType {
		return -1
	}
	return int(userType)
}

func (userType UserType) Values() []string {
	return userTypeStrings
}

func (userType UserType) String() string {
	if userType < Operator || Operator > userType {
		return ""
	}

	return userTypeStrings[userType]
}

func FromUserString(userType string) (UserType, error) {
	switch userType {
	case "Operator":
		return Operator, nil
	case "Seller":
		return Seller, nil
	case "Buyer":
		return Buyer, nil
	case "Scheduler":
		return Scheduler, nil
	default:
		return -1, errors.New("invalid UserType string")
	}
}
