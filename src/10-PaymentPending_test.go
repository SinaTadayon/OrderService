package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createPaymentRequestSampleFull() PaymentPendingRequest {
	var pr = PaymentPendingRequest{}

	pr.orderNumber = "102102"
	// buyer info
	pr.buyer.info.firstName = "farzan"
	pr.buyer.info.lastName = "dalaee"
	pr.buyer.info.email = "farzan.dalaee@gmail.com"
	pr.buyer.info.gender = "Male"
	pr.buyer.info.mobile = "+98912193870"
	pr.buyer.info.nationalId = "0012345678"
	// buyer address
	pr.buyer.address.title = "Home"
	pr.buyer.address.phone = "+98912193870"
	pr.buyer.address.zipCode = "1651764614"
	pr.buyer.address.country = "Iran"
	pr.buyer.address.state = "Tehran"
	pr.buyer.address.city = "Tehran"
	pr.buyer.address.address = "Sheikh bahaee, p 5"
	pr.buyer.address.lat = "10.1345664"
	pr.buyer.address.lan = "22.1345664"
	// buyer finance
	pr.buyer.finance.iban = "IR165411211001514313143545354134"
	pr.buyer.finance.bankName = "saman"
	pr.buyer.finance.cartNumber = "6014111122223333"
	// amount
	pr.amount.total = 200000
	pr.amount.discount = 40000
	pr.amount.paid = 160000
	// items
	pr.items = append(pr.items, Item{
		sku: "aaa000",
		amount: ItemAmount{
			total:            200000,
			paid:             160000,
			sellerCommission: 10000,
			systemCommission: 5000,
			discount:         40000,
		},
		quantity: 10,
		seller: ItemSeller{
			info: ItemSellerInfo{
				companyName: "digi",
			},
		},
		detail: ItemDetail{
			brand:       "Asus",
			categories:  "Electronic/laptop",
			description: "Asus G503 i7, 256SSD, 32G Ram",
		},
		shipment: ItemShipment{
			providerName: "Post",
		},
	})
	pr.items = append(pr.items, Item{
		sku: "aaa111",
		amount: ItemAmount{
			total:            300000,
			paid:             160000,
			sellerCommission: 10000,
			systemCommission: 5000,
			discount:         140000,
		},
		quantity: 1,
		seller: ItemSeller{
			info: ItemSellerInfo{
				companyName: "digi",
			},
		},
		detail: ItemDetail{
			brand:       "Asus",
			categories:  "Electronic/laptop",
			description: "Asus G503 i7, 256SSD, 32G Ram",
		},
		shipment: ItemShipment{
			providerName: "Post",
		},
	})

	return pr
}

func TestPaymentPendingMessageValidate(t *testing.T) {
	var pr = createPaymentRequestSampleFull()

	err := pr.validate()
	assert.Nil(t, err)
}
