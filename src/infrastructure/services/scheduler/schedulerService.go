package scheduler_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"time"
)

type StateConfig struct {
	State            states.IEnumState
	ScheduleInterval time.Duration
}

type ISchedulerService interface {
	Scheduler(ctx context.Context)
}
