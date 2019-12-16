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
	"time"
)

type CtxKey int

const (
	CtxUserID CtxKey = iota
	CtxAuthToken
)

const (
	DatabaseName   string = "orderService"
	CollectionName string = "orders"
)

var Globals struct {
	MongoDriver       *mongoadapter.Mongo
	Config            *configs.Config
	OrderRepository   order_repository.IOrderRepository
	PkgItemRepository pkg_repository.IPkgItemRepository
	SubPkgRepository  subpkg_repository.ISubpackageRepository
	Converter         converter.IConverter
	StockService      stock_service.IStockService
	PaymentService    payment_service.IPaymentService
	NotifyService     notify_service.INotificationService
	UserService       user_service.IUserService
	VoucherService    voucher_service.IVoucherService
}

func SetupMongoDriver(config configs.Config) (*mongoadapter.Mongo, error) {
	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     MainApp.Config.Mongo.Pass,
		ConnTimeout:     time.Duration(config.Mongo.ConnectionTimeout),
		ReadTimeout:     time.Duration(config.Mongo.ReadTimeout),
		WriteTimeout:    time.Duration(config.Mongo.WriteTimeout),
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime),
		MaxPoolSize:     uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(config.Mongo.MinPoolSize),
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("mongoadapter.NewMongo Mongo: %v", err.Error())
		return nil, errors.Wrap(err, "mongoadapter.NewMongo init failed")
	}

	_, err = mongoDriver.AddUniqueIndex(DatabaseName, CollectionName, "orderId")
	if err != nil {
		logger.Err("create orderId index failed, error: %s", err.Error())
		return nil, err
	}

	_, err = mongoDriver.AddTextV3Index(DatabaseName, CollectionName, "packages.pkgId")
	if err != nil {
		logger.Err("create packages.pkgId index failed, error: %s", err.Error())
		return nil, err
	}

	_, err = mongoDriver.AddUniqueIndex(DatabaseName, CollectionName, "packages.subpackages.sid")
	if err != nil {
		logger.Err("create packages.subpackages.sid index failed, error: %s", err.Error())
		return nil, err
	}

	return mongoDriver, nil
}