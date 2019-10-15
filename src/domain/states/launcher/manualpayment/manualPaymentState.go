package manual_payment_state

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
	stateName string = "Manual_Payment_Action_State"
	activeType = actives.ManualPaymentAction
)

type manualPaymentActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &manualPaymentActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &manualPaymentActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &manualPaymentActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (manualPayment manualPaymentActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, params ...interface{}) {
	panic("implementation required")
}

