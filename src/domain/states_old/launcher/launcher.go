package launcher_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type ILauncherState interface {
	states_old.IState
	ActiveType() actives.ActiveType
	ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture
}
