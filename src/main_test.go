package main

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	grpc_server "gitlab.faza.io/order-project/order-service/server/grpc"
	pg "gitlab.faza.io/protos/payment-gateway"
	"google.golang.org/grpc"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pb "gitlab.faza.io/protos/order"
)



//var EnvFile = ".env"
////var ConfigurationFile = "configuration.go"
//
//func AppendToFile() {
//	f, err := os.OpenFile(EnvFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer f.Close()
//
//	_, err = f.WriteString("\n__CTO__")
//	if err != nil {
//		log.Fatal(err)
//	}
//}
//func FixEnvFile() {
//	f, err := ioutil.ReadFile(EnvFile)
//	if err != nil {
//		log.Fatal(err)
//	}
//	newContent := bytes.ReplaceAll(f, []byte("\n__CTO__"), []byte{})
//	err = ioutil.WriteFile(EnvFile, newContent, os.ModePerm)
//	if err != nil {
//		log.Fatal(err)
//	}
//}

//func UpdateConfigurationFile() {
//	f, err := ioutil.ReadFile(ConfigurationFile)
//	if err != nil {
//		log.Fatal(err)
//	}
//	newContent := bytes.ReplaceAll(f, []byte("Port string `env:\"PORT\"`"), []byte("Port int `env:\"PORT\"`"))
//	err = ioutil.WriteFile(ConfigurationFile, newContent, os.ModePerm)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
//func FixConfigurationFile() {
//	f, err := ioutil.ReadFile(ConfigurationFile)
//	if err != nil {
//		log.Fatal(err)
//	}
//	newContent := bytes.ReplaceAll(f, []byte("Port int `env:\"PORT\"`"), []byte("Port string `env:\"PORT\"`"))
//	err = ioutil.WriteFile(ConfigurationFile, newContent, os.ModePerm)
//	if err != nil {
//		log.Fatal(err)
//	}
//}

//func TestMain(m *testing.M) {
//	FixEnvFile()
//	//FixConfigurationFile()
//	os.Exit(m.Run())
//}


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
		ConnTimeout:  time.Duration(App.Config.Mongo.ConnectionTimeout),
		ReadTimeout:  time.Duration(App.Config.Mongo.ReadTimeout),
		WriteTimeout: time.Duration(App.Config.Mongo.WriteTimeout),
		MaxConnIdleTime: time.Duration(App.Config.Mongo.MaxConnIdleTime),
		MaxPoolSize: uint64(App.Config.Mongo.MaxPoolSize),
		MinPoolSize: uint64(App.Config.Mongo.MinPoolSize),
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("NewOrderRepository Mongo: %v", err.Error())
		panic("mongo adapter creation failed, " + err.Error())
	}

	global.Singletons.OrderRepository ,err = order_repository.NewOrderRepository(mongoDriver)
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
	global.Singletons.StockService = stock_service.NewStockService(App.Config.StockService.Address, App.Config.StockService.Port)
	global.Singletons.PaymentService = payment_service.NewPaymentService(App.Config.PaymentGatewayService.Address,
		App.Config.PaymentGatewayService.Port)
	global.Singletons.NotifyService = notify_service.NewNotificationService()

	go App.grpcServer.Start()
}



