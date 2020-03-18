package calculate

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var err error
	if os.Getenv("APP_MODE") == "dev" {
		app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfigs("../../../testdata/.env", "")
	} else {
		app.Globals.Config, app.Globals.SMSTemplate, err = configs.LoadConfig("")
	}

	applog.GLog.ZapLogger = app.InitZap()
	applog.GLog.Logger = logger.NewZapLogger(applog.GLog.ZapLogger)

	app.Globals.ZapLogger = applog.GLog.ZapLogger
	app.Globals.Logger = applog.GLog.Logger

	if err != nil {
		applog.GLog.Logger.Error("LoadConfig of failed", "error", err)
		os.Exit(1)
	}

	mongoDriver, err := app.SetupMongoDriver(*app.Globals.Config)
	if err != nil {
		os.Exit(1)
	}

	app.Globals.OrderRepository = order_repository.NewOrderRepository(mongoDriver, app.Globals.Config.Mongo.Database, app.Globals.Config.Mongo.Collection)

	// Running Tests
	code := m.Run()
	os.Exit(code)
}

func createOrder() *entities.Order {

	timestamp := time.Now().UTC()
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
			Share:          nil,
			Voucher: &entities.Voucher{
				Percent: 0,
				AppliedPrice: &entities.Money{
					Amount:   "230000",
					Currency: "IRR",
				},
				RoundupAppliedPrice: &entities.Money{
					Amount:   "230000",
					Currency: "IRR",
				},
				RawShipmentAppliedPrice:     nil,
				RoundupShipmentAppliedPrice: nil,
				Price: &entities.Money{
					Amount:   "230000",
					Currency: "IRR",
				},
				Code: "Market",
				Details: &entities.VoucherDetails{
					Title:            "",
					Prefix:           "",
					UseLimit:         0,
					Count:            0,
					Length:           0,
					Categories:       nil,
					Products:         nil,
					Users:            nil,
					Sellers:          nil,
					IsFirstPurchase:  false,
					StartDate:        time.Now().UTC(),
					EndDate:          time.Now().UTC(),
					Type:             "Value",
					MaxDiscountValue: 1000,
					MinBasketValue:   13450,
					Extended:         nil,
				},
				Settlement:   "",
				SettlementAt: nil,
				Reserved:     "",
				ReservedAt:   nil,
				Extended:     nil,
			},
			CartRule: nil,
			SSO:      nil,
			VAT: &entities.VAT{
				Rate:         0,
				RawTotal:     nil,
				RoundupTotal: nil,
				CreatedAt:    &timestamp,
				UpdatedAt:    &timestamp,
			},
			TAX:      nil,
			Extended: nil,
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
					Share:    nil,
					Voucher:  nil,
					CartRule: nil,
					SSO: &entities.PackageSSO{
						Rate:         0,
						IsObliged:    false,
						RawTotal:     nil,
						RoundupTotal: nil,
						CreatedAt:    &timestamp,
						UpdatedAt:    &timestamp,
						Extended:     nil,
					},
					VAT:      nil,
					TAX:      nil,
					Extended: nil,
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
								Attributes: map[string]*entities.Attribute{
									"Color": {
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": {
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    0,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: false,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              0,
											IsObliged:         false,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              0,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
								},

								Extended: nil,
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
								Attributes: map[string]*entities.Attribute{
									"Color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    0,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: false,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              0,
											IsObliged:         false,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              0,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
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
								Attributes: map[string]*entities.Attribute{
									"Color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    0,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: true,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              0,
											IsObliged:         false,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              0,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
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
								Attributes: map[string]*entities.Attribute{
									"Color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    0,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: true,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              0,
											IsObliged:         false,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              0,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
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
					Share:    nil,
					Voucher:  nil,
					CartRule: nil,
					SSO: &entities.PackageSSO{
						Rate:         16.67,
						IsObliged:    true,
						RawTotal:     nil,
						RoundupTotal: nil,
						CreatedAt:    &timestamp,
						UpdatedAt:    &timestamp,
						Extended:     nil,
					},
					VAT:      nil,
					TAX:      nil,
					Extended: nil,
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
						IsVATObliged:             true,
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
								Attributes: map[string]*entities.Attribute{
									"Color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    10,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: false,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              9,
											IsObliged:         true,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              9,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
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
								Attributes: map[string]*entities.Attribute{
									"Color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    10,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: false,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              9,
											IsObliged:         true,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              9,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
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
								Attributes: map[string]*entities.Attribute{
									"Color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    10,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: true,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              9,
											IsObliged:         true,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              9,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
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
								Attributes: map[string]*entities.Attribute{
									"Color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
										ValueTranslate: map[string]string{
											"en": "رنگ",
											"fa": "رنگ",
										},
									},
									"dial_color": &entities.Attribute{
										KeyTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
										ValueTranslate: map[string]string{
											"fa": "رنگ صفحه",
											"en": "رنگ صفحه",
										},
									},
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
									SellerCommission: 0,
									Commission: &entities.ItemCommission{
										ItemCommission:    10,
										RawUnitPrice:      nil,
										RoundupUnitPrice:  nil,
										RawTotalPrice:     nil,
										RoundupTotalPrice: nil,
										CreatedAt:         &timestamp,
										UpdatedAt:         &timestamp,
										Extended:          nil,
									},
									Share:             nil,
									ApplicableVoucher: true,
									Voucher:           nil,
									CartRule:          nil,
									SSO:               nil,
									VAT: &entities.ItemVAT{
										SellerVat: &entities.SellerVAT{
											Rate:              9,
											IsObliged:         true,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										BusinessVat: &entities.BusinessVAT{
											Rate:              9,
											RawUnitPrice:      nil,
											RoundupUnitPrice:  nil,
											RawTotalPrice:     nil,
											RoundupTotalPrice: nil,
											CreatedAt:         &timestamp,
											UpdatedAt:         &timestamp,
											Extended:          nil,
										},
										Extended: nil,
									},
									TAX:      nil,
									Extended: nil,
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

func TestFinanceCalcEnum(t *testing.T) {
	for i, flag := range []FinanceCalcType{
		VOUCHER_CALC,
		SELLER_VAT_CALC,
		NET_COMMISSION_CALC,
		BUSINESS_VAT_CALC,
		SELLER_SSO_CALC,
		SHARE_CALC} {
		fmt.Printf("index: %d, flag: %v\n", i, flag)
	}

	fmt.Printf("value: %v\n", Set(SELLER_VAT_CALC, VOUCHER_CALC))

}

func TestFinanceCalcDecorator(t *testing.T) {
	newOrder := createOrder()
	ctx := context.Background()
	savedOrder, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err)
	_, e := New().FinanceCalc(ctx, *savedOrder, SELLER_VAT_CALC, SELLER_FINANCE)
	require.Nil(t, e)
}

func TestFinanceOrderFactory(t *testing.T) {
	newOrder := createOrder()
	ctx := context.Background()
	orderFinance, err := FactoryFromOrder(ctx, newOrder)
	require.Nil(t, err)
	require.NotNil(t, orderFinance)
}

func TestFinanceConvertToOrder(t *testing.T) {
	newOrder := createOrder()
	ctx := context.Background()
	orderFinance, err := FactoryFromOrder(ctx, newOrder)

	require.Nil(t, err)
	require.NotNil(t, orderFinance)

	err = ConvertToOrder(ctx, orderFinance, newOrder)
	require.Nil(t, err)
}

func TestFinanceOrderCalc(t *testing.T) {
	ctx := context.Background()
	newOrder := createOrder()

	newOrder.Status = "NEW"

	order, err := app.Globals.OrderRepository.Save(ctx, *newOrder)
	require.Nil(t, err)

	calcOrder, e := New().FinanceCalc(ctx, *order, Set(SHARE_CALC, VOUCHER_CALC), ORDER_FINANCE)
	require.Nil(t, e)
	require.NotNil(t, calcOrder)

	_, err = app.Globals.OrderRepository.Save(ctx, *calcOrder)
	require.Nil(t, err)
}

//
//func TestFinanceSellerVatCalc(t *testing.T) {
//	ctx := context.Background()
//	order, err := app.Globals.OrderRepository.FindById(ctx, 199277900)
//	require.Nil(t, err)
//
//	calcOrder, e := New().FinanceCalc(ctx, *order, SELLER_VAT_CALC, SELLER_FINANCE)
//	require.Nil(t, e)
//	require.NotNil(t, calcOrder)
//
//	_, err = app.Globals.OrderRepository.Save(ctx, *calcOrder)
//	require.Nil(t, err)
//}
//
//func TestFinanceSellerNetCommissionCalc(t *testing.T) {
//	ctx := context.Background()
//	order, err := app.Globals.OrderRepository.FindById(ctx, 199277900)
//	require.Nil(t, err)
//
//	calcOrder, e := New().FinanceCalc(ctx, *order, NET_COMMISSION_CALC, SELLER_FINANCE)
//	require.Nil(t, e)
//	require.NotNil(t, calcOrder)
//
//	_, err = app.Globals.OrderRepository.Save(ctx, *calcOrder)
//	require.Nil(t, err)
//}
//
//func TestFinanceSellerBusinessVatCalc(t *testing.T) {
//	ctx := context.Background()
//	order, err := app.Globals.OrderRepository.FindById(ctx, 199277900)
//	require.Nil(t, err)
//
//	calcOrder, e := New().FinanceCalc(ctx, *order, BUSINESS_VAT_CALC, SELLER_FINANCE)
//	require.Nil(t, e)
//	require.NotNil(t, calcOrder)
//
//	_, err = app.Globals.OrderRepository.Save(ctx, *calcOrder)
//	require.Nil(t, err)
//}
//
//func TestFinanceSellerSsoCalc(t *testing.T) {
//	ctx := context.Background()
//	order, err := app.Globals.OrderRepository.FindById(ctx, 199277900)
//	require.Nil(t, err)
//
//	calcOrder, e := New().FinanceCalc(ctx, *order, SELLER_SSO_CALC, SELLER_FINANCE)
//	require.Nil(t, e)
//	require.NotNil(t, calcOrder)
//
//	_, err = app.Globals.OrderRepository.Save(ctx, *calcOrder)
//	require.Nil(t, err)
//}
//
//func TestFinanceSellerShareCalc(t *testing.T) {
//	ctx := context.Background()
//	order, err := app.Globals.OrderRepository.FindById(ctx, 199277900)
//	require.Nil(t, err)
//
//	calcOrder, e := New().FinanceCalc(ctx, *order, SHARE_CALC, SELLER_FINANCE)
//	require.Nil(t, e)
//	require.NotNil(t, calcOrder)
//
//	_, err = app.Globals.OrderRepository.Save(ctx, *calcOrder)
//	require.Nil(t, err)
//}
