package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
	grpc_server "gitlab.faza.io/order-project/order-service/server/grpc"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
)

func init() {
	var err error
	if os.Getenv("APP_ENV") == "dev" {
		App.Config, err = configs.LoadConfig("./testdata/.env")
	} else {
		App.Config, err = configs.LoadConfig("")
	}
	if err != nil {
		logger.Err("LoadConfig of main init failed, %s ", err.Error())
		panic("LoadConfig of main init failed, " + err.Error())
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     App.Config.Mongo.Host,
		Port:     App.Config.Mongo.Port,
		Username: App.Config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:     time.Duration(App.Config.Mongo.ConnectionTimeout),
		ReadTimeout:     time.Duration(App.Config.Mongo.ReadTimeout),
		WriteTimeout:    time.Duration(App.Config.Mongo.WriteTimeout),
		MaxConnIdleTime: time.Duration(App.Config.Mongo.MaxConnIdleTime),
		MaxPoolSize:     uint64(App.Config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(App.Config.Mongo.MinPoolSize),
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("NewOrderRepository Mongo: %v", err.Error())
		panic("mongo adapter creation failed, " + err.Error())
	}

	global.Singletons.OrderRepository, err = order_repository.NewOrderRepository(mongoDriver)
	if err != nil {
		logger.Err("repository creation failed, %s ", err.Error())
		panic("order repository creation failed, " + err.Error())
	}

	// TODO create item repository
	App.flowManager, err = domain.NewFlowManager()
	if err != nil {
		logger.Err("flowManager creation failed, %s ", err.Error())
		panic("flowManager creation failed, " + err.Error())
	}

	App.grpcServer = grpc_server.NewServer(App.Config.GRPCServer.Address, uint16(App.Config.GRPCServer.Port), App.flowManager)

	global.Singletons.Converter = converter.NewConverter()

	if App.Config.StockService.MockEnabled {
		global.Singletons.StockService = stock_service.NewStockServiceMock()
	} else {
		global.Singletons.StockService = stock_service.NewStockService(App.Config.StockService.Address, App.Config.StockService.Port)
	}

	if App.Config.PaymentGatewayService.MockEnabled {
		global.Singletons.PaymentService = payment_service.NewPaymentServiceMock()
	} else {
		global.Singletons.PaymentService = payment_service.NewPaymentService(App.Config.PaymentGatewayService.Address, App.Config.PaymentGatewayService.Port)
	}

	if App.Config.VoucherService.MockEnabled {
		global.Singletons.VoucherService = voucher_service.NewVoucherServiceMock()
	} else {
		global.Singletons.VoucherService = voucher_service.NewVoucherService(App.Config.VoucherService.Address, App.Config.VoucherService.Port)
	}

	global.Singletons.NotifyService = notify_service.NewNotificationService()

	global.Singletons.UserService = user_service.NewUserService(App.Config.UserService.Address, App.Config.UserService.Port)

	if !checkTcpPort(App.Config.GRPCServer.Address, strconv.Itoa(App.Config.GRPCServer.Port)) {
		logger.Audit("Start GRPC Server for testing . . . ")
		go App.grpcServer.Start()
	}
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
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	iPromise := global.Singletons.UserService.UserLogin(ctx, "989100000002", "123456")
	futureData := iPromise.Data()

	if futureData.Ex != nil {
		return nil, futureData.Ex
	}

	loginTokens, ok := futureData.Data.(user_service.LoginTokens)
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
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance:         &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Amount.Total = 600000
	order.Amount.Subtotal = 550000
	order.Amount.Discount = 50000
	order.Amount.Currency = "IRR"
	order.Amount.PaymentMethod = "IPG"
	order.Amount.PaymentOption = "asanpardakht"
	order.Amount.ShipmentTotal = 700000
	order.Amount.Voucher = &pb.Voucher{
		Amount: 40000,
		Code:   "348",
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

	order.Buyer.ShippingAddress.Address = "Sheikh bahaee, p 5"
	order.Buyer.ShippingAddress.Province = "Tehran"
	order.Buyer.ShippingAddress.Phone = "+98912193870"
	order.Buyer.ShippingAddress.ZipCode = "1651764614"
	order.Buyer.ShippingAddress.City = "Tehran"
	order.Buyer.ShippingAddress.Country = "Iran"
	order.Buyer.ShippingAddress.Neighbourhood = "Seool"
	order.Buyer.ShippingAddress.Lat = "10.1345664"
	order.Buyer.ShippingAddress.Long = "22.1345664"

	item := pb.Item{
		Price:      &pb.PriceInfo{},
		Shipment:   &pb.ShippingSpec{},
		Attributes: make(map[string]string, 10),
		SellerId:   1000002,
	}

	item.InventoryId = "11111-22222"
	item.Brand = "Asus"
	item.Category = "Electronic/laptop"
	item.Title = "Asus G503 i7, 256SSD, 32G Ram"
	item.Guaranty = "ضمانت سلامت کالا"
	item.Image = "http://baman.io/image/asus.png"
	item.Returnable = true
	item.Quantity = 5

	item.Attributes["Quantity"] = "10"
	item.Attributes["Width"] = "8cm"
	item.Attributes["Height"] = "10cm"
	item.Attributes["Length"] = "15cm"
	item.Attributes["Weight"] = "20kg"
	item.Attributes["Color"] = "blue"
	item.Attributes["Materials"] = "stone"

	item.Price.Special = 200000
	item.Price.Total = 22000000
	item.Price.Original = 20000000
	item.Price.SellerCommission = 10
	item.Price.Unit = 100000
	item.Price.Currency = "IRR"

	//Standard, Express, Economy or Sameday.
	item.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item.Shipment.ShippingTime = 72
	item.Shipment.ReturnTime = 72
	item.Shipment.ReactionTime = 24
	item.Shipment.CarrierName = "Post"
	item.Shipment.CarrierProduct = "Post Express"
	item.Shipment.CarrierType = "standard"
	item.Shipment.ShippingCost = 100000
	item.Shipment.VoucherAmount = 0
	item.Shipment.Currency = "IRR"

	order.Items = append(order.Items, &item)

	item1 := pb.Item{
		Price:      &pb.PriceInfo{},
		Shipment:   &pb.ShippingSpec{},
		Attributes: make(map[string]string, 10),
		SellerId:   678912,
	}

	item1.InventoryId = "1111-33333"
	item1.Brand = "Lenovo"
	item1.Category = "Electronic/laptop"
	item1.Title = "Lenove G503 i7, 256SSD, 32G Ram"
	item1.Guaranty = "ضمانت سلامت کالا"
	item1.Image = "http://baman.io/image/asus.png"
	item1.Returnable = true
	item1.Quantity = 5

	item1.Attributes["Quantity"] = "10"
	item1.Attributes["Width"] = "8cm"
	item1.Attributes["Height"] = "10cm"
	item1.Attributes["Length"] = "15cm"
	item1.Attributes["Weight"] = "20kg"
	item1.Attributes["Color"] = "blue"
	item1.Attributes["Materials"] = "stone"

	item1.Price.Special = 250000
	item.Price.Total = 2500000
	item1.Price.Original = 200000
	item1.Price.SellerCommission = 10
	item1.Price.Unit = 200000
	item1.Price.Currency = "IRR"

	//Standard, Express, Economy or Sameday.
	item1.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item1.Shipment.ShippingTime = 72
	item1.Shipment.ReturnTime = 72
	item1.Shipment.ReactionTime = 24
	item1.Shipment.CarrierName = "Post"
	item1.Shipment.CarrierProduct = "Post Express"
	item1.Shipment.CarrierType = "standard"
	item1.Shipment.ShippingCost = 100000
	item1.Shipment.VoucherAmount = 0
	item1.Shipment.Currency = "IRR"

	order.Items = append(order.Items, &item1)
	return order
}

func updateOrderStatus(order *entities.Order, itemsId []uint64, orderStatus string, isUpdateOnlyOrderStatus bool, stepName string, stepIndex int) {

	if isUpdateOnlyOrderStatus == true {
		order.UpdatedAt = time.Now().UTC()
		order.Status = orderStatus
	} else {
		order.UpdatedAt = time.Now().UTC()
		order.Status = orderStatus
		findFlag := true
		if itemsId != nil && len(itemsId) > 0 {
			for _, id := range itemsId {
				findFlag = false
				for i := 0; i < len(order.Items); i++ {
					if order.Items[i].ItemId == id {
						doUpdateOrderStep(order, i, stepName, stepIndex)
						findFlag = true
						break
					}
				}
				if !findFlag {
					logger.Err("%s received itemId %d not exist in order, orderId: %d", stepName, id, order.OrderId)
				}
			}
		} else {
			for i := 0; i < len(order.Items); i++ {
				doUpdateOrderStep(order, i, stepName, stepIndex)
			}
		}
	}
}

func doUpdateOrderStep(order *entities.Order, index int, stepName string, stepIndex int) {
	order.Items[index].Progress.CreatedAt = time.Now().UTC()
	order.Items[index].Progress.CurrentStepName = stepName
	order.Items[index].Progress.CurrentStepIndex = stepIndex

	stepHistory := entities.StateHistory{
		Name:      stepName,
		Index:     stepIndex,
		CreatedAt: order.Items[index].Progress.CreatedAt,
		//ActionHistory: make([]entities.Action, 0, 1),
	}

	if order.Items[index].Progress.StepsHistory == nil || len(order.Items[index].Progress.StepsHistory) == 0 {
		order.Items[index].Progress.StepsHistory = make([]entities.StateHistory, 0, 5)
	}

	order.Items[index].Progress.StepsHistory = append(order.Items[index].Progress.StepsHistory, stepHistory)
}

func updateOrderItemsProgress(order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					doUpdateOrderItemsProgress(order, i, action, result, itemStatus)
					findFlag = true
				}
			}

			if !findFlag {
				logger.Err("received itemId %d not exist in order, orderId: %d", id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			doUpdateOrderItemsProgress(order, i, action, result, itemStatus)
		}
	}
}

