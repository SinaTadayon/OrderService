package pkg_repository

import (
	"context"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"strconv"

	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"testing"
	"time"
)

var pkgItemRepo IPkgItemRepository
var mongoAdapter *mongoadapter.Mongo
var config *configs.Config

func TestMain(m *testing.M) {
	var path string
	var err error
	if os.Getenv("APP_MODE") == "dev" {
		path = "../../../../testdata/.env"
	} else {
		path = ""
	}

	applog.GLog.ZapLogger = applog.InitZap()
	applog.GLog.Logger = logger.NewZapLogger(applog.GLog.ZapLogger)

	config, _, err = configs.LoadConfigs(path, "")
	if err != nil {
		applog.GLog.Logger.Error("configs.LoadConfig failed",
			"error", err)
		os.Exit(1)
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		// Host:     config.Mongo.Host,
		// Port:     config.Mongo.Port,
		ConnectUri: config.Mongo.Uri,
		Username:   config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:            time.Duration(config.Mongo.ConnectionTimeout) * time.Second,
		ReadTimeout:            time.Duration(config.Mongo.ReadTimeout) * time.Second,
		WriteTimeout:           time.Duration(config.Mongo.WriteTimeout) * time.Second,
		MaxConnIdleTime:        time.Duration(config.Mongo.MaxConnIdleTime) * time.Second,
		HeartbeatInterval:      time.Duration(config.Mongo.HeartBeatInterval) * time.Second,
		ServerSelectionTimeout: time.Duration(config.Mongo.ServerSelectionTimeout) * time.Second,
		RetryConnect:           uint64(config.Mongo.RetryConnect),
		MaxPoolSize:            uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:            uint64(config.Mongo.MinPoolSize),
		WriteConcernW:          config.Mongo.WriteConcernW,
		WriteConcernJ:          config.Mongo.WriteConcernJ,
		RetryWrites:            config.Mongo.RetryWrite,
	}

	mongoAdapter, err = mongoadapter.NewMongo(mongoConf)
	if err != nil {
		applog.GLog.Logger.Error("mongoadapter.NewMongo failed", "error", err)
		os.Exit(1)
	}

	pkgItemRepo = NewPkgItemRepository(mongoAdapter, config.Mongo.Database, config.Mongo.Collection)

	// Running Tests
	code := m.Run()
	removeCollection()
	os.Exit(code)
}

func TestUpdatePkgItemRepository_Failed(t *testing.T) {

	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")

	ctx, _ := context.WithCancel(context.Background())
	order.Packages[0].Version = 1
	order.Packages[0].Status = "Payment_Pending"
	_, err = pkgItemRepo.Update(ctx, *order.Packages[0])
	require.Error(t, err, "pkgItemRepo.Update failed")
	//require.Equal(t, uint64(1), packageItem.Version)
	//require.Equal(t, "Payment_Pending", packageItem.Status)
}

func TestUpdatePkgItemRepository_Success(t *testing.T) {

	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")

	ctx, _ := context.WithCancel(context.Background())
	order.Packages[1].Status = "Payment_Pending"
	packageItem, err := pkgItemRepo.Update(ctx, *order.Packages[1])
	require.Nil(t, err, "pkgItemRepo.Update failed")
	require.Equal(t, uint64(1), packageItem.Version)
	require.Equal(t, "Payment_Pending", packageItem.Status)
}

