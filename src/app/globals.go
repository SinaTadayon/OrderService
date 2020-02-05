package app

import (
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/subpackage"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
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
	FlowManagerSchedulerStateTimeUintConfig              string = "SchedulerStateTimeUint"
	FlowManagerSchedulerSellerReactionTimeConfig         string = "SchedulerSellerReactionTime"
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
	MongoDriver       *mongoadapter.Mongo
	Config            *configs.Config
	SMSTemplate       *configs.SmsTemplate
	ZapLogger         *zap.Logger
	Logger            logger.Logger
	OrderRepository   order_repository.IOrderRepository
	PkgItemRepository pkg_repository.IPkgItemRepository
	SubPkgRepository  subpkg_repository.ISubpackageRepository
	Converter         converter.IConverter
	StockService      stock_service.IStockService
	PaymentService    payment_service.IPaymentService
	NotifyService     notify_service.INotificationService
	UserService       user_service.IUserService
	VoucherService    voucher_service.IVoucherService
	FlowManagerConfig map[string]interface{}
}

func SetupMongoDriver(config configs.Config) (*mongoadapter.Mongo, error) {
	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     MainApp.Config.Mongo.Pass,
		ConnTimeout:     time.Duration(config.Mongo.ConnectionTimeout) * time.Second,
		ReadTimeout:     time.Duration(config.Mongo.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(config.Mongo.WriteTimeout) * time.Second,
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime) * time.Second,
		MaxPoolSize:     uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(config.Mongo.MinPoolSize),
		WriteConcernW:   config.Mongo.WriteConcernW,
		WriteConcernJ:   config.Mongo.WriteConcernJ,
		RetryWrites:     config.Mongo.RetryWrite,
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		Globals.Logger.Error("mongoadapter.NewMongo failed",
			"fn", "SetupMongoDriver",
			"Mongo", err)
		return nil, errors.Wrap(err, "mongoadapter.NewMongo init failed")
	}

	_, err = mongoDriver.AddUniqueIndex(config.Mongo.Database, config.Mongo.Collection, "orderId")
	if err != nil {
		Globals.Logger.Error("create orderId index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "createdAt")
	if err != nil {
		Globals.Logger.Error("create createdAt index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "updatedAt")
	if err != nil {
		Globals.Logger.Error("create updatedAt index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "packages.createdAt")
	if err != nil {
		Globals.Logger.Error("create packages.createdAt index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "packages.updatedAt")
	if err != nil {
		Globals.Logger.Error("create packages.updatedAt index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "packages.pkgId")
	if err != nil {
		Globals.Logger.Error("create packages.pkgId index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "packages.subpackages.sid")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.sid index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "packages.subpackages.createdAt")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.createdAt index failed",
			"fn", "SetupMongoDriver",
			"error", err)
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(config.Mongo.Database, config.Mongo.Collection, "packages.subpackages.updatedAt")
	if err != nil {
		Globals.Logger.Error("create packages.subpackages.updatedAt index failed",
			"fn", "SetupMongoDriver",
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