func doUpdateOrderItemsProgress(order *entities.Order, index int,
	actionName string, result bool, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
}

func addStock(ctx context.Context, requestNewOrder *pb.RequestNewOrder) error {
	if err := global.Singletons.StockService.ConnectToStockService(); err != nil {
		return err
	}

	request := stockProto.StockRequest{
		Quantity:    requestNewOrder.Items[0].Quantity + 100,
		InventoryId: requestNewOrder.Items[0].InventoryId,
	}

	if _, err := global.Singletons.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Add Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Items[1].Quantity + 100,
		InventoryId: requestNewOrder.Items[1].InventoryId,
	}

	if _, err := global.Singletons.StockService.GetStockClient().StockAllocate(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Add Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	return nil
}

func reservedStock(ctx context.Context, requestNewOrder *pb.RequestNewOrder) error {
	if err := global.Singletons.StockService.ConnectToStockService(); err != nil {
		return err
	}

	request := stockProto.StockRequest{
		Quantity:    requestNewOrder.Items[0].Quantity,
		InventoryId: requestNewOrder.Items[0].InventoryId,
	}

	if _, err := global.Singletons.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Reserved Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	request = stockProto.StockRequest{
		Quantity:    requestNewOrder.Items[1].Quantity,
		InventoryId: requestNewOrder.Items[1].InventoryId,
	}

	if _, err := global.Singletons.StockService.GetStockClient().StockReserve(ctx, &request); err != nil {
		return err
	} else {
		logger.Audit("Reserved Stock success, inventoryId: %s, quantity: %d", request.InventoryId, request.Quantity)
	}

	return nil
}

