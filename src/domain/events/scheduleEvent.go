package events

import "gitlab.faza.io/order-project/order-service/domain/actions"

type ISchedulerEvent interface {
	IEvent
	OrderId() uint64
	SellerId() uint64
	ItemsId() []uint64
	StateIndex() int
	Action() actions.IAction
}
