package stock_service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"os"
	"testing"
	"time"
)

var config *configs.Cfg
var stock *iStockServiceImpl

func createOrder() *entities.Order {

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
		Status:    "New",
		BuyerInfo: buyerInfo,
		Invoice: entities.Invoice{
			GrandTotal:     75400000,
			Subtotal:       73000000,
			Discount:       15600000,
			Currency:       "IRR",
			ShipmentTotal:  5700000,
			PaymentMethod:  "IPG",
			PaymentGateway: "APP",
			CartRule:       nil,
			Voucher: &entities.Voucher{
				Amount: 230000,
				Code:   "Market",
				Details: &entities.VoucherDetails{
					StartDate:        time.Now().UTC(),
					EndDate:          time.Now().UTC(),
					Type:             "Value",
					MaxDiscountValue: 1000,
					MinBasketValue:   13450,
				},
			},
		},
		Packages: []entities.PackageItem{
			{
				SellerId: 129384234,
				OrderId:  0,
				Version:  0,
				Invoice: entities.PackageInvoice{
					Subtotal:       2873423,
					Discount:       9283443,
					ShipmentAmount: 98734,
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
					ShippingCost:   1249348,
					VoucherAmount:  3242344,
					Currency:       "IRR",
					ReactionTime:   2,
					ShippingTime:   8,
					ReturnTime:     24,
					Details:        "no return",
				},
				Subpackages: []entities.Subpackage{
					{
						ItemId:   0,
						SellerId: 129384234,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
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
									Unit:              1270000,
									Total:             7450000,
									Original:          1270000,
									Special:           1000000,
									Discount:          23000,
									SellerCommission:  5334444,
									Currency:          "IRR",
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
									Unit:              3270000,
									Total:             87450000,
									Original:          21270000,
									Special:           10000000,
									Discount:          230000,
									SellerCommission:  5334444,
									Currency:          "IRR",
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
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
					{
						ItemId:   0,
						SellerId: 129384234,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
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
									Unit:              1270000,
									Total:             8750000,
									Original:          1270000,
									Special:           1000000,
									Discount:          2355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
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
									Unit:              3270000,
									Total:             12750000,
									Original:          2270000,
									Special:           100000,
									Discount:          2355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
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
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
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
				SellerId: 99988887777,
				OrderId:  0,
				Version:  0,
				Invoice: entities.PackageInvoice{
					Subtotal:       2873423,
					Discount:       9283443,
					ShipmentAmount: 98734,
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
					ShippingCost:   1249348,
					VoucherAmount:  3242344,
					Currency:       "IRR",
					ReactionTime:   2,
					ShippingTime:   8,
					ReturnTime:     24,
					Details:        "no return",
				},
				Subpackages: []entities.Subpackage{
					{
						ItemId:   0,
						SellerId: 99988887777,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
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
									Unit:              1270000,
									Total:             7340000,
									Original:          1270000,
									Special:           1000000,
									Discount:          23000,
									SellerCommission:  5334444,
									Currency:          "IRR",
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
									Unit:              2270000,
									Total:             6340000,
									Original:          4270000,
									Special:           100000,
									Discount:          2343000,
									SellerCommission:  533444,
									Currency:          "IRR",
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
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
					{
						ItemId:   0,
						SellerId: 99988887777,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
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
									Unit:              1270000,
									Total:             5646700,
									Original:          7340000,
									Special:           1000000,
									Discount:          2355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
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
									Unit:              7270000,
									Total:             4646700,
									Original:          2340000,
									Special:           1000000,
									Discount:          45355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
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
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							History: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
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
		panic("stockService.ConnectToPaymentService() failed")
	}

	defer stock.CloseConnection()

	order := createOrder()

	request := stockProto.StockRequest{
		Quantity:    5,
		InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId,
	}
	_, err := stock.stockService.StockAllocate(ctx, &request)
	assert.Nil(t, err)

	inventories := map[string]int{request.InventoryId: int(request.Quantity)}
	iFuture := stock.BatchStockActions(ctx, inventories, stock_action.New(stock_action.Reserve))
	futureData := iFuture.Get()
	assert.Nil(t, futureData.Error())

	response, err := stock.stockService.StockGet(ctx, &stockProto.GetRequest{InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId})
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
		panic("stockService.ConnectToPaymentService() failed")
	}

	defer func() {
		if err := stock.grpcConnection.Close(); err != nil {
		}
	}()

	order := createOrder()

	request := stockProto.StockRequest{
		Quantity:    5,
		InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId,
	}
	_, err := stock.stockService.StockAllocate(ctx, &request)
	assert.Nil(t, err)

	inventories := map[string]int{request.InventoryId: int(request.Quantity)}
	iFuture := stock.BatchStockActions(ctx, inventories, stock_action.New(stock_action.Reserve))
	futureData := iFuture.Get()
	assert.Nil(t, futureData.Error())

	iFuture = stock.BatchStockActions(ctx, inventories, stock_action.New(stock_action.Settlement))
	futureData = iFuture.Get()
	assert.Nil(t, futureData.Error())

	response, err := stock.stockService.StockGet(ctx, &stockProto.GetRequest{InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId})
	assert.Nil(t, err)
	logger.Audit("stockGet response: available: %d, reserved: %d", response.Available, response.Reserved)
	assert.Equal(t, response.Available, int32(0))
	assert.Equal(t, response.Reserved, int32(0))
}

func TestStockService_ReleaseSuccess(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	if err := stock.ConnectToStockService(); err != nil {
		logger.Err(err.Error())
		panic("stockService.ConnectToPaymentService() failed")
	}

	defer func() {
		if err := stock.grpcConnection.Close(); err != nil {
		}
	}()

	order := createOrder()

	request := stockProto.StockRequest{
		Quantity:    5,
		InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId,
	}
	_, err := stock.stockService.StockAllocate(ctx, &request)
	assert.Nil(t, err)

	inventories := map[string]int{request.InventoryId: int(request.Quantity)}
	iFuture := stock.BatchStockActions(ctx, inventories, stock_action.New(stock_action.Reserve))
	futureData := iFuture.Get()
	assert.Nil(t, futureData.Error())

	iFuture = stock.BatchStockActions(ctx, inventories, stock_action.New(stock_action.Release))
	futureData = iFuture.Get()
	assert.Nil(t, futureData.Error())

	response, err := stock.stockService.StockGet(ctx, &stockProto.GetRequest{InventoryId: order.Packages[0].Subpackages[0].Items[0].InventoryId})
	assert.Nil(t, err)
	logger.Audit("stockGet response: available: %d, reserved: %d", response.Available, response.Reserved)
	assert.Equal(t, response.Available, int32(5))
	assert.Equal(t, response.Reserved, int32(0))
}
