package scheduler_service

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	protoOrder "gitlab.faza.io/protos/order"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"sync"
	"time"
)

const (
	databaseName   string = "orderService"
	collectionName string = "orders"
)

type Order struct {
	Packages []Package
}

type Package struct {
	Subpackages []Subpackage
}

type Subpackage struct {
	SId       uint64
	Pid       uint64
	OrderId   uint64
	Sidx      int32
	Items     []Item
	Scheduler []Scheduler
}

type Scheduler struct {
	Name     string                 `bson:"name"`
	Group    string                 `bson:"group"`
	Action   string                 `bson:"action"`
	Index    int32                  `bson:"index"`
	Retry    int32                  `bson:"retry"`
	Cron     string                 `bson:"cron"`
	Start    *time.Time             `bson:"start"`
	End      *time.Time             `bson:"end"`
	Type     string                 `bson:"type"`
	Mode     string                 `bson:"mode"`
	Policy   interface{}            `bson:"policy"`
	Enabled  bool                   `bson:"enabled"`
	Data     interface{}            `bson:"data"`
	Extended map[string]interface{} `bson:"ext"`
}

type Item struct {
	InventoryId string
	Quantity    int32
}

type startWardFn func(ctx context.Context, pulseInterval time.Duration, scheduleInterval time.Duration, state states.IEnumState) (heartbeat <-chan interface{})
type startStewardFn func(ctx context.Context, pulseInterval time.Duration) (heartbeat <-chan interface{})

type SchedulerService struct {
	mongoAdapter            *mongoadapter.Mongo
	orderClient             protoOrder.OrderServiceClient
	grpcConnection          *grpc.ClientConn
	serverAddress           string
	serverPort              int
	states                  []StateConfig
	schedulerInterval       time.Duration
	schedulerStewardTimeout time.Duration
	schedulerWorkerTimeout  time.Duration
	waitGroup               sync.WaitGroup
}

func NewScheduler(mongoAdapter *mongoadapter.Mongo, address string, port int,
	schedulerInterval time.Duration, schedulerStewardTimeout time.Duration, schedulerWorkerTimeout time.Duration,
	states ...StateConfig) *SchedulerService {
	for i := 0; i < len(states); i++ {
		if states[i].ScheduleInterval == 0 {
			states[i].ScheduleInterval = schedulerInterval
		}
	}
	return &SchedulerService{mongoAdapter: mongoAdapter, serverAddress: address, serverPort: port,
		schedulerInterval: schedulerInterval, schedulerStewardTimeout: schedulerStewardTimeout, schedulerWorkerTimeout: schedulerWorkerTimeout,
		states: states}
}

func (scheduler *SchedulerService) ConnectToOrderService() error {
	if scheduler.grpcConnection == nil || scheduler.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		scheduler.grpcConnection, err = grpc.DialContext(ctx, scheduler.serverAddress+":"+fmt.Sprint(scheduler.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			logger.Err("GRPC connect dial to order service failed, err: %s", err.Error())
			return err
		}
		scheduler.orderClient = protoOrder.NewOrderServiceClient(scheduler.grpcConnection)
	}
	return nil
}

func (scheduler *SchedulerService) Scheduler(ctx context.Context) {

	for _, state := range scheduler.states {
		scheduler.waitGroup.Add(1)
		go scheduler.scheduleProcess(ctx, state)
	}
	scheduler.waitGroup.Wait()
}

func (scheduler *SchedulerService) scheduleProcess(ctx context.Context, config StateConfig) {

	stewardCtx, stewardCtxCancel := context.WithCancel(context.Background())
	stewardWorkerFn := scheduler.stewardFn(utils.ORContext(ctx, stewardCtx), scheduler.schedulerWorkerTimeout, config.ScheduleInterval, config.State, scheduler.worker)
	heartbeat := stewardWorkerFn(ctx, scheduler.schedulerStewardTimeout)
	stewardTimer := time.NewTimer(scheduler.schedulerStewardTimeout * 2)

	for {
		select {
		case <-ctx.Done():
			logger.Audit("scheduleProcess() => stewardWorkerFn goroutine context down!, state: %s", config.State.StateName())
			stewardTimer.Stop()
			scheduler.waitGroup.Done()
			return
		case _, ok := <-heartbeat:
			if ok == false {
				logger.Audit("scheduleProcess() => heartbeat of stewardWorkerFn closed, state: %s", config.State.StateName())
				stewardCtxCancel()
				stewardCtx, stewardCtxCancel = context.WithCancel(context.Background())
				stewardWorkerFn := scheduler.stewardFn(utils.ORContext(ctx, stewardCtx), scheduler.schedulerWorkerTimeout, config.ScheduleInterval, config.State, scheduler.worker)
				heartbeat = stewardWorkerFn(ctx, scheduler.schedulerStewardTimeout)
				stewardTimer.Reset(scheduler.schedulerStewardTimeout * 2)
			} else {
				//logger.Audit("scheduleProcess() => heartbeat stewardWorkerFn , state: %s", state.StateName())
				stewardTimer.Stop()
				stewardTimer.Reset(scheduler.schedulerStewardTimeout * 2)
			}

		case <-stewardTimer.C:
			logger.Audit("scheduleProcess() => stewardWorkerFn goroutine is not healthy!, state: %s", config.State.StateName())
			stewardCtxCancel()
			stewardCtx, stewardCtxCancel = context.WithCancel(context.Background())
			stewardWorkerFn := scheduler.stewardFn(utils.ORContext(ctx, stewardCtx), scheduler.schedulerWorkerTimeout, config.ScheduleInterval, config.State, scheduler.worker)
			heartbeat = stewardWorkerFn(ctx, scheduler.schedulerStewardTimeout)
			stewardTimer.Reset(scheduler.schedulerStewardTimeout * 2)
		}
	}
}

