package grpc_server

import (
	"context"
	"fmt"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/financeReport"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
	subpkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/subpackage"
	"gitlab.faza.io/order-project/order-service/domain/states"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
)

func TestMain(m *testing.M) {
	var err error
	if os.Getenv("APP_MODE") == "dev" {
		app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfigs("../../testdata/.env", "../../testdata/notification/sms/smsTemplate.txt")
	} else {
		app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfig("")
	}

	applog.GLog.ZapLogger = app.InitZap()
	applog.GLog.Logger = logger.NewZapLogger(applog.GLog.ZapLogger)

	app.Globals.ZapLogger = applog.GLog.ZapLogger
	app.Globals.Logger = applog.GLog.Logger

	if err != nil {
		applog.GLog.Logger.Error("LoadConfig of failed", "error", err)
		os.Exit(1)
	}

	mongoDriver, err := app.SetupMongoDriver(*app.Globals.Config)
	if err != nil {
		os.Exit(1)
	}

	app.Globals.OrderRepository = order_repository.NewOrderRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)
	app.Globals.PkgItemRepository = pkg_repository.NewPkgItemRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)
	app.Globals.SubPkgRepository = subpkg_repository.NewSubPkgRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)
	app.Globals.FinanceReportRepository = finance_repository.NewFinanceReportRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)

	// TODO create item repository
	flowManager, err := domain.NewFlowManager()
	if err != nil {
		applog.GLog.Logger.Error("flowManager creation failed", "error", err)
		os.Exit(1)
	}

	grpcServer := NewServer(app.Globals.Config.GRPCServer.Address, uint16(app.Globals.Config.GRPCServer.Port), flowManager)

	app.Globals.Converter = converter.NewConverter()
	app.Globals.StockService = stock_service.NewStockService(app.Globals.Config.StockService.Address, app.Globals.Config.StockService.Port, app.Globals.Config.StockService.Timeout)

	if app.Globals.Config.PaymentGatewayService.MockEnabled {
		app.Globals.PaymentService = payment_service.NewPaymentServiceMock()
	} else {
		app.Globals.PaymentService = payment_service.NewPaymentService(app.Globals.Config.PaymentGatewayService.Address, app.Globals.Config.PaymentGatewayService.Port, app.Globals.Config.PaymentGatewayService.CallbackTimeout, app.Globals.Config.PaymentGatewayService.PaymentResultTimeout)
	}

	app.Globals.VoucherService = voucher_service.NewVoucherService(app.Globals.Config.VoucherService.Address, app.Globals.Config.VoucherService.Port, app.Globals.Config.VoucherService.Timeout)
	app.Globals.NotifyService = notify_service.NewNotificationService(app.Globals.Config.NotifyService.Address, app.Globals.Config.NotifyService.Port, app.Globals.Config.NotifyService.NotifySeller, app.Globals.Config.NotifyService.NotifyBuyer, app.Globals.Config.NotifyService.Timeout)
	app.Globals.UserService = user_service.NewUserService(app.Globals.Config.UserService.Address, app.Globals.Config.UserService.Port, app.Globals.Config.UserService.Timeout)

	app.Globals.FlowManagerConfig = make(map[string]interface{}, 32)

	if app.Globals.Config.App.SchedulerStateTimeUint == "" {
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig] = app.HourTimeUnit
	} else {
		if app.Globals.Config.App.SchedulerStateTimeUint != "hour" &&
			app.Globals.Config.App.SchedulerStateTimeUint != "minute" {
			applog.GLog.Logger.Error("SchedulerStateTimeUint invalid, SchedulerStateTimeUint")
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig] = app.Globals.Config.App.SchedulerStateTimeUint
	}

	//if app.Globals.Config.App.SchedulerSellerReactionTime != "" {
	//	temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerSellerReactionTime)
	//	if err != nil {
	//		applog.GLog.Logger.Error("SchedulerSellerReactionTime invalid, SchedulerSellerReactionTime")
	//		os.Exit(1)
	//	}
	//	app.Globals.FlowManagerConfig[app.FlowManagerSchedulerSellerReactionTimeConfig] = temp
	//}

	if app.Globals.Config.App.SchedulerPaymentPendingState == "" {
		applog.GLog.Logger.Error("SchedulerPaymentPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerPaymentPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerPaymentPendingState invalid, SchedulerPaymentPendingState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerRetryPaymentPendingState == "" {
		applog.GLog.Logger.Error("SchedulerPaymentRetryPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerRetryPaymentPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerPaymentPendingState invalid, SchedulerRetryPaymentPendingState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerRetryPaymentPendingStateConfig] = int32(temp)
	}

	if app.Globals.Config.App.SchedulerApprovalPendingState == "" {
		applog.GLog.Logger.Error("SchedulerApprovalPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerApprovalPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerApprovalPendingState invalid, SchedulerApprovalPendingState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerApprovalPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerApprovalPendingState == "" {
		applog.GLog.Logger.Error("SchedulerApprovalPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerApprovalPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerApprovalPendingState invalid, SchedulerApprovalPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerApprovalPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerApprovalPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerShipmentPendingState == "" {
		applog.GLog.Logger.Error("SchedulerShipmentPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerShipmentPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerShipmentPendingState invalid, SchedulerShipmentPendingState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerShipmentPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerShippedState == "" {
		applog.GLog.Logger.Error("SchedulerShippedState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerShippedState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerShippedState invalid, SchedulerShippedState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerShippedStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerDeliveryPendingState == "" {
		applog.GLog.Logger.Error("SchedulerDeliveryPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerDeliveryPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerDeliveryPendingState invalid, SchedulerDeliveryPendingState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveryPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerNotifyDeliveryPendingState == "" {
		applog.GLog.Logger.Error("SchedulerNotifyDeliveryPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerNotifyDeliveryPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerNotifyDeliveryPendingState invalid, SchedulerNotifyDeliveryPendingState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerNotifyDeliveryPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerDeliveredState == "" {
		applog.GLog.Logger.Error("SchedulerDeliveredState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerDeliveredState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerDeliveredState invalid, SchedulerDeliveredState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveredStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnShippedState == "" {
		applog.GLog.Logger.Error("SchedulerReturnShippedState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnShippedState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerReturnShippedState invalid, SchedulerReturnShippedState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnShippedStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnRequestPendingState == "" {
		applog.GLog.Logger.Error("SchedulerReturnRequestPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnRequestPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerReturnRequestPendingState invalid, SchedulerReturnRequestPendingState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnRequestPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnShipmentPendingState == "" {
		applog.GLog.Logger.Error("SchedulerReturnShipmentPendingState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnShipmentPendingState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerReturnShipmentPendingState invalid, SchedulerReturnShipmentPendingState: %s, error: %s ", app.Globals.Config.App.SchedulerReturnShipmentPendingState, err.Error())
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnShipmentPendingStateConfig] = temp
	}

	if app.Globals.Config.App.SchedulerReturnDeliveredState == "" {
		applog.GLog.Logger.Error("SchedulerReturnDeliveredState is empty")
		os.Exit(1)
	} else {
		temp, err := strconv.Atoi(app.Globals.Config.App.SchedulerReturnDeliveredState)
		if err != nil {
			applog.GLog.Logger.Error("SchedulerReturnDeliveredState invalid, SchedulerReturnDeliveredState", "error", err)
			os.Exit(1)
		}
		app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnDeliveredStateConfig] = temp
	}

	if !checkTcpPort(app.Globals.Config.GRPCServer.Address, strconv.Itoa(app.Globals.Config.GRPCServer.Port)) {
		applog.GLog.Logger.Debug("Start GRPC Server for testing . . . ")
		go grpcServer.Start()
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
	futureData := app.Globals.UserService.UserLogin(ctx, "989200000000", "123456").Get()

	if futureData.Error() != nil {
		return nil, futureData.Error().Reason()
	}

	loginTokens, ok := futureData.Data().(user_service.LoginTokens)
	if ok != true {
		return nil, errors.New("data does not LoginTokens type")
	}

	var authorization = map[string]string{
		"authorization": fmt.Sprintf("Bearer %v", loginTokens.AccessToken),
		"userId":        "1000002",
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

	order.Buyer.BuyerId = 1000002
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
		SellerId: 1000002,
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

func addStock(ctx context.Context, requestNewOrder *pb.RequestNewOrder) error {
	if err := app.Globals.StockService.ConnectToStockService(); err != nil {
		return err
	}

	request := stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[0].Quantity + 500,
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
		Quantity:    requestNewOrder.Packages[0].Items[1].Quantity + 500,
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
		Quantity:    requestNewOrder.Packages[1].Items[0].Quantity + 500,
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
		Quantity:    requestNewOrder.Packages[1].Items[1].Quantity + 500,
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

func TestFinanceReport(t *testing.T) {

	startTimestamp := time.Now().UTC().Format(ISO8601)
	endTimestamp := time.Now().UTC().Add(time.Duration(time.Minute)).Format(ISO8601)
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PayToSeller)
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   "",
		ADT:    "",
		Method: "",
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:            0,
			UTP:            string(SchedulerUser),
			OID:            0,
			PID:            0,
			SIDs:           nil,
			Page:           1,
			PerPage:        10,
			IpAddress:      "",
			StartTimestamp: startTimestamp,
			EndTimestamp:   endTimestamp,
			Action:         nil,
			Sorts:          nil,
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(PayToSellerFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.FinanceOrderItems(ctx, request)
	require.Nil(t, err)

	var financeOrderItems pb.FinanceOrderItemDetailList
	err = ptypes.UnmarshalAny(response.Data, &financeOrderItems)
	require.Nil(t, err)

	require.NotNil(t, financeOrderItems)
	//require.Equal(t, 2, len(financeOrderItems.OrderItems))

}

func TestSellerOrderDetail(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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

	request := &pb.MessageRequest{
		Name:   string(SellerOrderDetail),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       1000002,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ApprovalPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderDetail pb.SellerOrderDetail
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderDetail)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderDetail)
	require.Equal(t, 2, len(sellerOrderDetail.Items))
	require.Equal(t, uint64(1000002), sellerOrderDetail.PID)
}

func TestSellerOrderDetailWithAllOrderFilter(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnCanceled)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PayToSeller)

	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderDetail),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       1000002,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(AllOrdersFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderDetail pb.SellerOrderDetail
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderDetail)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderDetail)
	require.Equal(t, 2, len(sellerOrderDetail.Items))
	require.Equal(t, uint64(1000002), sellerOrderDetail.PID)
}

func TestSellerOrderList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ApprovalPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderList pb.SellerOrderList
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderList)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderList)
	require.Equal(t, 1, len(sellerOrderList.Items))
	require.Equal(t, uint64(1000002), sellerOrderList.PID)
}

func TestSellerOrderListWithAllCancelFilter(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.CanceledByBuyer)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PayToBuyer)
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(AllCanceledFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderList pb.SellerOrderList
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderList)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderList)
	require.Equal(t, 1, len(sellerOrderList.Items))
	require.Equal(t, uint64(1000002), sellerOrderList.PID)
}

func TestSellerAllOrderList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(AllOrdersFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderList pb.SellerOrderList
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderList)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderList)
	require.Equal(t, 1, len(sellerOrderList.Items))
	require.Equal(t, uint64(1000002), sellerOrderList.PID)
}