func createRequestNewOrder() *pb.RequestNewOrder {
	order := &pb.RequestNewOrder{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance: &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Amount.Total = 600000
	order.Amount.Payable = 550000
	order.Amount.Discount = 50000
	order.Amount.Currency = "RR"
	order.Amount.PaymentMethod = "IPG"
	order.Amount.PaymentOption = "AAP"
	order.Amount.ShipmentTotal = 700000
	order.Amount.Voucher = &pb.Voucher{
		Amount: 40000,
		Code: "348",
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

	item := pb.Item {
		Price:    &pb.PriceInfo{},
		Shipment: &pb.ShippingSpec{},
		Attributes: make(map[string]string, 10),
		SellerId: "123456",
	}

	item.InventoryId = "11111-22222"
	item.Brand = "Asus"
	item.Categories = "Electronic/laptop"
	item.Title = "Asus G503 i7, 256SSD, 32G Ram"
	item.Guarantee = "ضمانت سلامت کالا"
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

	item.Price.Discount = 200000
	item.Price.Payable = 20000000
	item.Price.SellerCommission = 10
	item.Price.Unit = 100000
	item.Price.Currency = "RR"

	//Standard, Express, Economy or Sameday.
	item.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item.Shipment.ShippingTime = 72
	item.Shipment.ReturnTime = 72
	item.Shipment.ReactionTime = 24
	item.Shipment.CarrierName = "Post"
	item.Shipment.CarrierProduct = "Post Express"
	item.Shipment.CarrierType = "standard"
	item.Shipment.ShippingAmount = 100000
	item.Shipment.VoucherAmount = 0
	item.Shipment.Currency = "RR"

	order.Items = append(order.Items, &item)

	item1 := pb.Item {
		Price:    &pb.PriceInfo{},
		Shipment: &pb.ShippingSpec{},
		Attributes: make(map[string]string, 10),
		SellerId: "678912",
	}

	item1.InventoryId = "11111-33333"
	item1.Brand = "Lenovo"
	item1.Categories = "Electronic/laptop"
	item1.Title = "Lenove G503 i7, 256SSD, 32G Ram"
	item1.Guarantee = "ضمانت سلامت کالا"
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

	item1.Price.Discount = 250000
	item1.Price.Payable = 200000
	item1.Price.SellerCommission = 10
	item1.Price.Unit = 200000
	item1.Price.Currency = "RR"

	//Standard, Express, Economy or Sameday.
	item1.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item1.Shipment.ShippingTime = 72
	item1.Shipment.ReturnTime = 72
	item1.Shipment.ReactionTime = 24
	item1.Shipment.CarrierName = "Post"
	item1.Shipment.CarrierProduct = "Post Express"
	item1.Shipment.CarrierType = "standard"
	item1.Shipment.ShippingAmount = 100000
	item1.Shipment.VoucherAmount = 0
	item1.Shipment.Currency = "RR"

	order.Items = append(order.Items, &item1)
	return order
}

//func TestLoadConfig_AssertTrue(t *testing.T) {
//	err := os.Setenv("APP_ENV", "dev")
//	assert.Nil(t, err)
//	_, err = configs.LoadConfig("")
//	assert.Nil(t, err)
//}

func TestNewOrderRequest(t *testing.T) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address + ":" +
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure())
	assert.Nil(t, err)

	requestNewOrder := createRequestNewOrder()

	OrderService := pb.NewOrderServiceClient(grpcConn)
	resOrder, err := OrderService.NewOrder(ctx, requestNewOrder)

	assert.Nil(t, err)
	assert.NotEmpty(t, resOrder.CallbackUrl, "CallbackUrl is empty")
}

func TestPaymentGateway(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address + ":" +
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure())
	assert.Nil(t, err, "DialContext failed")

	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	newOrder.PaymentService = []entities.PaymentService{{
		PaymentRequest: &entities.PaymentRequest{
			Amount:    newOrder.Amount.Total,
			Currency:  "RR",
			Gateway:   "APP",
			CreatedAt: time.Now().UTC(),
		},
	}}

	order , err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	request :=  pg.PaygateHookRequest {
		OrderID: order.OrderId,
		PaymentId: "534545345",
		InvoiceId: 3434234234,
		Amount:	int64(order.Amount.Total),
		ReqBody: "request test url",
		ResBody: "response test url",
		CardMask: "293488374****7234",
		Result:	true,
	}

	paymentService := pg.NewBankResultHookClient(grpcConn)
	response, err := paymentService.PaymentGatewayHook(ctx, &request)

	assert.Nil(t, err)
	assert.True(t, response.Ok, "payment result false")
}

func TestSellerApprovalPending_Success(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address + ":" +
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure())
	assert.Nil(t, err)

	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	order , err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	for i:=0 ; i < len(order.Items); i++ {
		order.Items[i].Status = "20.Seller_Approval_Pending"
	}
	_ , err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	request := pb.RequestSellerOrderAction {
		OrderId: order.OrderId,
		SellerId: order.Items[0].SellerInfo.SellerId,
		ActionType: "Approved",
		Action: "success",
		Data: 	&pb.RequestSellerOrderAction_Success{
			Success: &pb.RequestSellerOrderActionSuccess{ShipmentMethod: "Post", TrackingId: "839832742"},
		},
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.SellerOrderAction(ctx, &request)

	assert.Nil(t, err)
	assert.True(t, result.Result)
}

func TestSellerApprovalPending_Failed(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	grpcConn, err := grpc.DialContext(ctx, App.Config.GRPCServer.Address + ":" +
		strconv.Itoa(int(App.Config.GRPCServer.Port)), grpc.WithInsecure())
	assert.Nil(t, err)

	requestNewOrder := createRequestNewOrder()
	value, err := global.Singletons.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	order , err := global.Singletons.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	for i:=0 ; i < len(order.Items); i++ {
		order.Items[i].Status = "20.Seller_Approval_Pending"
	}
	_ , err = global.Singletons.OrderRepository.Save(*order)
	assert.Nil(t, err, "save failed")

	request := pb.RequestSellerOrderAction {
		OrderId: order.OrderId,
		SellerId: order.Items[0].SellerInfo.SellerId,
		ActionType: "approved",
		Action: "failed",
		Data: 	&pb.RequestSellerOrderAction_Failed{
			Failed: &pb.RequestSellerOrderActionFailed{Reason: "Not Enough Stuff"},
		},
	}

	OrderService := pb.NewOrderServiceClient(grpcConn)
	result, err := OrderService.SellerOrderAction(ctx, &request)

	assert.Nil(t, err)
	assert.True(t, result.Result)
}