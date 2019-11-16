package events

type SchedulerEvent struct {
	OrderId    string
	SellerId   string
	ItemsId    []string
	StepIndex  int
	ActionName string
}