func TestOperatorOrderDetail(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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

	request := &pb.MessageRequest{
		Name:   string(OperatorOrderDetail),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       1000002,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var operatorOrderDetail pb.OperatorOrderDetail
	err = ptypes.UnmarshalAny(response.Data, &operatorOrderDetail)
	require.Nil(t, err)

	require.NotNil(t, operatorOrderDetail)
	require.Equal(t, order.OrderId, operatorOrderDetail.OrderId)
	//require.Equal(t, uint64(1000002), sellerOrderDetail.PID)
}

func TestOperatorOrderList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(OperatorOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ApprovalPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var operatorOrderList pb.OperatorOrderList
	err = ptypes.UnmarshalAny(response.Data, &operatorOrderList)
	require.Nil(t, err)

	require.NotNil(t, operatorOrderList)
	require.Equal(t, 1, len(operatorOrderList.Orders))
	//require.Equal(t, uint64(1000002), operatorOrderList.PID)
}

func TestOperatorGetOrderById(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	savedOrder, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(OperatorOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       savedOrder.OrderId,
			PID:       1000002,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ApprovalPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var operatorOrderList pb.OperatorOrderList
	err = ptypes.UnmarshalAny(response.Data, &operatorOrderList)
	require.Nil(t, err)

	require.NotNil(t, operatorOrderList)
	require.Equal(t, 1, len(operatorOrderList.Orders))
	//require.Equal(t, uint64(1000002), sellerOrderList.PID)
}

func TestSellerGetOrderById(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	savedOrder, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       savedOrder.OrderId,
			PID:       1000002,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ApprovalPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderList pb.SellerOrderList
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderList)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderList)
	require.Equal(t, 1, len(sellerOrderList.Items))
	require.Equal(t, uint64(1000002), sellerOrderList.PID)
}

