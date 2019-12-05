package states

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
)

type OrderStatus string
type PackageStatus string
type ActionResult string

const (
	OrderNewStatus        OrderStatus = "NEW"
	OrderInProgressStatus OrderStatus = "IN_PROGRESS"
	OrderClosedStatus     OrderStatus = "CLOSED"
)

const (
	PackageNewStatus        PackageStatus = "NEW"
	PackageInProgressStatus PackageStatus = "IN_PROGRESS"
	PackageClosedStatus     PackageStatus = "CLOSED"
)

const (
	ActionSuccess ActionResult = "Success"
	ActionFail    ActionResult = "Fail"
	ActionCancel  ActionResult = "Cancel"
)

type IState interface {
	Name() string
	Index() int
	Childes() []IState
	Parents() []IState
	Actions() []actions.IAction
	IsActionValid(actions.IAction) bool
	Process(ctx context.Context, frame frame.IFrame)
	//ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture
}
