package main

import (
	"context"
	"net"
	"net/http"

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
	if req.Amount != nil {
		ppr.amount.discount = float64(req.Amount.Discount)
		ppr.amount.payable = float64(req.Amount.Payable)
		ppr.amount.total = float64(req.Amount.Total)
	}
	if req.Buyer != nil {
		if req.Buyer.Info != nil {
			ppr.buyer.lastName = req.Buyer.Info.LastName
			ppr.buyer.firstName = req.Buyer.Info.FirstName
			ppr.buyer.email = req.Buyer.Info.Email
			ppr.buyer.mobile = req.Buyer.Info.Mobile
			ppr.buyer.nationalId = req.Buyer.Info.NationalId
		}
		if req.Buyer.Finance != nil {
			ppr.buyer.finance.iban = req.Buyer.Finance.Iban
		}
		if req.Buyer.Address != nil {
			ppr.buyer.address.address = req.Buyer.Address.Address
			ppr.buyer.address.state = req.Buyer.Address.State
			ppr.buyer.address.phone = req.Buyer.Address.Phone
			ppr.buyer.address.zipCode = req.Buyer.Address.ZipCode
			ppr.buyer.address.city = req.Buyer.Address.City
			ppr.buyer.address.country = req.Buyer.Address.Country
			ppr.buyer.address.lat = req.Buyer.Address.Lat
			ppr.buyer.address.lan = req.Buyer.Address.Lan
		}
	}
	if req.Items != nil {
		for _, item := range req.Items {
			var i = Item{}
			i.quantity = item.Quantity
			i.sku = item.Sku
			i.title = item.Title
			i.categories = item.Categories
			i.brand = item.Brand
			i.warranty = item.Warranty
			if item.Price != nil {
				i.price.total = float64(item.Price.Total)
				i.price.payable = float64(item.Price.Payable)
				i.price.discount = float64(item.Price.Discount)
				i.price.sellerCommission = float64(item.Price.SellerCommission)
				i.price.unit = float64(item.Price.Unit)
			}
			if item.Seller != nil {
				i.seller.title = item.Seller.Title
				i.seller.nationalId = item.Seller.NationalId
				i.seller.mobile = item.Seller.Mobile
				i.seller.email = item.Seller.Email
				i.seller.firstName = item.Seller.FirstName
				i.seller.lastName = item.Seller.LastName
				i.seller.companyName = item.Seller.CompanyName
				i.seller.economicCode = item.Seller.EconomicCode
				i.seller.registrationName = item.Seller.RegistrationName
				if item.Seller.Address != nil {
					i.seller.address.address = item.Seller.Address.Address
					i.seller.address.lan = item.Seller.Address.Lan
					i.seller.address.lat = item.Seller.Address.Lat
					i.seller.address.country = item.Seller.Address.Country
					i.seller.address.city = item.Seller.Address.City
					i.seller.address.zipCode = item.Seller.Address.ZipCode
					i.seller.address.title = item.Seller.Address.Title
					i.seller.address.phone = item.Seller.Address.Phone
					i.seller.address.state = item.Seller.Address.State
				}
				if item.Seller.Finance != nil {
					i.seller.finance.iban = item.Seller.Finance.Iban
				}
			}
			if item.Shipment != nil {
				i.shipment.providerName = item.Shipment.ProviderName
				i.shipment.reactionTime = item.Shipment.ReactionTime
				i.shipment.returnTime = item.Shipment.ReturnTime
				i.shipment.shippingTime = item.Shipment.ShippingTime
				i.shipment.shipmentDetail = item.Shipment.ShipmentDetail
			}
			ppr.items = append(ppr.items, i)
		}
	}
	ppr.orderNumber = req.OrderNumber

	// validate request
	err := ppr.validate()
	if err != nil {
		return &pb.OrderResponse{Status: string(http.StatusBadRequest)}, err
	}

	return &pb.OrderResponse{OrderNumber: ppr.orderNumber, Status: string(http.StatusOK), RedirectUrl: PaymentUrl}, nil
}
