package scheduler_service

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
	subpkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/subpackage"
	"gitlab.faza.io/order-project/order-service/domain/states"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	grpc_server "gitlab.faza.io/order-project/order-service/server/grpc"
	pb "gitlab.faza.io/protos/order"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

//var config *configs.Config
var schedulerService *SchedulerService

func TestMain(m *testing.M) {
	var err error
	var path string
	if os.Getenv("APP_MODE") == "dev" {
		path = "../../../testdata/.env"
	} else {
		path = ""
	}

	applog.GLog.ZapLogger = applog.InitZap()
	applog.GLog.Logger = logger.NewZapLogger(applog.GLog.ZapLogger)

	app.Globals.ZapLogger = applog.GLog.ZapLogger
	app.Globals.Logger = applog.GLog.Logger

	app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfigs(path, "../../../testdata/notification/sms/smsTemplate.txt")
	if err != nil {
		applog.GLog.Logger.Error("configs.LoadConfig failed",
			"error", err)
		os.Exit(1)
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     app.Globals.Config.Mongo.Host,
		Port:     app.Globals.Config.Mongo.Port,
		Username: app.Globals.Config.Mongo.User,
		//Password:     App.app.Globals.Config.Mongo.Pass,
		ConnTimeout:     time.Duration(app.Globals.Config.Mongo.ConnectionTimeout) * time.Second,
		ReadTimeout:     time.Duration(app.Globals.Config.Mongo.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(app.Globals.Config.Mongo.WriteTimeout) * time.Second,
		MaxConnIdleTime: time.Duration(app.Globals.Config.Mongo.MaxConnIdleTime) * time.Second,
		MaxPoolSize:     uint64(app.Globals.Config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(app.Globals.Config.Mongo.MinPoolSize),
		WriteConcernW:   app.Globals.Config.Mongo.WriteConcernW,
		WriteConcernJ:   app.Globals.Config.Mongo.WriteConcernJ,
		RetryWrites:     app.Globals.Config.Mongo.RetryWrite,
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		applog.GLog.Logger.Error("mongoadapter.NewMongo Mongo failed",
			"error", err.Error())
		os.Exit(1)
	}

	// TODO create item repository
	flowManager, err := domain.NewFlowManager()
	if err != nil {
		applog.GLog.Logger.Error("flowManager creation failed",
			"error", err.Error())
		os.Exit(1)
	}

	app.Globals.OrderRepository = order_repository.NewOrderRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)
	app.Globals.PkgItemRepository = pkg_repository.NewPkgItemRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)
	app.Globals.SubPkgRepository = subpkg_repository.NewSubPkgRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)
	app.Globals.StockService = stock_service.NewStockService(app.Globals.Config.StockService.Address, app.Globals.Config.StockService.Port, app.Globals.Config.StockService.Timeout)
	app.Globals.UserService = user_service.NewUserService(app.Globals.Config.UserService.Address, app.Globals.Config.UserService.Port, app.Globals.Config.UserService.Timeout)
	app.Globals.NotifyService = notify_service.NewNotificationService(app.Globals.Config.NotifyService.Address, app.Globals.Config.NotifyService.Port, app.Globals.Config.NotifyService.NotifySeller, app.Globals.Config.NotifyService.NotifyBuyer, app.Globals.Config.NotifyService.Timeout)
	app.Globals.Converter = converter.NewConverter()

	if app.Globals.Config.App.SchedulerInterval == "" ||
		app.Globals.Config.App.SchedulerParentWorkerTimeout == "" ||
		app.Globals.Config.App.SchedulerWorkerTimeout == "" {
		applog.GLog.Logger.Error("SchedulerInterval or SchedulerParentWorkerTimeout or SchedulerWorkerTimeout env is empty ")
		os.Exit(1)
	}

	if app.Globals.Config.App.SchedulerTimeUint != "hour" &&
		app.Globals.Config.App.SchedulerTimeUint != "minute" {
		applog.GLog.Logger.Error("SchedulerTimeUint env is invalid")
		os.Exit(1)
	}

	var stateList = make([]StateConfig, 0, 16)
	for _, stateConfig := range strings.Split(app.Globals.Config.App.SchedulerStates, ";") {
		values := strings.Split(stateConfig, ":")
		if len(values) == 1 {
			state := states.FromString(values[0])
			if state != nil {
				config := StateConfig{
					State:            state,
					ScheduleInterval: 0,
				}
				stateList = append(stateList, config)
			} else {
				applog.GLog.Logger.Error("state string SchedulerStates env is invalid")
				os.Exit(1)
			}
		} else if len(values) == 2 {
			state := states.FromString(values[0])
			temp, err := strconv.Atoi(values[1])
			var scheduleInterval time.Duration
			if err != nil {
				applog.GLog.Logger.Error("scheduleInterval of SchedulerStates env is invalid",
					"error", err)
				os.Exit(1)
			}
			if app.Globals.Config.App.SchedulerTimeUint == "hour" {
				scheduleInterval = time.Duration(temp) * time.Hour
			} else {
				scheduleInterval = time.Duration(temp) * time.Minute
			}
			if state != nil {
				config := StateConfig{
					State:            state,
					ScheduleInterval: scheduleInterval,
				}
				stateList = append(stateList, config)
			} else {
				applog.GLog.Logger.Error("state string SchedulerStates env is invalid")
				os.Exit(1)
			}
		} else {
			applog.GLog.Logger.Error("state string SchedulerStates env is invalid")
			os.Exit(1)
		}
	}

	app.Globals.FlowManagerConfig = make(map[string]interface{}, 32)

	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig] = app.DurationTimeUnit
	//app.Globals.FlowManagerConfig[app.FlowManagerSchedulerSellerReactionTimeConfig] = 2 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerApprovalPendingStateConfig] = 2 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerShipmentPendingStateConfig] = time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerShippedStateConfig] = 2 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveryPendingStateConfig] = 20 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerNotifyDeliveryPendingStateConfig] = 1 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveredStateConfig] = 2 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnShippedStateConfig] = 2 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnRequestPendingStateConfig] = 2 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnShipmentPendingStateConfig] = 2 * time.Second
	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnDeliveredStateConfig] = 2 * time.Second

	schedulerService = NewScheduler(mongoDriver,
		app.Globals.Config.Mongo.Database,
		app.Globals.Config.Mongo.Collection,
		app.Globals.Config.GRPCServer.Address,
		app.Globals.Config.GRPCServer.Port,
		10*time.Second,
		20*time.Second,
		5*time.Second,
		stateList...)

	grpcServer := grpc_server.NewServer(app.Globals.Config.GRPCServer.Address, uint16(app.Globals.Config.GRPCServer.Port), flowManager)

	if !checkTcpPort(app.Globals.Config.GRPCServer.Address, strconv.Itoa(app.Globals.Config.GRPCServer.Port)) {
		applog.GLog.Logger.Debug("Start GRPC Server for testing . . . ")
		go grpcServer.StartTest()
	}

	// Running Tests
	code := m.Run()
	removeCollection()
	os.Exit(code)
}

