package next_to_step_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher"
	"gitlab.faza.io/order-project/order-service/domain/steps"
)

const (
	stateName string = "Next_To_Step_Action_State"
	activeType = actives.NextToStepAction
)

type nextToStepActionLauncher struct {
	*launcher_state.BaseLauncherImpl
	actionStepMap 	map[actions.IEnumAction]steps.IStep
}

func New(index int, childes, parents []states.IState, actions actions.IAction, actionStepMap map[actions.IEnumAction]steps.IStep) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType), actionStepMap}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction, actionStepMap map[actions.IEnumAction]steps.IStep) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType), actionStepMap}
}

func NewFrom(base *launcher_state.BaseLauncherImpl, actionStepMap map[actions.IEnumAction]steps.IStep) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{base, actionStepMap}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

func (nextStep nextToStepActionLauncher) ActionStepMap() map[actions.IEnumAction]steps.IStep {
	return nextStep.actionStepMap
}

func (nextStep nextToStepActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, params ...interface{}) {
	panic("implementation required")
}

