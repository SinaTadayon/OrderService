package main

import (
	"context"
	"fmt"
	_ "github.com/devfeel/mapper"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	"net/http"
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
	if os.Getenv("APP_MODE") == "dev" {
		app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfig("./testdata/.env")
	} else if os.Getenv("APP_MODE") == "docker" {
		app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfig("/app/.docker-env")
	} else {
		app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfig("")
	}

	app.Globals.ZapLogger = app.InitZap()
	app.Globals.Logger = logger.NewZapLogger(app.Globals.ZapLogger)

	if err != nil {
		logger.Err("LoadConfig of main init failed, error: %v ", err)
		os.Exit(1)
	}

	mongoDriver, err := app.SetupMongoDriver(*app.Globals.Config)
	if err != nil {
		logger.Err("main SetupMongoDriver failed, configs: %v, error: %v ", app.Globals.Config.Mongo, err)
	}

	if app.Globals.Config.App.ServiceMode == "server" {
		logger.Audit("Order Service Run in Server Mode . . . ")

		app.Globals.FlowManagerConfig = make(map[string]interface{}, 32)

		if app.Globals.Config.App.OrderPaymentCallbackUrlSuccess == "" ||
			app.Globals.Config.App.OrderPaymentCallbackUrlFail == "" ||
			app.Globals.Config.App.OrderPaymentCallbackUrlAsanpardakhtSuccess == "" ||
			app.Globals.Config.App.OrderPaymentCallbackUrlAsanpardakhtFail == "" {
			logger.Err("OrderPaymentCallbackUrlSuccess/Fail or OrderPaymentCallbackUrlAsanpardakhtSuccess/Fail empty")
			os.Exit(1)
		}

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
				logger.Err("SchedulerSellerReactionTime invalid, SchedulerSellerReactionTime: %s, error: %v ", app.Globals.Config.App.SchedulerSellerReactionTime, err)
				os.Exit(1)
			}
			app.Globals.FlowManagerConfig[app.FlowManagerSchedulerSellerReactionTimeConfig] = temp
		}

		if app.Globals.Config.App.SchedulerPaymentPendingState == "" {
			logger.Err("SchedulerPaymentPendingState is empty")
			os.Exit(1)
		} else {
			temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerPaymentPendingState)
			if err != nil {
				logger.Err("SchedulerPaymentPendingState invalid, SchedulerPaymentPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerApprovalPendingState, err)
				os.Exit(1)
			}
			app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig] = temp
		}

		if app.Globals.Config.App.SchedulerRetryPaymentPendingState == "" {
			logger.Err("SchedulerPaymentRetryPendingState is empty")
			os.Exit(1)
		} else {
			temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerRetryPaymentPendingState)
			if err != nil {
				logger.Err("SchedulerPaymentPendingState invalid, SchedulerRetryPaymentPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerApprovalPendingState, err)
				os.Exit(1)
			}
			app.Globals.FlowManagerConfig[app.FlowManagerSchedulerRetryPaymentPendingStateConfig] = int32(temp)
		}

		if app.Globals.Config.App.SchedulerApprovalPendingState == "" {
			logger.Err("SchedulerApprovalPendingState is empty")
			os.Exit(1)
		} else {
			temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerApprovalPendingState)
			if err != nil {
				logger.Err("SchedulerApprovalPendingState invalid, SchedulerApprovalPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerApprovalPendingState, err)
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
				logger.Err("SchedulerApprovalPendingState invalid, SchedulerApprovalPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerApprovalPendingState, err)
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
				logger.Err("SchedulerShipmentPendingState invalid, SchedulerShipmentPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerShipmentPendingState, err)
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
				logger.Err("SchedulerShippedState invalid, SchedulerShippedState: %s, error: %v ", app.Globals.Config.App.SchedulerShippedState, err)
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
				logger.Err("SchedulerDeliveryPendingState invalid, SchedulerDeliveryPendingState: %v, error: %s ", app.Globals.Config.App.SchedulerDeliveryPendingState, err)
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
				logger.Err("SchedulerNotifyDeliveryPendingState invalid, SchedulerNotifyDeliveryPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerNotifyDeliveryPendingState, err)
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
				logger.Err("SchedulerDeliveredState invalid, SchedulerDeliveredState: %s, error: %v ", app.Globals.Config.App.SchedulerDeliveredState, err)
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
				logger.Err("SchedulerReturnShippedState invalid, SchedulerReturnShippedState: %s, error: %v ", app.Globals.Config.App.SchedulerDeliveredState, err)
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
				logger.Err("SchedulerReturnRequestPendingState invalid, SchedulerReturnRequestPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerReturnRequestPendingState, err)
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
				logger.Err("SchedulerReturnShipmentPendingState invalid, SchedulerReturnShipmentPendingState: %s, error: %v ", app.Globals.Config.App.SchedulerReturnShipmentPendingState, err)
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
				logger.Err("SchedulerReturnDeliveredState invalid, SchedulerReturnDeliveredState: %s, error: %v ", app.Globals.Config.App.SchedulerReturnDeliveredState, err)
				os.Exit(1)
			}
			app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnDeliveredStateConfig] = temp
		}

		app.Globals.OrderRepository = order_repository.NewOrderRepository(mongoDriver)
		app.Globals.PkgItemRepository = pkg_repository.NewPkgItemRepository(mongoDriver)
		app.Globals.SubPkgRepository = subpkg_repository.NewSubPkgRepository(mongoDriver)

		MainApp.flowManager, err = domain.NewFlowManager()
		if err != nil {
			logger.Err("flowManager creation failed, %v ", err)
			os.Exit(1)
		}

		MainApp.grpcServer = grpc_server.NewServer(app.Globals.Config.GRPCServer.Address, uint16(app.Globals.Config.GRPCServer.Port), MainApp.flowManager)
		app.Globals.Converter = converter.NewConverter()

		//app.Globals.StockService = stock_service.NewStockService(app.Globals.Config.StockService.Address, app.Globals.Config.StockService.Port)
		//app.Globals.OrderPayment = payment_service.NewPaymentService(app.Globals.Config.PaymentGatewayService.Address,
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

		// listen and serve prometheus scraper
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			promPort := fmt.Sprintf(":%d", app.Globals.Config.App.PrometheusPort)
			logger.Audit("prometheus port: %s", promPort)
			e := http.ListenAndServe(promPort, nil)
			if e != nil {
				logger.Err("error listening for prometheus: %v", e)
			}
		}()

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

		if app.Globals.Config.App.SchedulerTimeUint != "hour" &&
			app.Globals.Config.App.SchedulerTimeUint != "minute" {
			logger.Err("main() => SchedulerTimeUint env is invalid, %s ", app.Globals.Config.App.SchedulerTimeUint)
			os.Exit(1)
		}

		var stateList = make([]scheduler_service.StateConfig, 0, 16)
		for _, stateConfig := range strings.Split(app.Globals.Config.App.SchedulerStates, ";") {
			values := strings.Split(stateConfig, ":")
			if len(values) == 1 {
				state := states.FromString(values[0])
				if state != nil {
					config := scheduler_service.StateConfig{
						State:            state,
						ScheduleInterval: 0,
					}
					stateList = append(stateList, config)
				} else {
					logger.Err("main() => state string SchedulerStates env is invalid, state: %s", stateConfig)
					os.Exit(1)
				}
			} else if len(values) == 2 {
				state := states.FromString(values[0])
				temp, err := strconv.Atoi(values[1])
				var scheduleInterval time.Duration
				if err != nil {
					logger.Err("main() => scheduleInterval of SchedulerStates env is invalid, state: %s, err: %v", stateConfig, err)
					os.Exit(1)
				}
				if app.Globals.Config.App.SchedulerTimeUint == "hour" {
					scheduleInterval = time.Duration(temp) * time.Hour
				} else {
					scheduleInterval = time.Duration(temp) * time.Minute
				}
				if state != nil {
					config := scheduler_service.StateConfig{
						State:            state,
						ScheduleInterval: scheduleInterval,
					}
					stateList = append(stateList, config)
				} else {
					logger.Err("main() => state string SchedulerStates env is invalid, state: %s", stateConfig)
					os.Exit(1)
				}
			} else {
				logger.Err("main() => state string SchedulerStates env is invalid, state: %s", stateConfig)
				os.Exit(1)
			}
		}

		var schedulerInterval time.Duration
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerInterval)
		if err != nil {
			logger.Err("main() => SchedulerInterval env is invalid, %s ", app.Globals.Config.App.SchedulerInterval)
			os.Exit(1)
		} else {
			if temp <= 0 {
				logger.Err("main() => SchedulerInterval env is invalid, %s ", app.Globals.Config.App.SchedulerInterval)
				os.Exit(1)
			}
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
			if temp <= 0 {
				logger.Err("main() => SchedulerParentWorkerTimeout env is invalid, %s ", app.Globals.Config.App.SchedulerParentWorkerTimeout)
				os.Exit(1)
			}
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
			if temp <= 0 {
				logger.Err("main() => SchedulerWorkerTimeout env is invalid, %s ", app.Globals.Config.App.SchedulerWorkerTimeout)
				os.Exit(1)
			}
			if app.Globals.Config.App.SchedulerTimeUint == "hour" {
				schedulerWorkerTimeout = time.Duration(temp) * time.Hour
			} else {
				schedulerWorkerTimeout = time.Duration(temp) * time.Minute
			}
		}

		logger.Audit("Order Service Run in Schedulers Mode . . . ")

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
