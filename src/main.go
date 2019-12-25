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
	"gitlab.faza.io/order-project/order-service/domain/states"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	scheduler_service "gitlab.faza.io/order-project/order-service/infrastructure/services/scheduler"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
	grpc_server "gitlab.faza.io/order-project/order-service/server/grpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var MainApp struct {
	flowManager      domain.IFlowManager
	grpcServer       grpc_server.Server
	schedulerService scheduler_service.ISchedulerService
}

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

	app.Globals.FlowManagerConfig = make(map[string]interface{}, 32)

	if app.Globals.Config.App.SchedulerStateTimeUint == "" {
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig] = app.HourTimeUnit
	} else {
		if app.Globals.Config.App.SchedulerStateTimeUint != "hour" &&
			app.Globals.Config.App.SchedulerStateTimeUint != "minute" {
			logger.Err("SchedulerStateTimeUint invalid, SchedulerStateTimeUint: %s", app.Globals.Config.App.SchedulerApprovalPendingState)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig] = app.Globals.Config.App.SchedulerStateTimeUint
	}

	if app.Globals.Config.App.SchedulerSellerReactionTime != "" {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerSellerReactionTime)
		if err != nil {
			logger.Err("SchedulerSellerReactionTime invalid, SchedulerSellerReactionTime: %s, error: %s ", app.Globals.Config.App.SchedulerSellerReactionTime, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerSellerReactionTimeConfig] = temp
	}

	if app.Globals.Config.App.SchedulerApprovalPendingState == "" {
		logger.Err("SchedulerApprovalPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerApprovalPendingState)
		if err != nil {
			logger.Err("SchedulerApprovalPendingState invalid, SchedulerApprovalPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerApprovalPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerApprovalPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerApprovalPendingState == "" {
		logger.Err("SchedulerApprovalPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerApprovalPendingState)
		if err != nil {
			logger.Err("SchedulerApprovalPendingState invalid, SchedulerApprovalPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerApprovalPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerApprovalPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerShipmentPendingState == "" {
		logger.Err("SchedulerShipmentPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerShipmentPendingState)
		if err != nil {
			logger.Err("SchedulerShipmentPendingState invalid, SchedulerShipmentPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerShipmentPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerShipmentPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerShippedState == "" {
		logger.Err("SchedulerShippedState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerShippedState)
		if err != nil {
			logger.Err("SchedulerShippedState invalid, SchedulerShippedState: %s, error: %s ", app.Globals.Config.App.SchedulerShippedState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerShippedStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerDeliveryPendingState == "" {
		logger.Err("SchedulerDeliveryPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerDeliveryPendingState)
		if err != nil {
			logger.Err("SchedulerDeliveryPendingState invalid, SchedulerDeliveryPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerDeliveryPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveryPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerNotifyDeliveryPendingState == "" {
		logger.Err("SchedulerNotifyDeliveryPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerNotifyDeliveryPendingState)
		if err != nil {
			logger.Err("SchedulerNotifyDeliveryPendingState invalid, SchedulerNotifyDeliveryPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerNotifyDeliveryPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerNotifyDeliveryPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerDeliveredState == "" {
		logger.Err("SchedulerDeliveredState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerDeliveredState)
		if err != nil {
			logger.Err("SchedulerDeliveredState invalid, SchedulerDeliveredState: %s, error: %s ", app.Globals.Config.App.SchedulerDeliveredState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveredStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnShippedState == "" {
		logger.Err("SchedulerReturnShippedState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnShippedState)
		if err != nil {
			logger.Err("SchedulerReturnShippedState invalid, SchedulerReturnShippedState: %s, error: %s ", app.Globals.Config.App.SchedulerDeliveredState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnShippedStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnRequestPendingState == "" {
		logger.Err("SchedulerReturnRequestPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnRequestPendingState)
		if err != nil {
			logger.Err("SchedulerReturnRequestPendingState invalid, SchedulerReturnRequestPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerReturnRequestPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnRequestPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnShipmentPendingState == "" {
		logger.Err("SchedulerReturnShipmentPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnShipmentPendingState)
		if err != nil {
			logger.Err("SchedulerReturnShipmentPendingState invalid, SchedulerReturnShipmentPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerReturnShipmentPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnShipmentPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnDeliveredState == "" {
		logger.Err("SchedulerReturnDeliveredState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnDeliveredState)
		if err != nil {
			logger.Err("SchedulerReturnDeliveredState invalid, SchedulerReturnDeliveredState: %s, error: %s ", app.Globals.Config.App.SchedulerReturnDeliveredState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnDeliveredStateConfig] = temp
	}

	mongoDriver, err := app.SetupMongoDriver(*app.Globals.Config)
	if err != nil {
		logger.Err("main SetupMongoDriver failed, configs: %v, error: %s ", app.Globals.Config.Mongo, err.Error())
	}

	app.Globals.OrderRepository = order_repository.NewOrderRepository(mongoDriver)
	app.Globals.PkgItemRepository = pkg_repository.NewPkgItemRepository(mongoDriver)
	app.Globals.SubPkgRepository = subpkg_repository.NewSubPkgRepository(mongoDriver)

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

	if app.Globals.Config.App.ServiceMode == "server" {
		MainApp.grpcServer.Start()
	} else if app.Globals.Config.App.ServiceMode == "scheduler" {
		if app.Globals.Config.App.SchedulerStates == "" {
			logger.Err("main() => SchedulerState env is empty ")
			os.Exit(1)
		}

		if app.Globals.Config.App.SchedulerInterval == "" ||
			app.Globals.Config.App.SchedulerParentWorkerTimeout == "" ||
			app.Globals.Config.App.SchedulerWorkerTimeout == "" ||
			app.Globals.Config.App.SchedulerTimeUint == "" {
			logger.Err("main() => SchedulerTimeUint or SchedulerInterval or SchedulerParentWorkerTimeout or SchedulerWorkerTimeout env is empty ")
			os.Exit(1)
		}

		var stateList = make([]states.IEnumState, 0, 16)
		for _, strState := range strings.Split(app.Globals.Config.App.SchedulerStates, ";") {
			state := states.FromString(strState)
			if state != nil {
				stateList = append(stateList, state)
			} else {
				logger.Err("main() => state string SchedulerStates env is invalid, state: %s", strState)
				os.Exit(1)
			}
		}

		if app.Globals.Config.App.SchedulerTimeUint != "hour" &&
			app.Globals.Config.App.SchedulerTimeUint != "minute" {
			logger.Err("main() => SchedulerTimeUint env is invalid, %s ", app.Globals.Config.App.SchedulerTimeUint)
			os.Exit(1)
		}

		var schedulerInterval time.Duration
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerInterval)
		if err != nil {
			logger.Err("main() => SchedulerInterval env is invalid, %s ", app.Globals.Config.App.SchedulerInterval)
			os.Exit(1)
		} else {
			if app.Globals.Config.App.SchedulerTimeUint == "hour" {
				schedulerInterval = time.Duration(temp) * time.Hour
			} else {
				schedulerInterval = time.Duration(temp) * time.Minute
			}
		}

		var schedulerStewardTimeout time.Duration
		temp, err = strconv.Atoi(app.Globals.Config.App.SchedulerParentWorkerTimeout)
		if err != nil {
			logger.Err("main() => SchedulerParentWorkerTimeout env is invalid, %s ", app.Globals.Config.App.SchedulerParentWorkerTimeout)
			os.Exit(1)
		} else {
			if app.Globals.Config.App.SchedulerTimeUint == "hour" {
				schedulerStewardTimeout = time.Duration(temp) * time.Hour
			} else {
				schedulerStewardTimeout = time.Duration(temp) * time.Minute
			}
		}

		var schedulerWorkerTimeout time.Duration
		temp, err = strconv.Atoi(app.Globals.Config.App.SchedulerWorkerTimeout)
		if err != nil {
			logger.Err("main() => SchedulerWorkerTimeout env is invalid, %s ", app.Globals.Config.App.SchedulerWorkerTimeout)
			os.Exit(1)
		} else {
			if app.Globals.Config.App.SchedulerTimeUint == "hour" {
				schedulerStewardTimeout = time.Duration(temp) * time.Hour
			} else {
				schedulerStewardTimeout = time.Duration(temp) * time.Minute
			}
		}

		schedulerService := scheduler_service.NewScheduler(mongoDriver,
			app.Globals.Config.GRPCServer.Address,
			app.Globals.Config.GRPCServer.Port,
			schedulerInterval,
			schedulerStewardTimeout,
			schedulerWorkerTimeout,
			stateList...)
		schedulerService.Scheduler(context.Background())
	}
}
