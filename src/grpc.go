package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/rs/xid"

	"gitlab.faza.io/go-framework/logger"
	pb "gitlab.faza.io/protos/payment"
	"google.golang.org/grpc"
)

type PaymentServer struct{}

var GrpcStatesRules struct {
	SellerApprovalPending         map[string]bool
	ShipmentDetail                map[string]bool
	BuyerCancel                   map[string]bool
	Delivered                     map[string]bool
	ShipmentDeliveryDelayed       map[string]bool
	ReturnShipmentDeliveryDelayed map[string]bool
	ShipmentCanceled              map[string]bool
	ReturnShipmentCanceled        map[string]bool
	ShipmentDeliveryProblem       map[string]bool
	ReturnShipmentDeliveryProblem map[string]bool
	ShipmentSuccess               map[string]bool
	ReturnShipmentPending         map[string]bool
	ReturnShipmentDetail          map[string]bool
	ReturnShipmentDelivered       map[string]bool
	ReturnShipmentSuccess         map[string]bool
	PayToBuyerSuccess             map[string]bool
	PayToSellerSuccess            map[string]bool
	PayToMarketSuccess            map[string]bool
}

func addStateRule(s ...string) map[string]bool {
	m := make(map[string]bool)
	for _, status := range s {
		m[status] = true
	}
	return m
}

func addGrpcStateRule() {
	GrpcStatesRules.SellerApprovalPending = addStateRule(SellerApprovalPending)
	GrpcStatesRules.ShipmentDetail = addStateRule(ShipmentPending, ShipmentDetailDelayed)
	GrpcStatesRules.BuyerCancel = addStateRule(ShipmentDetailDelayed)
	GrpcStatesRules.Delivered = addStateRule(Shipped, ShipmentDeliveryDelayed)
	GrpcStatesRules.ShipmentDeliveryDelayed = addStateRule(ShipmentDeliveryPending)
	GrpcStatesRules.ReturnShipmentDeliveryDelayed = addStateRule(ReturnShipmentDeliveryPending)
	GrpcStatesRules.ShipmentCanceled = addStateRule(ShipmentDetailDelayed, ShipmentDeliveryDelayed)
	GrpcStatesRules.ReturnShipmentCanceled = addStateRule(ReturnShipmentDeliveryProblem, ReturnShipmentDeliveryDelayed)
	GrpcStatesRules.ShipmentDeliveryProblem = addStateRule(ShipmentDelivered)
	GrpcStatesRules.ReturnShipmentDeliveryProblem = addStateRule(ReturnShipmentDelivered)
	GrpcStatesRules.ShipmentSuccess = addStateRule(ReturnShipmentDetailDelayed, ShipmentDeliveryProblem, ShipmentDelivered)
	GrpcStatesRules.ReturnShipmentPending = addStateRule(ShipmentDelivered, ShipmentDeliveryProblem)
	GrpcStatesRules.ReturnShipmentDetail = addStateRule(ReturnShipmentPending, ReturnShipmentDetailDelayed)
	GrpcStatesRules.ReturnShipmentDelivered = addStateRule(ReturnShipmentDeliveryDelayed, ReturnShipmentDeliveryPending, ReturnShipped)
	GrpcStatesRules.ReturnShipmentSuccess = addStateRule(ReturnShipmentDeliveryProblem, ReturnShipmentDelivered)
	GrpcStatesRules.PayToBuyerSuccess = addStateRule(PayToBuyer, PayToBuyerFailed)
	GrpcStatesRules.PayToSellerSuccess = addStateRule(PayToSeller, PayToSellerFailed)
	GrpcStatesRules.PayToMarketSuccess = addStateRule(PayToMarket, PayToMarketFailed)
}

