package scheduler_service

import (
	"context"
	"errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const (
	databaseName  	string = "orderService"
	collectionName  string = "orders"
)


type fetchItemData struct {
	OrderId  		string
	ItemId	 		string
	SellerId		string
	StepName		string
	StepIndex		int
	actionHistory	[]fetchActionHistory
}

type fetchActionHistory struct {
	ActionName		string
	expiredTime		time.Time
	createdAt		time.Time
}

type iSchedulerServiceImpl struct {
	mongoAdapter  *mongoadapter.Mongo
	flowManager   domain.IFlowManager
	data  		  []ScheduleModel
}

func NewScheduler(mongoAdapter  *mongoadapter.Mongo, flowManager domain.IFlowManager) ISchedulerService {
	return &iSchedulerServiceImpl{mongoAdapter:mongoAdapter,data:nil}
}

func (scheduler *iSchedulerServiceImpl) Scheduler(ctx context.Context, schedulerData []ScheduleModel) error {

	if schedulerData == nil {
		return errors.New("schedulerData is nil")
	}

	scheduler.data = schedulerData
	go scheduler.doSchedule(ctx)

	return nil
}

func (scheduler *iSchedulerServiceImpl) doSchedule(ctx context.Context) {

	schCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
			case <-ctx.Done():
				return
		default:
		}

		for _, data := range scheduler.data {
			go scheduler.scheduleProcess(schCtx, data)
		}
	}
}

func (scheduler *iSchedulerServiceImpl) scheduleProcess(ctx context.Context, model ScheduleModel) {
	for {
		select {
		case <-ctx.Done():
			return
		case <- time.After(10 * time.Second):
			scheduler.doProcess(ctx, model)
		default:
		}
	}
}

// TODO Refactor
func (scheduler *iSchedulerServiceImpl) doProcess(ctx context.Context, data ScheduleModel) {
	pipeline := []bson.M{
		bson.M{ "$match": bson.M{"items.deletedAt": nil, "items.progress.currentStepName": data.step }},
		bson.M{ "$unwind": "$items"},
		bson.M{ "$unwind": bson.M{"path": "$items.progress.stepsHistory", "preserveNullAndEmptyArrays": false}},
		bson.M{ "$match": bson.M{"items.progress.stepsHistory.name": data.step }},
		bson.M{ "$project": bson.M{
				"_id": 0,
				"orderId": 1,
				"itemId": "$items.itemId",
				"sellerId": "$items.sellerInfo.sellerId",
				"stepName": "$items.progress.stepsHistory.name",
				"stepIndex": "$items.progress.stepsHistory.index",
				"actionHistory":
				bson.M{"$filter":
					bson.M{"input": "$items.progress.stepsHistory.actionHistory",
							"as": "action",
							"cond": bson.M{"$eq": bson.A{"$$action.name", data.action}},
						},
					},
				},
			},
		bson.M{ "$project": bson.M{
			"orderId": 1,
			"itemId": 1,
			"sellerId": 1,
			"stepName": 1,
			"stepIndex": 1,
			"actionHistory":
				bson.M{"$map":
					bson.M{"input": "$actionHistory",
						"as": "action",
						"in": bson.M{"actionName": "$$ah.name",
							"expiredTime": "$$ah.data.expiredTime",
							"createdAt":   "$$ah.createdAt"},
					},
				},
			}},
		}

	cursor, err := scheduler.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		logger.Err("scheduler.mongoAdapter.Aggregate failed, step: %s, action: %s,  error: %s",
			data.step, data.action, err)
		return
	}

	defer closeCursor(ctx, cursor)
	var expiredOrderMap = make(map[string]map[string]*SchedulerEvent, 64)

	// iterate through all documents
	for cursor.Next(ctx) {
		var fetchData fetchItemData
		// decode the document
		if err := cursor.Decode(&fetchData); err != nil {
			logger.Err("scheduler.mongoAdapter.Aggregate failed, step: %s, action: %s,  error: %s",
				data.step, data.action, err)
			return
		}

		if scheduler.checkExpiredTime(&fetchData) {
			if sellerMap, isFindOrder := expiredOrderMap[fetchData.OrderId]; isFindOrder {
				 if schedulerEvent, isFindSeller := sellerMap[fetchData.SellerId]; isFindSeller {
				 	schedulerEvent.ItemsId = append(schedulerEvent.ItemsId, fetchData.ItemId)
				 } else {
				 	newEvent := &SchedulerEvent {
						OrderId:    fetchData.OrderId,
						SellerId:   fetchData.SellerId,
						ItemsId:    nil,
						StepIndex:  fetchData.StepIndex,
						ActionName: fetchData.actionHistory[0].ActionName,
					}

				 	newEvent.ItemsId = make([]string,0, 16)
				 	newEvent.ItemsId = append(newEvent.ItemsId, fetchData.ItemId)
				 	expiredOrderMap[fetchData.OrderId][fetchData.SellerId] = newEvent
				 }
			} else {
				newEvent := &SchedulerEvent {
					OrderId:    fetchData.OrderId,
					SellerId:   fetchData.SellerId,
					ItemsId:    nil,
					StepIndex:  fetchData.StepIndex,
					ActionName: fetchData.actionHistory[0].ActionName,
				}

				newEvent.ItemsId = make([]string,0, 16)
				newEvent.ItemsId = append(newEvent.ItemsId, fetchData.ItemId)

				expiredOrderMap[fetchData.OrderId] = make(map[string]*SchedulerEvent, 16)
				expiredOrderMap[fetchData.OrderId][fetchData.SellerId] = newEvent
			}
		}
	}

	for _, sellerMap := range expiredOrderMap {
		for _, schedulerEvent := range sellerMap {
			scheduler.flowManager.SchedulerEvents(*schedulerEvent)
		}
	}
}

func (scheduler *iSchedulerServiceImpl) checkExpiredTime(fetchData *fetchItemData) bool {
	if fetchData.actionHistory[0].expiredTime.Before(time.Now()) {
		logger.Audit("action expired, " +
			"orderId: %s, itemId: %s, stepName: %s, stepIndex: %s, actionName: %s, expiredTime: %s ",
			fetchData.OrderId, fetchData.ItemId, fetchData.StepIndex, fetchData.StepName,
			fetchData.actionHistory[0].ActionName, fetchData.actionHistory[0].expiredTime)
		return true
	}

	return false
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		logger.Err("closeCursor() failed, error: %s", err)
	}
}