package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"

	pb "gitlab.faza.io/protos/payment"
)

func TestNewOrder(t *testing.T) {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnCoupon, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	PaymentService := pb.NewOrderServiceClient(grpcConnCoupon)

	req := createPaymentRequestSampleFull()

	order := &pb.OrderPaymentRequest{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance: &pb.BuyerFinance{},
			Address: &pb.BuyerAddress{},
		},
	}

	order.Amount.Total = float32(req.Amount.Total)
	order.Amount.Payable = float32(req.Amount.Payable)
	order.Amount.Discount = float32(req.Amount.Discount)

	order.Buyer.LastName = req.Buyer.LastName
	order.Buyer.FirstName = req.Buyer.FirstName
	order.Buyer.Email = req.Buyer.Email
	order.Buyer.Mobile = req.Buyer.Mobile
	order.Buyer.NationalId = req.Buyer.NationalId
	order.Buyer.Ip = req.Buyer.IP

	order.Buyer.Finance.Iban = req.Buyer.Finance.Iban

	order.Buyer.Address.Address = req.Buyer.Address.Address
	order.Buyer.Address.State = req.Buyer.Address.State
	order.Buyer.Address.Phone = req.Buyer.Address.Phone
	order.Buyer.Address.ZipCode = req.Buyer.Address.ZipCode
	order.Buyer.Address.City = req.Buyer.Address.City
	order.Buyer.Address.Country = req.Buyer.Address.Country
	order.Buyer.Address.Lat = req.Buyer.Address.Lat
	order.Buyer.Address.Lan = req.Buyer.Address.Lan

	i := pb.Item{
		Price:    &pb.ItemPrice{},
		Shipment: &pb.ItemShipment{},
		Seller: &pb.ItemSeller{
			Address: &pb.ItemSellerAddress{},
			Finance: &pb.ItemSellerFinance{},
		},
	}
	i.Sku = req.Items[0].Sku
	i.Brand = req.Items[0].Brand
	i.Categories = req.Items[0].Categories
	i.Title = req.Items[0].Title
	i.Warranty = req.Items[0].Warranty
	i.Quantity = req.Items[0].Quantity

	i.Price.Discount = float32(req.Items[0].Price.Discount)
	i.Price.Payable = float32(req.Items[0].Price.Payable)
	i.Price.Total = float32(req.Items[0].Price.Total)
	i.Price.SellerCommission = float32(req.Items[0].Price.SellerCommission)
	i.Price.Unit = float32(req.Items[0].Price.Unit)

	i.Shipment.ShipmentDetail = req.Items[0].Shipment.ShipmentDetail
	i.Shipment.ShippingTime = req.Items[0].Shipment.ShippingTime
	i.Shipment.ReturnTime = req.Items[0].Shipment.ReturnTime
	i.Shipment.ReactionTime = req.Items[0].Shipment.ReactionTime
	i.Shipment.ProviderName = req.Items[0].Shipment.ProviderName

	i.Seller.Title = req.Items[0].Seller.Title
	i.Seller.NationalId = req.Items[0].Seller.NationalId
	i.Seller.Mobile = req.Items[0].Seller.Mobile
	i.Seller.FirstName = req.Items[0].Seller.FirstName
	i.Seller.LastName = req.Items[0].Seller.LastName
	i.Seller.Email = req.Items[0].Seller.Email
	i.Seller.RegistrationName = req.Items[0].Seller.RegistrationName
	i.Seller.CompanyName = req.Items[0].Seller.CompanyName

	i.Seller.Address.Address = req.Items[0].Seller.Address.Address
	i.Seller.Address.Lan = req.Items[0].Seller.Address.Lan
	i.Seller.Address.Lat = req.Items[0].Seller.Address.Lat
	i.Seller.Address.Country = req.Items[0].Seller.Address.Country
	i.Seller.Address.City = req.Items[0].Seller.Address.City
	i.Seller.Address.ZipCode = req.Items[0].Seller.Address.ZipCode
	i.Seller.Address.Phone = req.Items[0].Seller.Address.Phone
	i.Seller.Address.State = req.Items[0].Seller.Address.State
	i.Seller.Address.Title = req.Items[0].Seller.Address.Title

	i.Seller.Finance.Iban = req.Items[0].Seller.Finance.Iban

	order.Items = append(order.Items, &i)

	resOrder, err := PaymentService.NewOrder(ctx, order)
	assert.Nil(t, err)
	assert.NotNil(t, resOrder)
}

