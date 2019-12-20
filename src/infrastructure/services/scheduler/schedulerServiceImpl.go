package scheduler_service

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	order "gitlab.faza.io/protos/order"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"time"
)

const (
	databaseName   string = "orderService"
	collectionName string = "orders"
)

type StateScheduler struct {
	OrderId uint64
	PId     uint64
	SId     uint64
	SIdx    int
	Items   []Item
	Data    Data
}

type Data struct {
	Name   string
	Value  time.Time
	Action string
}

type Item struct {
	InventoryId string
	Quantity    int32
}

type SchedulerService struct {
	mongoAdapter   *mongoadapter.Mongo
	orderClient    order.OrderServiceClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
	state          states.IEnumState
}

func NewScheduler(mongoAdapter *mongoadapter.Mongo, address string, port int, state states.IEnumState) *SchedulerService {
	return &SchedulerService{mongoAdapter: mongoAdapter, serverAddress: address, serverPort: port, state: state}
}

func (scheduler *SchedulerService) Scheduler(ctx context.Context) {
	go scheduler.doSchedule(ctx)
}

func (scheduler *SchedulerService) doSchedule(ctx context.Context) {

	schCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//for _, data := range schedulerData {
	go scheduler.scheduleProcess(schCtx)
	//}
	//for {
	select {
	case <-ctx.Done():
		return
		//default:
	}
	//}
}

func (scheduler *SchedulerService) scheduleProcess(ctx context.Context) {

	heartbeat := scheduler.worker(ctx, time.Duration(1*time.Minute), time.Duration(1*time.Hour))
	const timeout = 5 * time.Minute

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-heartbeat:
			if ok == false {
				logger.Audit("heartbeat of worker scheduler closed, state: %s", scheduler.state.StateName())
				//heartbeat = scheduler.worker(ctx, time.Duration(1 * time.Second), time.Duration(10 * time.Second), model)
				return
			}

			//logger.Audit("heartbeat pulse")

		case <-time.After(timeout):
			logger.Audit("worker goroutine is not healthy!, state: %s", scheduler.state.StateName())
			return
		}
	}
}

func (scheduler *SchedulerService) worker(ctx context.Context, pulseInterval time.Duration,
	scheduleInterval time.Duration) <-chan interface{} {

	var heartbeat = make(chan interface{}, 1)
	go func() {
		defer close(heartbeat)
		pulse := time.Tick(pulseInterval)
		schedule := time.Tick(scheduleInterval)
		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default:
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-pulse:
				sendPulse()
			case <-schedule:
				scheduler.doProcess(ctx)
			}
		}
	}()
	return heartbeat
}

