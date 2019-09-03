package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createPaymentRequestSampleFull() PaymentPendingRequest {
	var pr = PaymentPendingRequest{}

	pr.orderNumber = "102102"
	// buyer info
	pr.buyer.firstName = "farzan"
	pr.buyer.lastName = "dalaee"
	pr.buyer.email = "farzan.dalaee@gmail.com"
	pr.buyer.mobile = "+98912193870"
	pr.buyer.nationalId = "0012345678"
	// buyer address
	pr.buyer.address.phone = "+98912193870"
	pr.buyer.address.zipCode = "1651764614"
	pr.buyer.address.country = "Iran"
	pr.buyer.address.state = "Tehran"
	pr.buyer.address.city = "Tehran"
	pr.buyer.address.address = "Sheikh bahaee, p 5"
	pr.buyer.address.lat = "10.1345664"
	pr.buyer.address.lan = "22.1345664"
	// buyer finance
	pr.buyer.finance.iban = "IR165411211001514313143545"
	// amount
	pr.amount.total = 200000
	pr.amount.discount = 40000
	pr.amount.payable = 160000
	// items
	pr.items = append(pr.items, Item{
		sku:        "aaa000",
		quantity:   10,
		brand:      "Asus",
		categories: "Electronic/laptop",
		title:      "Asus G503 i7, 256SSD, 32G Ram",
		warranty:   "ضمانت سلامت کالا",
		price: ItemPrice{
			unit:             20000,
			total:            200000,
			payable:          160000,
			discount:         40000,
			sellerCommission: 10000,
		},
		seller: ItemSeller{
			companyName:      "digi",
			registrationName: "Digikala",
			lastName:         "hamid",
			firstName:        "hamid",
			email:            "info@digikala.com",
			mobile:           "09121112233",
			nationalId:       "0101010100",
			title:            "Digikala Shop",
			economicCode:     "dasdasdasasd",
			address: ItemSellerAddress{
				address: "address",
				title:   "office",
				state:   "Tehran",
				phone:   "0212222222",
				zipCode: "1651145864",
				city:    "Tehran",
				country: "Iran",
				lat:     "03221211",
				lan:     "23031121",
			},
			finance: ItemSellerFinance{
				iban: "IR165411211001514313143545354134",
			},
		},
		shipment: ItemShipment{
			reactionTime:   1,
			returnTime:     72,
			shippingTime:   72,
			providerName:   "Post",
			shipmentDetail: "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
	})
	pr.items = append(pr.items, Item{
		sku:        "aaa111",
		brand:      "Asus",
		categories: "Electronic/laptop",
		title:      "Asus G503 i7, 256SSD, 32G Ram",
		quantity:   1,
		warranty:   "صلامت کالا",
		price: ItemPrice{
			total:            300000,
			payable:          160000,
			discount:         140000,
			unit:             300000,
			sellerCommission: 20000,
		},
		seller: ItemSeller{
			companyName:      "digi",
			title:            "Digikala",
			economicCode:     "13211",
			nationalId:       "0010085555",
			mobile:           "09121112233",
			email:            "info@digikala.com",
			firstName:        "hamid",
			lastName:         "mohammadi",
			registrationName: "Digikala gostaran e shargh",
			finance: ItemSellerFinance{
				iban: "IR165411211001514313143545354134",
			},
			address: ItemSellerAddress{
				address: "Address",
				title:   "Office",
				lan:     "210313",
				lat:     "131533",
				country: "Iran",
				city:    "Tehran",
				zipCode: "113315",
				phone:   "021222222",
				state:   "Tehran",
			},
		},
		shipment: ItemShipment{
			reactionTime:   24,
			returnTime:     72,
			shippingTime:   72,
			providerName:   "Post",
			shipmentDetail: "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
		},
	})
	return pr
}

func TestPaymentPendingMessageValidate(t *testing.T) {
	var pr = createPaymentRequestSampleFull()

	err := pr.validate()
	assert.Nil(t, err)
}
