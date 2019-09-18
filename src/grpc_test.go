package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"

	pb "gitlab.faza.io/protos/payment"
)

func createOrderObject() *pb.OrderPaymentRequest {
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
	return order
}

func TestAddGrpcStateRule(t *testing.T) {
	GrpcStatesRules.SellerApprovalPending = addStateRule(PaymentSuccess)

	_, ok := GrpcStatesRules.SellerApprovalPending[PaymentSuccess]
	assert.True(t, ok)
}

// Grpc test
func TestNewOrder(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnNewOrder, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	PaymentService := pb.NewOrderServiceClient(grpcConnNewOrder)

	order := createOrderObject()

	resOrder, err := PaymentService.NewOrder(ctx, order)
	assert.Nil(t, err)
	assert.NotNil(t, resOrder)
}

func TestSellerApprovalPendingApproved(t *testing.T) {
	// Create ppr
	ppr := createPaymentRequestSampleFull()
	// Delete test order
	_, err := App.mongo.DeleteOne(MongoDB, Orders, bson.D{{"ordernumber", ppr.OrderNumber}})
	assert.Nil(t, err)
	statusHistory := StatusHistory{
		Status:    PaymentPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    PaymentSuccess,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    SellerApprovalPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "auto approval",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	ppr.Status.Current = ppr.Status.History[(len(ppr.Status.History) - 1)].Status

	_, err = App.mongo.InsertOne(MongoDB, Orders, ppr)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnOrderApproved, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConnOrderApproved)

	approveRequest := &pb.ApprovalRequest{
		OrderNumber: ppr.OrderNumber,
		Approval:    true,
		Reason:      "",
	}

	resApproval, err := OrderService.SellerApprovalPending(ctx, approveRequest)
	assert.Nil(t, err)
	assert.Equal(t, ppr.OrderNumber, resApproval.OrderNumber)

	savedOrder, err := GetOrder(ppr.OrderNumber)
	assert.Nil(t, err)
	assert.Equal(t, len(ppr.Status.History)+1, len(savedOrder.Status.History))
	assert.Equal(t, ShipmentPending, savedOrder.Status.Current)
}
func TestSellerApprovalPendingRejected(t *testing.T) {
	// Create ppr
	ppr := createPaymentRequestSampleFull()
	// Delete test order
	_, err := App.mongo.DeleteOne(MongoDB, Orders, bson.D{{"ordernumber", ppr.OrderNumber}})
	assert.Nil(t, err)
	statusHistory := StatusHistory{
		Status:    PaymentPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    PaymentSuccess,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    SellerApprovalPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "auto approval",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	ppr.Status.Current = SellerApprovalPending

	_, err = App.mongo.InsertOne(MongoDB, Orders, ppr)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnOrderApproved, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConnOrderApproved)

	approveRequest := &pb.ApprovalRequest{
		OrderNumber: ppr.OrderNumber,
		Approval:    false,
		Reason:      "out of stock",
	}

	resApproval, err := OrderService.SellerApprovalPending(ctx, approveRequest)
	assert.Nil(t, err)
	assert.Equal(t, ppr.OrderNumber, resApproval.OrderNumber)

	savedOrder, err := GetOrder(ppr.OrderNumber)
	assert.Nil(t, err)
	assert.Equal(t, len(ppr.Status.History)+2, len(savedOrder.Status.History))
	assert.Equal(t, PayToBuyer, savedOrder.Status.Current)
}
func TestShipmentDetail(t *testing.T) {
	// Create ppr
	ppr := createPaymentRequestSampleFull()
	// Delete test order
	_, err := App.mongo.DeleteOne(MongoDB, Orders, bson.D{{"ordernumber", ppr.OrderNumber}})
	assert.Nil(t, err)
	statusHistory := StatusHistory{
		Status:    PaymentPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    PaymentSuccess,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    SellerApprovalPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "auto approval",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    ShipmentPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "seller",
		Reason:    "hale",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)

	ppr.Status.Current = ppr.Status.History[(len(ppr.Status.History) - 1)].Status

	// insert to mongo
	_, err = App.mongo.InsertOne(MongoDB, Orders, ppr)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	// call grpc
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnOrderApproved, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConnOrderApproved)

	shipmentDetail := &pb.ShipmentDetailRequest{
		OrderNumber:            ppr.OrderNumber,
		ShipmentTrackingNumber: "Track1234",
		ShipmentProvider:       "SnappBox",
	}
	resDetail, err := OrderService.ShipmentDetail(ctx, shipmentDetail)
	assert.Nil(t, err)
	assert.Equal(t, ppr.OrderNumber, resDetail.OrderNumber)

	savedOrder, err := GetOrder(ppr.OrderNumber)
	assert.Nil(t, err)
	assert.Equal(t, len(ppr.Status.History)+1, len(savedOrder.Status.History))
	assert.Equal(t, Shipped, savedOrder.Status.Current)
}
func TestBuyerCancel(t *testing.T) {
	// Create ppr
	ppr := createPaymentRequestSampleFull()
	// Delete test order
	_, err := App.mongo.DeleteOne(MongoDB, Orders, bson.D{{"ordernumber", ppr.OrderNumber}})
	assert.Nil(t, err)
	statusHistory := StatusHistory{
		Status:    PaymentPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    PaymentSuccess,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    SellerApprovalPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "auto approval",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    ShipmentPending,
		CreatedAt: time.Now().UTC(),
		Agent:     "seller",
		Reason:    "hale",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	statusHistory = StatusHistory{
		Status:    ShipmentDetailDelayed,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "no action for x days",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	ppr.Status.Current = ppr.Status.History[(len(ppr.Status.History) - 1)].Status

	_, err = App.mongo.InsertOne(MongoDB, Orders, ppr)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnOrderApproved, err := grpc.DialContext(ctx, ":"+fmt.Sprint(App.config.App.Port), grpc.WithInsecure())
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConnOrderApproved)

	request := &pb.BuyerCancelRequest{
		OrderNumber: ppr.OrderNumber,
		Reason:      "its took soo much time",
	}

	resApproval, err := OrderService.BuyerCancel(ctx, request)
	assert.Nil(t, err)
	assert.Equal(t, ppr.OrderNumber, resApproval.OrderNumber)

	savedOrder, err := GetOrder(ppr.OrderNumber)
	assert.Nil(t, err)
	assert.Equal(t, len(ppr.Status.History)+1, len(savedOrder.Status.History))
	assert.Equal(t, ShipmentCanceled, savedOrder.Status.Current)
}
