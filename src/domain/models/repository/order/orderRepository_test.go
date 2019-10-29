package order_repository

import (
	"encoding/json"
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
		return
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:  time.Duration(config.Mongo.ConnectionTimeout),
		ReadTimeout:  time.Duration(config.Mongo.ReadTimeout),
		WriteTimeout: time.Duration(config.Mongo.WriteTimeout),
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime),
		MaxPoolSize: uint64(config.Mongo.MaxPoolSize),
		MinPoolSize: uint64(config.Mongo.MinPoolSize),
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
	res, _ := json.Marshal(order)
	logger.Audit("order model: %s",res)
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
	orders, _, err := orderRepository.FindAllWithPage(2,2)
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
	_, _, err = orderRepository.FindAllWithPage(1002,2000)
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

	orders, _, err := orderRepository.FindAllWithPageAndSort(1,2, "buyerInfo.firstName", 1)
	assert.Nil(t, err)
	assert.Equal(t, len(orders), 2)
	assert.Equal(t,orders[0].BuyerInfo.FirstName, "AAAA")
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
		return bson.D{{"buyerInfo.firstName", "Reza"},{"deletedAt", nil}}
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
		return bson.D{{"buyerInfo.firstName", "AAAA"},{"deletedAt", nil}}, "buyerInfo.firstName", 1
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
		return bson.D{{},{"deletedAt", nil}}
	}, 2,2)
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
		return bson.D{{},{"deletedAt", nil}}
	}, 20002,2000)
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
		return bson.D{{},{"deletedAt", nil}}, "buyerInfo.firstName", 1},1,2)
	assert.Nil(t, err)
	assert.Equal(t, len(orders), 2)
	assert.Equal(t,orders[0].BuyerInfo.FirstName, "AAAA")
}

func removeCollection() {
	orderRepository.RemoveAll()
}

