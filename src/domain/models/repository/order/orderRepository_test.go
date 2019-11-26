package order_repository

import (
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

var config *configs.Cfg
var orderRepository IOrderRepository

func init() {
	var err error
	var path string
	if os.Getenv("APP_ENV") == "dev" {
		path = "../../../../testdata/.env"
	} else {
		path = ""
	}

	config, err = configs.LoadConfig(path)
	if err != nil {
		logger.Err(err.Error())
		panic("configs.LoadConfig failed, " + err.Error())
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:     time.Duration(config.Mongo.ConnectionTimeout),
		ReadTimeout:     time.Duration(config.Mongo.ReadTimeout),
		WriteTimeout:    time.Duration(config.Mongo.WriteTimeout),
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime),
		MaxPoolSize:     uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(config.Mongo.MinPoolSize),
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("NewOrderRepository Mongo: %v", err.Error())
		panic("mongo adapter creation failed, " + err.Error())
	}

	orderRepository, err = NewOrderRepository(mongoDriver)
	if err != nil {
		panic("create order repository failed")
	}
}

func TestSaveOrderRepository(t *testing.T) {

	//defer removeCollection()
	order := createOrder()
	//res, _ := json.Marshal(order)
	//logger.Audit("order model: %s",res)
	order1, err := orderRepository.Save(order)
	assert.Nil(t, err, "orderRepository.Save failed")
	assert.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")
}

func TestUpdateOrderRepository(t *testing.T) {

	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Save(order)
	assert.Nil(t, err, "orderRepository.Save failed")
	assert.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")

	order1.BuyerInfo.FirstName = "Siamak"
	order1.BuyerInfo.LastName = "Marjoeee"

	order2, err := orderRepository.Save(*order1)
	assert.Nil(t, err, "orderRepository.Save failed")
	assert.Equal(t, "Siamak", order2.BuyerInfo.FirstName)
	assert.Equal(t, "Marjoeee", order2.BuyerInfo.LastName)
}

func TestUpdateOrderRepository_Failed(t *testing.T) {

	defer removeCollection()
	order := createOrder()
	timeTmp := time.Now().UTC()
	order.DeletedAt = &timeTmp
	order1, err := orderRepository.Save(order)
	assert.Nil(t, err, "orderRepository.Save failed")
	assert.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")

	order1.BuyerInfo.FirstName = "Siamak"
	_, err = orderRepository.Save(*order1)
	assert.Error(t, err)
	assert.Equal(t, err, errorUpdateFailed)
}

func TestInsertOrderRepository_Success(t *testing.T) {
	//defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	assert.Nil(t, err, "orderRepository.Save failed")
	assert.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")
}

func TestInsertOrderRepository_Failed(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	assert.Nil(t, err, "orderRepository.Save failed")
	assert.NotEmpty(t, order1.OrderId, "orderRepository.Save failed, order id not generated")
	_, err1 := orderRepository.Insert(*order1)
	assert.NotNil(t, err1)
}

func TestFindAllOrderRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)

	orders, err := orderRepository.FindAll()
	assert.Nil(t, err)
	assert.Equal(t, len(orders), 3)
}

func TestFindAllWithSortOrderRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)

	orders, err := orderRepository.FindAllWithSort("buyerInfo.firstName", 1)
	assert.Nil(t, err)
	assert.Equal(t, orders[0].BuyerInfo.FirstName, "AAAA")
}

func TestFindAllWithPageAndPerPageRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	orders, _, err := orderRepository.FindAllWithPage(2, 2)
	assert.Nil(t, err)
	assert.Equal(t, len(orders), 1)
}

func TestFindAllWithPageAndPerPageRepository_failed(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	_, _, err = orderRepository.FindAllWithPage(1002, 2000)
	assert.NotNil(t, err)
	assert.Equal(t, err, errorPageNotAvailable)
}

func TestFindAllWithPageAndPerPageAndSortRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)

	orders, _, err := orderRepository.FindAllWithPageAndSort(1, 2, "buyerInfo.firstName", 1)
	assert.Nil(t, err)
	assert.Equal(t, len(orders), 2)
	assert.Equal(t, orders[0].BuyerInfo.FirstName, "AAAA")
}

func TestFindByIdRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	assert.Nil(t, err)
	_, err1 := orderRepository.FindById(order1.OrderId)
	assert.Nil(t, err1)
}

func TestExistsByIdRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	assert.Nil(t, err)
	res, err1 := orderRepository.ExistsById(order1.OrderId)
	assert.Nil(t, err1)
	assert.Equal(t, res, true)
}

func TestCountRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)

	total, err := orderRepository.Count()
	assert.Nil(t, err)
	assert.Equal(t, total, int64(2))
}

func TestDeleteOrderRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	assert.Nil(t, err)
	order1, err1 := orderRepository.Delete(*order1)
	assert.Nil(t, err1)
	assert.NotNil(t, order1.DeletedAt)
}

