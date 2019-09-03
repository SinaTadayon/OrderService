package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pb "gitlab.faza.io/protos/payment"
	"google.golang.org/grpc"
)

func TestNewOrder(t *testing.T) {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnCoupon, err := grpc.DialContext(ctx, "localhost:"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	PaymentService := pb.NewOrderServiceClient(grpcConnCoupon)

	req := createPaymentRequestSampleFull()

	order := &pb.OrderPaymentRequest{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Info:    &pb.BuyerInfo{},
			Finance: &pb.BuyerFinance{},
			Address: &pb.BuyerAddress{},
		},
	}

	order.OrderNumber = req.orderNumber

	order.Amount.Total = float32(req.amount.total)
	order.Amount.Payable = float32(req.amount.payable)
	order.Amount.Discount = float32(req.amount.discount)

	order.Buyer.Info.LastName = req.buyer.lastName
	order.Buyer.Info.FirstName = req.buyer.firstName
	order.Buyer.Info.Email = req.buyer.email
	order.Buyer.Info.Mobile = req.buyer.mobile
	order.Buyer.Info.NationalId = req.buyer.nationalId

	order.Buyer.Finance.Iban = req.buyer.finance.iban

	order.Buyer.Address.Address = req.buyer.address.address
	order.Buyer.Address.State = req.buyer.address.state
	order.Buyer.Address.Phone = req.buyer.address.phone
	order.Buyer.Address.ZipCode = req.buyer.address.zipCode
	order.Buyer.Address.City = req.buyer.address.city
	order.Buyer.Address.Country = req.buyer.address.country
	order.Buyer.Address.Lat = req.buyer.address.lat
	order.Buyer.Address.Lan = req.buyer.address.lan

	i := pb.Item{
		Price:    &pb.ItemPrice{},
		Shipment: &pb.ItemShipment{},
		Seller: &pb.ItemSeller{
			Address: &pb.ItemSellerAddress{},
			Finance: &pb.ItemSellerFinance{},
		},
	}
	i.Sku = req.items[0].sku
	i.Brand = req.items[0].brand
	i.Categories = req.items[0].categories
	i.Title = req.items[0].title
	i.Warranty = req.items[0].warranty
	i.Quantity = req.items[0].quantity

	i.Price.Discount = float32(req.items[0].price.discount)
	i.Price.Payable = float32(req.items[0].price.payable)
	i.Price.Total = float32(req.items[0].price.total)
	i.Price.SellerCommission = float32(req.items[0].price.sellerCommission)
	i.Price.Unit = float32(req.items[0].price.unit)

	i.Shipment.ShipmentDetail = req.items[0].shipment.shipmentDetail
	i.Shipment.ShippingTime = req.items[0].shipment.shippingTime
	i.Shipment.ReturnTime = req.items[0].shipment.returnTime
	i.Shipment.ReactionTime = req.items[0].shipment.reactionTime
	i.Shipment.ProviderName = req.items[0].shipment.providerName

	i.Seller.Title = req.items[0].seller.title
	i.Seller.NationalId = req.items[0].seller.nationalId
	i.Seller.Mobile = req.items[0].seller.mobile
	i.Seller.FirstName = req.items[0].seller.firstName
	i.Seller.LastName = req.items[0].seller.lastName
	i.Seller.Email = req.items[0].seller.email
	i.Seller.RegistrationName = req.items[0].seller.registrationName
	i.Seller.CompanyName = req.items[0].seller.companyName

	i.Seller.Address.Address = req.items[0].seller.address.address
	i.Seller.Address.Lan = req.items[0].seller.address.lan
	i.Seller.Address.Lat = req.items[0].seller.address.lat
	i.Seller.Address.Country = req.items[0].seller.address.country
	i.Seller.Address.City = req.items[0].seller.address.city
	i.Seller.Address.ZipCode = req.items[0].seller.address.zipCode
	i.Seller.Address.Phone = req.items[0].seller.address.phone
	i.Seller.Address.State = req.items[0].seller.address.state
	i.Seller.Address.Title = req.items[0].seller.address.title

	i.Seller.Finance.Iban = req.items[0].seller.finance.iban

	order.Items = append(order.Items, &i)

	//json, err := json.Marshal(order)
	assert.Nil(t, err)
	//fmt.Println(string(json))

	_, err = PaymentService.NewOrder(ctx, order)
	//assert.NotNil(t, resOrder)
	//assert.Equal(t, string(http.StatusOK), resOrder.Status)
	//fmt.Println(resOrder)
	assert.Nil(t, err)
}