// TODO Refactor
func (scheduler *SchedulerService) doProcess(ctx context.Context) {
	logger.Audit("doProcess called . . .")
	time.Sleep(5 * time.Second)

	var perPage = int64(16)
	//var page = 1

	totalCount, err := scheduler.getTotalCount(ctx)
	if err != nil {
		logger.Err("scheduler worker doProcess() => getTotalCount failed, state: %s", scheduler.state.StateName())
		return
	}

	for page := int64(1); page < totalCount/perPage; page++ {
		statesList, total, err := scheduler.findAllWithPage(ctx, page, perPage)
		if err != nil {
			logger.Err("scheduler worker doProcess() => findAllWithPage failed, state: %s", scheduler.state.StateName())
			return
		}

		for _, stateSchedule := range statesList {
			subpackages := make([]*order.ActionData_Subpackage, 0, len(actionRequest.Data.Subpackages))
			for _, subPkgRequest := range actionRequest.Data.Subpackages {
				subpackageItems := make([]*order.ActionData_Subpackage_Item, 0, len(subPkgRequest.Items))
				for _, subPkgItem := range subPkgRequest.Items {
					subpackageItem := &order.ActionData_Subpackage_Item{
						InventoryId: subPkgItem.InventoryId,
						Quantity:    int32(subPkgItem.Quantity),
						Reasons:     subPkgItem.Reasons,
					}
					subpackageItems = append(subpackageItems, subpackageItem)
				}

				subpackage := &order.ActionData_Subpackage{
					SID:   subPkgRequest.SID,
					Items: subpackageItems,
				}

				subpackages = append(subpackages, subpackage)
			}

			actionData := &order.ActionData{
				Subpackages:    subpackages,
				Carrier:        actionRequest.Data.Carrier,
				TrackingNumber: actionRequest.Data.TrackingNumber,
			}

			serializedData, err := proto.Marshal(actionData)
			if err != nil {
				logger.Err("OrderAction() => could not serialize pbOrder.ActionData, request: %v, error:%s", actionRequest, err)
				return
			}

			msgReq := &order.MessageRequest{
				Name:   "",
				Type:   string(ActionReqType),
				ADT:    string(SingleType),
				Method: string(PostMethod),
				Time:   ptypes.TimestampNow(),
				Meta: &order.RequestMetadata{
					UID:     actionRequest.UID,
					UTP:     string(utp),
					OID:     actionRequest.OID,
					PID:     actionRequest.PID,
					SIDs:    nil,
					Page:    0,
					PerPage: 0,
					//IpAddress: ipAddress,
					Action: &order.MetaAction{
						ActionType:  actionRequest.Type,
						ActionState: string(action),
						StateIndex:  int32(actionRequest.SIdx),
					},
					Sorts:   nil,
					Filters: nil,
				},
				Data: &any.Any{
					TypeUrl: "baman.io/" + proto.MessageName(actionData),
					Value:   serializedData,
				},
			}

			response, err := scheduler.orderClient.SchedulerMessageHandler(ctx, msgReq)
			if err != nil {
				logger.Err(err.Error())
				return
			}

			var actionResponse order.ActionResponse
			if err := ptypes.UnmarshalAny(response.Data, &actionResponse); err != nil {
				logger.Err("Could not unmarshal actionResponse from response anything field, request: %v, error %s", msgReq, err)
				return
			}
		}

		if total != totalCount {
			page = 1
			totalCount = total
		}
	}

	select {
	case <-ctx.Done():
		return
	default:
	}

	//var expiredOrderMap = make(map[uint64]map[uint64]*events.ISchedulerEvent, 64)

	// iterate through all documents
	//for cursor.Next(ctx) {
	//	var fetchData fetchItemData
	//	// decode the document
	//	if err := cursor.Decode(&fetchData); err != nil {
	//		logger.Err("scheduler.mongoAdapter.Aggregate failed, step: %s, action: %s,  error: %s",
	//			data.Step, data.Action, err)
	//		return
	//	}

	if scheduler.checkExpiredTime(&fetchData) {
		if sellerMap, isFindOrder := expiredOrderMap[fetchData.OrderId]; isFindOrder {
			if schedulerEvent, isFindSeller := sellerMap[fetchData.PId]; isFindSeller {
				schedulerEvent.ItemsId = append(schedulerEvent.ItemsId, fetchData.ItemId)
			} else {
				newEvent := &events.ISchedulerEvent{
					OrderId:    fetchData.OrderId,
					PId:        fetchData.PId,
					ItemsId:    nil,
					StateIndex: fetchData.StepIndex,
					Action:     fetchData.ActionHistory[0].ActionName,
				}

				newEvent.ItemsId = make([]uint64, 0, 16)
				newEvent.ItemsId = append(newEvent.ItemsId, fetchData.ItemId)
				expiredOrderMap[fetchData.OrderId][fetchData.PId] = newEvent
			}
		} else {
			newEvent := &events.ISchedulerEvent{
				OrderId:    fetchData.OrderId,
				PId:        fetchData.PId,
				ItemsId:    nil,
				StateIndex: fetchData.StepIndex,
				Action:     fetchData.ActionHistory[0].ActionName,
			}

			newEvent.ItemsId = make([]uint64, 0, 16)
			newEvent.ItemsId = append(newEvent.ItemsId, fetchData.ItemId)

			expiredOrderMap[fetchData.OrderId] = make(map[uint64]*events.ISchedulerEvent, 16)
			expiredOrderMap[fetchData.OrderId][fetchData.PId] = newEvent
		}
	}
	//}

	for _, sellerMap := range expiredOrderMap {
		for _, schedulerEvent := range sellerMap {
			scheduler.flowManager.SchedulerEvents(*schedulerEvent)
		}
	}
}

