package converter

import (
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	pb "gitlab.faza.io/protos/order"
	"testing"
)

func TestOrderConverter(t *testing.T) {
	converter := NewConverter()
	RequestNewOrder := createRequestNewOrder()
	out, err := converter.Map(RequestNewOrder, entities.Order{})
	assert.NoError(t, err, "mapping order request to order failed")
	order, ok := out.(*entities.Order)
	assert.True(t, ok, "mapping order request to order failed")
	assert.NotEmpty(t, order.Invoice.GrandTotal)
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

	order.Invoice.PaymentMethod = "IPG"
	order.Invoice.PaymentGateway = "AAP"
	order.Invoice.PaymentOption = nil
	order.Invoice.ShipmentTotal = &pb.Money{
		Amount:   "700000",
		Currency: "IRR",
	}
	order.Invoice.Voucher = &pb.Voucher{
		Percent: 0,
		Price: &pb.Money{
			Amount:   "40000",
			Currency: "IRR",
		},
		Code: "348",
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
			ShippingCost: &pb.Money{
				Amount:   "100000",
				Currency: "IRR",
			},
			VoucherPrice: nil,
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

			SellerCommission: 10,
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
			SellerCommission: 5,
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

			VoucherPrice: &pb.Money{
				Amount:   "60000",
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

			SellerCommission: 8,
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

			SellerCommission: 3,
		},
	}

	pkg.Items = append(pkg.Items, item)
	return order
}
