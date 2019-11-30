package states

import (
	"gitlab.faza.io/order-project/order-service/domain/states_old"
)

type IBaseStep interface {
	BaseStep() *BaseStepImpl
	StatesMap() map[int]states_old.IState
}
