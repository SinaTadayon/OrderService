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
	order , ok := out.(*entities.Order)
	assert.True(t, ok, "mapping order request to order failed")
	assert.NotEmpty(t, order.Amount.Total)
}


func createRequestNewOrder() *pb.RequestNewOrder {
	order := &pb.RequestNewOrder{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance: &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Amount.Total = 600000
	order.Amount.Original = 550000
	order.Amount.Special = 50000
	order.Amount.Currency = "RR"
	order.Amount.PaymentMethod = "IPG"
	order.Amount.PaymentOption = "AAP"
	order.Amount.ShipmentTotal = 700000
	order.Amount.Voucher = &pb.Voucher{
		Amount: 40000,
		Code: "348",
	}

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

	item := pb.Item {
		Price:    &pb.PriceInfo{},
		Attributes: make(map[string]string, 10),
		Shipment: &pb.ShippingSpec{},
		SellerId: "6546345",
	}

	item.InventoryId = "453564554435345"
	item.Brand = "Asus"
	item.Category = "Electronic/laptop"
	item.Title = "Asus G503 i7, 256SSD, 32G Ram"
	item.Guaranty = "ضمانت سلامت کالا"
	item.Image = "http://baman.io/image/asus.png"
	item.Returnable = true

	item.Price.Special = 200000
	item.Price.Original = 20000000
	item.Price.SellerCommission = 10
	item.Price.Unit = 100000
	item.Price.Currency = "RR"

	
	item.Attributes["Quantity"] = "10"
	item.Attributes["Width"] = "8cm"
	item.Attributes["Height"] = "10cm"
	item.Attributes["Length"] = "15cm"
	item.Attributes["Weight"] = "20kg"
	item.Attributes["Color"] = "blue"
	item.Attributes["Materials"] = "stone"

	//Standard, Express, Economy or Sameday.
	item.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item.Shipment.ShippingTime = 72
	item.Shipment.ReturnTime = 72
	item.Shipment.ReactionTime = 24
	item.Shipment.CarrierName = "Post"
	item.Shipment.CarrierProduct = "Post Express"
	item.Shipment.CarrierType = "standard"
	item.Shipment.ShippingCost = 100000
	item.Shipment.VoucherAmount = 0
	item.Shipment.Currency = "RR"

	item.SellerId = "345346343"
	order.Items = append(order.Items, &item)
	return order
}