func TestNewOrderRequest(t *testing.T) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()
	defer removeCollection()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	//ctx, err = createAuthenticatedContext()
	//assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	resOrder, err := OrderService.NewOrder(ctx, requestNewOrder)

	assert.Nil(t, err)
	assert.NotEmpty(t, resOrder.CallbackUrl, "CallbackUrl is empty")
}

func TestNewOrderRequestWithZeroAmountAndVoucher(t *testing.T) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()
	defer removeCollection()

	requestNewOrder := createRequestNewOrder()
	requestNewOrder.Amount.Total = 0
	requestNewOrder.Amount.Voucher.Amount = 1000000
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	resOrder, err := OrderService.NewOrder(ctx, requestNewOrder)

	assert.Nil(t, err)
	assert.NotEmpty(t, resOrder.CallbackUrl, "CallbackUrl is empty")
}

func TestPaymentGateway(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err, "DialContext failed")
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	newOrder.PaymentService = []entities.PaymentService{{
		PaymentRequest: &entities.PaymentRequest{
			Amount:    newOrder.Invoice.Total,
			Currency:  "IRR",
			Gateway:   "APP",
			CreatedAt: time.Now().UTC(),
		},
	}}

	updateOrderStatus(newOrder, nil, states.NewStatus, false, "0.New_Order", 0)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := pg.PaygateHookRequest{
		OrderID:   strconv.Itoa(int(order.OrderId)),
		PaymentId: "534545345",
		InvoiceId: 3434234234,
		Amount:    int64(order.Invoice.Total),
		ReqBody:   "request test url",
		ResBody:   "response test url",
		CardMask:  "293488374****7234",
		Result:    true,
	}

	paymentService := pg.NewBankResultHookClient(grpcConn)
	response, err := paymentService.PaymentGatewayHook(ctx, &request)

	assert.Nil(t, err)
	assert.True(t, response.Ok, "payment result false")
}

