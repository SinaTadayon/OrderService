package app

import (
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
	worker_pool "gitlab.faza.io/order-project/order-service/infrastructure/workerPool"
	"go.uber.org/zap"
	"time"
)

type CtxKey int

const (
	CtxUserID CtxKey = iota
	CtxAuthToken
)

const (
	HourTimeUnit     string = "hour"
	MinuteTimeUnit   string = "minute"
	SecondTimeUnit   string = "second"
	DurationTimeUnit string = "duration"
)

const (
	FlowManagerSchedulerStateTimeUintConfig string = "SchedulerStateTimeUint"
	//FlowManagerSchedulerSellerReactionTimeConfig         string = "SchedulerSellerReactionTime"
	FlowManagerSchedulerPaymentPendingStateConfig        string = "SchedulerPaymentPendingState"
	FlowManagerSchedulerRetryPaymentPendingStateConfig   string = "SchedulerRetryPaymentPendingState"
	FlowManagerSchedulerApprovalPendingStateConfig       string = "SchedulerApprovalPendingState"
	FlowManagerSchedulerShipmentPendingStateConfig       string = "SchedulerShipmentPendingState"
	FlowManagerSchedulerShippedStateConfig               string = "SchedulerShippedState"
	FlowManagerSchedulerDeliveryPendingStateConfig       string = "SchedulerDeliveryPendingState"
	FlowManagerSchedulerNotifyDeliveryPendingStateConfig string = "SchedulerNotifyDeliveryPendingState"
	FlowManagerSchedulerDeliveredStateConfig             string = "SchedulerDeliveredState"
	FlowManagerSchedulerReturnShippedStateConfig         string = "SchedulerReturnShippedState"
	FlowManagerSchedulerReturnRequestPendingStateConfig  string = "SchedulerReturnRequestPendingState"
	FlowManagerSchedulerReturnShipmentPendingStateConfig string = "SchedulerReturnShipmentPendingState"
	FlowManagerSchedulerReturnDeliveredStateConfig       string = "SchedulerReturnDeliveredState"
)

var Globals struct {
	CmdMongoDriver    *mongoadapter.Mongo
	QueryMongoDriver  *mongoadapter.Mongo
	Config            *configs.Config
	SMSTemplate       *configs.SmsTemplate
	ZapLogger         *zap.Logger
	Logger            logger.Logger
	CQRSRepository    cqrs.ICQRSRepository
	Converter         converter.IConverter
	StockService      stock_service.IStockService
	PaymentService    payment_service.IPaymentService
	NotifyService     notify_service.INotificationService
	UserService       user_service.IUserService
	VoucherService    voucher_service.IVoucherService
	WorkerPool        worker_pool.IWorkerPool
	FlowManagerConfig map[string]interface{}
}

func SetupCmdMongoDriver(config configs.Config) (*mongoadapter.Mongo, error) {
	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Username: config.CmdMongo.User,
		//Password:     MainApp.Config.Mongo.Pass,
		ConnTimeout:            time.Duration(config.CmdMongo.ConnectionTimeout) * time.Second,
		ReadTimeout:            time.Duration(config.CmdMongo.ReadTimeout) * time.Second,
		WriteTimeout:           time.Duration(config.CmdMongo.WriteTimeout) * time.Second,
		MaxConnIdleTime:        time.Duration(config.CmdMongo.MaxConnIdleTime) * time.Second,
		HeartbeatInterval:      time.Duration(config.CmdMongo.HeartBeatInterval) * time.Second,
		ServerSelectionTimeout: time.Duration(config.CmdMongo.ServerSelectionTimeout) * time.Second,
		RetryConnect:           uint64(config.CmdMongo.RetryConnect),
		MaxPoolSize:            uint64(config.CmdMongo.MaxPoolSize),
		MinPoolSize:            uint64(config.CmdMongo.MinPoolSize),
		WriteConcernW:          config.CmdMongo.WriteConcernW,
		WriteConcernJ:          config.CmdMongo.WriteConcernJ,
		RetryWrites:            config.CmdMongo.RetryWrite,
		ReadConcern:            config.CmdMongo.ReadConcern,
		ReadPreference:         config.CmdMongo.ReadPreferred,
		ConnectUri:             config.CmdMongo.Uri,
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		Globals.Logger.Error("mongoadapter.NewMongo failed",
			"fn", "SetupCmdMongoDriver",
			"CmdMongo", err)
		return nil, errors.Wrap(err, "mongoadapter.NewMongo init failed")
	}

	_, err = mongoDriver.AddUniqueIndex(config.CmdMongo.Database, config.CmdMongo.Collection, "orderId")
	if err != nil {
		Globals.Logger.Error("create orderId index failed",
			"fn", "SetupCmdMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.subpackages.sid")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.sid index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	return mongoDriver, nil
}

