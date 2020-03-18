package converter

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	pb "gitlab.faza.io/protos/order"
	"testing"
)

func TestOrderConverter(t *testing.T) {
	ctx := context.Background()
	converter := NewConverter()
	RequestNewOrder := createRequestNewOrder()
	out, err := converter.Map(ctx, RequestNewOrder, entities.Order{})
	require.NoError(t, err, "mapping order request to order failed")
	order, ok := out.(*entities.Order)
	require.True(t, ok, "mapping order request to order failed")
	require.NotEmpty(t, order.Invoice.GrandTotal)
}

func createRequestNewOrder() *pb.RequestNewOrder {
	order := &pb.RequestNewOrder{
		Platform: "PWA",
		Invoice:  &pb.Invoice{},
		Buyer: &pb.Buyer{
			Finance:         &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Invoice.GrandTotal = &pb.Money{
		Amount:   "600000",
		Currency: "IRR",
	}
	order.Invoice.Subtotal = &pb.Money{
		Amount:   "550000",
		Currency: "IRR",
	}
	order.Invoice.Discount = &pb.Money{
		Amount:   "50000",
		Currency: "IRR",
	}

	order.Invoice.Vat = &pb.Invoice_BusinessVAT{
		Value: 9,
	}

	order.Invoice.PaymentMethod = "IPG"
	order.Invoice.PaymentGateway = "AAP"
	order.Invoice.PaymentOption = nil
	order.Invoice.ShipmentTotal = &pb.Money{
		Amount:   "700000",
		Currency: "IRR",
	}
	order.Invoice.Voucher = &pb.Voucher{
		Percent: 0,
		RawAppliedPrice: &pb.Money{
			Amount:   "40000",
			Currency: "IRR",
		},
		RoundupAppliedPrice: &pb.Money{
			Amount:   "40000",
			Currency: "IRR",
		},
		Price: &pb.Money{
			Amount:   "40000",
			Currency: "IRR",
		},
		Code: "348",
		Details: &pb.VoucherDetails{
			StartDate:        "2019-12-28T14:32:46-0700",
			EndDate:          "2020-01-20T00:00:00-0000",
			Type:             "",
			MaxDiscountValue: 0,
			MinBasketValue:   0,
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
		},
	}

	order.Buyer.BuyerId = 1000001
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

	order.Buyer.ShippingAddress.FirstName = "sina"
	order.Buyer.ShippingAddress.LastName = "tadayon"
	order.Buyer.ShippingAddress.Address = "Sheikh bahaee, p 5"
	order.Buyer.ShippingAddress.Province = "Tehran"
	order.Buyer.ShippingAddress.Mobile = "+98912193870"
	order.Buyer.ShippingAddress.Phone = "+98218475644"
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
			ShippingCost: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			ReactionTime: 24,
			ShippingTime: 72,
			ReturnTime:   72,
			Details:      "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
		Invoice: &pb.PackageInvoice{
			Subtotal: &pb.Money{
				Amount:   "9238443",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "9734234",
				Currency: "IRR",
			},
			ShipmentPrice: &pb.Money{
				Amount:   "23123",
				Currency: "IRR",
			},
			Sso: &pb.PackageInvoice_SellerSSO{
				Value:     9,
				IsObliged: true,
			},
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
		Attributes: map[string]*pb.Attribute{
			"Quantity": &pb.Attribute{
				KeyTrans: map[string]string{
					"en": "Quantity",
				},
				ValueTrans: map[string]string{
					"en": "10",
				},
			},
			"Width": &pb.Attribute{
				KeyTrans: map[string]string{
					"en": "Width",
				},
				ValueTrans: map[string]string{
					"en": "10",
				},
			},
		},
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "20000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "220000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "20000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
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
		Attributes:  nil,
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "10000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "120000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "10000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
		},
	}
	pkg.Items = append(pkg.Items, item)

	pkg = &pb.Package{
		SellerId: 111122223333,
		Shipment: &pb.ShippingSpec{
			CarrierNames:   []string{"Post"},
			CarrierProduct: "Post Express",
			CarrierType:    "standard",
			ShippingCost: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},

			ReactionTime: 24,
			ShippingTime: 72,
			ReturnTime:   72,
			Details:      "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
		Invoice: &pb.PackageInvoice{
			Subtotal: &pb.Money{
				Amount:   "9238443",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "9734234",
				Currency: "IRR",
			},
			ShipmentPrice: &pb.Money{
				Amount:   "23123",
				Currency: "IRR",
			},
			Sso: &pb.PackageInvoice_SellerSSO{
				Value:     16.67,
				IsObliged: true,
			},
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
		Attributes:  nil,
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "20000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "220000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "200000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "20000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
		},
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
		Attributes:  nil,
		Invoice: &pb.ItemInvoice{
			Unit: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Total: &pb.Money{
				Amount:   "10000000",
				Currency: "IRR",
			},
			Original: &pb.Money{
				Amount:   "120000",
				Currency: "IRR",
			},
			Special: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			Discount: &pb.Money{
				Amount:   "10000",
				Currency: "IRR",
			},
			ItemCommission: 10,
			Vat: &pb.ItemInvoice_SellerVAT{
				Value:     9,
				IsObliged: true,
			},
		},
	}

	pkg.Items = append(pkg.Items, item)
	return order
}