func (scheduler *SchedulerService) stewardFn(ctx context.Context, wardPulseInterval time.Duration, wardScheduleInterval time.Duration, state states.IEnumState, startWorker startWardFn) startStewardFn {
	return func(ctx context.Context, stewardPulse time.Duration) <-chan interface{} {
		heartbeat := make(chan interface{}, 1)
		go func() {
			defer close(heartbeat)

			var wardCtx context.Context
			var wardCtxCancel context.CancelFunc
			var wardHeartbeat <-chan interface{}
			startWard := func() {
				wardCtx, wardCtxCancel = context.WithCancel(context.Background())
				wardHeartbeat = startWorker(utils.ORContext(ctx, wardCtx), wardPulseInterval, wardScheduleInterval, state)
			}
			startWard()
			pulseTimer := time.NewTimer(stewardPulse)
			wardTimer := time.NewTimer(wardPulseInterval * 2)

			for {
				select {
				case <-pulseTimer.C:
					select {
					case heartbeat <- struct{}{}:
					default:
					}
					pulseTimer.Reset(stewardPulse)

				case <-wardHeartbeat:
					//logger.Audit("wardHeartbeat , state: %s", state.StateName())
					wardTimer.Stop()
					wardTimer.Reset(wardPulseInterval * 2)

				case <-wardTimer.C:
					logger.Err("stewardFn() => ward unhealthy; restarting ward, state: %s", state.StateName())
					wardCtxCancel()
					startWard()
					wardTimer.Reset(wardPulseInterval * 2)

				case <-ctx.Done():
					wardTimer.Stop()
					logger.Audit("stewardFn() => context done . . ., state: %s, cause: %s", state.StateName(), ctx.Err())
					return
				}
			}
		}()
		return heartbeat
	}
}

func (scheduler *SchedulerService) worker(ctx context.Context, pulseInterval time.Duration,
	scheduleInterval time.Duration, state states.IEnumState) <-chan interface{} {

	logger.Audit("worker() => pulse: %d, schedule: %d , state: %s,", pulseInterval, scheduleInterval, state.StateName())
	var heartbeat = make(chan interface{}, 1)
	go func() {
		defer close(heartbeat)
		pulseTimer := time.NewTimer(pulseInterval)
		scheduleTimer := time.NewTimer(scheduleInterval)
		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default:
			}
		}

		for {
			select {
			case <-ctx.Done():
				pulseTimer.Stop()
				scheduleTimer.Stop()
				logger.Audit("worker() => context down, state: %s, cause: %s", state.StateName(), ctx.Err())
				return
			case <-pulseTimer.C:
				//logger.Audit("worker() => send pulse, state: %s", state.StateName())
				sendPulse()
				pulseTimer.Reset(pulseInterval)
			case <-scheduleTimer.C:
				//logger.Audit("worker() => schedule, state: %s", state.StateName())
				scheduler.doProcess(ctx, state)
				scheduleTimer.Reset(scheduleInterval)
			}
		}
	}()
	return heartbeat
}

