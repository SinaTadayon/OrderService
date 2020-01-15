package scheduler_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"time"
)

type StateInterval struct {
	State    states.IEnumState
	Interval time.Duration
}

type ISchedulerService interface {
	Scheduler(ctx context.Context)
}
