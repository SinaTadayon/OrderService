package grpc_server

import (
	message "gitlab.faza.io/protos/order"
	pb "gitlab.faza.io/protos/order"

	//"github.com/rs/xid"

	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
)

var config *configs.Cfg

func init() {
	var err error
	config, err = configs.LoadConfig("../../testdata/.env")
	if err != nil {
		logger.Err(err.Error())
		return
	}

	server := NewServer(config.GRPCServer.Address, uint16(config.GRPCServer.Port), nil)
	go server.Start()
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
	order.Amount.Subtotal = 550000
	order.Amount.Discount = 50000
	order.Amount.Currency = "IRR"
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
	item.Price.Total = 20000000
	item.Price.Original = 1800000
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

	item1 := pb.Item {
		Price:    &pb.PriceInfo{},
		Shipment: &pb.ShippingSpec{},
		Attributes: make(map[string]string, 10),
		SellerId: "678912",
	}

	item1.InventoryId = "11111-33333"
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
	item1.Price.Total = 1750000
	item1.Price.Original = 1500000
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

func createMetaDataRequest() *message.RequestMetadata {
	var metadata = &message.RequestMetadata{
		Page:                 1,
		PerPage:              25,
		Sorts:                []*message.MetaSorts{
			{
				Name:      "mobile",
				Direction: 0,
			}, {
				Name:      "name",
				Direction: 1,
			},
		},
		Filters:              []*message.MetaFilter{
			{
				Name: "mobile",
				Opt: "eq",
				Value: "012933434",
			},
		},
	}

	return metadata
}

// Grpc test
//func TestOrderRequestsHandler(t *testing.T) {
//
//	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
//	grpcConnNewOrder, err := grpc.DialContext(ctx, config.GRPCServer.Address + ":" +
//		strconv.Itoa(int(config.GRPCServer.Port)), grpc.WithInsecure())
//	assert.Nil(t, err)
//	OrderService := pb.NewOrderServiceClient(grpcConnNewOrder)
//
//	RequestNewOrder := createRequestNewOrder()
//	metadata := createMetaDataRequest()
//
//	serializedOrder, err := proto.Marshal(RequestNewOrder)
//	if err != nil {
//		logger.Err("could not serialize timestamp")
//	}
//
//	orderId := entities.GenerateOrderId()
//	request := message.MessageRequest {
//		OrderId: orderId,
//		//ItemId: orderId + strconv.Itoa(int(entities.GenerateRandomNumber())),
//		Time: ptypes.TimestampNow(),
//		Meta: metadata,
//		Data: &any.Any{
//			TypeUrl: "baman.io/" + proto.MessageName(RequestNewOrder),
//			Value:   serializedOrder,
//		},
//	}
//
//	resOrder, err := OrderService.OrderRequestsHandler(ctx, &request)
//
//	if err != nil {
//		st := status.Convert(err)
//		for _, detail := range st.Details() {
//			switch t := detail.(type) {
//			case *message.ErrorDetails:
//				fmt.Println("Oops! Your request was rejected by the server.")
//				for _, validate := range t.Validation {
//					fmt.Printf("The %q field was wrong:\n", validate.GetField())
//					fmt.Printf("\t%s\n", validate.GetDesc())
//				}
//			}
//		}
//	}
//
//	//assert.Nil(t, err)
//	assert.NotNil(t, resOrder)
//}

//func TestNewOrder(t *testing.T) {
//
//	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
//	grpcConnNewOrder, err := grpc.DialContext(ctx, config.GRPCServer.Address + ":" +
//		strconv.Itoa(int(config.GRPCServer.Port)), grpc.WithInsecure())
//	assert.Nil(t, err)
//
//	requestNewOrder := createRequestNewOrder()
//
//	defer grpcConnNewOrder.Close()
//	OrderService := pb.NewOrderServiceClient(grpcConnNewOrder)
//	resOrder, err := OrderService.NewOrder(ctx, requestNewOrder)
//
//	assert.Nil(t, err)
//	assert.NotEmpty(t, resOrder.CallbackUrl, "CallbackUrl is empty")
//}