func createOrder() entities.Order {
	paymentRequest := entities.PaymentRequest {
		Amount:	     	75400000,
		Currency:		"RR",
		Gateway: 		"AAP",
		CreatedAt:   	time.Now().UTC(),
	}

	paymentResponse	:= entities.PaymentResponse {
		Result:			true,
		Reason:			"",
		Description:	"",
		CallBackUrl:	"http://baman.io/payment-service",
		InvoiceId:		12345678946,
		PaymentId:		"r3r434ef45d",
		CreatedAt:   	time.Now().UTC(),
	}

	paymentResult := entities.PaymentResult {
		Result:			true,
		Reason:			"",
		PaymentId:      "r3r434ef45d",
		InvoiceId:		12345678946,
		Amount:    		75400000,
		ReqBody:   		"",
		ResBody:  		"",
		CardNumMask: 	"545498******4553",
		CreatedAt:   	time.Now().UTC(),
	}

	buyerInfo := entities.BuyerInfo {
		FirstName:			"Sina",
		LastName:   		"Tadayon",
		Mobile:     		"09123343534",
		Email:      		"sina.tadayon@baman.io",
		NationalId: 		"00598342521",
		Gender:				"male",
		IP:         		"127.0.0.1",
		FinanceInfo:   		entities.FinanceInfo {
			Iban:			"IR9450345802934803",
			CardNumber:		"4444555533332222",
			AccountNumber:	"293.6000.9439283.1",
			BankName:		"passargad",
		},
		ShippingAddress: 	entities.AddressInfo {
			Address:		"Tehran, Narmak, Golestan.st",
			Phone:   		"0217734873",
			Country: 		"Iran",
			City: 			"Tehran",
			Province: 		"Tehran",
			Neighbourhood:	"Chizar",
			Location:		entities.Location{
				Type:        "Point",
				Coordinates: []float64{-72.7738706, 41.6332836},
			},
			ZipCode: 		"1645630586",
		},
	}

	newOrder := entities.Order{
		OrderId: "",
		PaymentService: []entities.PaymentService{{
			PaymentRequest:  paymentRequest,
			PaymentResponse: paymentResponse,
			PaymentResult:   paymentResult,
		}},
		SystemPayment: entities.SystemPayment{
			PayToBuyerInfo: []entities.PayToBuyerInfo{{
				PaymentRequest:  paymentRequest,
				PaymentResponse: paymentResponse,
				PaymentResult:   paymentResult,
			}},
			PayToSellerInfo: []entities.PayToSellerInfo{{
				PaymentRequest:  paymentRequest,
				PaymentResponse: paymentResponse,
				PaymentResult:   paymentResult,
			}},
			PayToMarket: []entities.PayToMarket{{
				PaymentRequest:  paymentRequest,
				PaymentResponse: paymentResponse,
				PaymentResult:   paymentResult,
			}},
		},
		BuyerInfo: buyerInfo,
		Amount: entities.Amount {
			Total:    75400000,
			Payable:  73000000,
			Discount: 15600000,
			Currency: "RR",
			ShipmentTotal: 5700000,
			PaymentMethod:	"IPG",
			PaymentOption:	"APP",
			Voucher:		&entities.Voucher{
				Amount:  230000,
				Code:    "Market",
				Details: nil,
			},
		},
		Items: []entities.Item{
			{
				ItemId:		 "",
				InventoryId: "1111111111",
				Title:       "Mobile",
				Brand:       "Nokia",
				Warranty:    "Sazegar",
				Categories:  "Electronic",
				Image:       "",
				Returnable:  false,
				DeletedAt:   nil,
				BuyerInfo:   buyerInfo,
				Attributes:	 entities.Attributes{
					Quantity:  0,
					Width:     "5cm",
					Height:    "7cm",
					Length:    "2m",
					Weight:    "5kg",
					Color:     "Blue",
					Materials: "Stone",
					Extra:     nil,
				},
				SellerInfo: entities.SellerInfo{
					SellerId: 		  "129384234",
					Profile:            &entities.SellerProfile {
						Title:            "Sazgar",
						FirstName:        "Shahidi",
						LastName:         "nezhad",
						Mobile:           "019124343",
						Email:            "shahidi@samsong.com",
						NationalId:       "9793287434",
						CompanyName:      "Samservice",
						RegistrationName: "Sazgar",
						EconomicCode:     "342346434343",
						Finance: entities.FinanceInfo{
							Iban:          "IR92347299384782734",
							CardNumber:    "8888777766665555",
							AccountNumber: "983.234.2948723894.2",
							BankName:      "saderat",
						},
						ShippingAddress: entities.AddressInfo{
							Address:       "Tehran, Jordan",
							Phone:         "01249874345",
							Country:       "Iran",
							City:          "Tehran",
							Province:      "Tehran",
							Neighbourhood: "Narmak",
							Location: entities.Location{
								Type:        "Point",
								Coordinates: []float64{-104.7738706, 54.6332836},
							},
							ZipCode: "947534586",
						},
					},
				},
				PriceInfo: entities.PriceInfo{
					Unit:             1270000,
					Total:            8340000,
					Payable:          7340000,
					Discount:         1000000,
					SellerCommission: 5334444,
					Currency:         "RR",
				},
				ShipmentSpec: entities.ShipmentSpec {
					CarrierName:		"Post",
					CarrierProduct:		"Post Express",
					CarrierType:		"Standard",
					ShippingAmount:		1249348,
					VoucherAmount:		3242344,
					Currency:			"RR",
					ReactionTime: 		2,
					ShippingTime: 		8,
					ReturnTime:   		24,
					Details:      		"no return",
				},
				ShipmentDetails: entities.ShipmentDetails{
					SellerShipmentDetail: 	entities.ShipmentDetail{
						CarrierName: 		"Post",
						TrackingNumber:   	"545349534958349",
						Image:            	"",
						Description:      	"",
						CreatedAt:        	time.Now().UTC(),
					},
					BuyerReturnShipmentDetail: entities.ShipmentDetail{
						CarrierName: 			"Post",
						TrackingNumber:   		"545349534958349",
						Image:                  "",
						Description:            "",
						CreatedAt:              time.Now().UTC(),
					},
				},
				OrderStep: entities.OrderStep{
					CurrentName:  "0.NewOrder",
					CurrentIndex: 0,
					CurrentState: entities.State{
						Name:  "0.New_Order_Process_State",
						Index: 0,
						Action: entities.Action{
							Name:           "Success",
							Type:           "NewOrder",
							Base:           "Active",
							Data:           "",
							DispatchedTime: time.Now().UTC(),
						},
						ActionResult: false,
						Reason:       "",
						CreatedAt:    time.Now().UTC(),
					},
					CreatedAt: time.Now().UTC(),
					StepsHistory: []entities.StepHistory{{
						Name:      "0.NewOrder",
						Index:     0,
						CreatedAt: time.Now().UTC(),
						StatesHistory: []entities.State{{
							Name:  "0.New_Order_Process_State",
							Index: 0,
							Action: entities.Action{
								Name:           "Success",
								Type:           "NewOrder",
								Base:           "Active",
								Data:           "",
								DispatchedTime: time.Now().UTC(),
							},
							ActionResult: false,
							Reason:       "",
							CreatedAt:    time.Now().UTC(),
						}},
					}},
				},
			},
			{
				ItemId:		 "",
				InventoryId: "2222222222",
				Title:       "Laptop",
				Brand:       "Lenovo",
				Warranty:    "Iranargham",
				Categories:  "Electronic",
				Image:       "",
				Returnable:  true,
				DeletedAt:   nil,
				BuyerInfo:   buyerInfo,
				Attributes:	 entities.Attributes{
					Quantity:  0,
					Width:     "5cm",
					Height:    "7cm",
					Length:    "2m",
					Weight:    "5kg",
					Color:     "Blue",
					Materials: "Stone",
					Extra:     nil,
				},
				SellerInfo: entities.SellerInfo{
					SellerId:			"2384723083",
					Profile:			&entities.SellerProfile{
						Title:            "Avazhang",
						FirstName:        "Mostafavi",
						LastName:         "Rezaii",
						Mobile:           "0394739844",
						Email:            "mostafavi@samsong.com",
						NationalId:       "39458979455",
						CompanyName:      "Avazhang",
						RegistrationName: "Avazeh",
						EconomicCode:     "3045988273784",
						Finance: entities.FinanceInfo{
							Iban:          "IR209345882374",
							CardNumber:    "92384787263443443",
							AccountNumber: "983.234.2293452434.2",
							BankName:      "saderat",
						},
						ShippingAddress: entities.AddressInfo{
							Address:       "Tehran, Jordan",
							Phone:         "01249874345",
							Country:       "Iran",
							City:          "Tehran",
							Province:      "Tehran",
							Neighbourhood: "Navab",
							Location: entities.Location{
								Type:        "Point",
								Coordinates: []float64{-104.7738706, 54.6332836},
							},
							ZipCode: "947534586",
						},
					},
				},
				PriceInfo: entities.PriceInfo{
					Unit:             1270000,
					Total:            8340000,
					Payable:          7340000,
					Discount:         1000000,
					SellerCommission: 5334444,
					Currency:         "RR",
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierName:		"Post",
					CarrierProduct:		"Post Express",
					CarrierType:		"Standard",
					ShippingAmount:		1249348,
					VoucherAmount:		3242344,
					Currency:			"RR",
					ReactionTime: 		2,
					ShippingTime: 		8,
					ReturnTime:   		24,
					Details:      		"no return",
				},
				ShipmentDetails: entities.ShipmentDetails{
					SellerShipmentDetail: 	entities.ShipmentDetail{
						CarrierName: 		"Post",
						TrackingNumber:   	"545349534958349",
						Image:            	"",
						Description:      	"",
						CreatedAt:        	time.Now().UTC(),
					},
					BuyerReturnShipmentDetail: entities.ShipmentDetail{
						CarrierName: 			"Post",
						TrackingNumber:   		"545349534958349",
						Image:                  "",
						Description:            "",
						CreatedAt:              time.Now().UTC(),
					},
				},
				OrderStep: entities.OrderStep{
					CurrentName:  "0.NewOrder",
					CurrentIndex: 0,
					CurrentState: entities.State{
						Name:  "0.New_Order_Process_State",
						Index: 0,
						Action: entities.Action{
							Name:           "Success",
							Type:           "NewOrder",
							Base:           "Active",
							Data:           "",
							DispatchedTime: time.Now().UTC(),
						},
						ActionResult: false,
						Reason:       "",
						CreatedAt:    time.Now().UTC(),
					},
					CreatedAt: time.Now().UTC(),
					StepsHistory: []entities.StepHistory{{
						Name:      "0.NewOrder",
						Index:     0,
						CreatedAt: time.Now().UTC(),
						StatesHistory: []entities.State{{
							Name:  "0.New_Order_Process_State",
							Index: 0,
							Action: entities.Action{
								Name:           "Success",
								Type:           "NewOrder",
								Base:           "Active",
								Data:           "",
								DispatchedTime: time.Now().UTC(),
							},
							ActionResult: false,
							Reason:       "",
							CreatedAt:    time.Now().UTC(),
						}},
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