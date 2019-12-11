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
	order.Invoice.PaymentOption = "AAP"
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

func createMetaDataRequest() *message.RequestMetadata {
	var metadata = &message.RequestMetadata{
		Page:    1,
		PerPage: 25,
		Sorts: []*message.MetaSorts{
			{
				Name:      "mobile",
				Direction: "ASC",
			}, {
				Name:      "name",
				Direction: "DSE",
			},
		},
		Filters: []*message.MetaFilter{
			{
				Type:  "mobile",
				Opt:   "eq",
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
//		//SId: orderId + strconv.Itoa(int(entities.GenerateRandomNumber())),
//		Time: ptypes.TimestampNow(),
//		Meta: metadata,
//		Get: &any.Any{
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