func (scheduler *SchedulerService) doProcess(ctx context.Context, state states.IEnumState) {
	logger.Audit("doProcess() => state: %s", state.StateName())
	var perPage = int64(25)

	totalCount, err := scheduler.getTotalCount(ctx, state)
	if err != nil {
		logger.Err("scheduler worker doProcess() => getTotalCount failed, states: %s", state.StateName())
		return
	}

	for page := int64(1); page <= (totalCount/perPage)+1; page++ {
		orderList, _, err := scheduler.findAllWithPage(ctx, state, page, perPage)
		if err != nil {
			logger.Err("scheduler worker doProcess() => findAllWithPage failed, states: %s", state.StateName())
			return
		}

		if len(orderList) == 0 {
			return
		}

		var orderRequestList []*protoOrder.SchedulerActionRequest_Order = nil
		for i := 0; i < len(orderList); i++ {
			var packageList []*protoOrder.SchedulerActionRequest_Order_Package = nil
			var orderReq *protoOrder.SchedulerActionRequest_Order = nil
			for j := 0; j < len(orderList[i].Packages); j++ {
				var subpackageList []*protoOrder.SchedulerActionRequest_Order_Package_Subpackage = nil
				var pkg *protoOrder.SchedulerActionRequest_Order_Package = nil
				for k := 0; k < len(orderList[i].Packages[j].Subpackages); k++ {
					scheduler := scheduler.checkExpiredTime(orderList[i].Packages[j].Subpackages[k])
					if scheduler == nil {
						continue
					}

					if packageList == nil {
						packageList = make([]*protoOrder.SchedulerActionRequest_Order_Package, 0, len(orderList[i].Packages))
					}

					if orderReq == nil {
						orderReq = &protoOrder.SchedulerActionRequest_Order{
							OID:         orderList[i].Packages[j].Subpackages[k].OrderId,
							ActionType:  "",
							ActionState: scheduler.Action,
							StateIndex:  orderList[i].Packages[j].Subpackages[k].Sidx,
							Packages:    packageList,
						}
					}

					if subpackageList == nil {
						subpackageList = make([]*protoOrder.SchedulerActionRequest_Order_Package_Subpackage, 0, len(orderList[i].Packages[j].Subpackages))
					}

					if pkg == nil {
						pkg = &protoOrder.SchedulerActionRequest_Order_Package{
							PID:         orderList[i].Packages[j].Subpackages[k].Pid,
							Subpackages: subpackageList,
						}
					}

					subpkg := &protoOrder.SchedulerActionRequest_Order_Package_Subpackage{
						SID:   orderList[i].Packages[j].Subpackages[k].SId,
						Items: nil,
					}

					itemList := make([]*protoOrder.SchedulerActionRequest_Order_Package_Subpackage_Item, 0, len(orderList[i].Packages[j].Subpackages[k].Items))
					for z := 0; z < len(orderList[i].Packages[j].Subpackages[k].Items); z++ {
						subpackageItem := &protoOrder.SchedulerActionRequest_Order_Package_Subpackage_Item{
							InventoryId: orderList[i].Packages[j].Subpackages[k].Items[z].InventoryId,
							Quantity:    orderList[i].Packages[j].Subpackages[k].Items[z].Quantity,
						}

						itemList = append(itemList, subpackageItem)
					}

					subpkg.Items = itemList
					subpackageList = append(subpackageList, subpkg)
					pkg.Subpackages = subpackageList
				}
				if packageList != nil {
					packageList = append(packageList, pkg)
					if orderReq != nil {
						orderReq.Packages = packageList
					}
				}
			}

			if orderReq != nil {
				if orderRequestList == nil {
					orderRequestList = make([]*protoOrder.SchedulerActionRequest_Order, 0, len(orderList))
				}

				orderRequestList = append(orderRequestList, orderReq)
			}
		}

		if orderRequestList == nil {
			continue
		}

		request := &protoOrder.SchedulerActionRequest{
			Orders: orderRequestList,
		}

		serializedData, err := proto.Marshal(request)
		if err != nil {
			logger.Err("scheduler worker doProcess() => could not serialize protoOrder.SchedulerActionRequest, state: %s error:%s", state.StateName(), err)
			return
		}

		msgReq := &protoOrder.MessageRequest{
			Name:   "",
			Type:   "Action",
			ADT:    "List",
			Method: "",
			Time:   ptypes.TimestampNow(),
			Meta: &protoOrder.RequestMetadata{
				UID:     0,
				UTP:     "Schedulers",
				OID:     0,
				PID:     0,
				SIDs:    nil,
				Page:    0,
				PerPage: 0,
				//IpAddress: ipAddress,
				Action:  nil,
				Sorts:   nil,
				Filters: nil,
			},
			Data: &any.Any{
				TypeUrl: "baman.io/" + proto.MessageName(request),
				Value:   serializedData,
			},
		}

		err = scheduler.ConnectToOrderService()
		if err != nil {
			logger.Err("scheduler worker doProcess() => scheduler.ConnectToOrderService failed, error: %s", err)
			return
		}

		_, err = scheduler.orderClient.SchedulerMessageHandler(ctx, msgReq)
		if err != nil {
			logger.Err("scheduler worker doProcess() => scheduler.orderClient.SchedulerMessageHandler failed, error: %s", err)
			return
		}

		select {
		case <-ctx.Done():
			logger.Audit("scheduler worker doProcess() => context down, state: %s, cause: %s", state.StateName(), ctx.Err())
			return
		default:
		}
	}

	//if total != totalCount {
	//	page = 1
	//	totalCount = total
	//}
	//
}

