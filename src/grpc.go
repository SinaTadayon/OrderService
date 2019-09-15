package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"gitlab.faza.io/go-framework/kafkaadapter"

	"github.com/rs/xid"

	"gitlab.faza.io/go-framework/logger"
	pb "gitlab.faza.io/protos/payment"
	"google.golang.org/grpc"
)

type PaymentServer struct{}

func startGrpc() {
	lis, err := net.Listen("tcp", ":"+App.config.App.Port)
	if err != nil {
		logger.Err("Failed to listen to TCP on port " + App.config.App.Port + err.Error())
	}
	logger.Audit("app started at " + App.config.App.Port)

	// Start GRPC server and register the server
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, &PaymentServer{})
	if err := grpcServer.Serve(lis); err != nil {
		logger.Err("Failed to listen to gRPC server. " + err.Error())
	}
}

func (PaymentServer *PaymentServer) NewOrder(ctx context.Context, req *pb.OrderPaymentRequest) (*pb.OrderResponse, error) {
	ppr := PaymentPendingRequest{}
	ppr.OrderNumber = generateOrderNumber()
	ppr.CreatedAt = time.Now().UTC()
	ppr.Status.CreatedAt = time.Now().UTC()
	ppr.Status.Current = PaymentSuccess
	ppr.Status.History = []StatusHistory{}
	// validate request & convert to PaymentPendingRequest
	if req.Amount != nil {
		ppr.Amount.Discount = float64(req.Amount.Discount)
		ppr.Amount.Payable = float64(req.Amount.Payable)
		ppr.Amount.Total = float64(req.Amount.Total)
	}
	if req.Buyer != nil {
		ppr.Buyer.LastName = req.Buyer.LastName
		ppr.Buyer.FirstName = req.Buyer.FirstName
		ppr.Buyer.Email = req.Buyer.Email
		ppr.Buyer.Mobile = req.Buyer.Mobile
		ppr.Buyer.NationalId = req.Buyer.NationalId
		ppr.Buyer.IP = req.Buyer.Ip
		if req.Buyer.Finance != nil {
			ppr.Buyer.Finance.Iban = req.Buyer.Finance.Iban
		}
		if req.Buyer.Address != nil {
			ppr.Buyer.Address.Address = req.Buyer.Address.Address
			ppr.Buyer.Address.State = req.Buyer.Address.State
			ppr.Buyer.Address.Phone = req.Buyer.Address.Phone
			ppr.Buyer.Address.ZipCode = req.Buyer.Address.ZipCode
			ppr.Buyer.Address.City = req.Buyer.Address.City
			ppr.Buyer.Address.Country = req.Buyer.Address.Country
			ppr.Buyer.Address.Lat = req.Buyer.Address.Lat
			ppr.Buyer.Address.Lan = req.Buyer.Address.Lan
		}
	}
	if req.Items != nil {
		for _, item := range req.Items {
			var i = Item{}
			i.Quantity = item.Quantity
			i.Sku = item.Sku
			i.Title = item.Title
			i.Categories = item.Categories
			i.Brand = item.Brand
			i.Warranty = item.Warranty
			if item.Price != nil {
				i.Price.Total = float64(item.Price.Total)
				i.Price.Payable = float64(item.Price.Payable)
				i.Price.Discount = float64(item.Price.Discount)
				i.Price.SellerCommission = float64(item.Price.SellerCommission)
				i.Price.Unit = float64(item.Price.Unit)
			}
			if item.Seller != nil {
				i.Seller.Title = item.Seller.Title
				i.Seller.NationalId = item.Seller.NationalId
				i.Seller.Mobile = item.Seller.Mobile
				i.Seller.Email = item.Seller.Email
				i.Seller.FirstName = item.Seller.FirstName
				i.Seller.LastName = item.Seller.LastName
				i.Seller.CompanyName = item.Seller.CompanyName
				i.Seller.EconomicCode = item.Seller.EconomicCode
				i.Seller.RegistrationName = item.Seller.RegistrationName
				if item.Seller.Address != nil {
					i.Seller.Address.Address = item.Seller.Address.Address
					i.Seller.Address.Lan = item.Seller.Address.Lan
					i.Seller.Address.Lat = item.Seller.Address.Lat
					i.Seller.Address.Country = item.Seller.Address.Country
					i.Seller.Address.City = item.Seller.Address.City
					i.Seller.Address.ZipCode = item.Seller.Address.ZipCode
					i.Seller.Address.Title = item.Seller.Address.Title
					i.Seller.Address.Phone = item.Seller.Address.Phone
					i.Seller.Address.State = item.Seller.Address.State
				}
				if item.Seller.Finance != nil {
					i.Seller.Finance.Iban = item.Seller.Finance.Iban
				}
			}
			if item.Shipment != nil {
				i.Shipment.ProviderName = item.Shipment.ProviderName
				i.Shipment.ReactionTime = item.Shipment.ReactionTime
				i.Shipment.ReturnTime = item.Shipment.ReturnTime
				i.Shipment.ShippingTime = item.Shipment.ShippingTime
				i.Shipment.ShipmentDetail = item.Shipment.ShipmentDetail
			}
			ppr.Items = append(ppr.Items, i)
		}
	}
	// validate payment pending request
	err := ppr.validate()
	if err != nil {
		return &pb.OrderResponse{Status: string(http.StatusBadRequest)}, err
	}
	// insert into mongo
	_, err = App.mongo.InsertOne(MongoDB, Orders, ppr)
	if err != nil {
		return &pb.OrderResponse{OrderNumber: "", Status: string(http.StatusInternalServerError), RedirectUrl: ""}, err
	}

	// @todo: remove this mock - start
	App.kafka = kafkaadapter.NewKafka(brokers, "payment-success")
	App.kafka.Config.Producer.Return.Successes = true
	jsons, err := json.Marshal(ppr)
	if err != nil {
		logger.Err("cant convert to json: %v", err)
	}

	_, _, err = App.kafka.SendOne("", jsons)
	if err != nil {
		logger.Err("cant insert to kafka: %v", err)
	}
	// @todo: remove this mock - end

	return &pb.OrderResponse{OrderNumber: ppr.OrderNumber, Status: string(http.StatusOK), RedirectUrl: PaymentUrl}, nil
}
func (PaymentServer *PaymentServer) SellerApprovalPending(ctx context.Context, req *pb.ApprovalRequest) (*pb.ApprovalResponse, error) {
	ppr, err := GetOrder(req.GetOrderNumber())
	if err != nil {
		logger.Err("can't get order: %v", err)
	}

	if req.Approval {
		err = SellerApprovalPendingApproved(ppr)
		if err != nil {
			logger.Err("seller approval pending approved failed: %v", err)
			return &pb.ApprovalResponse{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
				Message: err.Error()}, err
		}
	} else {
		err = SellerApprovalPendingRejected(ppr, req.Reason)
		if err != nil {
			logger.Err("seller approval pending rejected failed: %v", err)
			return &pb.ApprovalResponse{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
				Message: err.Error()}, err
		}
	}

	return &pb.ApprovalResponse{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}

func generateOrderNumber() string {
	id := xid.New()
	return id.String()
}
