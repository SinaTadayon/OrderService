package launcher_state

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
)

type ILauncherState interface {
	states.IState
	ActiveType() actives.ActiveType
	ActionLauncher(ctx context.Context, order entities.Order, params ...interface{})
}