func TestSellerApprovalPending(t *testing.T) {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnCoupon, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	PaymentService := pb.NewOrderServiceClient(grpcConnCoupon)

	req := createPaymentRequestSampleFull()

	order := &pb.OrderPaymentRequest{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance: &pb.BuyerFinance{},
			Address: &pb.BuyerAddress{},
		},
	}

	order.Amount.Total = float32(req.Amount.Total)
	order.Amount.Payable = float32(req.Amount.Payable)
	order.Amount.Discount = float32(req.Amount.Discount)

	order.Buyer.LastName = req.Buyer.LastName
	order.Buyer.FirstName = req.Buyer.FirstName
	order.Buyer.Email = req.Buyer.Email
	order.Buyer.Mobile = req.Buyer.Mobile
	order.Buyer.NationalId = req.Buyer.NationalId
	order.Buyer.Ip = req.Buyer.IP

	order.Buyer.Finance.Iban = req.Buyer.Finance.Iban

	order.Buyer.Address.Address = req.Buyer.Address.Address
	order.Buyer.Address.State = req.Buyer.Address.State
	order.Buyer.Address.Phone = req.Buyer.Address.Phone
	order.Buyer.Address.ZipCode = req.Buyer.Address.ZipCode
	order.Buyer.Address.City = req.Buyer.Address.City
	order.Buyer.Address.Country = req.Buyer.Address.Country
	order.Buyer.Address.Lat = req.Buyer.Address.Lat
	order.Buyer.Address.Lan = req.Buyer.Address.Lan

	i := pb.Item{
		Price:    &pb.ItemPrice{},
		Shipment: &pb.ItemShipment{},
		Seller: &pb.ItemSeller{
			Address: &pb.ItemSellerAddress{},
			Finance: &pb.ItemSellerFinance{},
		},
	}
	i.Sku = req.Items[0].Sku
	i.Brand = req.Items[0].Brand
	i.Categories = req.Items[0].Categories
	i.Title = req.Items[0].Title
	i.Warranty = req.Items[0].Warranty
	i.Quantity = req.Items[0].Quantity

	i.Price.Discount = float32(req.Items[0].Price.Discount)
	i.Price.Payable = float32(req.Items[0].Price.Payable)
	i.Price.Total = float32(req.Items[0].Price.Total)
	i.Price.SellerCommission = float32(req.Items[0].Price.SellerCommission)
	i.Price.Unit = float32(req.Items[0].Price.Unit)

	i.Shipment.ShipmentDetail = req.Items[0].Shipment.ShipmentDetail
	i.Shipment.ShippingTime = req.Items[0].Shipment.ShippingTime
	i.Shipment.ReturnTime = req.Items[0].Shipment.ReturnTime
	i.Shipment.ReactionTime = req.Items[0].Shipment.ReactionTime
	i.Shipment.ProviderName = req.Items[0].Shipment.ProviderName

	i.Seller.Title = req.Items[0].Seller.Title
	i.Seller.NationalId = req.Items[0].Seller.NationalId
	i.Seller.Mobile = req.Items[0].Seller.Mobile
	i.Seller.FirstName = req.Items[0].Seller.FirstName
	i.Seller.LastName = req.Items[0].Seller.LastName
	i.Seller.Email = req.Items[0].Seller.Email
	i.Seller.RegistrationName = req.Items[0].Seller.RegistrationName
	i.Seller.CompanyName = req.Items[0].Seller.CompanyName

	i.Seller.Address.Address = req.Items[0].Seller.Address.Address
	i.Seller.Address.Lan = req.Items[0].Seller.Address.Lan
	i.Seller.Address.Lat = req.Items[0].Seller.Address.Lat
	i.Seller.Address.Country = req.Items[0].Seller.Address.Country
	i.Seller.Address.City = req.Items[0].Seller.Address.City
	i.Seller.Address.ZipCode = req.Items[0].Seller.Address.ZipCode
	i.Seller.Address.Phone = req.Items[0].Seller.Address.Phone
	i.Seller.Address.State = req.Items[0].Seller.Address.State
	i.Seller.Address.Title = req.Items[0].Seller.Address.Title

	i.Seller.Finance.Iban = req.Items[0].Seller.Finance.Iban

	order.Items = append(order.Items, &i)

	resOrder, err := PaymentService.NewOrder(ctx, order)
	assert.Nil(t, err)
	assert.NotNil(t, resOrder)

	time.Sleep(2 * time.Second)

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnOrder, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConnOrder)

	approveRequest := &pb.ApprovalRequest{
		OrderNumber: resOrder.OrderNumber,
		Approval:    true,
		Reason:      "",
	}

	resApproval, err := OrderService.SellerApprovalPending(ctx, approveRequest)
	assert.Nil(t, err)
	assert.Equal(t, resOrder.OrderNumber, resApproval.OrderNumber)
}