func (scheduler *SchedulerService) checkExpiredTime(subpackage Subpackage) *Scheduler {
	if len(subpackage.Scheduler) == 1 {
		if subpackage.Scheduler[0].Data.(primitive.DateTime).Time().Before(time.Now().UTC()) && subpackage.Scheduler[0].Enabled {
			logger.Audit("action expired, "+
				"orderId: %d, sid: %d, stateName: %s, stateIndex: %d, actionName: %s, expiredTime: %s ",
				subpackage.OrderId, subpackage.SId, states.FromIndex(int32(subpackage.Sidx)).StateName(),
				subpackage.Sidx, subpackage.Scheduler[0].Action, subpackage.Scheduler[0].Data)
			return &subpackage.Scheduler[0]
		}
	} else {
		sortedScheduler := make([]*Scheduler, 0, len(subpackage.Scheduler))
		for i := 0; i < len(subpackage.Scheduler); i++ {
			sortedScheduler = append(sortedScheduler, &subpackage.Scheduler[i])
			for j := i + 1; j < len(subpackage.Scheduler); j++ {
				if sortedScheduler[i].Index > subpackage.Scheduler[j].Index {
					sortedScheduler[i] = &subpackage.Scheduler[j]
				}
			}
		}

		var sche *Scheduler = nil
		for i := 0; i < len(sortedScheduler); i++ {
			if sortedScheduler[i].Data.(primitive.DateTime).Time().Before(time.Now().UTC()) && sortedScheduler[i].Enabled {
				sche = sortedScheduler[i]
			}
		}

		return sche
	}

	return nil
}

func (scheduler *SchedulerService) findAllWithPage(ctx context.Context, state states.IEnumState, page, perPage int64) ([]*Order, int64, error) {

	if page <= 0 || perPage <= 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = scheduler.getTotalCount(ctx, state)
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
		{"$match": bson.M{"deletedAt": nil, "packages.subpackages.status": state.StateName()}},
		{"$skip": offset},
		{"$limit": perPage},
		{"$unwind": "$packages"},
		{"$unwind": "$packages.subpackages"},
		{"$match": bson.M{"packages.subpackages.status": state.StateName()}},
		{"$project": bson.M{
			"packages.subpackages.pid":                       1,
			"packages.subpackages.orderId":                   1,
			"packages.subpackages.sid":                       1,
			"packages.subpackages.items":                     1,
			"packages.subpackages.tracking.state.index":      1,
			"packages.subpackages.tracking.state.schedulers": 1,
		}},
		{"$replaceRoot": bson.M{"newRoot": "$packages.subpackages"}},
		{"$project": bson.M{
			"sidx":      "$tracking.state.index",
			"scheduler": "$tracking.state.schedulers",
			"items":     1,
			"sid":       1,
			"pid":       1,
			"orderId":   1,
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
		{"$group": bson.M{"_id": bson.M{"oid": "$orderId", "pid": "$pid"},
			"subpackages": bson.M{"$push": "$$ROOT"}},
		},
		{"$project": bson.M{"oid": "$_id.oid", "subpackages": 1, "_id": 0}},
		{"$group": bson.M{"_id": "$oid", "packages": bson.M{"$push": "$$ROOT"}}},
		{"$project": bson.M{"_id": 0, "packages.oid": 0}},
	}

	cursor, err := scheduler.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	orders := make([]*Order, 0, perPage)

	for cursor.Next(ctx) {
		var oneOrder Order
		if err := cursor.Decode(&oneOrder); err != nil {
			return nil, 0, errors.Wrap(err, "cursor.Decode failed")
		}

		orders = append(orders, &oneOrder)
	}

	return orders, totalCount, nil
}

func (scheduler *SchedulerService) getTotalCount(ctx context.Context, state states.IEnumState) (int64, error) {
	var total struct {
		Count int
	}

	totalCountPipeline := []bson.M{
		{"$match": bson.M{"deletedAt": nil, "packages.subpackages.status": state.StateName()}},
		{"$unwind": "$packages"},
		{"$unwind": "$packages.subpackages"},
		{"$match": bson.M{"packages.subpackages.status": state.StateName()}},
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