func TestDeleteAllRepository(t *testing.T) {
	defer removeCollection()
	var order entities.Order
	order = createOrder()
	_, err := orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	err = orderRepository.DeleteAll()
	assert.Nil(t, err)
}

func TestRemoveOrderRepository(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	assert.Nil(t, err)
	err = orderRepository.Remove(*order1)
	assert.Nil(t, err)
}

func TestFindByFilterRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	order.BuyerInfo.FirstName = "Reza"
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "Hosein"
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)

	orders, err := orderRepository.FindByFilter(func() interface{} {
		return bson.D{{"buyerInfo.firstName", "Reza"}, {"deletedAt", nil}}
	})

	assert.Nil(t, err)
	assert.Equal(t, orders[0].BuyerInfo.FirstName, "Reza")
}

func TestFindByFilterWithSortOrderRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)

	orders, err := orderRepository.FindByFilterWithSort(func() (interface{}, string, int) {
		return bson.D{{"buyerInfo.firstName", "AAAA"}, {"deletedAt", nil}}, "buyerInfo.firstName", 1
	})
	assert.Nil(t, err)
	assert.Equal(t, orders[0].BuyerInfo.FirstName, "AAAA")
}

func TestFindByFilterWithPageAndPerPageRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	orders, _, err := orderRepository.FindByFilterWithPage(func() interface{} {
		return bson.D{{}, {"deletedAt", nil}}
	}, 2, 2)
	assert.Nil(t, err)
	assert.Equal(t, len(orders), 1)
}

func TestFindByFilterWithPageAndPerPageRepository_failed(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	_, _, err = orderRepository.FindByFilterWithPage(func() interface{} {
		return bson.D{{}, {"deletedAt", nil}}
	}, 20002, 2000)
	assert.NotNil(t, err)
	assert.Equal(t, err, errorPageNotAvailable)
}

func TestFindByFilterWithPageAndPerPageAndSortRepository_success(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)
	order = createOrder()
	order.BuyerInfo.FirstName = "AAAA"
	_, err = orderRepository.Insert(order)
	assert.Nil(t, err)

	orders, _, err := orderRepository.FindByFilterWithPageAndSort(func() (interface{}, string, int) {
		return bson.D{{}, {"deletedAt", nil}}, "buyerInfo.firstName", 1
	}, 1, 2)
	assert.Nil(t, err)
	assert.Equal(t, len(orders), 2)
	assert.Equal(t, orders[0].BuyerInfo.FirstName, "AAAA")
}

func removeCollection() {
	orderRepository.RemoveAll()
}

