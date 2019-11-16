package actives

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type IActiveAction interface {
	actions.IAction
	ActionEnums() []actions.IEnumAction
	ActiveType() ActiveType
}