func checkTcpPort(host string, port string) bool {

	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		//	fmt.Println("Connecting error:", err)
		return false
	}
	if conn != nil {
		defer func() {
			if err := conn.Close(); err != nil {
			}
		}()
		//	fmt.Println("Opened", net.JoinHostPort(host, port))
	}
	return true
}

func createAuthenticatedContext() (context.Context, error) {
	ctx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	futureData := app.Globals.UserService.UserLogin(ctx, "989100000002", "123456").Get()

	if futureData.Error() != nil {
		return nil, futureData.Error().Reason()
	}

	loginTokens, ok := futureData.Data().(user_service.LoginTokens)
	if ok != true {
		return nil, errors.New("data does not LoginTokens type")
	}

	var authorization = map[string]string{
		"authorization": fmt.Sprintf("Bearer %v", loginTokens.AccessToken),
		"userId":        "1000001",
	}
	md := metadata.New(authorization)
	ctxToken := metadata.NewOutgoingContext(ctx, md)

	return ctxToken, nil
}

func createRequestNewOrder() *pb.RequestNewOrder {
	order := &pb.RequestNewOrder{
		Platform: "PWA",
		Invoice:  &pb.Invoice{},
		Buyer: &pb.Buyer{
			Finance:         &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Invoice.GrandTotal = &pb.Money{
		Amount:   "600000",
		Currency: "IRR",
	}
	order.Invoice.Subtotal = &pb.Money{
		Amount:   "550000",
		Currency: "IRR",
	}
	order.Invoice.Discount = &pb.Money{
		Amount:   "50000",
		Currency: "IRR",
	}

	order.Invoice.Vat = &pb.Invoice_BusinessVAT{
		Value: 9,
	}

	order.Invoice.PaymentMethod = "IPG"
	order.Invoice.PaymentGateway = "AAP"
	order.Invoice.PaymentOption = nil
	order.Invoice.ShipmentTotal = &pb.Money{
		Amount:   "700000",
		Currency: "IRR",
	}
	order.Invoice.Voucher = &pb.Voucher{
		Percent: 0,
		RawAppliedPrice: &pb.Money{
			Amount:   "40000",
			Currency: "IRR",
		},
		RoundupAppliedPrice: &pb.Money{
			Amount:   "40000",
			Currency: "IRR",
		},
		Price: &pb.Money{
			Amount:   "40000",
			Currency: "IRR",
		},
		Code: "348",
		Details: &pb.VoucherDetails{
			StartDate:        "2019-12-28T14:32:46-0700",
			EndDate:          "2020-01-20T00:00:00-0000",
			Type:             "",
			MaxDiscountValue: 0,
			MinBasketValue:   0,
			Title:            "",
			Prefix:           "",
			UseLimit:         0,
			Count:            0,
			Length:           0,
			IsFirstPurchase:  false,
			Info:             nil,
			VoucherType:      pb.VoucherDetails_PURCHASE,
			VoucherSponsor:   pb.VoucherDetails_BAZLIA,
		},
	}

	order.Buyer.BuyerId = 1000001
	order.Buyer.LastName = "Tadayon"
	order.Buyer.FirstName = "Sina"
	order.Buyer.Email = "Sina.Tadayon@baman.io"
	order.Buyer.Mobile = "09124566788"
	order.Buyer.NationalId = "005938404734"
	order.Buyer.Ip = "127.0.0.1"
	order.Buyer.Gender = "male"

	order.Buyer.Finance.Iban = "IR165411211001514313143545"
	order.Buyer.Finance.AccountNumber = "303.100.1269574.1"
	order.Buyer.Finance.CardNumber = "4345345423533453"
	order.Buyer.Finance.BankName = "pasargad"

	order.Buyer.ShippingAddress.FirstName = "sina"
	order.Buyer.ShippingAddress.LastName = "tadayon"
	order.Buyer.ShippingAddress.Address = "Sheikh bahaee, p 5"
	order.Buyer.ShippingAddress.Province = "Tehran"
	order.Buyer.ShippingAddress.Mobile = "+98912193870"
	order.Buyer.ShippingAddress.Phone = "+98218475644"
	order.Buyer.ShippingAddress.ZipCode = "1651764614"
	order.Buyer.ShippingAddress.City = "Tehran"
	order.Buyer.ShippingAddress.Country = "Iran"
	order.Buyer.ShippingAddress.Neighbourhood = "Seool"
	order.Buyer.ShippingAddress.Lat = "10.1345664"
	order.Buyer.ShippingAddress.Long = "22.1345664"

	order.Packages = make([]*pb.Package, 0, 2)

	var pkg = &pb.Package{
		SellerId: 6546345,
		ShopName: "sazgar",
		Shipment: &pb.ShippingSpec{
			CarrierNames:   []string{"Post"},
			CarrierProduct: "Post Express",
			CarrierType:    "standard",
			ShippingCost: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			ReactionTime: 24,
			ShippingTime: 72,
			ReturnTime:   72,
			Details:      "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
		Invoice: &pb.PackageInvoice{
			Subtotal: &pb.Money{
				Amount:   "9238443",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "9734234",
				Currency: "IRR",
			},
			ShipmentPrice: &pb.Money{
				Amount:   "23123",
				Currency: "IRR",
			},
			Sso: &pb.PackageInvoice_SellerSSO{
				Value:     9,
				IsObliged: true,
			},
		},
	}
	order.Packages = append(order.Packages, pkg)
	pkg.Items = make([]*pb.Item, 0, 2)
	var item = &pb.Item{
		Sku:         "53456-2342",
		InventoryId: "1243444",
		Title:       "Asus",
		Brand:       "Electronic/laptop",
		Category:    "Asus G503 i7, 256SSD, 32G Ram",
		Guaranty:    "ضمانت سلامت کالا",
		Image:       "http://baman.io/image/asus.png",
		Returnable:  true,
		Quantity:    5,
		Attributes: map[string]*pb.Attribute{
			"Quantity": &pb.Attribute{
				KeyTrans: map[string]string{
					"en": "Quantity",
				},
				ValueTrans: map[string]string{
					"en": "10",
				},
			},
			"Width": &pb.Attribute{
				KeyTrans: map[string]string{
					"en": "Width",
				},
				ValueTrans: map[string]string{
					"en": "10",
				},
			},
		},
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "20000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "220000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "20000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
		},
	}
	pkg.Items = append(pkg.Items, item)
	item = &pb.Item{
		Sku:         "dfg34534",
		InventoryId: "57834534",
		Title:       "Nexus",
		Brand:       "Electronic/laptop",
		Category:    "Nexus G503 i7, 256SSD, 32G Ram",
		Guaranty:    "ضمانت سلامت کالا",
		Image:       "http://baman.io/image/nexus.png",
		Returnable:  true,
		Quantity:    8,
		Attributes:  nil,
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "10000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "120000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "10000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
		},
	}
	pkg.Items = append(pkg.Items, item)

	pkg = &pb.Package{
		SellerId: 111122223333,
		Shipment: &pb.ShippingSpec{
			CarrierNames:   []string{"Post"},
			CarrierProduct: "Post Express",
			CarrierType:    "standard",
			ShippingCost: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},

			ReactionTime: 24,
			ShippingTime: 72,
			ReturnTime:   72,
			Details:      "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
		Invoice: &pb.PackageInvoice{
			Subtotal: &pb.Money{
				Amount:   "9238443",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "9734234",
				Currency: "IRR",
			},
			ShipmentPrice: &pb.Money{
				Amount:   "23123",
				Currency: "IRR",
			},
			Sso: &pb.PackageInvoice_SellerSSO{
				Value:     16.67,
				IsObliged: true,
			},
		},
	}
	order.Packages = append(order.Packages, pkg)
	pkg.Items = make([]*pb.Item, 0, 2)
	item = &pb.Item{
		Sku:         "gffd-4534",
		InventoryId: "7684034234",
		Title:       "Asus",
		Brand:       "Electronic/laptop",
		Category:    "Asus G503 i7, 256SSD, 32G Ram",
		Guaranty:    "ضمانت سلامت کالا",
		Image:       "http://baman.io/image/asus.png",
		Returnable:  true,
		Quantity:    2,
		Attributes:  nil,
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "20000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "220000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "20000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
		},
	}
	pkg.Items = append(pkg.Items, item)
	item = &pb.Item{
		Sku:         "dfg-54322",
		InventoryId: "443353563463",
		Title:       "Nexus",
		Brand:       "Electronic/laptop",
		Category:    "Nexus G503 i7, 256SSD, 32G Ram",
		Guaranty:    "ضمانت سلامت کالا",
		Image:       "http://baman.io/image/nexus.png",
		Returnable:  true,
		Quantity:    6,
		Attributes:  nil,
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "10000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "120000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "10000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
		},
	}

	pkg.Items = append(pkg.Items, item)
	return order
}

