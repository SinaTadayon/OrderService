package grpc_server

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
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
	subpkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/subpackage"
	"gitlab.faza.io/order-project/order-service/domain/states"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
)

func TestMain(m *testing.M) {
	var err error
	if os.Getenv("APP_ENV") == "dev" {
		app.Globals.Config, err = configs.LoadConfig("../../testdata/.env")
	} else {
		app.Globals.Config, err = configs.LoadConfig("")
	}
	if err != nil {
		logger.Err("LoadConfig of main init failed, %s ", err.Error())
		os.Exit(1)
	}

	mongoDriver, err := app.SetupMongoDriver(*app.Globals.Config)
	if err != nil {
		os.Exit(1)
	}

	app.Globals.OrderRepository = order_repository.NewOrderRepository(mongoDriver)
	app.Globals.PkgItemRepository = pkg_repository.NewPkgItemRepository(mongoDriver)
	app.Globals.SubPkgRepository = subpkg_repository.NewSubPkgRepository(mongoDriver)

	// TODO create item repository
	flowManager, err := domain.NewFlowManager()
	if err != nil {
		logger.Err("flowManager creation failed, %s ", err.Error())
		panic("flowManager creation failed, " + err.Error())
	}

	grpcServer := NewServer(app.Globals.Config.GRPCServer.Address, uint16(app.Globals.Config.GRPCServer.Port), flowManager)

	app.Globals.Converter = converter.NewConverter()
	app.Globals.StockService = stock_service.NewStockService(app.Globals.Config.StockService.Address, app.Globals.Config.StockService.Port)

	if app.Globals.Config.PaymentGatewayService.MockEnabled {
		app.Globals.PaymentService = payment_service.NewPaymentServiceMock()
	} else {
		app.Globals.PaymentService = payment_service.NewPaymentService(app.Globals.Config.PaymentGatewayService.Address, app.Globals.Config.PaymentGatewayService.Port)
	}

	app.Globals.VoucherService = voucher_service.NewVoucherService(app.Globals.Config.VoucherService.Address, app.Globals.Config.VoucherService.Port)
	app.Globals.NotifyService = notify_service.NewNotificationService(app.Globals.Config.NotifyService.Address, app.Globals.Config.NotifyService.Port)
	app.Globals.UserService = user_service.NewUserService(app.Globals.Config.UserService.Address, app.Globals.Config.UserService.Port)

	if !checkTcpPort(app.Globals.Config.GRPCServer.Address, strconv.Itoa(app.Globals.Config.GRPCServer.Port)) {
		logger.Audit("Start GRPC Server for testing . . . ")
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
	ctx, _ := context.WithCancel(context.Background())
	futureData := app.Globals.UserService.UserLogin(ctx, "989100000002", "123456").Get()

	if futureData.Error() != nil {
		return nil, futureData.Error().Reason()
	}

	loginTokens, ok := futureData.Data().(user_service.LoginTokens)
	if ok != true {
		return nil, errors.New("data does not LoginTokens type")
	}

	var authorization = map[string]string{"authorization": fmt.Sprintf("Bearer %v", loginTokens.AccessToken)}
	md := metadata.New(authorization)
	ctxToken := metadata.NewOutgoingContext(ctx, md)

	return ctxToken, nil
}

func createRequestNewOrder() *pb.RequestNewOrder {
	order := &pb.RequestNewOrder{
		Invoice: &pb.Invoice{},
		Buyer: &pb.Buyer{
			Finance:         &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Invoice.GrandTotal = 600000
	order.Invoice.Subtotal = 550000
	order.Invoice.Discount = 50000
	order.Invoice.Currency = "IRR"
	order.Invoice.PaymentMethod = "IPG"
	order.Invoice.PaymentGateway = "AAP"
	order.Invoice.ShipmentTotal = 700000
	order.Invoice.Voucher = &pb.Voucher{
		Amount: 40000,
		Code:   "348",
	}

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

	order.Buyer.ShippingAddress.Address = "Sheikh bahaee, p 5"
	order.Buyer.ShippingAddress.Province = "Tehran"
	order.Buyer.ShippingAddress.Phone = "+98912193870"
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
			ShippingCost:   100000,
			VoucherAmount:  0,
			Currency:       "IRR",
			ReactionTime:   24,
			ShippingTime:   72,
			ReturnTime:     72,
			Details:        "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
		Invoice: &pb.PackageInvoice{
			Subtotal:       9238443,
			Discount:       9734234,
			ShipmentAmount: 23123,
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
		Attributes: map[string]string{
			"Quantity":  "10",
			"Width":     "8cm",
			"Height":    "10cm",
			"Length":    "15cm",
			"Weight":    "20kg",
			"Color":     "blue",
			"Materials": "stone",
		},
		Invoice: &pb.ItemInvoice{
			Unit:             200000,
			Total:            20000000,
			Original:         220000,
			Special:          200000,
			Discount:         20000,
			SellerCommission: 10,
			Currency:         "IRR",
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
		Attributes: map[string]string{
			"Quantity":  "20",
			"Width":     "8cm",
			"Height":    "10cm",
			"Length":    "15cm",
			"Weight":    "20kg",
			"Color":     "blue",
			"Materials": "stone",
		},
		Invoice: &pb.ItemInvoice{
			Unit:             100000,
			Total:            10000000,
			Original:         120000,
			Special:          100000,
			Discount:         10000,
			SellerCommission: 5,
			Currency:         "IRR",
		},
	}
	pkg.Items = append(pkg.Items, item)

	pkg = &pb.Package{
		SellerId: 111122223333,
		Shipment: &pb.ShippingSpec{
			CarrierNames:   []string{"Post"},
			CarrierProduct: "Post Express",
			CarrierType:    "standard",
			ShippingCost:   100000,
			VoucherAmount:  0,
			Currency:       "IRR",
			ReactionTime:   24,
			ShippingTime:   72,
			ReturnTime:     72,
			Details:        "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
		Invoice: &pb.PackageInvoice{
			Subtotal:       9238443,
			Discount:       9734234,
			ShipmentAmount: 23123,
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
		Attributes: map[string]string{
			"Quantity":  "10",
			"Width":     "8cm",
			"Height":    "10cm",
			"Length":    "15cm",
			"Weight":    "20kg",
			"Color":     "blue",
			"Materials": "stone",
		},
		Invoice: &pb.ItemInvoice{
			Unit:             200000,
			Total:            20000000,
			Original:         220000,
			Special:          200000,
			Discount:         20000,
			SellerCommission: 8,
			Currency:         "IRR",
		},
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
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
		Attributes: map[string]string{
			"Quantity":  "20",
			"Width":     "8cm",
			"Height":    "10cm",
			"Length":    "15cm",
			"Weight":    "20kg",
			"Color":     "blue",
			"Materials": "stone",
		},
		Invoice: &pb.ItemInvoice{
			Unit:             100000,
			Total:            10000000,
			Original:         120000,
			Special:          100000,
			Discount:         10000,
			SellerCommission: 3,
			Currency:         "IRR",
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
		Quantity:    requestNewOrder.Packages[0].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Add Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Add Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Add Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Add Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
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
		logger.Audit("Reserve Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Reserve Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Reserve success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Reserve stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
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
		logger.Audit("Release Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[0].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[0].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockRelease(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Release Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[0].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[0].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockRelease(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Release Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Packages[1].Items[1].Quantity,
		InventoryId: requestNewOrder.Packages[1].Items[1].InventoryId,
	}

	if _, err := app.Globals.StockService.GetStockClient().StockRelease(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Release stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
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
			UpdateSubPackage(ctx, subPkgState, &order.Packages[i].Subpackages[j], nil)
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
					UpdateSubPackage(ctx, subPkgState, &order.Packages[i].Subpackages[j], actions[z])
				}
			}
		} else {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				UpdateSubPackage(ctx, subPkgState, &order.Packages[i].Subpackages[j], nil)
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

func TestNewOrderRequest(t *testing.T) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()
	defer removeCollection()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)
	//ctx, err = createAuthenticatedContext()
	//assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	resOrder, err := OrderService.NewOrder(ctx, requestNewOrder)

	require.Nil(t, err)
	require.NotEmpty(t, resOrder.CallbackUrl, "CallbackUrl is empty")
}

func TestNewOrderRequestWithZeroAmountAndVoucher(t *testing.T) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, app.Globals.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(app.Globals.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	require.Nil(t, err)
	defer grpcConn.Close()
	defer removeCollection()

	requestNewOrder := createRequestNewOrder()
	requestNewOrder.Invoice.GrandTotal = 0
	requestNewOrder.Invoice.Voucher.Amount = 1000000
	err = addStock(ctx, requestNewOrder)
	require.Nil(t, err)

	defer releaseStock(ctx, requestNewOrder)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	resOrder, err := OrderService.NewOrder(ctx, requestNewOrder)

	require.Nil(t, err)
	require.NotEmpty(t, resOrder.CallbackUrl, "CallbackUrl is empty")
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

	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	newOrder.PaymentService = []entities.PaymentService{{
		PaymentRequest: &entities.PaymentRequest{
			Amount:    newOrder.Invoice.GrandTotal,
			Currency:  "IRR",
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
		Amount:    int64(order.Invoice.GrandTotal),
		ReqBody:   "request test url",
		ResBody:   "response test url",
		CardMask:  "293488374****7234",
		Result:    true,
	}

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

	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
	require.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	newOrder.PaymentService = []entities.PaymentService{{
		PaymentRequest: &entities.PaymentRequest{
			Amount:    newOrder.Invoice.GrandTotal,
			Currency:  "IRR",
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
		Amount:    int64(order.Invoice.GrandTotal),
		ReqBody:   "request test url",
		ResBody:   "response test url",
		CardMask:  "293488374****7234",
		Result:    false,
	}

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

	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
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
		Method: string(GetMethod),
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

	require.Equal(t, states.ApprovalPending.StateName(), lastOrder.Packages[0].Subpackages[0].Status)
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

	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
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
		Method: string(GetMethod),
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
			require.Equal(t, 3, subpackage.Items[0].Quantity)
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

	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
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
		Method: string(GetMethod),
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

	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
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
		Method: string(GetMethod),
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

func TestApprovalPending_BuyerReject_All(t *testing.T) {
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

	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
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
		Method: string(GetMethod),
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

//
//func TestOperatorShipmentPending_Failed(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//	err = addStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//	err = reservedStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "32.Shipment_Delivered", 32)
//	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	//for i:=0 ; i < len(order.Items); i++ {
//	//	order.Items[i].Status = "32.Shipment_Delivered"
//	//}
//	_, err = app.Globals.OrderRepository.Save(*order)
//	assert.Nil(t, err, "save failed")
//
//	defer removeCollection()
//
//	request := pb.RequestBackOfficeOrderAction{
//		ItemId:     order.Items[0].ItemId,
//		ActionType: "shipmentDelivered",
//		Action:     "cancel",
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.BackOfficeOrderAction(ctx, &request)
//
//	assert.Nil(t, err)
//
//	lastOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
//	assert.Nil(t, err, "failed")
//
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 80)
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "CANCELED")
//	assert.True(t, result.Result)
//}
//
//func TestSellerapp.GlobalsrovalPending_Success(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//
//	err = addStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	err = reservedStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "20.Seller_app.Globalsroval_Pending", 20)
//	updateOrderItemsProgress(newOrder, nil, "app.GlobalsrovalPending", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	//for i:=0 ; i < len(order.Items); i++ {
//	//	order.Items[i].Status = "20.Seller_app.Globalsroval_Pending"
//	//}
//	_, err = app.Globals.OrderRepository.Save(*order)
//	assert.Nil(t, err, "save failed")
//
//	defer removeCollection()
//
//	request := pb.RequestSellerOrderAction{
//		OrderId:    order.OrderId,
//		PId:   order.Items[0].SellerInfo.PId,
//		ActionType: "approved",
//		Action:     "success",
//		Data:       nil,
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.SellerOrderAction(ctx, &request)
//
//	assert.Nil(t, err)
//	lastOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
//	assert.Nil(t, err, "failed")
//
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 30)
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "SellerShipmentPending")
//
//	assert.True(t, result.Result)
//}
//
//func TestSellerapp.GlobalsrovalPending_Failed(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//	err = addStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	err = reservedStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	updateOrderStatus(order, nil, "IN_PROGRESS", false, "20.Seller_app.Globalsroval_Pending", 20)
//	updateOrderItemsProgress(order, nil, "app.GlobalsrovalPending", true, states.OrderInProgressStatus)
//	_, err = app.Globals.OrderRepository.Save(*order)
//	assert.Nil(t, err, "save failed")
//
//	//for i:=0 ; i < len(order.Items); i++ {
//	//	order.Items[i].Status = "20.Seller_app.Globalsroval_Pending"
//	//}
//	_, err = app.Globals.OrderRepository.Save(*order)
//	assert.Nil(t, err, "save failed")
//
//	defer removeCollection()
//
//	request := pb.RequestSellerOrderAction{
//		OrderId:    order.OrderId,
//		PId:   order.Items[0].SellerInfo.PId,
//		ActionType: "approved",
//		Action:     "failed",
//		Data: &pb.RequestSellerOrderAction_Failed{
//			Failed: &pb.RequestSellerOrderActionFailed{Reason: "Not Enough Stuff"},
//		},
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.SellerOrderAction(ctx, &request)
//
//	assert.Nil(t, err)
//
//	lastOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
//	assert.Nil(t, err, "failed")
//
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 80)
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "CANCELED")
//	assert.True(t, result.Result)
//}
//
//func TestShipmentPending_Success(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//	err = addStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	err = reservedStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	//for i:=0 ; i < len(order.Items); i++ {
//	//	order.Items[i].Status = "30.Shipment_Pending"
//	//}
//	_, err = app.Globals.OrderRepository.Save(*order)
//	assert.Nil(t, err, "save failed")
//
//	defer removeCollection()
//	request := pb.RequestSellerOrderAction{
//		OrderId:    order.OrderId,
//		PId:   order.Items[0].SellerInfo.PId,
//		ActionType: "shipped",
//		Action:     "success",
//		Data: &pb.RequestSellerOrderAction_Success{
//			Success: &pb.RequestSellerOrderActionSuccess{ShipmentMethod: "Post", TrackingId: "839832742"},
//		},
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.SellerOrderAction(ctx, &request)
//
//	assert.Nil(t, err)
//
//	lastOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
//	assert.Nil(t, err, "failed")
//
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 32)
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "ShipmentDeliveredPending")
//
//	assert.True(t, result.Result)
//}
//
//func TestShipmentPending_Failed(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//	err = addStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	err = reservedStock(ctx, requestNewOrder)
//	assert.Nil(t, err)
//
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	//for i:=0 ; i < len(order.Items); i++ {
//	//	order.Items[i].Status = "30.Shipment_Pending"
//	//}
//	_, err = app.Globals.OrderRepository.Save(*order)
//	assert.Nil(t, err, "save failed")
//
//	defer removeCollection()
//
//	request := pb.RequestSellerOrderAction{
//		OrderId:    order.OrderId,
//		PId:   order.Items[0].SellerInfo.PId,
//		ActionType: "shipped",
//		Action:     "failed",
//		Data: &pb.RequestSellerOrderAction_Failed{
//			Failed: &pb.RequestSellerOrderActionFailed{Reason: "Post Failed"},
//		},
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.SellerOrderAction(ctx, &request)
//
//	assert.Nil(t, err)
//
//	lastOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
//	assert.Nil(t, err, "failed")
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 80)
//	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "CANCELED")
//	assert.True(t, result.Result)
//}
//
//func TestSellerFindAllItems(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	request := &pb.RequestIdentifier{
//		Id: strconv.Itoa(int(order.Items[0].SellerInfo.PId)),
//	}
//
//	defer removeCollection()
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.SellerFindAllItems(ctx, request)
//
//	assert.Nil(t, err)
//	assert.Equal(t, result.Items[0].Quantity, int32(5))
//}
//
//func TestBuyerFindAllOrders(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//
//	defer grpcConn.Close()
//	requestNewOrder := createRequestNewOrder()
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	defer removeCollection()
//
//	request := &pb.RequestIdentifier{
//		Id: strconv.Itoa(int(order.BuyerInfo.BuyerId)),
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.BuyerFindAllOrders(ctx, request)
//
//	assert.Nil(t, err)
//	assert.Equal(t, len(result.Orders), 1)
//
//}
//
//func TestBackOfficeOrdersListView(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//
//	defer grpcConn.Close()
//	requestNewOrder := createRequestNewOrder()
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.OrderInProgressStatus)
//	_, err = app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	time.Sleep(100 * time.Millisecond)
//
//	requestNewOrder2 := createRequestNewOrder()
//	value2, err2 := app.Globals.Converter.Map(*requestNewOrder2, entities.Order{})
//	assert.Nil(t, err2, "Converter failed")
//	newOrder2 := value2.(*entities.Order)
//
//	updateOrderStatus(newOrder2, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder2, nil, "Shipped", true, states.OrderInProgressStatus)
//	_, err = app.Globals.OrderRepository.Save(*newOrder2)
//	assert.Nil(t, err2, "save failed")
//
//	time.Sleep(100 * time.Millisecond)
//
//	requestNewOrder1 := createRequestNewOrder()
//	value1, err1 := app.Globals.Converter.Map(*requestNewOrder1, entities.Order{})
//	assert.Nil(t, err1, "Converter failed")
//	newOrder1 := value1.(*entities.Order)
//
//	updateOrderStatus(newOrder1, nil, "IN_PROGRESS", false, "20.Seller_app.Globalsroval_Pending", 20)
//	updateOrderItemsProgress(newOrder1, nil, "app.Globalsroved", true, states.OrderInProgressStatus)
//	_, err = app.Globals.OrderRepository.Save(*newOrder1)
//	assert.Nil(t, err1, "save failed")
//
//	time.Sleep(100 * time.Millisecond)
//
//	requestNewOrder3 := createRequestNewOrder()
//	value3, err3 := app.Globals.Converter.Map(*requestNewOrder3, entities.Order{})
//	assert.Nil(t, err3, "Converter failed")
//	newOrder3 := value3.(*entities.Order)
//
//	updateOrderStatus(newOrder3, nil, "IN_PROGRESS", false, "20.Seller_app.Globalsroval_Pending", 20)
//	updateOrderItemsProgress(newOrder3, nil, "app.Globalsroved", true, states.OrderInProgressStatus)
//	_, err = app.Globals.OrderRepository.Save(*newOrder3)
//	assert.Nil(t, err3, "save failed")
//
//	defer removeCollection()
//
//	request := &pb.RequestBackOfficeOrdersList{
//		Page:      1,
//		PerPage:   3,
//		Sort:      "createdAt",
//		Direction: -1,
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.BackOfficeOrdersListView(ctx, request)
//
//	assert.Nil(t, err)
//	assert.Equal(t, len(result.Orders), 3)
//}
//
//func TestBackOfficeOrderDetailView(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//
//	defer grpcConn.Close()
//	requestNewOrder := createRequestNewOrder()
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.OrderInProgressStatus)
//	newOrder, err = app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//
//	defer removeCollection()
//
//	request := &pb.RequestIdentifier{
//		Id: strconv.Itoa(int(newOrder.OrderId)),
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	result, err := OrderService.BackOfficeOrderDetailView(ctx, request)
//
//	assert.Nil(t, err)
//	assert.Equal(t, result.OrderId, newOrder.OrderId)
//}
//
//func TestSellerReportOrders(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//	defer removeCollection()
//
//	request := &pb.RequestSellerReportOrders{
//		StartDateTime: order.CreatedAt.Unix() - 10,
//		PId:      order.Items[0].SellerInfo.PId,
//		Status:        order.Items[0].Status,
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	downloadStream, err := OrderService.SellerReportOrders(ctx, request)
//	assert.Nil(t, err)
//	defer downloadStream.CloseSend()
//
//	f, err := os.Create("/tmp/SellerReportOrder.csv")
//	assert.Nil(t, err)
//	defer f.Close()
//	defer os.Remove("/tmp/" + "SellerReportOrder.csv")
//
//	for {
//		res, err := downloadStream.Recv()
//		if err != nil {
//			if err == io.EOF {
//				break
//			}
//			break
//		}
//		_, err = f.Write(res.Data)
//		assert.Nil(t, err)
//	}
//
//	assert.Nil(t, err)
//}
//
//func TestBackOfficeReportOrderItems(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	grpcConn, err := grpc.DialContext(ctx, app.Globals.config.GRPCServer.Address+":"+
//		strconv.Itoa(int(app.Globals.config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
//	assert.Nil(t, err)
//	defer grpcConn.Close()
//
//	requestNewOrder := createRequestNewOrder()
//	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
//	assert.Nil(t, err, "Converter failed")
//	newOrder := value.(*entities.Order)
//
//	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
//	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.OrderInProgressStatus)
//	order, err := app.Globals.OrderRepository.Save(*newOrder)
//	assert.Nil(t, err, "save failed")
//	defer removeCollection()
//
//	request := &pb.RequestBackOfficeReportOrderItems{
//		StartDateTime: uint64(order.CreatedAt.Unix() - 10),
//		EndDataTime:   uint64(order.CreatedAt.Unix() + 10),
//	}
//
//	ctx, err = createAuthenticatedContext()
//	assert.Nil(t, err)
//
//	OrderService := pb.NewOrderServiceClient(grpcConn)
//	downloadStream, err := OrderService.BackOfficeReportOrderItems(ctx, request)
//	assert.Nil(t, err)
//	defer downloadStream.CloseSend()
//
//	f, err := os.Create("/tmp/BackOfficeReportOrderItems.csv")
//	assert.Nil(t, err)
//	defer f.Close()
//	defer os.Remove("/tmp/" + "BackOfficeReportOrderItems.csv")
//
//	for {
//		res, err := downloadStream.Recv()
//		if err != nil {
//			if err == io.EOF {
//				break
//			}
//			break
//		}
//		_, err = f.Write(res.Data)
//		assert.Nil(t, err)
//	}
//
//	assert.Nil(t, err)
//}

func removeCollection() {
	if err := app.Globals.OrderRepository.RemoveAll(context.Background()); err != nil {
	}
}
