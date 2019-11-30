package next_to_step_state

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/states"
)

type INextToStep interface {
	ActionStepMap() map[actions.IEnumAction]states.IStep
}
