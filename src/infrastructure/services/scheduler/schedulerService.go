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