func (scheduler *SchedulerService) checkExpiredTime(fetchData *fetchItemData) bool {
	if fetchData.ActionHistory[0].ExpiredTime.Before(time.Now().UTC()) {
		logger.Audit("action expired, "+
			"orderId: %d, sid: %d, stepName: %s, stepIndex: %d, actionName: %s, expiredTime: %s ",
			fetchData.OrderId, fetchData.ItemId, fetchData.StepName, fetchData.StepIndex,
			fetchData.ActionHistory[0].ActionName, fetchData.ActionHistory[0].ExpiredTime)
		return true
	}

	return false
}

func (scheduler *SchedulerService) findAllWithPage(ctx context.Context, page, perPage int64) ([]*StateScheduler, int64, error) {

	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = scheduler.getTotalCount(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindAllWithPage Subpackage Failed")
	}

	if totalCount == 0 {
		return nil, 0, nil
	}

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, totalCount, errors.New("ErrorPageNotAvailable")
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, errors.New("ErrorTotalCountExceeded")
	}

	pipeline := []bson.M{
		{"$match": bson.M{"deletedAt": nil, "packages.subpackages.tracking.state.name": scheduler.state.StateName()}},
		{"$unwind": "$packages"},
		{"$unwind": "$packages.subpackages"},
		{"$match": bson.M{"packages.subpackages.tracking.state.name": scheduler.state.StateName()}},
		{"$project": bson.M{
			"_id":     0,
			"orderId": "$packages.subpackages.orderId",
			"pid":     "$packages.subpackages.pid",
			"sid":     "$packages.subpackages.sid",
			"sidx":    "$packages.subpackages.tracking.state.index",
			"data":    "$packages.subpackages.tracking.state.data.scheduler",
			"items":   "$packages.subpackages.items",
		}},
		{"$project": bson.M{
			"items.extended":   0,
			"items.invoice":    0,
			"items.attributes": 0,
			"items.reasons":    0,
			"items.returnable": 0,
			"items.image":      0,
			"items.category":   0,
			"items.guaranty":   0,
			"items.brand":      0,
			"items.title":      0,
			"items.sku":        0,
		}},
	}

	cursor, err := scheduler.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	stateSchedulers := make([]*StateScheduler, 0, perPage)

	for cursor.Next(ctx) {
		var stateScheduler StateScheduler
		if err := cursor.Decode(&stateScheduler); err != nil {
			return nil, 0, errors.Wrap(err, "cursor.Decode failed")
		}

		stateSchedulers = append(stateSchedulers, &stateScheduler)
	}

	return stateSchedulers, totalCount, nil
}

func (scheduler *SchedulerService) getTotalCount(ctx context.Context) (int64, error) {
	var total struct {
		Count int
	}

	totalCountPipeline := []bson.M{
		{"$match": bson.M{"deletedAt": nil, "packages.subpackages.tracking.state.name": scheduler.state.StateName()}},
		{"$unwind": "$packages"},
		{"$unwind": "$packages.subpackages"},
		{"$match": bson.M{"packages.subpackages.tracking.state.name": scheduler.state.StateName()}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0, "count": 1}},
	}

	cursor, err := scheduler.mongoAdapter.Aggregate(databaseName, collectionName, totalCountPipeline)
	if err != nil {
		return 0, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&total); err != nil {
			return 0, errors.Wrap(err, "cursor.Decode failed")
		}
	}

	return int64(total.Count), nil
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		logger.Err("closeCursor() failed, error: %s", err)
	}
}