func UpdateOrderAllStatus(ctx context.Context, order *entities.Order,
	status states.OrderStatus, pkgStatus states.PackageStatus, subPkgState states.IEnumState, actions ...*entities.Action) {
	order.UpdatedAt = time.Now().UTC()
	order.Status = string(status)
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].UpdatedAt = time.Now().UTC()
		order.Packages[i].Status = string(pkgStatus)
		//for z := 0; z < len(actions); z++ {
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			UpdateSubPackage(ctx, subPkgState, order.Packages[i].Subpackages[j], nil)
		}
		//}
	}
}

func UpdateOrderAllSubPkg(ctx context.Context, subPkgState states.IEnumState, order *entities.Order, actions ...*entities.Action) {
	order.UpdatedAt = time.Now().UTC()
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].UpdatedAt = time.Now().UTC()
		if actions != nil && len(actions) > 0 {
			for z := 0; z < len(actions); z++ {
				for j := 0; j < len(order.Packages[i].Subpackages); j++ {
					UpdateSubPackage(ctx, subPkgState, order.Packages[i].Subpackages[j], actions[z])
				}
			}
		} else {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				UpdateSubPackage(ctx, subPkgState, order.Packages[i].Subpackages[j], nil)
			}
		}
	}
}

func UpdateSubPackage(ctx context.Context, subPkgState states.IEnumState, subpackage *entities.Subpackage, action *entities.Action) {
	subpackage.UpdatedAt = time.Now().UTC()
	subpackage.Status = subPkgState.StateName()
	subpackage.Tracking.Action = action
	if subpackage.Tracking.State == nil {
		state := entities.State{
			Name:      subPkgState.StateName(),
			Index:     subPkgState.StateIndex(),
			Data:      nil,
			Actions:   nil,
			CreatedAt: time.Now().UTC(),
		}
		if action != nil {
			state.Actions = make([]entities.Action, 0, 8)
			state.Actions = append(state.Actions, *action)
		}

		if subpackage.Tracking.History == nil {
			subpackage.Tracking.History = make([]entities.State, 0, 32)
		}
		subpackage.Tracking.State = &state
		subpackage.Tracking.History = append(subpackage.Tracking.History, state)
	} else {
		if subpackage.Tracking.State.Index != subPkgState.StateIndex() {
			newState := entities.State{
				Name:      subPkgState.StateName(),
				Index:     subPkgState.StateIndex(),
				Data:      nil,
				Actions:   nil,
				CreatedAt: time.Now().UTC(),
			}
			if action != nil {
				newState.Actions = make([]entities.Action, 0, 8)
				newState.Actions = append(newState.Actions, *action)
			}
			//if subpackage.Tracking.History == nil {
			//	subpackage.Tracking.History = make([]entities.State, 0, 32)
			//}
			subpackage.Tracking.State = &newState
			subpackage.Tracking.History = append(subpackage.Tracking.History, newState)
		} else {
			if action != nil {
				subpackage.Tracking.State.Actions = append(subpackage.Tracking.State.Actions, *action)
				subpackage.Tracking.Action = action
			}
			subpackage.Tracking.History[len(subpackage.Tracking.History)-1] = *subpackage.Tracking.State
		}
	}
}