func TestOperatorShipmentPending_Success(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, states.InProgressStatus, false, "32.Shipment_Delivered", 32)
	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	//for i:=0 ; i < len(order.Items); i++ {
	//	order.Items[i].Status = "32.Shipment_Delivered"
	//}
	_, err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := pb.RequestBackOfficeOrderAction{
		ItemId:     order.Items[0].ItemId,
		ActionType: "shipmentDelivered",
		Action:     "success",
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.BackOfficeOrderAction(ctx, &request)

	assert.Nil(t, err)

	lastOrder, err := global.Singletons.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err, "failed")

	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 90)
	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "DELIVERED")

	assert.True(t, result.Result)
}

func TestOperatorShipmentPending_Failed(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)
	err = reservedStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "32.Shipment_Delivered", 32)
	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	//for i:=0 ; i < len(order.Items); i++ {
	//	order.Items[i].Status = "32.Shipment_Delivered"
	//}
	_, err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := pb.RequestBackOfficeOrderAction{
		ItemId:     order.Items[0].ItemId,
		ActionType: "shipmentDelivered",
		Action:     "cancel",
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.BackOfficeOrderAction(ctx, &request)

	assert.Nil(t, err)

	lastOrder, err := global.Singletons.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err, "failed")

	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 80)
	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "CANCELED")
	assert.True(t, result.Result)
}

func TestSellerApprovalPending_Success(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()

	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "20.Seller_Approval_Pending", 20)
	updateOrderItemsProgress(newOrder, nil, "ApprovalPending", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	//for i:=0 ; i < len(order.Items); i++ {
	//	order.Items[i].Status = "20.Seller_Approval_Pending"
	//}
	_, err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := pb.RequestSellerOrderAction{
		OrderId:    order.OrderId,
		SellerId:   order.Items[0].SellerInfo.SellerId,
		ActionType: "approved",
		Action:     "success",
		Data:       nil,
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.SellerOrderAction(ctx, &request)

	assert.Nil(t, err)
	lastOrder, err := global.Singletons.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err, "failed")

	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 30)
	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "SellerShipmentPending")

	assert.True(t, result.Result)
}