func TestSellerOrderList_ShipmentDelayedFilter(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ShipmentDelayedFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderList pb.SellerOrderList
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderList)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderList)
	require.Equal(t, 1, len(sellerOrderList.Items))
	require.Equal(t, uint64(1000002), sellerOrderList.PID)
}

func TestSellerReturnOrderDetailList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerReturnOrderList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ReturnRequestPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerReturnOrderList pb.SellerReturnOrderList
	err = ptypes.UnmarshalAny(response.Data, &sellerReturnOrderList)
	require.Nil(t, err)

	require.NotNil(t, sellerReturnOrderList)
	require.Equal(t, 1, len(sellerReturnOrderList.Items))
	require.Equal(t, uint64(1000002), sellerReturnOrderList.PID)
}

func TestSellerReturnOrderDetail(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	savedOrder, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerReturnOrderDetail),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       savedOrder.OrderId,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ReturnRequestPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerReturnOrderDetail pb.SellerReturnOrderDetail
	err = ptypes.UnmarshalAny(response.Data, &sellerReturnOrderDetail)
	require.Nil(t, err)

	require.NotNil(t, sellerReturnOrderDetail)
	require.Equal(t, uint64(1000002), sellerReturnOrderDetail.PID)
}

func TestSellerOrderReturnReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderReturnReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderReturnReports pb.SellerOrderReturnReports
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderReturnReports)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderReturnReports)
	require.Equal(t, uint32(1), sellerOrderReturnReports.ReturnDelivered)
	require.Equal(t, uint64(1000002), sellerOrderReturnReports.SellerId)
}

func TestSellerOrderDashboardReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderDashboardReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderDashboardReports pb.SellerOrderDashboardReports
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderDashboardReports)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderDashboardReports)
	require.Equal(t, uint32(1), sellerOrderDashboardReports.ShipmentDelayed)
	require.Equal(t, uint64(1000002), sellerOrderDashboardReports.SellerId)
}

func TestSellerOrderShipmentReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderShipmentReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderShipmentReports pb.SellerOrderShipmentReports
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderShipmentReports)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderShipmentReports)
	require.Equal(t, uint32(1), sellerOrderShipmentReports.ShipmentDelayed)
	require.Equal(t, uint64(1000002), sellerOrderShipmentReports.SellerId)
}

func TestSellerOrderDeliveredReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderDeliveredReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderDeliveredReports pb.SellerOrderDeliveredReports
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderDeliveredReports)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderDeliveredReports)
	require.Equal(t, uint32(1), sellerOrderDeliveredReports.DeliveryDelayed)
	require.Equal(t, uint64(1000002), sellerOrderDeliveredReports.SellerId)
}

func TestSellerOrderApprovalPendingReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerApprovalPendingOrderReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerApprovalPendingReports pb.SellerApprovalPendingReports
	err = ptypes.UnmarshalAny(response.Data, &sellerApprovalPendingReports)
	require.Nil(t, err)

	require.NotNil(t, sellerApprovalPendingReports)
	require.Equal(t, uint32(1), sellerApprovalPendingReports.ApprovalPending)
	require.Equal(t, uint64(1000002), sellerApprovalPendingReports.SellerId)
}

func TestSellerAllOrderReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.CanceledBySeller)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PayToBuyer)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerAllOrderReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerAllOrderReports pb.SellerAllOrderReports
	err = ptypes.UnmarshalAny(response.Data, &sellerAllOrderReports)
	require.Nil(t, err)

	require.NotNil(t, sellerAllOrderReports)
	require.Equal(t, uint32(1), sellerAllOrderReports.CancelReport.CanceledBySeller)
	require.Equal(t, uint64(1000002), sellerAllOrderReports.SellerId)
}

func TestSellerOrderCancelReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.CanceledBySeller)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PayToBuyer)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(SellerOrderCancelReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var sellerOrderCancelReports pb.SellerOrderCancelReports
	err = ptypes.UnmarshalAny(response.Data, &sellerOrderCancelReports)
	require.Nil(t, err)

	require.NotNil(t, sellerOrderCancelReports)
	require.Equal(t, uint32(1), sellerOrderCancelReports.CanceledBySeller)
	require.Equal(t, uint64(1000002), sellerOrderCancelReports.SellerId)
}

func TestBuyerOrderDetailList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentSuccess)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ApprovalPending)
	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(BuyerOrderDetailList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ApprovalPendingFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var buyerOrderDetailList pb.BuyerOrderDetailList
	err = ptypes.UnmarshalAny(response.Data, &buyerOrderDetailList)
	require.Nil(t, err)

	require.NotNil(t, buyerOrderDetailList)
	require.Equal(t, 1, len(buyerOrderDetailList.OrderDetails))
	require.Equal(t, uint64(1000002), buyerOrderDetailList.BuyerId)
}

func TestBuyerGetOrderByIdDetailList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	ctx, _ = context.WithCancel(context.Background())
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.PaymentFailed)
	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(BuyerOrderDetailList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       order.OrderId,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var buyerOrderDetailList pb.BuyerOrderDetailList
	err = ptypes.UnmarshalAny(response.Data, &buyerOrderDetailList)
	require.Nil(t, err)

	require.NotNil(t, buyerOrderDetailList)
	require.Equal(t, 1, len(buyerOrderDetailList.OrderDetails))
	require.Equal(t, uint64(1000002), buyerOrderDetailList.BuyerId)
}

