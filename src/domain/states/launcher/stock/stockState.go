package stock_action_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher"
)

const (
	stateName string = "Stock_Action_State"
	activeType = actives.StockAction
)

type stockActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &stockActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction,
	launcherType actives.ActiveType) launcher_state.ILauncherState {
	return &stockActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, launcherType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &stockActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

func (stock stockActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, param interface{}) promise.IPromise {
	panic("implementation required")
	return
}