func TestSellerApprovalPending_Failed(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	updateOrderStatus(order, nil, "IN_PROGRESS", false, "20.Seller_Approval_Pending", 20)
	updateOrderItemsProgress(order, nil, "ApprovalPending", true, states.InProgressStatus)
	_, err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	//for i:=0 ; i < len(order.Items); i++ {
	//	order.Items[i].Status = "20.Seller_Approval_Pending"
	//}
	_, err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := pb.RequestSellerOrderAction{
		OrderId:    order.OrderId,
		SellerId:   order.Items[0].SellerInfo.SellerId,
		ActionType: "approved",
		Action:     "failed",
		Data: &pb.RequestSellerOrderAction_Failed{
			Failed: &pb.RequestSellerOrderActionFailed{Reason: "Not Enough Stuff"},
		},
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.SellerOrderAction(ctx, &request)

	assert.Nil(t, err)

	lastOrder, err := global.Singletons.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err, "failed")

	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 80)
	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "CANCELED")
	assert.True(t, result.Result)
}

func TestShipmentPending_Success(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	//for i:=0 ; i < len(order.Items); i++ {
	//	order.Items[i].Status = "30.Shipment_Pending"
	//}
	_, err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	defer removeCollection()
	request := pb.RequestSellerOrderAction{
		OrderId:    order.OrderId,
		SellerId:   order.Items[0].SellerInfo.SellerId,
		ActionType: "shipped",
		Action:     "success",
		Data: &pb.RequestSellerOrderAction_Success{
			Success: &pb.RequestSellerOrderActionSuccess{ShipmentMethod: "Post", TrackingId: "839832742"},
		},
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.SellerOrderAction(ctx, &request)

	assert.Nil(t, err)

	lastOrder, err := global.Singletons.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err, "failed")

	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 32)
	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "ShipmentDeliveredPending")

	assert.True(t, result.Result)
}

func TestShipmentPending_Failed(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	err = addStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	err = reservedStock(ctx, requestNewOrder)
	assert.Nil(t, err)

	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	//for i:=0 ; i < len(order.Items); i++ {
	//	order.Items[i].Status = "30.Shipment_Pending"
	//}
	_, err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := pb.RequestSellerOrderAction{
		OrderId:    order.OrderId,
		SellerId:   order.Items[0].SellerInfo.SellerId,
		ActionType: "shipped",
		Action:     "failed",
		Data: &pb.RequestSellerOrderAction_Failed{
			Failed: &pb.RequestSellerOrderActionFailed{Reason: "Post Failed"},
		},
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.SellerOrderAction(ctx, &request)

	assert.Nil(t, err)

	lastOrder, err := global.Singletons.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err, "failed")
	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].Index, 80)
	assert.Equal(t, lastOrder.Items[0].Progress.StepsHistory[len(lastOrder.Items[0].Progress.StepsHistory)-1].ActionHistory[0].Name, "CANCELED")
	assert.True(t, result.Result)
}

func TestSellerFindAllItems(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	request := &pb.RequestIdentifier{
		Id: strconv.Itoa(int(order.Items[0].SellerInfo.SellerId)),
	}

	defer removeCollection()

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.SellerFindAllItems(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, result.Items[0].Quantity, int32(5))
}

func TestBuyerFindAllOrders(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)

	defer grpcConn.Close()
	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := &pb.RequestIdentifier{
		Id: strconv.Itoa(int(order.BuyerInfo.BuyerId)),
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.BuyerFindAllOrders(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, len(result.Orders), 1)

}

