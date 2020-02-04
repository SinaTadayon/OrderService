package states

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
)

type OrderStatus string
type PackageStatus string
type ActionResult string
type SchedulerType string

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

const (
	SchedulerSubpackageStateExpire SchedulerType = "SP_EXPIRATION"
	SchedulerSubpackageStateNotify SchedulerType = "SP_NOTIFICATION"
)

const (
	SchedulerJobName   string = "SCH_SP_JOB"
	SchedulerGroupName string = "SCH_SP_GROUP"
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
