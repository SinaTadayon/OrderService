package scheduler_service

import (
	"context"
)

type ISchedulerService interface {
	Scheduler(ctx context.Context)
}
