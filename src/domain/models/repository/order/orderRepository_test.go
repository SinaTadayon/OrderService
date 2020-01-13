package order_repository

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

var orderRepository IOrderRepository

func TestMain(m *testing.M) {
	var err error
	var path string
	if os.Getenv("APP_MODE") == "dev" {
		path = "../../../../testdata/.env"
	} else {
		path = ""
	}

	config, _, err := configs.LoadConfigs(path, "")
	if err != nil {
		logger.Err("configs.LoadConfig failed, %s", err.Error())
		os.Exit(1)
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:     time.Duration(config.Mongo.ConnectionTimeout) * time.Second,
		ReadTimeout:     time.Duration(config.Mongo.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(config.Mongo.WriteTimeout) * time.Second,
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime) * time.Second,
		MaxPoolSize:     uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(config.Mongo.MinPoolSize),
		WriteConcernW:   config.Mongo.WriteConcernW,
		WriteConcernJ:   config.Mongo.WriteConcernJ,
		RetryWrites:     config.Mongo.RetryWrite,
	}

	mongoAdapter, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("IPkgItemRepository Mongo: %v", err.Error())
		os.Exit(1)
	}

	orderRepository = NewOrderRepository(mongoAdapter)

	// Running Tests
	code := m.Run()
	//removeCollection()
	os.Exit(code)
}

func TestSaveOrderRepository(t *testing.T) {

	//defer removeCollection()
	order := createOrder()
	//res, _ := json.Marshal(order)
	//logger.Audit("order model: %s",res)
	ctx, _ := context.WithCancel(context.Background())
	order1, err := orderRepository.Save(ctx, *order)
	require.Nil(t, err, "orderRepository.Save failed")
	require.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")
}

func TestUpdateOrderRepository(t *testing.T) {

	defer removeCollection()
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	order1, err := orderRepository.Save(ctx, *order)
	require.Nil(t, err, "orderRepository.Save failed")
	require.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")

	order1.BuyerInfo.FirstName = "Siamak"
	order1.BuyerInfo.LastName = "Marjoeee"

	order2, err := orderRepository.Save(ctx, *order1)
	require.Nil(t, err, "orderRepository.Save failed")
	require.Equal(t, order2.BuyerInfo.FirstName, "Siamak")
	require.Equal(t, order2.BuyerInfo.LastName, "Marjoeee")
}

func TestUpdateOrderRepository_Failed(t *testing.T) {

	defer removeCollection()
	order := createOrder()
	timeTmp := time.Now().UTC()
	order.DeletedAt = &timeTmp
	ctx, _ := context.WithCancel(context.Background())
	order1, err := orderRepository.Save(ctx, *order)
	require.Nil(t, err, "orderRepository.Save failed")
	require.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")

	order1.BuyerInfo.FirstName = "Siamak"
	_, err = orderRepository.Save(ctx, *order1)
	require.Error(t, err)
	//require.Equal(t, repository.ErrorUpdateFailed, err.Reason())
}

func TestInsertOrderRepository_Success(t *testing.T) {
	//defer removeCollection()
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	order, err := orderRepository.Insert(ctx, *order)
	require.Nil(t, err, "orderRepository.Save failed")
	require.NotEmpty(t, order.OrderId, "orderRepository.Save failed, order id not generated")
}

func TestInsertOrderRepository_Failed(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	order, err := orderRepository.Insert(ctx, *order)
	require.Nil(t, err, "orderRepository.Save failed")
	require.NotEmpty(t, order.OrderId, "orderRepository.Save failed, order id not generated")
	_, err1 := orderRepository.Insert(ctx, *order)
	require.NotNil(t, err1)
}

func TestFindAllOrderRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)

	orders, err := orderRepository.FindAll(ctx)
	require.Nil(t, err)
	require.Equal(t, 3, len(orders))
}

func TestFindAllWithSortOrderRepository(t *testing.T) {

	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)

	orders, err := orderRepository.FindAllWithSort(ctx, "buyerInfo.firstName", 1)
	require.Nil(t, err)
	require.Equal(t, orders[0].BuyerInfo.FirstName, "AAAA")
}

func TestFindAllWithPageAndPerPageRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	orders, _, err := orderRepository.FindAllWithPage(ctx, 2, 2)
	require.Nil(t, err)
	require.Equal(t, 1, len(orders))
}

func TestFindAllWithPageAndPerPageRepository_failed(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	_, _, e := orderRepository.FindAllWithPage(ctx, 1002, 2000)
	require.NotNil(t, e)
	require.Equal(t, repository.ErrorPageNotAvailable, e.Reason())
}

func TestFindAllWithPageAndPerPageAndSortRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)

	orders, _, err := orderRepository.FindAllWithPageAndSort(ctx, 1, 2, "buyerInfo.firstName", 1)
	require.Nil(t, err)
	require.Equal(t, 2, len(orders))
	require.Equal(t, "AAAA", orders[0].BuyerInfo.FirstName)
}

func TestFindByIdRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	order, err := orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	_, err1 := orderRepository.FindById(ctx, order.OrderId)
	require.Nil(t, err1)
}

func TestExistsByIdRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	order, err := orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	res, err1 := orderRepository.ExistsById(ctx, order.OrderId)
	require.Nil(t, err1)
	require.Equal(t, true, res)
}

func TestCountRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)

	total, err := orderRepository.Count(ctx)
	require.Nil(t, err)
	require.Equal(t, int64(2), total)
}

func TestDeleteOrderRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	order, err := orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order1, err1 := orderRepository.Delete(ctx, *order)
	require.Nil(t, err1)
	require.NotNil(t, order1.DeletedAt)
}

func TestDeleteAllRepository(t *testing.T) {
	defer removeCollection()
	var order *entities.Order
	order = createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err := orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	err = orderRepository.DeleteAll(ctx)
	require.Nil(t, err)
}

func TestRemoveOrderRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	order, err := orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	err = orderRepository.Remove(ctx, *order)
	require.Nil(t, err)
}

func TestFindByFilterRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	order.BuyerInfo.FirstName = "Reza"
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "Hosein"
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)

	orders, err := orderRepository.FindByFilter(ctx, func() interface{} {
		return bson.D{{"buyerInfo.firstName", "Reza"}, {"deletedAt", nil}}
	})

	require.Nil(t, err)
	require.Equal(t, "Reza", orders[0].BuyerInfo.FirstName)
}

func TestFindByFilterWithSortOrderRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)

	orders, err := orderRepository.FindByFilterWithSort(ctx, func() (interface{}, string, int) {
		return bson.D{{"buyerInfo.firstName", "AAAA"}, {"deletedAt", nil}}, "buyerInfo.firstName", 1
	})
	require.Nil(t, err)
	require.Equal(t, "AAAA", orders[0].BuyerInfo.FirstName)
}

func TestFindByFilterWithPageAndPerPageRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	orders, _, err := orderRepository.FindByFilterWithPage(ctx, func() interface{} {
		return bson.D{{}, {"deletedAt", nil}}
	}, 2, 2)
	require.Nil(t, err)
	require.Equal(t, 1, len(orders))
}

func TestFindByFilterWithPageAndPerPageRepository_failed(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	_, _, e := orderRepository.FindByFilterWithPage(ctx, func() interface{} {
		return bson.D{{}, {"deletedAt", nil}}
	}, 20002, 2000)
	require.NotNil(t, e)
	require.Equal(t, repository.ErrorPageNotAvailable, e.Reason())
}

func TestFindByFilterWithPageAndPerPageAndSortRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	ctx, _ := context.WithCancel(context.Background())
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(ctx, *order)
	require.Nil(t, err)

	orders, _, err := orderRepository.FindByFilterWithPageAndSort(ctx, func() (interface{}, string, int) {
		return bson.D{{}, {"deletedAt", nil}}, "buyerInfo.firstName", 1
	}, 1, 2)
	require.Nil(t, err)
	require.Equal(t, 2, len(orders))
	require.Equal(t, "AAAA", orders[0].BuyerInfo.FirstName)
}

func removeCollection() {
	ctx, _ := context.WithCancel(context.Background())
	if err := orderRepository.RemoveAll(ctx); err != nil {
	}
}