func TestBackOfficeOrdersListView(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)

	defer grpcConn.Close()
	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.InProgressStatus)
	_, err = global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	time.Sleep(100 * time.Millisecond)

	requestNewOrder2 := createRequestNewOrder()
	value2, err2 := global.Singletons.Converter.Map(*requestNewOrder2, entities.Order{})
	assert.Nil(t, err2, "Converter failed")
	newOrder2 := value2.(*entities.Order)

	updateOrderStatus(newOrder2, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder2, nil, "Shipped", true, states.InProgressStatus)
	_, err = global.Singletons.OrderRepository.Save(*newOrder2)
	assert.Nil(t, err2, "save failed")

	time.Sleep(100 * time.Millisecond)

	requestNewOrder1 := createRequestNewOrder()
	value1, err1 := global.Singletons.Converter.Map(*requestNewOrder1, entities.Order{})
	assert.Nil(t, err1, "Converter failed")
	newOrder1 := value1.(*entities.Order)

	updateOrderStatus(newOrder1, nil, "IN_PROGRESS", false, "20.Seller_Approval_Pending", 20)
	updateOrderItemsProgress(newOrder1, nil, "Approved", true, states.InProgressStatus)
	_, err = global.Singletons.OrderRepository.Save(*newOrder1)
	assert.Nil(t, err1, "save failed")

	time.Sleep(100 * time.Millisecond)

	requestNewOrder3 := createRequestNewOrder()
	value3, err3 := global.Singletons.Converter.Map(*requestNewOrder3, entities.Order{})
	assert.Nil(t, err3, "Converter failed")
	newOrder3 := value3.(*entities.Order)

	updateOrderStatus(newOrder3, nil, "IN_PROGRESS", false, "20.Seller_Approval_Pending", 20)
	updateOrderItemsProgress(newOrder3, nil, "Approved", true, states.InProgressStatus)
	_, err = global.Singletons.OrderRepository.Save(*newOrder3)
	assert.Nil(t, err3, "save failed")

	defer removeCollection()

	request := &pb.RequestBackOfficeOrdersList{
		Page:      1,
		PerPage:   3,
		Sort:      "createdAt",
		Direction: -1,
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.BackOfficeOrdersListView(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, len(result.Orders), 3)
}

func TestBackOfficeOrderDetailView(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)

	defer grpcConn.Close()
	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.InProgressStatus)
	newOrder, err = global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	request := &pb.RequestIdentifier{
		Id: strconv.Itoa(int(newOrder.OrderId)),
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.BackOfficeOrderDetailView(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, result.OrderId, newOrder.OrderId)
}

func TestSellerReportOrders(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")
	defer removeCollection()

	request := &pb.RequestSellerReportOrders{
		StartDateTime: order.CreatedAt.Unix() - 10,
		SellerId:      order.Items[0].SellerInfo.SellerId,
		Status:        order.Items[0].Status,
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	downloadStream, err := OrderService.SellerReportOrders(ctx, request)
	assert.Nil(t, err)
	defer downloadStream.CloseSend()

	f, err := os.Create("/tmp/SellerReportOrder.csv")
	assert.Nil(t, err)
	defer f.Close()
	defer os.Remove("/tmp/" + "SellerReportOrder.csv")

	for {
		res, err := downloadStream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		_, err = f.Write(res.Data)
		assert.Nil(t, err)
	}

	assert.Nil(t, err)
}

func TestBackOfficeReportOrderItems(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address+":"+
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer grpcConn.Close()

	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "Shipped", true, states.InProgressStatus)
	order, err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")
	defer removeCollection()

	request := &pb.RequestBackOfficeReportOrderItems{
		StartDateTime: uint64(order.CreatedAt.Unix() - 10),
		EndDataTime:   uint64(order.CreatedAt.Unix() + 10),
	}

	ctx, err = createAuthenticatedContext()
	assert.Nil(t, err)

	OrderService := pb.NewOrderServiceClient(grpcConn)
	downloadStream, err := OrderService.BackOfficeReportOrderItems(ctx, request)
	assert.Nil(t, err)
	defer downloadStream.CloseSend()

	f, err := os.Create("/tmp/BackOfficeReportOrderItems.csv")
	assert.Nil(t, err)
	defer f.Close()
	defer os.Remove("/tmp/" + "BackOfficeReportOrderItems.csv")

	for {
		res, err := downloadStream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		_, err = f.Write(res.Data)
		assert.Nil(t, err)
	}

	assert.Nil(t, err)
}

func removeCollection() {
	if err := global.Singletons.OrderRepository.RemoveAll(); err != nil {
	}
}