func TestBuyerReturnOrderDetailList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(BuyerReturnOrderDetailList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(ReturnDeliveredFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var buyerReturnOrderDetailList pb.BuyerReturnOrderDetailList
	err = ptypes.UnmarshalAny(response.Data, &buyerReturnOrderDetailList)
	require.Nil(t, err)

	require.NotNil(t, buyerReturnOrderDetailList)
	require.Equal(t, 1, len(buyerReturnOrderDetailList.ReturnOrderDetail))
	require.Equal(t, uint64(1000002), buyerReturnOrderDetailList.BuyerId)
}

func TestBuyerReturnAllOrderDetailList(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(BuyerReturnOrderDetailList),
		Type:   string(DataReqType),
		ADT:    string(ListType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts: []*pb.MetaSorts{
				{
					Name:      "createdAt",
					Direction: "ASC",
				},
			},
			Filters: []*pb.MetaFilter{
				{
					Type:  string(OrderStateFilterType),
					Opt:   "eq",
					Value: string(AllOrdersFilter),
				},
			},
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var buyerReturnOrderDetailList pb.BuyerReturnOrderItemDetailList
	err = ptypes.UnmarshalAny(response.Data, &buyerReturnOrderDetailList)
	require.Nil(t, err)

	require.NotNil(t, buyerReturnOrderDetailList)
	require.Equal(t, 4, len(buyerReturnOrderDetailList.ReturnOrderItemDetailList))
	require.Equal(t, uint64(1000002), buyerReturnOrderDetailList.BuyerId)
}

func TestBuyerReturnOrderReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(BuyerReturnOrderReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var buyerReturnOrderReports pb.BuyerReturnOrderReports
	err = ptypes.UnmarshalAny(response.Data, &buyerReturnOrderReports)
	require.Nil(t, err)

	require.NotNil(t, buyerReturnOrderReports)
	require.Equal(t, int32(2), buyerReturnOrderReports.ReturnDelivered)
	require.Equal(t, uint64(1000002), buyerReturnOrderReports.BuyerId)
}

func TestBuyerAllOrderReports(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)

	_, err = app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   string(BuyerAllOrderReports),
		Type:   string(DataReqType),
		ADT:    string(SingleType),
		Method: string(GetMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       0,
			PID:       1000002,
			SIDs:      nil,
			Page:      1,
			PerPage:   2,
			IpAddress: "",
			Action:    nil,
			Sorts:     nil,
			Filters:   nil,
		},
		Data: nil,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	var buyerAllOrderReports pb.BuyerAllOrderReports
	err = ptypes.UnmarshalAny(response.Data, &buyerAllOrderReports)
	require.Nil(t, err)

	require.NotNil(t, buyerAllOrderReports)
	require.Equal(t, int32(2), buyerAllOrderReports.ReturnOrders)
	require.Equal(t, uint64(1000002), buyerAllOrderReports.BuyerId)
}

func TestNewOrderRequest(t *testing.T) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()
	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)
	//ctx, err = createAuthenticatedContext()
	//assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.NewOrder(ctx, requestNewOrder)

	require.Nil(t, err)
	require.NotEmpty(t, response.Response.(*pb.ResponseNewOrder_Ipg).Ipg.CallbackUrl, "CallbackUrl is empty")
}

func TestNewOrderRequestWithZeroAmountAndVoucher(t *testing.T) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()
	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	requestNewOrder := createRequestNewOrder()
	requestNewOrder.Invoice.GrandTotal = &pb.Money{
		Amount:   "0",
		Currency: "IRR",
	}
	requestNewOrder.Invoice.Voucher.Price = &pb.Money{
		Amount:   "1000000",
		Currency: "IRR",
	}
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.NewOrder(ctx, requestNewOrder)

	require.Nil(t, err)
	require.NotEmpty(t, response.Response.(*pb.ResponseNewOrder_Ipg).Ipg.CallbackUrl, "CallbackUrl is empty")
}

func TestPaymentPending_PaymentGatewayNotRespond(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err, "DialContext failed")
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	newOrder.OrderPayment = []entities.PaymentService{{
		PaymentRequest: &entities.PaymentRequest{
			Price: &entities.Money{
				Amount:   newOrder.Invoice.GrandTotal.Amount,
				Currency: "IRR",
			},
			Gateway:   "APP",
			CreatedAt: time.Now().UTC(),
		},
	}}

	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	for i := 0; i < len(newOrder.Packages); i++ {
		newOrder.Packages[i].UpdatedAt = time.Now().UTC()
		for j := 0; j < len(newOrder.Packages[i].Subpackages); j++ {
			newOrder.Packages[i].Subpackages[j].Tracking.State.Schedulers = []*entities.SchedulerData{
				{
					newOrder.Packages[i].Subpackages[j].OrderId,
					newOrder.Packages[i].Subpackages[j].PId,
					newOrder.Packages[i].Subpackages[j].SId,
					newOrder.Packages[i].Subpackages[j].Tracking.State.Name,
					newOrder.Packages[i].Subpackages[j].Tracking.State.Index,
					states.SchedulerJobName,
					states.SchedulerGroupName,
					scheduler_action.PaymentFail.ActionName(),
					0,
					app.Globals.FlowManagerConfig[app.FlowManagerSchedulerRetryPaymentPendingStateConfig].(int32),
					"",
					nil,
					nil,
					string(states.SchedulerSubpackageStateExpire),
					"",
					nil,
					true,
					time.Now().UTC(),
					time.Now().UTC(),
					time.Now().UTC(),
					nil,
					nil,
				},
			}
		}
	}

	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	request := &pb.SchedulerActionRequest{
		Orders: []*pb.SchedulerActionRequest_Order{
			{
				OID:         order.OrderId,
				ActionType:  "",
				ActionState: scheduler_action.PaymentFail.ActionName(),
				StateIndex:  int32(states.PaymentPending.StateIndex()),
				Packages:    nil,
			},
		},
	}

	serializedData, err := proto.Marshal(request)
	require.Nil(t, err)

	msgReq := &pb.MessageRequest{
		Name:   "",
		Type:   "Action",
		ADT:    "List",
		Method: "",
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
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

	OrderService := pb.NewOrderServiceClient(grpcConn)
	_, err = OrderService.SchedulerMessageHandler(ctx, msgReq)

	require.Nil(t, err)
	//require.True(t, resOrder., "payment result false")
}