func createOrder() entities.Order {
	//currentTime := time.Now().UTC()

	paymentRequest := entities.PaymentRequest{
		Amount:    75400000,
		Currency:  "IRR",
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
		Result:      true,
		Reason:      "",
		PaymentId:   "r3r434ef45d",
		InvoiceId:   12345678946,
		Amount:      75400000,
		ReqBody:     "",
		ResBody:     "",
		CardNumMask: "545498******4553",
		CreatedAt:   time.Now().UTC(),
	}

	buyerInfo := entities.BuyerInfo{
		FirstName:  "Sina",
		LastName:   "Tadayon",
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
			Address:       "Tehran, Narmak, Golestan.st",
			Phone:         "0217734873",
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
		PaymentService: []entities.PaymentService{{
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
			PayToSeller: []entities.PayToSellerInfo{{
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
		BuyerInfo: buyerInfo,
		Invoice: entities.Invoice{
			Total:         75400000,
			Subtotal:      73000000,
			Discount:      15600000,
			Currency:      "IRR",
			ShipmentTotal: 5700000,
			PaymentMethod: "IPG",
			PaymentOption: "APP",
			Voucher: &entities.Voucher{
				Amount:  230000,
				Code:    "Market",
				Details: nil,
			},
		},
		Items: []entities.Item{
			{
				ItemId:      0,
				InventoryId: "1111111111",
				Title:       "Mobile",
				Brand:       "Nokia",
				Guaranty:    "Sazegar",
				Category:    "Electronic",
				Image:       "",
				Returnable:  false,
				DeletedAt:   nil,
				Attributes: map[string]string{
					"Quantity":  "0",
					"Width":     "5cm",
					"Height":    "7cm",
					"Length":    "2m",
					"Weight":    "5kg",
					"Color":     "Blue",
					"Materials": "Stone",
				},
				SellerInfo: entities.SellerInfo{
					SellerId: 129384234,
					Profile: &entities.SellerProfile{
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
				},
				Invoice: entities.ItemInvoice{
					Unit:             1270000,
					Original:         7340000,
					Special:          1000000,
					SellerCommission: 5334444,
					Currency:         "IRR",
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierNames:   "Post",
					CarrierProduct: "Post Express",
					CarrierType:    "Standard",
					ShippingCost:   1249348,
					VoucherAmount:  3242344,
					Currency:       "IRR",
					ReactionTime:   2,
					ShippingTime:   8,
					ReturnTime:     24,
					Details:        "no return",
				},
				ShipmentDetails: entities.ShipmentDetails{
					ShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
					ReturnShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
				},
				Progress: entities.Progress{
					StepName:  "0.NewOrder",
					StepIndex: 0,
					//CurrentState: entities.State {
					//	Name:  "0.New_Order_Process_State",
					//	Index: 0,
					//	Type: "LauncherAction",
					//	Actions: []entities.Action {{
					//		Name: "Success",
					//		Type: "NewOrder",
					//		Base: "Active",
					//		Data: nil,
					//		Time: &currentTime,
					//	}},
					//	AcceptedAction:entities.Action {
					//		Name: "Success",
					//		Type: "NewOrder",
					//		Base: "Active",
					//		Data: nil,
					//		Time: &currentTime,
					//	},
					//	Result: false,
					//	Reason:       "",
					//	CreatedAt:    time.Now().UTC(),
					//},
					CreatedAt: time.Now().UTC(),
					StepsHistory: []entities.StepHistory{{
						Name:      "0.NewOrder",
						Index:     0,
						CreatedAt: time.Now().UTC(),
						//StatesHistory: []entities.StateHistory{{
						//	Name:  "0.New_Order_Process_State",
						//	Index: 0,
						//	Type: "ListenerAction",
						//	Action: entities.Action{
						//		Name:           "Success",
						//		Type:           "NewOrder",
						//		Base:           "Active",
						//		Data:           nil,
						//		Time: 			&currentTime,
						//	},
						//	Result: 	  false,
						//	Reason:       "",
						//	CreatedAt:    time.Now().UTC(),
						//}},
					}},
				},
			},
			{
				ItemId:      0,
				InventoryId: "2222222222",
				Title:       "Laptop",
				Brand:       "Lenovo",
				Guaranty:    "Iranargham",
				Category:    "Electronic",
				Image:       "",
				Returnable:  true,
				DeletedAt:   nil,
				Attributes: map[string]string{
					"Quantity":  "0",
					"Width":     "5cm",
					"Height":    "7cm",
					"Length":    "2m",
					"Weight":    "5kg",
					"Color":     "Blue",
					"Materials": "Stone",
				},
				SellerInfo: entities.SellerInfo{
					SellerId: 2384723083,
					Profile: &entities.SellerProfile{
						SellerId:        2384723083,
						GeneralInfo:     nil,
						CorporationInfo: nil,
						IndividualInfo:  nil,
						ReturnInfo:      nil,
						ContactPerson:   nil,
						ShipmentInfo:    nil,
						FinanceData:     nil,
						CreatedAt:       time.Now().UTC(),
						UpdatedAt:       time.Now().UTC(),
					},
				},
				Invoice: entities.ItemInvoice{
					Unit:             1270000,
					Original:         7340000,
					Special:          1000000,
					SellerCommission: 5334444,
					Currency:         "IRR",
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierNames:   "Post",
					CarrierProduct: "Post Express",
					CarrierType:    "Standard",
					ShippingCost:   1249348,
					VoucherAmount:  3242344,
					Currency:       "IRR",
					ReactionTime:   2,
					ShippingTime:   8,
					ReturnTime:     24,
					Details:        "no return",
				},
				ShipmentDetails: entities.ShipmentDetails{
					ShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
					ReturnShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
				},
				Progress: entities.Progress{
					StepName:  "0.NewOrder",
					StepIndex: 0,
					//CurrentState: entities.State{
					//	Name:  "0.New_Order_Process_State",
					//	Index: 0,
					//	Actions: []entities.Action{{
					//		Name:           "Success",
					//		Type:           "NewOrder",
					//		Base:           "Active",
					//		Data:           nil,
					//		Time: 			&currentTime,
					//	}},
					//	AcceptedAction: entities.Action{
					//		Name:           "Success",
					//		Type:           "NewOrder",
					//		Base:           "Active",
					//		Data:           nil,
					//		Time: 			&currentTime,
					//	},
					//	Result: false,
					//	Reason:       "",
					//	CreatedAt:    time.Now().UTC(),
					//},
					CreatedAt: time.Now().UTC(),
					StepsHistory: []entities.StepHistory{{
						Name:      "0.NewOrder",
						Index:     0,
						CreatedAt: time.Now().UTC(),
						//StatesHistory: []entities.StateHistory{{
						//	Name:  "0.New_Order_Process_State",
						//	Index: 0,
						//	Type: "ListenerAction",
						//	Action: entities.Action{
						//		Name:           "Success",
						//		Type:           "NewOrder",
						//		Base:           "Active",
						//		Data:           nil,
						//		Time: 			&currentTime,
						//	},
						//
						//	Result: false,
						//	Reason:       "",
						//	CreatedAt:    time.Now().UTC(),
						//}},
					}},
				},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		DeletedAt: nil,
	}

	return newOrder
}
