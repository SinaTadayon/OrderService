package launcher_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type ILauncherState interface {
	states.IState
	ActiveType() actives.ActiveType
	ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise
}