func TestUpdatePkgItemWithNewSubPkgRepository(t *testing.T) {

	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	subpackage := &entities.Subpackage{
		SId:     0,
		PId:     order.Packages[0].PId,
		OrderId: order.OrderId,
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
				Attributes:  nil,
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
		},
		Shipments: &entities.Shipment{
			ShipmentDetail: &entities.ShippingDetail{
				CourierName:    "Post",
				ShippingMethod: "Normal",
				TrackingNumber: "545349534958349",
				Image:          "",
				Description:    "",
				ShippedAt:      nil,
				CreatedAt:      time.Now().UTC(),
			},
			ReturnShipmentDetail: &entities.ReturnShippingDetail{
				CourierName:    "Post",
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
				Name:       "1.New",
				Index:      1,
				Schedulers: nil,
				Data:       nil,
				Actions: []entities.Action{
					{
						Name:      "BuyerCancel",
						Type:      "",
						UId:       0,
						UTP:       "OrderBuyerCancel",
						Perm:      "",
						Priv:      "",
						Policy:    "",
						Result:    "Success",
						Reasons:   nil,
						Note:      "",
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					},
				},
				CreatedAt: time.Now().UTC(),
				Extended:  nil,
			},
			Action: &entities.Action{
				Name:      "BuyerCancel",
				Type:      "",
				UId:       0,
				UTP:       "OrderBuyerCancel",
				Perm:      "",
				Priv:      "",
				Policy:    "",
				Result:    "Success",
				Reasons:   nil,
				Note:      "",
				Data:      nil,
				CreatedAt: time.Now().UTC(),
				Extended:  nil,
			},
			History: []entities.State{
				{
					Name:  "1.New",
					Index: 1,
					Data:  nil,
					Actions: []entities.Action{
						{
							Name:      "BuyerCancel",
							Type:      "",
							UId:       0,
							UTP:       "OrderBuyerCancel",
							Perm:      "",
							Priv:      "",
							Policy:    "",
							Result:    "Success",
							Reasons:   nil,
							Note:      "",
							Data:      nil,
							CreatedAt: time.Now().UTC(),
							Extended:  nil,
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
	}
	order.Packages[1].Subpackages = append(order.Packages[1].Subpackages, subpackage)
	ctx, _ := context.WithCancel(context.Background())
	order.Packages[1].Status = "Payment_Pending"
	_, err = pkgItemRepo.Update(ctx, *order.Packages[1])
	require.Nil(t, err, "pkgItemRepo.Update failed")
	packageItem, err := pkgItemRepo.FindById(ctx, order.OrderId, order.Packages[1].PId)
	require.Nil(t, err, "pkgItemRepo.find failed")
	require.Equal(t, 3, len(packageItem.Subpackages))
	require.Equal(t, uint64(1), packageItem.Version)
	require.Equal(t, "Payment_Pending", packageItem.Status)
}

func TestIPkgItemRepositoryImpl_FindPkgItmBuyinfById(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	pkgItem, buyId, err := pkgItemRepo.FindPkgItmBuyinfById(ctx, order.OrderId, order.Packages[0].PId)
	require.Nil(t, err, "pkgItemRepo.FindById failed")
	require.Equal(t, order.BuyerInfo.BuyerId, buyId)
	require.Equal(t, order.Packages[0].PId, pkgItem.PId)
	require.Equal(t, uint64(0), pkgItem.Version)
	require.Equal(t, "NEW", pkgItem.Status)
}

func TestFindById(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	packageItem, err := pkgItemRepo.FindById(ctx, order.OrderId, order.Packages[1].PId)
	require.Nil(t, err, "pkgItemRepo.FindById failed")
	require.Equal(t, order.Packages[1].PId, packageItem.PId)
	require.Equal(t, uint64(0), packageItem.Version)
	require.Equal(t, "NEW", packageItem.Status)
}

func TestExitsById_Success(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	result, err := pkgItemRepo.ExistsById(ctx, order.OrderId, order.Packages[1].PId)
	require.Nil(t, err, "pkgItemRepo.ExistsById failed")
	require.True(t, result)
}

func TestExitsById_Failed(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	result, err := pkgItemRepo.ExistsById(ctx, order.OrderId, 1235356)
	require.Nil(t, err, "pkgItemRepo.ExistsById failed")
	require.False(t, result)
}

func TestCount(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	result, err := pkgItemRepo.Count(ctx, order.Packages[0].PId)
	require.Nil(t, err, "pkgItemRepo.Quantity failed")
	require.Equal(t, int64(1), result)
}

func TestCountWithFilter(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	result, err := pkgItemRepo.CountWithFilter(ctx, func() (filter interface{}) {
		return []bson.M{
			{"$match": bson.M{"packages.pid": order.Packages[1].PId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": order.Packages[1].PId, "deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	})

	require.Nil(t, err, "pkgItemRepo.CountWithFilter failed")
	require.Equal(t, int64(1), result)
}

func TestFindByFilter(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	packageItem, err := pkgItemRepo.FindByFilter(ctx, func() (filter interface{}) {
		return []bson.M{
			{"$match": bson.M{"orderId": order.OrderId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": order.Packages[0].PId}},
			{"$project": bson.M{"_id": 0, "packages": 1}},
			{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		}
	})
	require.Equal(t, order.Packages[0].PId, packageItem[0].PId)
	require.Equal(t, uint64(0), packageItem[0].Version)
	require.Equal(t, "NEW", packageItem[0].Status)
}

func removeCollection() {
	if _, err := mongoAdapter.DeleteMany(config.Mongo.Database, config.Mongo.Collection, bson.M{}); err != nil {
	}
}

func insert(order *entities.Order) (*entities.Order, error) {

	if order.OrderId == 0 {
		order.OrderId = entities.GenerateOrderId()
		mapItemIds := make(map[int]uint64, 64)

		for i := 0; i < len(order.Packages); i++ {
			order.Packages[i].OrderId = order.OrderId
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				for {
					random := int(entities.GenerateRandomNumber())
					if _, ok := mapItemIds[random]; ok {
						continue
					}
					mapItemIds[random] = order.Packages[i].PId
					sid, _ := strconv.Atoi(strconv.Itoa(int(order.OrderId)) + strconv.Itoa(random))
					order.Packages[i].Subpackages[j].SId = uint64(sid)
					order.Packages[i].Subpackages[j].PId = order.Packages[i].PId
					order.Packages[i].Subpackages[j].OrderId = order.OrderId
					order.Packages[i].Subpackages[j].CreatedAt = time.Now().UTC()
					order.Packages[i].Subpackages[j].UpdatedAt = time.Now().UTC()
					break
				}
			}
		}

		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = mongoAdapter.InsertOne(config.Mongo.Database, config.Mongo.Collection, &order)
		if err != nil {
			if mongoAdapter.IsDupError(err) {
				for mongoAdapter.IsDupError(err) {
					insertOneResult, err = mongoAdapter.InsertOne(config.Mongo.Database, config.Mongo.Collection, &order)
				}
			} else {
				return nil, err
			}
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	} else {
		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = mongoAdapter.InsertOne(config.Mongo.Database, config.Mongo.Collection, &order)
		if err != nil {
			return nil, err
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	}
	return order, nil
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
		Response: entities.PaymentIPGResponse{
			CallBackUrl: "http://baman.io/payment-service",
			InvoiceId:   12345678946,
			PaymentId:   "r3r434ef45d",
			Extended:    nil,
		},
		CreatedAt: time.Now().UTC(),
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
					VoucherType:      "PURCHASE",
					VoucherSponsor:   "BAZLIA",
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
								Attributes:  nil,
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
								Attributes:  nil,
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
								CourierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CourierName:    "Post",
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
								Name:       "1.New",
								Index:      1,
								Schedulers: nil,
								Data:       nil,
								Actions: []entities.Action{
									{
										Name:      "BuyerCancel",
										Type:      "",
										UId:       0,
										UTP:       "OrderBuyerCancel",
										Perm:      "",
										Priv:      "",
										Policy:    "",
										Result:    "Success",
										Reasons:   nil,
										Note:      "",
										Data:      nil,
										CreatedAt: time.Now().UTC(),
										Extended:  nil,
									},
								},
								CreatedAt: time.Now().UTC(),
								Extended:  nil,
							},
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "",
								UId:       0,
								UTP:       "OrderBuyerCancel",
								Perm:      "",
								Priv:      "",
								Policy:    "",
								Result:    "Success",
								Reasons:   nil,
								Note:      "",
								Data:      nil,
								CreatedAt: time.Now().UTC(),
								Extended:  nil,
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Data:  nil,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "",
											UId:       0,
											UTP:       "OrderBuyerCancel",
											Perm:      "",
											Priv:      "",
											Policy:    "",
											Result:    "Success",
											Reasons:   nil,
											Note:      "",
											Data:      nil,
											CreatedAt: time.Now().UTC(),
											Extended:  nil,
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
								Attributes:  nil,
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
								Attributes:  nil,
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
								CourierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CourierName:    "Post",
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
								Name:       "1.New",
								Index:      1,
								Schedulers: nil,
								Data:       nil,
								Actions: []entities.Action{
									{
										Name:      "BuyerCancel",
										Type:      "",
										UId:       0,
										UTP:       "OrderBuyerCancel",
										Perm:      "",
										Priv:      "",
										Policy:    "",
										Result:    "Success",
										Reasons:   nil,
										Note:      "",
										Data:      nil,
										CreatedAt: time.Now().UTC(),
										Extended:  nil,
									},
								},
								CreatedAt: time.Now().UTC(),
								Extended:  nil,
							},
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "",
								UId:       0,
								UTP:       "OrderBuyerCancel",
								Perm:      "",
								Priv:      "",
								Policy:    "",
								Result:    "Success",
								Reasons:   nil,
								Note:      "",
								Data:      nil,
								CreatedAt: time.Now().UTC(),
								Extended:  nil,
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Data:  nil,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "",
											UId:       0,
											UTP:       "OrderBuyerCancel",
											Perm:      "",
											Priv:      "",
											Policy:    "",
											Result:    "Success",
											Reasons:   nil,
											Note:      "",
											Data:      nil,
											CreatedAt: time.Now().UTC(),
											Extended:  nil,
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
								Attributes:  nil,
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
								Attributes:  nil,
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
								CourierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CourierName:    "Post",
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
								Name:       "1.New",
								Index:      1,
								Schedulers: nil,
								Data:       nil,
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
								Extended:  nil,
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
								Attributes:  nil,
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
								Attributes:  nil,
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
								CourierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedAt:      nil,
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ReturnShippingDetail{
								CourierName:    "Post",
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
								Name:       "1.New",
								Index:      1,
								Schedulers: nil,
								Data:       nil,
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
								Extended:  nil,
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

func createOrderAndSave() (*entities.Order, error) {
	return insert(createOrder())
}
