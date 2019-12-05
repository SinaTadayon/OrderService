package pay_to_market_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

const (
	stateName  string = "Pay_To_Market_Action_State"
	activeType        = actives.PayToMarketAction
)

type payToMarketActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &payToMarketActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &payToMarketActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &payToMarketActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

func (payToMarket payToMarketActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
	panic("implementation required")
}
