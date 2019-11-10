package scheduler_service

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain"
)

type iSchedulerServiceImpl struct {
	mongoAdapter  *mongoadapter.Mongo
	flowManager   domain.IFlowManager
	data  		  []ScheduleModel
}

func NewScheduler(mongoAdapter  *mongoadapter.Mongo, flowManager domain.IFlowManager) ISchedulerService {
	return &iSchedulerServiceImpl{mongoAdapter:mongoAdapter,data:nil}
}

func (scheduler *iSchedulerServiceImpl) Scheduler(ctx context.Context, schedulerData []ScheduleModel) error {
	//for _, data := range schedulerData {
	//
	//}
	return nil
}

func (scheduler *iSchedulerServiceImpl) worker(ctx context.Context, data ScheduleModel) {

}