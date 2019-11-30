package states_old

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

type IState interface {
	Name() string
	Index() int
	Childes() []IState
	Parents() []IState
	Actions() actions.IAction
}