func startGrpc() {
	addGrpcStateRule()

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

// response-error: StatusNotAcceptable, StatusInternalServerError, StatusOK
func (PaymentServer *PaymentServer) NewOrder(ctx context.Context, req *pb.OrderPaymentRequest) (*pb.OrderResponse, error) {
	// @TODO: add grpc context validation for all
	ppr := PaymentPendingRequest{}
	ppr.OrderNumber = generateOrderNumber()
	ppr.CreatedAt = time.Now().UTC()
	ppr.Status.CreatedAt = time.Now().UTC()
	ppr.Status.Current = PaymentPending
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
	statusHistory := StatusHistory{
		Status:    ppr.Status.Current,
		CreatedAt: time.Now().UTC(),
		Agent:     "system",
		Reason:    "",
	}
	ppr.Status.History = append(ppr.Status.History, statusHistory)
	// insert into mongo
	_, err = App.mongo.InsertOne(MongoDB, Orders, ppr)
	if err != nil {
		return &pb.OrderResponse{OrderNumber: "", Status: string(http.StatusInternalServerError), RedirectUrl: ""}, err
	}

	// @todo: remove this mock - start
	err = MoveOrderToNewState("system", "", PaymentSuccess, "payment-success", ppr)
	if err != nil {
		logger.Err(err.Error())
	}
	// @todo: remove this mock - end

	return &pb.OrderResponse{OrderNumber: ppr.OrderNumber, Status: string(http.StatusOK), RedirectUrl: PaymentUrl}, nil
}
func (PaymentServer *PaymentServer) SellerApprovalPending(ctx context.Context, req *pb.ApprovalRequest) (*pb.Response, error) {
	//userClient.NewClient()
	ppr, err := GetOrder(req.GetOrderNumber())
	if err != nil {
		logger.Err("can't get order: %v", err)
		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
			Message: "can't get order: " + err.Error()}, err
	}
	// check grpc status with state machine rules
	if _, ok := GrpcStatesRules.SellerApprovalPending[ppr.Status.Current]; !ok {
		logger.Err("seller approval pending no allowed for this order: %v", ppr.OrderNumber)
		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
			Message: "seller approval pending no allowed for this order: " + ppr.OrderNumber}, err
	}

	if req.Approval {
		err = SellerApprovalPendingApproved(ppr)
		if err != nil {
			logger.Err("seller approval pending approved failed: %v", err)
			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
				Message: err.Error()}, err
		}
	} else {
		err = SellerApprovalPendingRejected(ppr, req.Reason)
		if err != nil {
			logger.Err("seller approval pending rejected failed: %v", err)
			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
				Message: err.Error()}, err
		}
	}

	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ShipmentDetail(ctx context.Context, req *pb.ShipmentDetailRequest) (*pb.Response, error) {
	ppr, err := GetOrder(req.GetOrderNumber())
	if err != nil {
		logger.Err("can't get order: %v", err)
		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
			Message: "can't get order: " + err.Error()}, err
	}
	// check grpc status with state machine rules
	if _, ok := GrpcStatesRules.ShipmentDetail[ppr.Status.Current]; !ok {
		logger.Err("shipment detail no allowed for this order: %v", ppr.OrderNumber)
		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
			Message: "shipment detail no allowed for this order: " + ppr.OrderNumber}, err
	}

	if req.GetShipmentProvider() == "" {
		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
			Message: "shipment provider not defined"}, errors.New("shipment provider not defined")
	}

	err = ShipmentPendingEnteredDetail(ppr, req)
	if err != nil {
		logger.Err("shipment detail enter failed: %v", err)
		if err.Error() == StateMachineNextStateNotAvailable {
			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
				Message: err.Error()}, err
		} else {
			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
				Message: err.Error()}, err
		}
	}

	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}

func (PaymentServer *PaymentServer) BuyerCancel(ctx context.Context, req *pb.BuyerCancelRequest) (*pb.Response, error) {
	ppr, err := GetOrder(req.GetOrderNumber())
	if err != nil {
		logger.Err("can't get order: %v", err)
		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
			Message: "can't get order: " + err.Error()}, err
	}
	// check grpc status with state machine rules
	if _, ok := GrpcStatesRules.ShipmentDetail[ppr.Status.Current]; !ok {
		logger.Err("shipment detail no allowed for this order: %v", ppr.OrderNumber)
		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
			Message: "shipment detail no allowed for this order: " + ppr.OrderNumber}, err
	}

	err = BuyerCancel(ppr, req)
	if err != nil {
		logger.Err("buyer cancel failed: %v", err)
		if err.Error() == StateMachineNextStateNotAvailable {
			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
				Message: err.Error()}, err
		} else {
			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
				Message: err.Error()}, err
		}
	}
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) Delivered(ctx context.Context, req *pb.DeliveredRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ShipmentDeliveryDelayed(ctx context.Context, req *pb.DeliveryDelayedRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ReturnShipmentDeliveryDelayed(ctx context.Context, req *pb.ReturnDeliveryDelayedRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ShipmentCanceled(ctx context.Context, req *pb.ShipmentCanceledRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ReturnShipmentCanceled(ctx context.Context, req *pb.ReturnShipmentCanceledRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ShipmentDeliveryProblem(ctx context.Context, req *pb.ShipmentDeliveryProblemRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ReturnShipmentDeliveryProblem(ctx context.Context, req *pb.ReturnShipmentDeliveryProblemRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ShipmentSuccess(ctx context.Context, req *pb.ShipmentSuccessRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ReturnShipmentPending(ctx context.Context, req *pb.ReturnShipmentPendingRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ReturnShipmentDetail(ctx context.Context, req *pb.ReturnShipmentDetailRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ReturnShipmentDelivered(ctx context.Context, req *pb.ReturnShipmentDeliveredRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) ReturnShipmentSuccess(ctx context.Context, req *pb.ReturnShipmentSuccessRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) PayToBuyerSuccess(ctx context.Context, req *pb.PayToBuyerSuccessRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) PayToSellerSuccess(ctx context.Context, req *pb.PayToSellerSuccessRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}
func (PaymentServer *PaymentServer) PayToMarketSuccess(ctx context.Context, req *pb.PayToMarketSuccessRequest) (*pb.Response, error) {
	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
}

func generateOrderNumber() string {
	id := xid.New()
	return id.String()
}