func createOrder() *entities.Order {

	paymentRequest := entities.PaymentRequest{
		Price: &entities.Money{
			Amount:   "75400000",
			Currency: "IRR",
		},
		Gateway:   "AAP",
		CreatedAt: time.Now().UTC(),
	}

	paymentResponse := entities.PaymentResponse{
		Result:      true,
		Reason:      "",
		Description: "",
		CallBackUrl: "http://baman.io/payment-service",
		InvoiceId:   12345678946,
		PaymentId:   "r3r434ef45d",
		CreatedAt:   time.Now().UTC(),
	}

	paymentResult := entities.PaymentResult{
		Result:    true,
		Reason:    "",
		PaymentId: "r3r434ef45d",
		InvoiceId: 12345678946,
		Price: &entities.Money{
			Amount:   "75400000",
			Currency: "IRR",
		},
		CardNumMask: "545498******4553",
		CreatedAt:   time.Now().UTC(),
	}

	buyerInfo := entities.BuyerInfo{
		BuyerId:    6453563,
		FirstName:  "Sina",
		LastName:   "Tadayon",
		Phone:      "09124234234",
		Mobile:     "09123343534",
		Email:      "sina.tadayon@baman.io",
		NationalId: "00598342521",
		Gender:     "male",
		IP:         "127.0.0.1",
		FinanceInfo: entities.FinanceInfo{
			Iban:          "IR9450345802934803",
			CardNumber:    "4444555533332222",
			AccountNumber: "293.6000.9439283.1",
			BankName:      "passargad",
		},
		ShippingAddress: entities.AddressInfo{
			FirstName:     "Ali Reza",
			LastName:      "Rastegar",
			Address:       "Tehran, Narmak, Golestan.st",
			Phone:         "0217734873",
			Mobile:        "091284345",
			Country:       "Iran",
			City:          "Tehran",
			Province:      "Tehran",
			Neighbourhood: "Chizar",
			Location: &entities.Location{
				Type:        "Point",
				Coordinates: []float64{-72.7738706, 41.6332836},
			},
			ZipCode: "1645630586",
		},
	}

	newOrder := entities.Order{
		OrderId: 0,
		Version: 0,
		OrderPayment: []entities.PaymentService{{
			PaymentRequest:  &paymentRequest,
			PaymentResponse: &paymentResponse,
			PaymentResult:   &paymentResult,
		}},
		SystemPayment: entities.SystemPayment{
			PayToBuyer: []entities.PayToBuyerInfo{{
				PaymentRequest:  &paymentRequest,
				PaymentResponse: &paymentResponse,
				PaymentResult:   &paymentResult,
			}},
			PayToMarket: []entities.PayToMarket{{
				PaymentRequest:  &paymentRequest,
				PaymentResponse: &paymentResponse,
				PaymentResult:   &paymentResult,
			}},
		},
		Status:    "New",
		BuyerInfo: buyerInfo,
		Invoice: entities.Invoice{
			GrandTotal: entities.Money{
				Amount:   "75400000",
				Currency: "IRR",
			},

			Subtotal: entities.Money{
				Amount:   "73000000",
				Currency: "IRR",
			},

			Discount: entities.Money{
				Amount:   "15600000",
				Currency: "IRR",
			},

			ShipmentTotal: entities.Money{
				Amount:   "5700000",
				Currency: "IRR",
			},
			PaymentMethod:  "IPG",
			PaymentGateway: "APP",
			PaymentOption:  nil,
			CartRule:       nil,
			Voucher: &entities.Voucher{
				Percent: 0,
				Price: &entities.Money{
					Amount:   "230000",
					Currency: "IRR",
				},
				Code: "Market",
				Details: &entities.VoucherDetails{
					StartDate:        time.Now().UTC(),
					EndDate:          time.Now().UTC(),
					Type:             "Value",
					MaxDiscountValue: 1000,
					MinBasketValue:   13450,
				},
			},
		},
		Packages: []*entities.PackageItem{
			{
				PId:      129384234,
				OrderId:  0,
				Version:  0,
				ShopName: "Sazagar",
				Invoice: entities.PackageInvoice{
					Subtotal: entities.Money{
						Amount:   "2873423",
						Currency: "IRR",
					},

					Discount: entities.Money{
						Amount:   "9283443",
						Currency: "IRR",
					},

					ShipmentAmount: entities.Money{
						Amount:   "98734",
						Currency: "IRR",
					},
				},
				SellerInfo: &entities.SellerProfile{
					SellerId: 129384234,
					GeneralInfo: &entities.GeneralSellerInfo{
						ShopDisplayName:          "Sazgar",
						Type:                     "",
						Email:                    "info@sazgar.com",
						LandPhone:                "02834709",
						MobilePhone:              "1836491827346",
						Website:                  "www.sazgar.com",
						Province:                 "tehran",
						City:                     "tehran",
						Neighborhood:             "joradan",
						PostalAddress:            "jordan, shaghayegh",
						PostalCode:               "1254754",
						IsVATObliged:             false,
						VATCertificationImageURL: "http://test.faza.io",
					},
					CorporationInfo: &entities.CorporateSellerInfo{
						CompanyRegisteredName:     "avazhang",
						CompanyRegistrationNumber: "10237128366",
						CompanyRationalId:         "1823128434",
						TradeNumber:               "19293712937",
					},
					IndividualInfo: &entities.IndividualSellerInfo{
						FirstName:          "Sazgar",
						FamilyName:         "Sazgar",
						NationalId:         "3254534334",
						NationalIdFrontURL: "http://adkuhfadlf",
						NationalIdBackURL:  "http://adkuhfadlf",
					},
					ReturnInfo: &entities.ReturnInfo{
						Country:       "Iran",
						Province:      "Tehran",
						City:          "Tehran",
						Neighborhood:  "Tehran",
						PostalAddress: "joradan",
						PostalCode:    "28349394332",
					},
					ContactPerson: &entities.SellerContactPerson{
						FirstName:   "Sazgar",
						FamilyName:  "Sazgar",
						MobilePhone: "9324729348",
						Email:       "sazgar@sazgar.com",
					},
					ShipmentInfo: &entities.SellerShipmentInfo{
						SameCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
						DifferentCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
					},
					FinanceData: &entities.SellerFinanceData{
						Iban:                    "405872058724850",
						AccountHolderFirstName:  "sazgar",
						AccountHolderFamilyName: "sazgar",
					},
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierNames:   []string{"Post", "Snap"},
					CarrierProduct: "Post Express",
					CarrierType:    "Standard",
					ShippingCost: &entities.Money{
						Amount:   "1249348",
						Currency: "IRR",
					},
					ReactionTime: 2,
					ShippingTime: 8,
					ReturnTime:   24,
					Details:      "no return",
				},
				Subpackages: []*entities.Subpackage{
					{
						SId:     0,
						PId:     129384234,
						OrderId: 0,
						Version: 0,
						Items: []*entities.Item{
							{
								SKU:         "yt545-34",
								InventoryId: "1111111111",
								Title:       "Mobile",
								Brand:       "Nokia",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    5,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "0",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "1270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "7450000",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "1270000",
										Currency: "IRR",
									},

									Special: entities.Money{
										Amount:   "1000000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "23000",
										Currency: "IRR",
									},

									SellerCommission:  5334444,
									ApplicableVoucher: false,
								},
							},
							{
								SKU:         "ye-564634",
								InventoryId: "11111999999",
								Title:       "TV",
								Brand:       "Nokia",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    2,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "2",
									"Width":     "120cm",
									"Height":    "110cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "3270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "87450000",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "21270000",
										Currency: "IRR",
									},

									Special: entities.Money{
										Amount:   "10000000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "230000",
										Currency: "IRR",
									},

									SellerCommission:  87,
									ApplicableVoucher: false,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								RequestedAt:    nil,
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							State: &entities.State{
								Name:  "1.New",
								Index: 1,
								Data:  nil,
								Actions: []entities.Action{
									{
										Name:      "BuyerCancel",
										UTP:       "OrderBuyerCancel",
										Result:    "Success",
										Reasons:   nil,
										CreatedAt: time.Now().UTC(),
									},
								},
								CreatedAt: time.Now().UTC(),
							},
							Action: &entities.Action{
								Name:      "BuyerCancel",
								UTP:       "OrderBuyerCancel",
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Data:  nil,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											UTP:       "OrderBuyerCancel",
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "1.New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
					{
						SId:     0,
						PId:     129384234,
						OrderId: 0,
						Version: 0,
						Items: []*entities.Item{
							{
								SKU:         "gd534-34344",
								InventoryId: "2222222222",
								Title:       "Laptop",
								Brand:       "Lenovo",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    5,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "0",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "1270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "8750000",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "1270000",
										Currency: "IRR",
									},

									Special: entities.Money{
										Amount:   "1000000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "2355434",
										Currency: "IRR",
									},

									SellerCommission:  56,
									ApplicableVoucher: true,
								},
							},
							{
								SKU:         "5645-yer434",
								InventoryId: "22222888888",
								Title:       "AllInOne",
								Brand:       "Lazada",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    2,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "2",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "3270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "12750000",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "2270000",
										Currency: "IRR",
									},

									Special: entities.Money{
										Amount:   "100000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "2355434",
										Currency: "IRR",
									},

									SellerCommission:  34,
									ApplicableVoucher: true,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								RequestedAt:    nil,
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							State: &entities.State{
								Name:  "1.New",
								Index: 1,
								Data:  nil,
								Actions: []entities.Action{
									{
										Name:      "BuyerCancel",
										UTP:       "OrderBuyerCancel",
										Result:    "Success",
										Reasons:   nil,
										CreatedAt: time.Now().UTC(),
									},
								},
								CreatedAt: time.Now().UTC(),
							},
							Action: &entities.Action{
								Name:      "BuyerCancel",
								UTP:       "OrderBuyerCancel",
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Data:  nil,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											UTP:       "OrderBuyerCancel",
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "1.New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
				},
				Status:    "NEW",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				DeletedAt: nil,
			},
			{
				PId:      99988887777,
				OrderId:  0,
				Version:  0,
				ShopName: "Sazgar",
				Invoice: entities.PackageInvoice{
					Subtotal: entities.Money{
						Amount:   "2873423",
						Currency: "IRR",
					},

					Discount: entities.Money{
						Amount:   "9283443",
						Currency: "IRR",
					},

					ShipmentAmount: entities.Money{
						Amount:   "98734",
						Currency: "IRR",
					},
				},
				SellerInfo: &entities.SellerProfile{
					SellerId: 99988887777,
					GeneralInfo: &entities.GeneralSellerInfo{
						ShopDisplayName:          "Sazgar",
						Type:                     "",
						Email:                    "info@sazgar.com",
						LandPhone:                "02834709",
						MobilePhone:              "1836491827346",
						Website:                  "www.sazgar.com",
						Province:                 "tehran",
						City:                     "tehran",
						Neighborhood:             "joradan",
						PostalAddress:            "jordan, shaghayegh",
						PostalCode:               "1254754",
						IsVATObliged:             false,
						VATCertificationImageURL: "http://test.faza.io",
					},
					CorporationInfo: &entities.CorporateSellerInfo{
						CompanyRegisteredName:     "avazhang",
						CompanyRegistrationNumber: "10237128366",
						CompanyRationalId:         "1823128434",
						TradeNumber:               "19293712937",
					},
					IndividualInfo: &entities.IndividualSellerInfo{
						FirstName:          "Sazgar",
						FamilyName:         "Sazgar",
						NationalId:         "3254534334",
						NationalIdFrontURL: "http://adkuhfadlf",
						NationalIdBackURL:  "http://adkuhfadlf",
					},
					ReturnInfo: &entities.ReturnInfo{
						Country:       "Iran",
						Province:      "Tehran",
						City:          "Tehran",
						Neighborhood:  "Tehran",
						PostalAddress: "joradan",
						PostalCode:    "28349394332",
					},
					ContactPerson: &entities.SellerContactPerson{
						FirstName:   "Sazgar",
						FamilyName:  "Sazgar",
						MobilePhone: "9324729348",
						Email:       "sazgar@sazgar.com",
					},
					ShipmentInfo: &entities.SellerShipmentInfo{
						SameCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
						DifferentCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
					},
					FinanceData: &entities.SellerFinanceData{
						Iban:                    "405872058724850",
						AccountHolderFirstName:  "sazgar",
						AccountHolderFamilyName: "sazgar",
					},
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierNames:   []string{"Post", "Snap"},
					CarrierProduct: "Post Express",
					CarrierType:    "Standard",
					ShippingCost: &entities.Money{
						Amount:   "1249348",
						Currency: "IRR",
					},
					ReactionTime: 2,
					ShippingTime: 8,
					ReturnTime:   24,
					Details:      "no return",
				},
				Subpackages: []*entities.Subpackage{
					{
						SId:     0,
						PId:     99988887777,
						OrderId: 0,
						Version: 0,
						Items: []*entities.Item{
							{
								SKU:         "trrer-5343fdf",
								InventoryId: "55555555555",
								Title:       "Mobile",
								Brand:       "Nokia",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    5,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "0",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "1270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "7340000",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "1270000",
										Currency: "IRR",
									},
									Special: entities.Money{
										Amount:   "1000000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "23000",
										Currency: "IRR",
									},

									SellerCommission:  34,
									ApplicableVoucher: false,
								},
							},
							{
								SKU:         "uer5434-5343",
								InventoryId: "555554444444",
								Title:       "MobileMini",
								Brand:       "Mac",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    3,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "3",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "2270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "6340000",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "4270000",
										Currency: "IRR",
									},

									Special: entities.Money{
										Amount:   "100000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "2343000",
										Currency: "IRR",
									},

									SellerCommission:  23,
									ApplicableVoucher: false,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								RequestedAt:    nil,
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							State: &entities.State{
								Name:  "1.New",
								Index: 1,
								Data:  nil,
								Actions: []entities.Action{
									{
										Name:      "BuyerCancel",
										UTP:       "OrderBuyerCancel",
										Result:    "Success",
										Reasons:   nil,
										CreatedAt: time.Now().UTC(),
									},
								},
								CreatedAt: time.Now().UTC(),
							},
							Action: &entities.Action{
								Name:      "BuyerCancel",
								UTP:       "OrderBuyerCancel",
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Data:  nil,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											UTP:       "OrderBuyerCancel",
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "1.New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
					{
						SId:     0,
						PId:     99988887777,
						OrderId: 0,
						Version: 0,
						Items: []*entities.Item{
							{
								SKU:         "5456",
								InventoryId: "3333333333333",
								Title:       "PC",
								Brand:       "HP",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    3,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "3",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "1270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "5646700",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "7340000",
										Currency: "IRR",
									},

									Special: entities.Money{
										Amount:   "1000000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "2355434",
										Currency: "IRR",
									},

									SellerCommission:  23,
									ApplicableVoucher: true,
								},
							},
							{
								SKU:         "uet-54634",
								InventoryId: "33333666666",
								Title:       "ZeroClient",
								Brand:       "Lardan",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    3,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "3",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit: entities.Money{
										Amount:   "7270000",
										Currency: "IRR",
									},

									Total: entities.Money{
										Amount:   "4646700",
										Currency: "IRR",
									},

									Original: entities.Money{
										Amount:   "2340000",
										Currency: "IRR",
									},

									Special: entities.Money{
										Amount:   "1000000",
										Currency: "IRR",
									},

									Discount: entities.Money{
										Amount:   "45355434",
										Currency: "IRR",
									},

									SellerCommission:  34,
									ApplicableVoucher: true,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								RequestedAt:    nil,
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							State: &entities.State{
								Name:  "1.New",
								Index: 1,
								Data:  nil,
								Actions: []entities.Action{
									{
										Name:      "BuyerCancel",
										UTP:       "OrderBuyerCancel",
										Result:    "Success",
										Reasons:   nil,
										CreatedAt: time.Now().UTC(),
									},
								},
								CreatedAt: time.Now().UTC(),
							},
							Action: &entities.Action{
								Name:      "BuyerCancel",
								UTP:       "OrderBuyerCancel",
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Data:  nil,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											UTP:       "OrderBuyerCancel",
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "1.New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
				},
				Status:    "NEW",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				DeletedAt: nil,
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		DeletedAt: nil,
	}

	return &newOrder
}