func TestPaymentGateway_Success(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err, "DialContext failed")
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	newOrder.OrderPayment = []entities.PaymentService{{
		PaymentRequest: &entities.PaymentRequest{
			Price: &entities.Money{
				Amount:   newOrder.Invoice.GrandTotal.Amount,
				Currency: "IRR",
			},
			Gateway:   "APP",
			CreatedAt: time.Now().UTC(),
		},
	}}

	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	request := pg.PaygateHookRequest{
		OrderID:   strconv.Itoa(int(order.OrderId)),
		PaymentId: "534545345",
		InvoiceId: 3434234234,
		Amount:    0,
		CardMask:  "293488374****7234",
		Result:    true,
	}

	amount, err := strconv.Atoi(order.Invoice.GrandTotal.Amount)
	require.Nil(t, err)
	request.Amount = int64(amount)

	paymentService := pg.NewBankResultHookClient(grpcConn)
	response, err := paymentService.PaymentGatewayHook(ctx, &request)

	require.Nil(t, err)
	require.True(t, response.Ok, "payment result false")
}

func TestPaymentGateway_Fail(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err, "DialContext failed")
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	value, err := app.Globals.Converter.Map(ctx, requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	newOrder.OrderPayment = []entities.PaymentService{{
		PaymentRequest: &entities.PaymentRequest{
			Price: &entities.Money{
				Amount:   "newOrder.Invoice.GrandTotal",
				Currency: "IRR",
			},
			Gateway:   "APP",
			CreatedAt: time.Now().UTC(),
		},
	}}

	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.NewOrder)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderNewStatus, states.PackageNewStatus, states.PaymentPending)
	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err, "save failed")

	defer removeCollection()

	request := pg.PaygateHookRequest{
		OrderID:   strconv.Itoa(int(order.OrderId)),
		PaymentId: "534545345",
		InvoiceId: 3434234234,
		Amount:    0,
		CardMask:  "293488374****7234",
		Result:    false,
	}

	amount, err := strconv.Atoi(order.Invoice.GrandTotal.Amount)
	require.Nil(t, err)
	request.Amount = int64(amount)

	paymentService := pg.NewBankResultHookClient(grpcConn)
	response, err := paymentService.PaymentGatewayHook(ctx, &request)

	require.Nil(t, err)
	require.True(t, response.Ok, "payment result false")
}

func TestApprovalPending_SellerApproved_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(ApproveAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ShipmentPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestApprovalPending_SellerCancel_All(t *testing.T) {
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

	newOrder.Packages[1] = nil
	newOrder.Packages = newOrder.Packages[:len(newOrder.Packages)-1] // Truncate slice.

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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(RejectAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, string(states.OrderClosedStatus), lastOrder.Status)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestApprovalPending_SellerApproved_Diff(t *testing.T) {
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
		Quantity:    3,
		Reasons:     nil,
	}
	subpackageItems = append(subpackageItems, subpackageItem)

	subpackageItem = &pb.ActionData_Subpackage_Item{
		InventoryId: order.Packages[0].Subpackages[0].Items[1].InventoryId,
		Quantity:    3,
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(ApproveAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity - 3,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity - 3,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ApprovalPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 2, len(lastOrder.Packages[0].Subpackages))
	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)

	for _, subpackage := range lastOrder.Packages[0].Subpackages {
		if subpackage.SId == order.Packages[0].Subpackages[0].SId {
			require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], subpackage.Items[0].Quantity)
			require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], subpackage.Items[1].Quantity)
		} else if subpackage.SId == actionResponse.SIDs[0] {
			require.Equal(t, int32(3), subpackage.Items[0].Quantity)
		}
	}

	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
}

func TestApprovalPending_SellerApproved_DiffAndFullItem(t *testing.T) {
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
		Quantity:    3,
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(ApproveAction),
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
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ApprovalPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 2, len(lastOrder.Packages[0].Subpackages))
	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)

	for _, subpackage := range lastOrder.Packages[0].Subpackages {
		if len(subpackage.Items) == 1 {
			require.Equal(t, order.Packages[0].Subpackages[0].Items[1].Quantity-3,
				subpackage.Items[0].Quantity)
			continue
		} else {
			for _, item := range subpackage.Items {
				if order.Packages[0].Subpackages[0].Items[0].InventoryId == item.InventoryId {
					require.Equal(t, order.Packages[0].Subpackages[0].Items[0].Quantity, item.Quantity)
					break
				}
			}
			continue
		}
	}

	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
}

