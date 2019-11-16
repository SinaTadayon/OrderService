package stock_service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"os"
	"testing"
	"time"
)

var config *configs.Cfg
var stock *iStockServiceImpl

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
		OrderId: 123456789,
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
		Amount: entities.Amount{
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
				ItemId:      123456789123,
				InventoryId: "1111111111",
				Title:       "Mobile",
				Brand:       "Nokia",
				Guaranty:    "Sazegar",
				Category:    "Electronic",
				Image:       "",
				Returnable:  false,
				Quantity:    5,
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
				Price: entities.Price{
					Unit:             1270000,
					Original:         7340000,
					Special:          1000000,
					SellerCommission: 5334444,
					Currency:         "IRR",
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierName:    "Post",
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
					SellerShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
					BuyerReturnShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
				},
				Progress: entities.Progress{
					CurrentStepName:  "0.NewOrder",
					CurrentStepIndex: 0,
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
				ItemId:      123456789567,
				InventoryId: "2222222222",
				Title:       "Laptop",
				Brand:       "Lenovo",
				Guaranty:    "Iranargham",
				Category:    "Electronic",
				Image:       "",
				Returnable:  true,
				Quantity:    5,
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
				Price: entities.Price{
					Unit:             1270000,
					Original:         7340000,
					Special:          1000000,
					SellerCommission: 5334444,
					Currency:         "IRR",
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierName:    "Post",
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
					SellerShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
					BuyerReturnShipmentDetail: entities.ShipmentDetail{
						CarrierName:    "Post",
						TrackingNumber: "545349534958349",
						Image:          "",
						Description:    "",
						CreatedAt:      time.Now().UTC(),
					},
				},
				Progress: entities.Progress{
					CurrentStepName:  "0.NewOrder",
					CurrentStepIndex: 0,
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

func init() {
	var err error
	var path string
	if os.Getenv("APP_ENV") == "dev" {
		path = "../../../testdata/.env"
	} else {
		path = ""
	}

	config, err = configs.LoadConfig(path)
	if err != nil {
		logger.Err(err.Error())
		panic("configs.LoadConfig failed")
	}

	stock = &iStockServiceImpl{nil, nil,
		config.StockService.Address, config.StockService.Port}
}

func TestStockService_ReservedSuccess(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	if err := stock.ConnectToStockService(); err != nil {
		logger.Err(err.Error())
		panic("stockService.ConnectToStockService() failed")
	}

	defer stock.CloseConnection()

	order := createOrder()

	request := stockProto.StockRequest{
		Quantity:    5,
		InventoryId: order.Items[0].InventoryId,
	}
	_, err := stock.stockService.StockAllocate(ctx, &request)
	assert.Nil(t, err)

	itemsId := []uint64{order.Items[0].ItemId}
	promise := stock.BatchStockActions(ctx, order, itemsId, "StockReserved")
	futureData := promise.Data()
	assert.Nil(t, futureData.Ex)

	response, err := stock.stockService.StockGet(ctx, &stockProto.GetRequest{InventoryId: order.Items[0].InventoryId})
	assert.Nil(t, err)
	logger.Audit("stockGet response: available: %d, reserved: %d", response.Available, response.Reserved)
	assert.Equal(t, response.Available, int32(0))
	assert.Equal(t, response.Reserved, int32(5))
	_, err = stock.stockService.StockRelease(ctx, &request)
	assert.Nil(t, err)
}

func TestStockService_SettlementSuccess(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	if err := stock.ConnectToStockService(); err != nil {
		logger.Err(err.Error())
		panic("stockService.ConnectToStockService() failed")
	}

	defer func() {
		if err := stock.grpcConnection.Close(); err != nil {
		}
	}()

	order := createOrder()

	request := stockProto.StockRequest{
		Quantity:    5,
		InventoryId: order.Items[0].InventoryId,
	}
	_, err := stock.stockService.StockAllocate(ctx, &request)
	assert.Nil(t, err)

	itemsId := []uint64{order.Items[0].ItemId}
	promise := stock.BatchStockActions(ctx, order, itemsId, "StockReserved")
	futureData := promise.Data()
	assert.Nil(t, futureData.Ex)

	promise = stock.BatchStockActions(ctx, order, itemsId, "StockSettlement")
	futureData = promise.Data()
	assert.Nil(t, futureData.Ex)

	response, err := stock.stockService.StockGet(ctx, &stockProto.GetRequest{InventoryId: order.Items[0].InventoryId})
	assert.Nil(t, err)
	logger.Audit("stockGet response: available: %d, reserved: %d", response.Available, response.Reserved)
	assert.Equal(t, response.Available, int32(0))
	assert.Equal(t, response.Reserved, int32(0))
}

func TestStockService_ReleaseSuccess(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	if err := stock.ConnectToStockService(); err != nil {
		logger.Err(err.Error())
		panic("stockService.ConnectToStockService() failed")
	}

	defer func() {
		if err := stock.grpcConnection.Close(); err != nil {
		}
	}()

	order := createOrder()

	request := stockProto.StockRequest{
		Quantity:    5,
		InventoryId: order.Items[0].InventoryId,
	}
	_, err := stock.stockService.StockAllocate(ctx, &request)
	assert.Nil(t, err)

	itemsId := []uint64{order.Items[0].ItemId}
	promise := stock.BatchStockActions(ctx, order, itemsId, "StockReserved")
	futureData := promise.Data()
	assert.Nil(t, futureData.Ex)

	promise = stock.BatchStockActions(ctx, order, itemsId, "StockReleased")
	futureData = promise.Data()
	assert.Nil(t, futureData.Ex)

	response, err := stock.stockService.StockGet(ctx, &stockProto.GetRequest{InventoryId: order.Items[0].InventoryId})
	assert.Nil(t, err)
	logger.Audit("stockGet response: available: %d, reserved: %d", response.Available, response.Reserved)
	assert.Equal(t, response.Available, int32(5))
	assert.Equal(t, response.Reserved, int32(0))
}