func TestSellerApprovalPendingRejected(t *testing.T) {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnCoupon, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	PaymentService := pb.NewOrderServiceClient(grpcConnCoupon)

	req := createPaymentRequestSampleFull()

	order := &pb.OrderPaymentRequest{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance: &pb.BuyerFinance{},
			Address: &pb.BuyerAddress{},
		},
	}

	order.Amount.Total = float32(req.Amount.Total)
	order.Amount.Payable = float32(req.Amount.Payable)
	order.Amount.Discount = float32(req.Amount.Discount)

	order.Buyer.LastName = req.Buyer.LastName
	order.Buyer.FirstName = req.Buyer.FirstName
	order.Buyer.Email = req.Buyer.Email
	order.Buyer.Mobile = req.Buyer.Mobile
	order.Buyer.NationalId = req.Buyer.NationalId
	order.Buyer.Ip = req.Buyer.IP

	order.Buyer.Finance.Iban = req.Buyer.Finance.Iban

	order.Buyer.Address.Address = req.Buyer.Address.Address
	order.Buyer.Address.State = req.Buyer.Address.State
	order.Buyer.Address.Phone = req.Buyer.Address.Phone
	order.Buyer.Address.ZipCode = req.Buyer.Address.ZipCode
	order.Buyer.Address.City = req.Buyer.Address.City
	order.Buyer.Address.Country = req.Buyer.Address.Country
	order.Buyer.Address.Lat = req.Buyer.Address.Lat
	order.Buyer.Address.Lan = req.Buyer.Address.Lan

	i := pb.Item{
		Price:    &pb.ItemPrice{},
		Shipment: &pb.ItemShipment{},
		Seller: &pb.ItemSeller{
			Address: &pb.ItemSellerAddress{},
			Finance: &pb.ItemSellerFinance{},
		},
	}
	i.Sku = req.Items[0].Sku
	i.Brand = req.Items[0].Brand
	i.Categories = req.Items[0].Categories
	i.Title = req.Items[0].Title
	i.Warranty = req.Items[0].Warranty
	i.Quantity = req.Items[0].Quantity

	i.Price.Discount = float32(req.Items[0].Price.Discount)
	i.Price.Payable = float32(req.Items[0].Price.Payable)
	i.Price.Total = float32(req.Items[0].Price.Total)
	i.Price.SellerCommission = float32(req.Items[0].Price.SellerCommission)
	i.Price.Unit = float32(req.Items[0].Price.Unit)

	i.Shipment.ShipmentDetail = req.Items[0].Shipment.ShipmentDetail
	i.Shipment.ShippingTime = req.Items[0].Shipment.ShippingTime
	i.Shipment.ReturnTime = req.Items[0].Shipment.ReturnTime
	i.Shipment.ReactionTime = req.Items[0].Shipment.ReactionTime
	i.Shipment.ProviderName = req.Items[0].Shipment.ProviderName

	i.Seller.Title = req.Items[0].Seller.Title
	i.Seller.NationalId = req.Items[0].Seller.NationalId
	i.Seller.Mobile = req.Items[0].Seller.Mobile
	i.Seller.FirstName = req.Items[0].Seller.FirstName
	i.Seller.LastName = req.Items[0].Seller.LastName
	i.Seller.Email = req.Items[0].Seller.Email
	i.Seller.RegistrationName = req.Items[0].Seller.RegistrationName
	i.Seller.CompanyName = req.Items[0].Seller.CompanyName

	i.Seller.Address.Address = req.Items[0].Seller.Address.Address
	i.Seller.Address.Lan = req.Items[0].Seller.Address.Lan
	i.Seller.Address.Lat = req.Items[0].Seller.Address.Lat
	i.Seller.Address.Country = req.Items[0].Seller.Address.Country
	i.Seller.Address.City = req.Items[0].Seller.Address.City
	i.Seller.Address.ZipCode = req.Items[0].Seller.Address.ZipCode
	i.Seller.Address.Phone = req.Items[0].Seller.Address.Phone
	i.Seller.Address.State = req.Items[0].Seller.Address.State
	i.Seller.Address.Title = req.Items[0].Seller.Address.Title

	i.Seller.Finance.Iban = req.Items[0].Seller.Finance.Iban

	order.Items = append(order.Items, &i)

	resOrder, err := PaymentService.NewOrder(ctx, order)
	assert.Nil(t, err)
	assert.NotNil(t, resOrder)

	time.Sleep(2 * time.Second)

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnOrder, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConnOrder)

	approveRequest := &pb.ApprovalRequest{
		OrderNumber: resOrder.OrderNumber,
		Approval:    false,
		Reason:      "out of stock",
	}

	resApproval, err := OrderService.SellerApprovalPending(ctx, approveRequest)
	assert.Nil(t, err)
	assert.Equal(t, resOrder.OrderNumber, resApproval.OrderNumber)
}
