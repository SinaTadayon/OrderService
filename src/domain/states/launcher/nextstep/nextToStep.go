package next_to_step_state

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/steps"
)

type INextToStep interface {
	ActionStepMap() map[actions.IEnumAction]steps.IStep
}