func TestApprovalPending_SellerReject_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(RejectAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestApprovalPending_BuyerCancel_All(t *testing.T) {
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
		Reasons: []*pb.Reason{
			&pb.Reason{
				Key:         "change_of_mind",
				Description: "",
			},
		},
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestApprovalPending_BuyerCancel_All_InvalidReason(t *testing.T) {
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
		Reasons: []*pb.Reason{
			&pb.Reason{
				Key:         "change_of_mindsdgsdf",
				Description: "",
			},
		},
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.NotNil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")
	// lastOrder.Packages[0].Subpackages[0]

	require.Equal(t, states.ApprovalPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	require.Nil(t, response)
	// var actionResponse pb.ActionResponse
	// err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	// require.Nil(t, err)
	// require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	// require.Equal(t, 1, len(actionResponse.SIDs))
	// require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipmentPending_SellerShipmentDetail_All(t *testing.T) {
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
		Carrier:        "Post",
		TrackingNumber: "1234567899999000",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(EnterShipmentDetailAction),
				StateIndex:  30,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.Shipped.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipmentPending_SellerCancel_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
				StateIndex:  30,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipmentPending_SchedulerCancel_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
				StateIndex:  30,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ShipmentDelayed.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipmentDelayed_SellerShipmentDetail_All(t *testing.T) {
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
		Carrier:        "Post",
		TrackingNumber: "1234567899999000",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(EnterShipmentDetailAction),
				StateIndex:  33,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.Shipped.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipmentDelayed_SellerCancel_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
				StateIndex:  33,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipmentDelayed_BuyerCancel_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
				StateIndex:  33,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipped_SchedulerDeliveryPending_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryPendingAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.DeliveryPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipped_SellerShipmentDetail_All(t *testing.T) {
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
		Carrier:        "Post",
		TrackingNumber: "1234567899999000",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(EnterShipmentDetailAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.Shipped.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestShipped_OperatorDeliveryDelayed_All(t *testing.T) {
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryDelayAction),
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

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.DeliveryDelayed.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDeliveryPending_SchedulerDelivered_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliverAction),
				StateIndex:  34,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.Delivered.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDeliveryPending_SchedulerNotification_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	newOrder.Packages[0].Subpackages[0].Tracking.State.Schedulers = []*entities.SchedulerData{
		{
			newOrder.Packages[0].Subpackages[0].OrderId,
			newOrder.Packages[0].Subpackages[0].PId,
			newOrder.Packages[0].Subpackages[0].SId,
			newOrder.Packages[0].Subpackages[0].Tracking.State.Name,
			newOrder.Packages[0].Subpackages[0].Tracking.State.Index,
			states.SchedulerJobName,
			states.SchedulerGroupName,
			scheduler_action.Notification.ActionName(),
			0,
			0,
			"",
			nil,
			nil,
			string(states.SchedulerSubpackageStateNotify),
			"",
			nil,
			true,
			time.Now().UTC(),
			time.Now().UTC(),
			time.Now().UTC(),
			nil,
			nil,
		},
		{
			newOrder.Packages[0].Subpackages[0].OrderId,
			newOrder.Packages[0].Subpackages[0].PId,
			newOrder.Packages[0].Subpackages[0].SId,
			newOrder.Packages[0].Subpackages[0].Tracking.State.Name,
			newOrder.Packages[0].Subpackages[0].Tracking.State.Index,
			states.SchedulerJobName,
			states.SchedulerGroupName,
			scheduler_action.Deliver.ActionName(),
			1,
			0,
			"",
			nil,
			nil,
			string(states.SchedulerSubpackageStateExpire),
			"",
			nil,
			true,
			time.Now().UTC().Add(1 * time.Minute),
			time.Now().UTC(),
			time.Now().UTC(),
			nil,
			nil,
		},
	}
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: "Notification",
				StateIndex:  34,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.DeliveryPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDeliveryPending_SellerEnterShipmentDetail_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
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
		Carrier:        "Post",
		TrackingNumber: "1234567899999000",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(EnterShipmentDetailAction),
				StateIndex:  34,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.DeliveryPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDeliveryPending_OperatorDeliveryDelayed_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryDelayAction),
				StateIndex:  34,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.DeliveryDelayed.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDeliveryDelayed_OperatorDelivery_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliverAction),
				StateIndex:  35,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.Delivered.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDeliveryDelayed_OperatorDeliveryFailed_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryFailAction),
				StateIndex:  35,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDelivered_BuyerSubmitReturnRequest_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(SubmitReturnRequestAction),
				StateIndex:  32,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnRequestPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDelivered_OperatorDeliveryDelayed_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryDelayAction),
				StateIndex:  32,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.DeliveryDelayed.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestDelivered_SchedulerClose_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CloseAction),
				StateIndex:  32,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToSeller.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRequestPending_SellerReject_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(RejectAction),
				StateIndex:  40,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnRequestRejected.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRequestPending_BuyerCancel_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
				StateIndex:  40,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToSeller.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRequestPending_SchedulerAccept_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(AcceptAction),
				StateIndex:  40,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnShipmentPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRequestPending_SellerAccept_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(AcceptAction),
				StateIndex:  40,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnShipmentPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnShipmentPending_BuyerEnterShipment_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
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
		Carrier:        "Post",
		TrackingNumber: "66666655555544444",
	}

	serializedData, err := proto.Marshal(actionData)
	require.Nil(t, err)

	request := &pb.MessageRequest{
		Name:   "",
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(BuyerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(EnterShipmentDetailAction),
				StateIndex:  50,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnShipped.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnShipmentPending_SchedulerClose_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(CancelAction),
				StateIndex:  50,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToSeller.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRequestReject_OperatorAccept_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(AcceptAction),
				StateIndex:  41,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnShipmentPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRequestReject_OperatorReject_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(RejectAction),
				StateIndex:  41,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToSeller.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnShipped_SellerDeliver_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliverAction),
				StateIndex:  51,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnDelivered.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnShipped_SchedulerDeliveryPending_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryPendingAction),
				StateIndex:  51,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnDeliveryPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnDelivered_SchedulerClose_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SchedulerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(AcceptAction),
				StateIndex:  52,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnDelivered_SellerReject_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(RejectAction),
				StateIndex:  52,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnRejected.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnDeliveryPending_SellerDeliver_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliverAction),
				StateIndex:  53,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnDelivered.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnDeliveryPending_SellerDeliveryFailed_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryPending)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(SellerUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryFailAction),
				StateIndex:  53,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnDeliveryDelayed.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnDeliveryDelayed_OperatorDeliver_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryDelayed)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliverAction),
				StateIndex:  54,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.ReturnDelivered.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnDeliveryDelayed_OperatorDeliveryFailed_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryDelayed)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(DeliveryFailAction),
				StateIndex:  54,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToSeller.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRejected_OperatorAccept_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRejected)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(AcceptAction),
				StateIndex:  55,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToBuyer.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func TestReturnRejected_OperatorReject_All(t *testing.T) {
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
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.DeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.Delivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRequestRejected)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipmentPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnShipped)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryPending)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDeliveryDelayed)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnDelivered)
	UpdateOrderAllStatus(ctx, newOrder, states.OrderInProgressStatus, states.PackageInProgressStatus, states.ReturnRejected)
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
		Type:   string(ActionReqType),
		ADT:    string(SingleType),
		Method: string(PostMethod),
		Time:   ptypes.TimestampNow(),
		Meta: &pb.RequestMetadata{
			UID:       1000002,
			UTP:       string(OperatorUser),
			OID:       order.OrderId,
			PID:       order.Packages[0].PId,
			SIDs:      nil,
			Page:      0,
			PerPage:   0,
			IpAddress: "",
			Action: &pb.MetaAction{
				ActionType:  "",
				ActionState: string(RejectAction),
				StateIndex:  55,
			},
			Sorts:   nil,
			Filters: nil,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionData),
			Value:   serializedData,
		},
	}

	var validation = map[string]int32{
		order.Packages[0].Subpackages[0].Items[0].InventoryId: order.Packages[0].Subpackages[0].Items[0].Quantity,
		order.Packages[0].Subpackages[0].Items[1].InventoryId: order.Packages[0].Subpackages[0].Items[1].Quantity,
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	response, err := OrderService.RequestHandler(ctx, request)
	require.Nil(t, err)

	lastOrder, err := app.Globals.OrderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err, "failed")

	require.Equal(t, states.PayToSeller.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, 1, len(lastOrder.Packages[0].Subpackages))
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[0].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[0].Quantity)
	require.Equal(t, validation[order.Packages[0].Subpackages[0].Items[1].InventoryId], lastOrder.Packages[0].Subpackages[0].Items[1].Quantity)

	var actionResponse pb.ActionResponse
	err = ptypes.UnmarshalAny(response.Data, &actionResponse)
	require.Nil(t, err)
	require.Equal(t, lastOrder.OrderId, actionResponse.OID)
	require.Equal(t, 1, len(actionResponse.SIDs))
	require.Equal(t, lastOrder.Packages[0].Subpackages[0].SId, actionResponse.SIDs[0])
}

func removeCollection() {
	if err := app.Globals.OrderRepository.RemoveAll(context.Background()); err != nil {
	}
}
