package payment_pending_step

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func createPaymentRequestSampleFull() PaymentPendingRequest {
//	var pr = PaymentPendingRequest{}
//
//	pr.OrderNumber = "TESTTEST"
//	// buyer info
//	pr.Buyer.FirstName = "farzan"
//	pr.Buyer.LastName = "dalaee"
//	pr.Buyer.Email = "farzan.dalaee@gmail.com"
//	pr.Buyer.Mobile = "+98912193870"
//	pr.Buyer.NationalId = "0012345678"
//	pr.Buyer.IP = "127.0.0.1"
//	// buyer address
//	pr.Buyer.Address.Phone = "+98912193870"
//	pr.Buyer.Address.ZipCode = "1651764614"
//	pr.Buyer.Address.Country = "Iran"
//	pr.Buyer.Address.State = "Tehran"
//	pr.Buyer.Address.City = "Tehran"
//	pr.Buyer.Address.Address = "Sheikh bahaee, p 5"
//	pr.Buyer.Address.Lat = "10.1345664"
//	pr.Buyer.Address.Lan = "22.1345664"
//	// buyer finance
//	pr.Buyer.Finance.Iban = "IR165411211001514313143545"
//	// amount
//	pr.Amount.total = 200000
//	pr.Amount.Discount = 40000
//	pr.Amount.Subtotal = 160000
//	// items
//	pr.Items = append(pr.Items, Item{
//		Sku:        "aaa000",
//		Quantity:   10,
//		Brand:      "Asus",
//		Category: "Electronic/laptop",
//		Title:      "Asus G503 i7, 256SSD, 32G Ram",
//		Guaranty:   "ضمانت سلامت کالا",
//		Price: ItemPrice{
//			Unit:             20000,
//			total:            200000,
//			Subtotal:          160000,
//			Discount:         40000,
//			SellerCommission: 10000,
//		},
//		SellerInfo: ItemSeller{
//			CompanyName:      "digi",
//			RegistrationName: "Digikala",
//			LastName:         "hamid",
//			FirstName:        "hamid",
//			Email:            "info@digikala.com",
//			Mobile:           "09121112233",
//			NationalId:       "0101010100",
//			Title:            "Digikala Shop",
//			EconomicCode:     "dasdasdasasd",
//			Address: ItemSellerAddress{
//				Address: "address",
//				Title:   "office",
//				State:   "Tehran",
//				Phone:   "0212222222",
//				ZipCode: "1651145864",
//				City:    "Tehran",
//				Country: "Iran",
//				Lat:     "03221211",
//				Lan:     "23031121",
//			},
//			Finance: ItemSellerFinance{
//				Iban: "IR165411211001514313143545354134",
//			},
//		},
//		Shipment: ItemShipment{
//			ReactionTime:   1,
//			ReturnTime:     72,
//			ShippingTime:   72,
//			ProviderName:   "Post",
//			ShipmentDetail: "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
//		},
//	})
//	pr.Items = append(pr.Items, Item{
//		Sku:        "aaa111",
//		Brand:      "Asus",
//		Category: "Electronic/laptop",
//		Title:      "Asus G503 i7, 256SSD, 32G Ram",
//		Quantity:   1,
//		Guaranty:   "صلامت کالا",
//		Price: ItemPrice{
//			total:            300000,
//			Subtotal:          160000,
//			Discount:         140000,
//			Unit:             300000,
//			SellerCommission: 20000,
//		},
//		SellerInfo: ItemSeller{
//			CompanyName:      "digi",
//			Title:            "Digikala",
//			EconomicCode:     "13211",
//			NationalId:       "0010085555",
//			Mobile:           "09121112233",
//			Email:            "info@digikala.com",
//			FirstName:        "hamid",
//			LastName:         "mohammadi",
//			RegistrationName: "Digikala gostaran e shargh",
//			Finance: ItemSellerFinance{
//				Iban: "IR165411211001514313143545354134",
//			},
//			Address: ItemSellerAddress{
//				Address: "Address",
//				Title:   "Office",
//				Lan:     "210313",
//				Lat:     "131533",
//				Country: "Iran",
//				City:    "Tehran",
//				ZipCode: "113315",
//				Phone:   "021222222",
//				State:   "Tehran",
//			},
//		},
//		Shipment: ItemShipment{
//			ReactionTime:   24,
//			ReturnTime:     72,
//			ShippingTime:   72,
//			ProviderName:   "Post",
//			ShipmentDetail: "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد",
//		},
//	})
//	return pr
//}
//
//func TestPaymentPendingMessageValidate(t *testing.T) {
//	var pr = createPaymentRequestSampleFull()
//
//	err := pr.validate()
//	assert.Nil(t, err)
//}
