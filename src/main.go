package main

import (
	"context"
	_ "github.com/devfeel/mapper"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/subpackage"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	scheduler_service "gitlab.faza.io/order-project/order-service/infrastructure/services/scheduler"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
	grpc_server "gitlab.faza.io/order-project/order-service/server/grpc"
	"os"
)

var MainApp struct {
	flowManager      domain.IFlowManager
	grpcServer       grpc_server.Server
	schedulerService scheduler_service.ISchedulerService
}

// TODO Add worker scheduler and start from main
func main() {
	var err error
	if os.Getenv("APP_ENV") == "dev" {
		app.Globals.Config, err = configs.LoadConfig("./testdata/.env")
	} else {
		app.Globals.Config, err = configs.LoadConfig("")
	}
	if err != nil {
		logger.Err("LoadConfig of main init failed, error: %s ", err.Error())
		os.Exit(1)
	}

	mongoDriver, err := app.SetupMongoDriver(*app.Globals.Config)
	if err != nil {
		logger.Err("main SetupMongoDriver failed, configs: %v, error: %s ", app.Globals.Config.Mongo, err.Error())
	}

	app.Globals.OrderRepository = order_repository.NewOrderRepository(mongoDriver)
	app.Globals.PkgItemRepository = pkg_repository.NewPkgItemRepository(mongoDriver)
	app.Globals.SubPkgRepository = subpkg_repository.NewSubPkgRepository(mongoDriver)

	// TODO create item repository
	MainApp.flowManager, err = domain.NewFlowManager()
	if err != nil {
		logger.Err("flowManager creation failed, %s ", err.Error())
		os.Exit(1)
	}

	MainApp.grpcServer = grpc_server.NewServer(app.Globals.Config.GRPCServer.Address, uint16(app.Globals.Config.GRPCServer.Port), MainApp.flowManager)

	app.Globals.Converter = converter.NewConverter()

	//app.Globals.StockService = stock_service.NewStockService(app.Globals.Config.StockService.Address, app.Globals.Config.StockService.Port)
	//app.Globals.PaymentService = payment_service.NewPaymentService(app.Globals.Config.PaymentGatewayService.Address,
	//	app.Globals.Config.PaymentGatewayService.Port)

	//if app.Globals.Config.StockService.MockEnabled {
	//	app.Globals.StockService = stock_service.NewStockServiceMock()
	//} else {
	app.Globals.StockService = stock_service.NewStockService(app.Globals.Config.StockService.Address, app.Globals.Config.StockService.Port)
	//}

	if app.Globals.Config.PaymentGatewayService.MockEnabled {
		app.Globals.PaymentService = payment_service.NewPaymentServiceMock()
	} else {
		app.Globals.PaymentService = payment_service.NewPaymentService(app.Globals.Config.PaymentGatewayService.Address, app.Globals.Config.PaymentGatewayService.Port)
	}

	//if app.Globals.Config.VoucherService.MockEnabled {
	//	app.Globals.VoucherService = voucher_service.NewVoucherServiceMock()
	//} else {
	app.Globals.VoucherService = voucher_service.NewVoucherService(app.Globals.Config.VoucherService.Address, app.Globals.Config.VoucherService.Port)
	//}

	app.Globals.NotifyService = notify_service.NewNotificationService(app.Globals.Config.NotifyService.Address, app.Globals.Config.NotifyService.Port)
	app.Globals.UserService = user_service.NewUserService(app.Globals.Config.UserService.Address, app.Globals.Config.UserService.Port)
	//MainApp.schedulerService = scheduler_service.NewScheduler(mongoDriver, MainApp.flowManager)

	//brokers = strings.Split(app.Globals.Config.Kafka.Brokers, ",")
	//if app.Globals.Config.MainApp.Port == "" {
	//	logger.Err("grpc PORT env not defined")
	//	//return errors.New("grpc PORT env not defined")
	//}

	//// store in mongo
	//mongoConf := &mongoadapter.MongoConfig{
	//	Host:     app.Globals.Config.Mongo.Host,
	//	Port:     app.Globals.Config.Mongo.Port,
	//	Username: app.Globals.Config.Mongo.User,
	//	//Password:     MainApp.Config.Mongo.Pass,
	//	ConnTimeout:  time.Duration(app.Globals.Config.Mongo.ConnectionTimeout),
	//	ReadTimeout:  time.Duration(app.Globals.Config.Mongo.ReadTimeout),
	//	WriteTimeout: time.Duration(app.Globals.Config.Mongo.WriteTimeout),
	//}
	//
	//MainApp.mongo, err = mongoadapter.NewMongo(mongoConf)
	//if err != nil {
	//	logger.Err("New Mongo: %v", err.Error())
	//}
	//_, err = MainApp.mongo.AddUniqueIndex(MongoDB, Orders, "orderId")
	//if err != nil {
	//	logger.Err(err.Error())
	//}

	//err = initTopics()
	//if err != nil {
	//	logger.Err(err.Error())
	//	os.Exit(1)
	//}

	if app.Globals.Config.App.ServiceMode == "server" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		scheduleDataList := []scheduler_service.ScheduleModel{
			{
				Step:   "20.Seller_Approval_Pending",
				Action: "ApprovalPending",
			},
			{
				Step:   "30.Shipment_Pending",
				Action: "SellerShipmentPending",
			},
			{
				Step:   "32.Shipment_Delivered",
				Action: "ShipmentDeliveredPending",
			},
		}

		if err := MainApp.schedulerService.Scheduler(ctx, scheduleDataList); err != nil {
			logger.Err("SchedulerService.Scheduler failed, error: %s", err)
			return
		}
		MainApp.grpcServer.Start()
	}

	//switch MainApp.Config.Kafka.ConsumerTopic {
	//case "payment-pending":
	//	logger.Audit("starting grpc ...")
	//	server.startGrpc()
	//case "payment-success":
	//	logger.Audit("starting " + MainApp.Config.Kafka.ConsumerTopic)
	//	startPaymentSuccess(MainApp.Config.Kafka.Version, MainApp.Config.Kafka.ConsumerTopic)
	//case "seller-approval-pending":
	//	logger.Audit("starting " + MainApp.Config.Kafka.ConsumerTopic)
	//	server.startGrpc()
	//default:
	//	logger.Err("consumer topic env is wrong:" + MainApp.Config.Kafka.ConsumerTopic)
	//}
}
