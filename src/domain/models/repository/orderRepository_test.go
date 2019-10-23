package repository

import (
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

var config *configs.Cfg
var orderRepository IOrderRepository

func init() {
	var err error
	config, err = configs.LoadConfigWithPath("../../../testdata/.env")
	if err != nil {
		logger.Err(err.Error())
		return
	}

	orderRepository, err = NewOrderRepository(config)
	if err != nil {
		panic("create order repository failed")
	}
}

func TestSaveOrderRepository(t *testing.T) {

	order := createOrder()
	order1, err := orderRepository.Save(order)
	if err != nil {
		t.Fatal("orderRepository.Save failed", err)
	}

	if len(order1.OrderId) == 0 {
		t.Fatal("orderRepository.Save failed, order id not generated")
	}

	assert.Nil(t, orderRepository.RemoveAll())
}

func TestUpdateOrderRepository(t *testing.T) {

	order := createOrder()
	order1, err := orderRepository.Save(order)
	if err != nil {
		t.Fatal("orderRepository.Save failed", err)
	}

	if len(order1.OrderId) == 0 {
		t.Fatal("orderRepository.Save failed, order id not generated")
	}

	order1.BuyerInfo.FirstName = "Siamak"
	order1.BuyerInfo.LastName = "Marjoeee"

	order2, err := orderRepository.Save(*order1)
	if err != nil {
		t.Fatal("orderRepository.Save failed", err)
	}

	assert.Equal(t, "Siamak", order2.BuyerInfo.FirstName)
	assert.Equal(t, "Marjoeee", order2.BuyerInfo.LastName)
	assert.Nil(t, orderRepository.RemoveAll())
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
	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	if err != nil {
		t.Fatal("orderRepository.Save failed", err)
	}

	if len(order1.OrderId) == 0 {
		t.Fatal("orderRepository.Save failed, order id not generated")
	}
}

func TestInsertOrderRepository_Failed(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	order1, err := orderRepository.Insert(order)
	if err != nil {
		t.Fatal("orderRepository.Save failed", err)
	}

	if len(order1.OrderId) == 0 {
		t.Fatal("orderRepository.Save failed, order id not generated")
	}

	_, err1 := orderRepository.Insert(*order1)
	assert.NotNil(t, err1)
}

func TestFindAllOrderRepository(t *testing.T) {
	defer removeCollection()
	var err error
	order := createOrder()
	_, err = orderRepository.Insert(order)
	order = createOrder()
	_, err = orderRepository.Insert(order)
	order = createOrder()
	_, err = orderRepository.Insert(order)

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
		CreatedAt:   	time.Now().UTC(),
	}

	buyerInfo := entities.BuyerInfo {
		FirstName:			"Sina",
		LastName:   		"Tadayon",
		Mobile:     		"09123343534",
		Email:      		"sina.tadayon@baman.io",
		NationalId: 		"00598342521",
		IP:         		"127.0.0.1",
		Finance:    		entities.FinanceInfo {
			Iban:			"IR9450345802934803",
			CardNumber:		"4444555533332222",
			AccountNumber:	"293.6000.9439283.1",
			BankName:		"passargad",
			Gateway:		"AAP",
		},
		Address:    		entities.AddressInfo {
			Address:		"Tehran, Narmak, Golestan.st",
			Phone:   		"0217734873",
			Country: 		"Iran",
			City: 			"Tehran",
			State: 			"Tehran",
			Lat:			"-72.7738706",
			Lan:			"41.6332836",
			Location:		entities.Location{
				Type:        "point",
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
		Amount: entities.Amount{
			Total:    75400000,
			Payable:  73000000,
			Discount: 15600000,
			Currency: "RR",
		},
		Items: []entities.Item{
			{
				ProductId:  "1111111111",
				Title:      "Mobile",
				Quantity:   1,
				Brand:      "Nokia",
				Warranty:   "Sazegar",
				Categories: "Electronic",
				Image:      "",
				Returnable: false,
				DeletedAt:  nil,
				BuyerInfo:  buyerInfo,
				SellerInfo: entities.SellerInfo{
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
						Gateway:       "AAP",
					},
					Address: entities.AddressInfo{
						Address: "Tehran, Jordan",
						Phone:   "01249874345",
						Country: "Iran",
						City:    "Tehran",
						State:   "Tehran",
						Lat:     "-104.7738706",
						Lan:     "54.6332836",
						Location: entities.Location{
							Type:        "point",
							Coordinates: []float64{-104.7738706, 54.6332836},
						},
						ZipCode: "947534586",
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
				ShipmentSpecInfo: entities.ShipmentSpecInfo{
					ProviderName: "Post",
					ReactionTime: 2,
					ShippingTime: 8,
					ReturnTime:   24,
					Details:      "no return",
				},
				ShipmentDetails: entities.ShipmentDetails{
					ShipmentDetail: entities.ShipmentDetail{
						ShipmentProvider:       "Post",
						ShipmentTrackingNumber: "545349534958349",
						Image:                  "",
						Description:            "",
						CreatedAt:              time.Now().UTC(),
					},
					ReturnShipmentDetail: entities.ReturnShipmentDetail{
						ShipmentProvider:       "Post",
						ShipmentTrackingNumber: "545349534958349",
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
				ProductId:  "2222222222",
				Title:      "Laptop",
				Quantity:   1,
				Brand:      "Lenovo",
				Warranty:   "Iranargham",
				Categories: "Electronic",
				Image:      "",
				Returnable: true,
				DeletedAt:  nil,
				BuyerInfo:  buyerInfo,
				SellerInfo: entities.SellerInfo{
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
						Gateway:       "AAP",
					},
					Address: entities.AddressInfo{
						Address: "Tehran, Jordan",
						Phone:   "01249874345",
						Country: "Iran",
						City:    "Tehran",
						State:   "Tehran",
						Lat:     "-104.7738706",
						Lan:     "54.6332836",
						Location: entities.Location{
							Type:        "point",
							Coordinates: []float64{-104.7738706, 54.6332836},
						},
						ZipCode: "947534586",
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
				ShipmentSpecInfo: entities.ShipmentSpecInfo{
					ProviderName: "Post",
					ReactionTime: 2,
					ShippingTime: 8,
					ReturnTime:   24,
					Details:      "no return",
				},
				ShipmentDetails: entities.ShipmentDetails{
					ShipmentDetail: entities.ShipmentDetail{
						ShipmentProvider:       "Post",
						ShipmentTrackingNumber: "545349534958349",
						Image:                  "",
						Description:            "",
						CreatedAt:              time.Now().UTC(),
					},
					ReturnShipmentDetail: entities.ReturnShipmentDetail{
						ShipmentProvider:       "Post",
						ShipmentTrackingNumber: "545349534958349",
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