package steps

import "gitlab.faza.io/order-project/order-service/domain/states"

type IBaseStep interface {
	BaseStep() 		*BaseStepImpl
	StatesMap()		map[int]states.IState
}
