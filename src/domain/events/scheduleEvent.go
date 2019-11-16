package events

type SchedulerEvent struct {
	OrderId    uint64
	SellerId   uint64
	ItemsId    []uint64
	StepIndex  int
	ActionName string
}