func SetupQueryMongoDriver(config configs.Config) (*mongoadapter.Mongo, error) {
	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Username: config.QueryMongo.User,
		//Password:     MainApp.Config.Mongo.Pass,
		ConnTimeout:            time.Duration(config.QueryMongo.ConnectionTimeout) * time.Second,
		ReadTimeout:            time.Duration(config.QueryMongo.ReadTimeout) * time.Second,
		WriteTimeout:           time.Duration(config.QueryMongo.WriteTimeout) * time.Second,
		MaxConnIdleTime:        time.Duration(config.QueryMongo.MaxConnIdleTime) * time.Second,
		HeartbeatInterval:      time.Duration(config.QueryMongo.HeartBeatInterval) * time.Second,
		ServerSelectionTimeout: time.Duration(config.QueryMongo.ServerSelectionTimeout) * time.Second,
		RetryConnect:           uint64(config.QueryMongo.RetryConnect),
		MaxPoolSize:            uint64(config.QueryMongo.MaxPoolSize),
		MinPoolSize:            uint64(config.QueryMongo.MinPoolSize),
		WriteConcernW:          config.QueryMongo.WriteConcernW,
		WriteConcernJ:          config.QueryMongo.WriteConcernJ,
		RetryWrites:            config.QueryMongo.RetryWrite,
		ReadConcern:            config.QueryMongo.ReadConcern,
		ReadPreference:         config.QueryMongo.ReadPreferred,
		ConnectUri:             config.QueryMongo.Uri,
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		Globals.Logger.Error("mongoadapter.NewMongo failed",
			"fn", "SetupQueryMongoDriver",
			"QueryMongo", err)
		return nil, errors.Wrap(err, "mongoadapter.NewMongo init failed")
	}

	_, err = mongoDriver.AddUniqueIndex(config.QueryMongo.Database, config.QueryMongo.Collection, "orderId")
	if err != nil {
		Globals.Logger.Error("create orderId index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "createdAt")
	if err != nil {
		Globals.Logger.Error("create createdAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "updatedAt")
	if err != nil {
		Globals.Logger.Error("create updatedAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "deletedAt")
	if err != nil {
		Globals.Logger.Error("create deletedAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "buyerInfo.mobile")
	if err != nil {
		Globals.Logger.Error("create buyerInfo.mobile index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "buyerInfo.buyerId")
	if err != nil {
		Globals.Logger.Error("create buyerInfo.buyerId index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.pid")
	if err != nil {
		Globals.Logger.Error("create packages.pid index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.createdAt")
	if err != nil {
		Globals.Logger.Error("create packages.createdAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.updatedAt")
	if err != nil {
		Globals.Logger.Error("create packages.updatedAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.deletedAt")
	if err != nil {
		Globals.Logger.Error("create packages.deletedAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.subpackages.sid")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.sid index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.subpackages.createdAt")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.createdAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.subpackages.updatedAt")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.updatedAt index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.subpackages.status")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.status index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.QueryMongo.Database, config.QueryMongo.Collection, "packages.subpackages.tracking.history.name")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.tracking.history.name index failed",
			"fn", "SetupQueryMongoDriver",
			"error", err)
		return nil, err
	}

	return mongoDriver, nil
}

func InitZap() (zapLogger *zap.Logger) {
	conf := zap.NewProductionConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	conf.DisableCaller = true
	conf.DisableStacktrace = true
	zapLogger, e := conf.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	// zapLogger, e := conf.Build()
	// zapLogger, e := zap.NewProduction(zap.AddCallerSkip(3))
	if e != nil {
		panic(e)
	}
	return
}