func addStock(ctx context.Context, requestNewOrder *pb.RequestNewOrder) error {
	if err := app.Globals.StockService.ConnectToStockService(); err != nil {
		return err
	}

	request := stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[0].Quantity + 300,
		InventoryId: requestNewOrder.Packages[0].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Add Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[1].Quantity + 300,
		InventoryId: requestNewOrder.Packages[0].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Add Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[0].Quantity + 300,
		InventoryId: requestNewOrder.Packages[1].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Add Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[1].Quantity + 300,
		InventoryId: requestNewOrder.Packages[1].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Add Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	return nil
}

func reservedStock(ctx context.Context, requestNewOrder *pb.RequestNewOrder) error {
	if err := app.Globals.StockService.ConnectToStockService(); err != nil {
		return err
	}

	request := stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Reserve Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Reserve Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Reserve success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Reserve stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	return nil
}

func releaseStock(ctx context.Context, requestNewOrder *pb.RequestNewOrder) error {
	if err := app.Globals.StockService.ConnectToStockService(); err != nil {
		return err
	}

	request := stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockRelease(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Release Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockRelease(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Release Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockRelease(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Release Stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockRelease(ctx, &request); err != nil {
		return err
	} else {
		applog.GLog.Logger.Debug("Release stock success",
			"inventoryId", request.InventoryId,
			"quantity", request.Quantity)
	}

	return nil
}

func TestSchedulerSellerShipmentPending(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	err = reservedStock(ctx, requestNewOrder)
	require.Nil(t, err)

	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	subpackages := make([]*pb.ActionData_Subpackage, 0, 1)
	subpackageItems := make([]*pb.ActionData_Subpackage_Item, 0, 2)

	subpackageItem := &pb.ActionData_Subpackage_Item{
		InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId,
		Quantity:    order.Packages[0].Subpackages[0].Items[0].Quantity,
		Reasons:     nil,
	}
	subpackageItems = append(subpackageItems, subpackageItem)

	subpackageItem = &pb.ActionData_Subpackage_Item{
		InventoryId: order.Packages[0].Subpackages[0].Items[1].InventoryId,
		Quantity:    order.Packages[0].Subpackages[0].Items[1].Quantity,
		Reasons:     nil,
	}
	subpackageItems = append(subpackageItems, subpackageItem)

	subpackage := &pb.ActionData_Subpackage{
		SID:   order.Packages[0].Subpackages[0].SId,
		Items: subpackageItems,
	}

	subpackages = append(subpackages, subpackage)

	actionData := &pb.ActionData{
		Subpackages:    subpackages,
		Carrier:        "",
		TrackingNumber: "",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   "Action",
		ADT:    "Single",
		Method: "Post",
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000001,
			UTP:       "Seller",
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: "Approve",
				StateIndex:  20,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	_, err = OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	time.Sleep(3 * time.Second)
	schedulerService.doProcess(ctx, states.ShipmentPending)

	time.Sleep(3 * time.Second)
	changedOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err)
	require.Equal(t, states.ShipmentDelayed.StateName(), changedOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, string(states.OrderInProgressStatus), changedOrder.Status)
}

func TestSchedulerDeliveryPending_Notification(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	err = reservedStock(ctx, requestNewOrder)
	require.Nil(t, err)

	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ShipmentDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Shipped)
	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	subpackages := make([]*pb.ActionData_Subpackage, 0, 1)
	subpackageItems := make([]*pb.ActionData_Subpackage_Item, 0, 2)

	subpackageItem := &pb.ActionData_Subpackage_Item{
		InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId,
		Quantity:    order.Packages[0].Subpackages[0].Items[0].Quantity,
		Reasons:     nil,
	}
	subpackageItems = append(subpackageItems, subpackageItem)

	subpackageItem = &pb.ActionData_Subpackage_Item{
		InventoryId: order.Packages[0].Subpackages[0].Items[1].InventoryId,
		Quantity:    order.Packages[0].Subpackages[0].Items[1].Quantity,
		Reasons:     nil,
	}
	subpackageItems = append(subpackageItems, subpackageItem)

	subpackage := &pb.ActionData_Subpackage{
		SID:   order.Packages[0].Subpackages[0].SId,
		Items: subpackageItems,
	}

	subpackages = append(subpackages, subpackage)

	actionData := &pb.ActionData{
		Subpackages:    subpackages,
		Carrier:        "",
		TrackingNumber: "",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   "Action",
		ADT:    "Single",
		Method: "Post",
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000001,
			UTP:       "Schedulers",
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: "DeliveryPending",
				StateIndex:  31,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	_, err = OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	time.Sleep(3 * time.Second)
	schedulerService.doProcess(ctx, states.DeliveryPending)

	time.Sleep(3 * time.Second)
	changedOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err)
	require.Equal(t, states.DeliveryPending.StateName(), changedOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, string(states.OrderInProgressStatus), changedOrder.Status)
}

func TestSchedulerDeliveryPending_Delivered(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	err = reservedStock(ctx, requestNewOrder)
	require.Nil(t, err)

	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ShipmentDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Shipped)
	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	subpackages := make([]*pb.ActionData_Subpackage, 0, 1)
	subpackageItems := make([]*pb.ActionData_Subpackage_Item, 0, 2)

	subpackageItem := &pb.ActionData_Subpackage_Item{
		InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId,
		Quantity:    order.Packages[0].Subpackages[0].Items[0].Quantity,
		Reasons:     nil,
	}
	subpackageItems = append(subpackageItems, subpackageItem)

	subpackageItem = &pb.ActionData_Subpackage_Item{
		InventoryId: order.Packages[0].Subpackages[0].Items[1].InventoryId,
		Quantity:    order.Packages[0].Subpackages[0].Items[1].Quantity,
		Reasons:     nil,
	}
	subpackageItems = append(subpackageItems, subpackageItem)

	subpackage := &pb.ActionData_Subpackage{
		SID:   order.Packages[0].Subpackages[0].SId,
		Items: subpackageItems,
	}

	subpackages = append(subpackages, subpackage)

	actionData := &pb.ActionData{
		Subpackages:    subpackages,
		Carrier:        "",
		TrackingNumber: "",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   "Action",
		ADT:    "Single",
		Method: "Post",
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000001,
			UTP:       "Schedulers",
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: "DeliveryPending",
				StateIndex:  31,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	_, err = OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	time.Sleep(25 * time.Second)
	schedulerService.doProcess(ctx, states.DeliveryPending)

	time.Sleep(3 * time.Second)
	changedOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err)
	require.Equal(t, states.Delivered.StateName(), changedOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, string(states.OrderInProgressStatus), changedOrder.Status)
}

func removeCollection() {
	if err := app.Globals.OrderRepository.RemoveAll(context.Background()); err != nil {
	}
}
