package states

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

const (
	NewStatus        = "NEW"
	InProgressStatus = "IN_PROGRESS"
	ClosedStatus     = "CLOSED"
)

type IState interface {
	Name() string
	Index() int
	Childes() []IState
	Parents() []IState
	Actions() []actions.IAction
	Process(ctx context.Context, frame frame.IFrame) future.IFuture
	//ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture
}
