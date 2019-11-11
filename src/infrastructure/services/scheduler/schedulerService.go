package scheduler_service

import (
	"context"
)

type ISchedulerService interface {
	Scheduler(ctx context.Context, data []ScheduleModel) error
}

type ScheduleModel struct {
	step 	string
	action  string
}

type SchedulerEvent struct {
	OrderId 	string
	SellerId	string
	ItemsId		[]string
	StepIndex	int
	ActionName	string
}