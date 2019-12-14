package scheduler_service

//
//import (
//	"context"
//	"github.com/pkg/errors"
//	"gitlab.faza.io/go-framework/logger"
//	"gitlab.faza.io/go-framework/mongoadapter"
//	"gitlab.faza.io/order-project/order-service/domain"
//	"gitlab.faza.io/order-project/order-service/domain/events"
//	"go.mongodb.org/mongo-driver/bson"
//	"go.mongodb.org/mongo-driver/mongo"
//	"time"
//)
//
//const (
//	databaseName   string = "orderService"
//	collectionName string = "orders"
//)
//
//type fetchItemData struct {
//	OrderId       uint64
//	ItemId        uint64
//	PId      uint64
//	StepName      string
//	StepIndex     int
//	ActionHistory []fetchActionHistory
//}
//
//type fetchActionHistory struct {
//	ActionName  string
//	ExpiredTime time.Time
//	CreatedAt   time.Time
//}
//
//type iSchedulerServiceImpl struct {
//	mongoAdapter *mongoadapter.Mongo
//	flowManager  domain.IFlowManager
//}
//
//func NewScheduler(mongoAdapter *mongoadapter.Mongo, flowManager domain.IFlowManager) ISchedulerService {
//	return &iSchedulerServiceImpl{mongoAdapter: mongoAdapter, flowManager: flowManager}
//}
//
//func (scheduler *iSchedulerServiceImpl) Scheduler(ctx context.Context, schedulerData []ScheduleModel) error {
//
//	if schedulerData == nil || len(schedulerData) == 0 {
//		return errors.New("schedulerData is nil")
//	}
//
//	go scheduler.doSchedule(ctx, schedulerData)
//	return nil
//}
//
//func (scheduler *iSchedulerServiceImpl) doSchedule(ctx context.Context, schedulerData []ScheduleModel) {
//
//	schCtx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	for _, data := range schedulerData {
//		go scheduler.scheduleProcess(schCtx, data)
//	}
//	//for {
//	select {
//	case <-ctx.Done():
//		return
//		//default:
//	}
//	//}
//}
//
//func (scheduler *iSchedulerServiceImpl) scheduleProcess(ctx context.Context, model ScheduleModel) {
//
//	heartbeat := scheduler.worker(ctx, time.Duration(1*time.Minute), time.Duration(1*time.Hour), model)
//	const timeout = 5 * time.Minute
//
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case _, ok := <-heartbeat:
//			if ok == false {
//				logger.Audit("heartbeat of worker scheduler closed, step: %s, action: %s ", model.Step, model.Action)
//				//heartbeat = scheduler.worker(ctx, time.Duration(1 * time.Second), time.Duration(10 * time.Second), model)
//				return
//			}
//
//			//logger.Audit("heartbeat pulse")
//
//		case <-time.After(timeout):
//			logger.Audit("worker goroutine is not healthy!, step: %s, action: %s ", model.Step, model.Action)
//			return
//		}
//	}
//}
//
//func (scheduler *iSchedulerServiceImpl) worker(ctx context.Context, pulseInterval time.Duration,
//	scheduleInterval time.Duration, data ScheduleModel) <-chan interface{} {
//
//	var heartbeat = make(chan interface{}, 1)
//	go func() {
//		defer close(heartbeat)
//		pulse := time.Tick(pulseInterval)
//		schedule := time.Tick(scheduleInterval)
//		sendPulse := func() {
//			select {
//			case heartbeat <- struct{}{}:
//			default:
//			}
//		}
//
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-pulse:
//				sendPulse()
//			case <-schedule:
//				scheduler.doProcess(ctx, data)
//			}
//		}
//	}()
//	return heartbeat
//}
//
//// TODO Refactor
//func (scheduler *iSchedulerServiceImpl) doProcess(ctx context.Context, data ScheduleModel) {
//	logger.Audit("doProcess called . . .")
//	time.Sleep(5 * time.Second)
//	pipeline := []bson.M{
//		{"$match": bson.M{"items.deletedAt": nil, "items.progress.currentStepName": data.Step}},
//		{"$unwind": "$items"},
//		{"$unwind": bson.M{"path": "$items.progress.stepsHistory", "preserveNullAndEmptyArrays": false}},
//		{"$match": bson.M{"items.progress.stepsHistory.name": data.Step}},
//		{"$project": bson.M{
//			"_id":       0,
//			"orderId":   1,
//			"sid":       "$items.sid",
//			"sellerId":  "$items.sellerInfo.sellerId",
//			"stepName":  "$items.progress.stepsHistory.name",
//			"stepIndex": "$items.progress.stepsHistory.index",
//			"actionHistory": bson.M{"$filter": bson.M{"input": "$items.progress.stepsHistory.actionHistory",
//				"as":   "action",
//				"cond": bson.M{"$eq": bson.A{"$$action.name", data.Action}},
//			},
//			},
//		},
//		},
//		{"$project": bson.M{
//			"orderId":   1,
//			"sid":       1,
//			"sellerId":  1,
//			"stepName":  1,
//			"stepIndex": 1,
//			"actionHistory": bson.M{"$map": bson.M{"input": "$actionHistory",
//				"as": "action",
//				"in": bson.M{"actionName": "$$action.name",
//					"expiredTime": "$$action.data.expiredTime",
//					"createdAt":   "$$action.createdAt"},
//			},
//			},
//		}},
//	}
//
//	cursor, err := scheduler.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
//	if err != nil {
//		logger.Err("scheduler.mongoAdapter.Aggregate failed, step: %s, action: %s,  error: %s",
//			data.Step, data.Action, err)
//		return
//	}
//
//	defer closeCursor(ctx, cursor)
//
//	select {
//	case <-ctx.Done():
//		return
//	default:
//	}
//
//	var expiredOrderMap = make(map[uint64]map[uint64]*events.ISchedulerEvent, 64)
//
//	// iterate through all documents
//	for cursor.Next(ctx) {
//		var fetchData fetchItemData
//		// decode the document
//		if err := cursor.Decode(&fetchData); err != nil {
//			logger.Err("scheduler.mongoAdapter.Aggregate failed, step: %s, action: %s,  error: %s",
//				data.Step, data.Action, err)
//			return
//		}
//
//		if scheduler.checkExpiredTime(&fetchData) {
//			if sellerMap, isFindOrder := expiredOrderMap[fetchData.OrderId]; isFindOrder {
//				if schedulerEvent, isFindSeller := sellerMap[fetchData.PId]; isFindSeller {
//					schedulerEvent.ItemsId = append(schedulerEvent.ItemsId, fetchData.ItemId)
//				} else {
//					newEvent := &events.ISchedulerEvent{
//						OrderId:    fetchData.OrderId,
//						PId:   fetchData.PId,
//						ItemsId:    nil,
//						StateIndex: fetchData.StepIndex,
//						Action:     fetchData.ActionHistory[0].ActionName,
//					}
//
//					newEvent.ItemsId = make([]uint64, 0, 16)
//					newEvent.ItemsId = append(newEvent.ItemsId, fetchData.ItemId)
//					expiredOrderMap[fetchData.OrderId][fetchData.PId] = newEvent
//				}
//			} else {
//				newEvent := &events.ISchedulerEvent{
//					OrderId:    fetchData.OrderId,
//					PId:   fetchData.PId,
//					ItemsId:    nil,
//					StateIndex: fetchData.StepIndex,
//					Action:     fetchData.ActionHistory[0].ActionName,
//				}
//
//				newEvent.ItemsId = make([]uint64, 0, 16)
//				newEvent.ItemsId = append(newEvent.ItemsId, fetchData.ItemId)
//
//				expiredOrderMap[fetchData.OrderId] = make(map[uint64]*events.ISchedulerEvent, 16)
//				expiredOrderMap[fetchData.OrderId][fetchData.PId] = newEvent
//			}
//		}
//	}
//
//	for _, sellerMap := range expiredOrderMap {
//		for _, schedulerEvent := range sellerMap {
//			scheduler.flowManager.SchedulerEvents(*schedulerEvent)
//		}
//	}
//}
//
//func (scheduler *iSchedulerServiceImpl) checkExpiredTime(fetchData *fetchItemData) bool {
//	if fetchData.ActionHistory[0].ExpiredTime.Before(time.Now().UTC()) {
//		logger.Audit("action expired, "+
//			"orderId: %d, sid: %d, stepName: %s, stepIndex: %d, actionName: %s, expiredTime: %s ",
//			fetchData.OrderId, fetchData.ItemId, fetchData.StepName, fetchData.StepIndex,
//			fetchData.ActionHistory[0].ActionName, fetchData.ActionHistory[0].ExpiredTime)
//		return true
//	}
//
//	return false
//}
//
//func closeCursor(context context.Context, cursor *mongo.Cursor) {
//	err := cursor.Close(context)
//	if err != nil {
//		logger.Err("closeCursor() failed, error: %s", err)
//	}
//}
