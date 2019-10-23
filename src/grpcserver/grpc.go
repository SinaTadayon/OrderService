package grpcserver

import (
	"context"
	"github.com/golang/protobuf/ptypes"

	//"errors"
	"net"
	//"net/http"
	//"time"

	_ "github.com/devfeel/mapper"
	"gitlab.faza.io/go-framework/logger"
	pb "gitlab.faza.io/protos/order"
	"gitlab.faza.io/protos/order/general"
	"google.golang.org/grpc"
)

type OrderServer struct{
	pb.UnimplementedOrderServiceServer
}

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

// TODO error handling
// TODO mongo query for id
// TODO mapping from order request to order model
// TODO Test Response
func (orderSrv *OrderServer) OrderRequestsHandler(ctx context.Context, req *message.Request) (*message.Response, error) {
	logger.Audit("Req Id: %v", req.GetId())
	logger.Audit("Req timestamp: %v", req.GetTime())
	logger.Audit("Req Meta Page: %v", req.GetMeta().GetPage())
	logger.Audit("Req Meta PerPage %v", req.GetMeta().GetPerPage())
	logger.Audit("Req Meta Sorts: ")
	for index, sort := range req.GetMeta().GetSorts() {
		logger.Audit("Req Meta Sort[%d].name: %v", index, sort.GetName())
		logger.Audit("Req Meta Sort[%d].name: %v", index, sort.GetDirection())
	}
	logger.Audit("Req Meta Filters: ")
	for index, filter := range req.GetMeta().GetFilters() {
		logger.Audit("Req Meta Filter.filters[%d].name = %v", index, filter.GetName())
		logger.Audit("Req Meta Filter.filters[%d].opt = %v", index, filter.GetOpt())
		logger.Audit("Req Meta Filter.filters[%d].values = %v", index, filter.GetValue())
	}

	var newOrderRequest pb.NewOrderRequest
	if err := ptypes.UnmarshalAny(req.Data, &newOrderRequest); err != nil {
		logger.Err("Could not unmarshal OrderRequest from anything field: %s", err)
		return &message.Response{}, err
	}

	logger.Audit("Req NewOrderRequest Buyer.firstName: %v", newOrderRequest.GetBuyer().GetFirstName())
	logger.Audit("Req NewOrderRequest Buyer.lastName: %v", newOrderRequest.GetBuyer().GetLastName())
	logger.Audit("Req NewOrderRequest Buyer.finance: %v", newOrderRequest.GetBuyer().GetFinance())
	logger.Audit("Req NewOrderRequest Buyer.Address: %v", newOrderRequest.GetBuyer().GetAddress())

	return &message.Response{}, nil
}

//func addStateRule(s ...string) map[string]bool {
//	m := make(map[string]bool)
//	for _, status := range s {
//		m[status] = true
//	}
//	return m
//}
//
//func addGrpcStateRule() {
//	GrpcStatesRules.SellerApprovalPending = addStateRule(SellerApprovalPending)
//	GrpcStatesRules.ShipmentDetail = addStateRule(ShipmentPending, ShipmentDetailDelayed)
//	GrpcStatesRules.BuyerCancel = addStateRule(ShipmentDetailDelayed)
//	GrpcStatesRules.Delivered = addStateRule(Shipped, ShipmentDeliveryDelayed)
//	GrpcStatesRules.ShipmentDeliveryDelayed = addStateRule(ShipmentDeliveryPending)
//	GrpcStatesRules.ReturnShipmentDeliveryDelayed = addStateRule(ReturnShipmentDeliveryPending)
//	GrpcStatesRules.ShipmentCanceled = addStateRule(ShipmentDetailDelayed, ShipmentDeliveryDelayed)
//	GrpcStatesRules.ReturnShipmentCanceled = addStateRule(ReturnShipmentDeliveryProblem, ReturnShipmentDeliveryDelayed)
//	GrpcStatesRules.ShipmentDeliveryProblem = addStateRule(ShipmentDelivered)
//	GrpcStatesRules.ReturnShipmentDeliveryProblem = addStateRule(ReturnShipmentDelivered)
//	GrpcStatesRules.ShipmentSuccess = addStateRule(ReturnShipmentDetailDelayed, ShipmentDeliveryProblem, ShipmentDelivered)
//	GrpcStatesRules.ReturnShipmentPending = addStateRule(ShipmentDelivered, ShipmentDeliveryProblem)
//	GrpcStatesRules.ReturnShipmentDetail = addStateRule(ReturnShipmentPending, ReturnShipmentDetailDelayed)
//	GrpcStatesRules.ReturnShipmentDelivered = addStateRule(ReturnShipmentDeliveryDelayed, ReturnShipmentDeliveryPending, ReturnShipped)
//	GrpcStatesRules.ReturnShipmentSuccess = addStateRule(ReturnShipmentDeliveryProblem, ReturnShipmentDelivered)
//	GrpcStatesRules.PayToBuyerSuccess = addStateRule(PayToBuyer, PayToBuyerFailed)
//	GrpcStatesRules.PayToSellerSuccess = addStateRule(PayToSeller, PayToSellerFailed)
//	GrpcStatesRules.PayToMarketSuccess = addStateRule(PayToMarket, PayToMarketFailed)
//}

func startGrpc(port string) {
	//addGrpcStateRule()

	lis, err := net.Listen("tcp", ":"+ port)
	if err != nil {
		logger.Err("Failed to listen to TCP on port " + port + err.Error())
	}
	logger.Audit("app started at " + port)

	// Start GRPC server and register the server
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, &OrderServer{})
	if err := grpcServer.Serve(lis); err != nil {
		logger.Err("Failed to listen to gRPC server. " + err.Error())
	}
}

// TODO Check ACL and Security with Mostafa SDK
// TODO Check Order Owner
// TODO: add grpc context validation for all
// TODO: Request / Response Payment Service
// TODO: Add notifications - SMS -farzan SDK
// TODO: Add Product id to Add RPC Order Request / Response
// TODO: API Server GRPC impl
// response-error: StatusNotAcceptable, StatusInternalServerError, StatusOK
//func (PaymentServer *PaymentServer) NewOrder(ctx context.Context, req *pb.OrderPaymentRequest) (*pb.OrderResponse, error) {
//	ppr := PaymentPendingRequest{}
//	ppr.OrderNumber = generateOrderNumber()
//	ppr.CreatedAt = time.Now().UTC()
//	ppr.Status.CreatedAt = time.Now().UTC()
//	ppr.Status.Current = PaymentPending
//	ppr.Status.History = []StatusHistory{}
//	// validate request & convert to PaymentPendingRequest
//	if req.Amount != nil {
//		ppr.Amount.Discount = float64(req.Amount.Discount)
//		ppr.Amount.Payable = float64(req.Amount.Payable)
//		ppr.Amount.Total = float64(req.Amount.Total)
//	}
//	if req.Buyer != nil {
//		ppr.Buyer.LastName = req.Buyer.LastName
//		ppr.Buyer.FirstName = req.Buyer.FirstName
//		ppr.Buyer.Email = req.Buyer.Email
//		ppr.Buyer.Mobile = req.Buyer.Mobile
//		ppr.Buyer.NationalId = req.Buyer.NationalId
//		ppr.Buyer.IP = req.Buyer.Ip
//		if req.Buyer.Finance != nil {
//			ppr.Buyer.Finance.Iban = req.Buyer.Finance.Iban
//		}
//		if req.Buyer.Address != nil {
//			ppr.Buyer.Address.Address = req.Buyer.Address.Address
//			ppr.Buyer.Address.State = req.Buyer.Address.State
//			ppr.Buyer.Address.Phone = req.Buyer.Address.Phone
//			ppr.Buyer.Address.ZipCode = req.Buyer.Address.ZipCode
//			ppr.Buyer.Address.City = req.Buyer.Address.City
//			ppr.Buyer.Address.Country = req.Buyer.Address.Country
//			ppr.Buyer.Address.Lat = req.Buyer.Address.Lat
//			ppr.Buyer.Address.Lan = req.Buyer.Address.Lan
//		}
//	}
//	if req.Items != nil {
//		for _, item := range req.Items {
//			var i = Item{}
//			i.Quantity = item.Quantity
//			i.Sku = item.Sku
//			i.Title = item.Title
//			i.Categories = item.Categories
//			i.Brand = item.Brand
//			i.Warranty = item.Warranty
//			if item.Price != nil {
//				i.Price.Total = float64(item.Price.Total)
//				i.Price.Payable = float64(item.Price.Payable)
//				i.Price.Discount = float64(item.Price.Discount)
//				i.Price.SellerCommission = float64(item.Price.SellerCommission)
//				i.Price.Unit = float64(item.Price.Unit)
//			}
//			if item.Seller != nil {
//				i.Seller.Title = item.Seller.Title
//				i.Seller.NationalId = item.Seller.NationalId
//				i.Seller.Mobile = item.Seller.Mobile
//				i.Seller.Email = item.Seller.Email
//				i.Seller.FirstName = item.Seller.FirstName
//				i.Seller.LastName = item.Seller.LastName
//				i.Seller.CompanyName = item.Seller.CompanyName
//				i.Seller.EconomicCode = item.Seller.EconomicCode
//				i.Seller.RegistrationName = item.Seller.RegistrationName
//				if item.Seller.Address != nil {
//					i.Seller.Address.Address = item.Seller.Address.Address
//					i.Seller.Address.Lan = item.Seller.Address.Lan
//					i.Seller.Address.Lat = item.Seller.Address.Lat
//					i.Seller.Address.Country = item.Seller.Address.Country
//					i.Seller.Address.City = item.Seller.Address.City
//					i.Seller.Address.ZipCode = item.Seller.Address.ZipCode
//					i.Seller.Address.Title = item.Seller.Address.Title
//					i.Seller.Address.Phone = item.Seller.Address.Phone
//					i.Seller.Address.State = item.Seller.Address.State
//				}
//				if item.Seller.Finance != nil {
//					i.Seller.Finance.Iban = item.Seller.Finance.Iban
//				}
//			}
//			if item.Shipment != nil {
//				i.Shipment.ProviderName = item.Shipment.ProviderName
//				i.Shipment.ReactionTime = item.Shipment.ReactionTime
//				i.Shipment.ReturnTime = item.Shipment.ReturnTime
//				i.Shipment.ShippingTime = item.Shipment.ShippingTime
//				i.Shipment.ShipmentDetail = item.Shipment.ShipmentDetail
//			}
//			ppr.Items = append(ppr.Items, i)
//		}
//	}
//	// validate payment pending request
//	err := ppr.validate()
//	if err != nil {
//		return &pb.OrderResponse{Status: string(http.StatusBadRequest)}, err
//	}
//	statusHistory := StatusHistory{
//		Status:    ppr.Status.Current,
//		CreatedAt: time.Now().UTC(),
//		Agent:     "system",
//		Reason:    "",
//	}
//	ppr.Status.History = append(ppr.Status.History, statusHistory)
//	// insert into mongo
//	_, err = App.mongo.InsertOne(MongoDB, Orders, ppr)
//	if err != nil {
//		return &pb.OrderResponse{OrderNumber: "", Status: string(http.StatusInternalServerError), RedirectUrl: ""}, err
//	}
//
//	// @todo: remove this mock - start
//	err = MoveOrderToNewState("system", "", PaymentSuccess, "payment-success", ppr)
//	if err != nil {
//		logger.Err(err.Error())
//	}
//	// @todo: remove this mock - end
//
//	return &pb.OrderResponse{OrderNumber: ppr.OrderNumber, Status: string(http.StatusOK), RedirectUrl: PaymentUrl}, nil
//}
//func (PaymentServer *PaymentServer) SellerApprovalPending(ctx context.Context, req *pb.ApprovalRequest) (*pb.Response, error) {
//	//userClient.NewClient()
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.SellerApprovalPending[ppr.Status.Current]; !ok {
//		logger.Err("seller approval pending no allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "seller approval pending no allowed for this order: " + ppr.OrderNumber},
//			errors.New("seller approval pending no allowed for this order: " + ppr.OrderNumber)
//	}
//
//	if req.Approval {
//		err = SellerApprovalPendingApproved(ppr)
//		if err != nil {
//			logger.Err("seller approval pending approved failed: %v", err)
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	} else {
//		err = SellerApprovalPendingRejected(ppr, req.Reason)
//		if err != nil {
//			logger.Err("seller approval pending rejected failed: %v", err)
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ShipmentDetail(ctx context.Context, req *pb.ShipmentDetailRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ShipmentDetail[ppr.Status.Current]; !ok {
//		logger.Err("shipment detail no allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "shipment detail no allowed for this order: " + ppr.OrderNumber},
//			errors.New("shipment detail no allowed for this order: " + ppr.OrderNumber)
//	}
//
//	if req.GetShipmentProvider() == "" {
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "shipment provider not defined"}, errors.New("shipment provider not defined")
//	}
//
//	err = ShipmentPendingEnteredDetail(ppr, req)
//	if err != nil {
//		logger.Err("shipment detail enter failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) BuyerCancel(ctx context.Context, req *pb.BuyerCancelRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.BuyerCancel[ppr.Status.Current]; !ok {
//		logger.Err("buyer cancel not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "buyer cancel not allowed for this order: " + ppr.OrderNumber},
//			errors.New("buyer cancel not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = BuyerCancel(ppr, req)
//	if err != nil {
//		logger.Err("buyer cancel failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ShipmentDelivered(ctx context.Context, req *pb.ShipmentDeliveredRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.Delivered[ppr.Status.Current]; !ok {
//		logger.Err("delivered not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "delivered not allowed for this order: " + ppr.OrderNumber},
//			errors.New("delivered not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ShipmentDeliveredAction(ppr, req)
//	if err != nil {
//		logger.Err("delivered failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ShipmentDeliveryDelayed(ctx context.Context, req *pb.ShipmentDeliveryDelayedRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ShipmentDeliveryDelayed[ppr.Status.Current]; !ok {
//		logger.Err("shipment delivery delayed not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "shipment delivery delayed not allowed for this order: " + ppr.OrderNumber},
//			errors.New("shipment delivery delayed not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ShipmentDeliveryDelay(ppr, req)
//	if err != nil {
//		logger.Err("shipment delivered failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ReturnShipmentDeliveryDelayed(ctx context.Context, req *pb.ReturnShipmentDeliveryDelayedRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ReturnShipmentDeliveryDelayed[ppr.Status.Current]; !ok {
//		logger.Err("return shipment delivery delayed not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "return shipment delivery delayed not allowed for this order: " + ppr.OrderNumber},
//			errors.New("return shipment delivery delayed not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ReturnShipmentDeliveryDelay(ppr, req)
//	if err != nil {
//		logger.Err("return shipment delivered delayed failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ShipmentCanceled(ctx context.Context, req *pb.ShipmentCanceledRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ShipmentCanceled[ppr.Status.Current]; !ok {
//		logger.Err("shipment canceled not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "shipment canceled not allowed for this order: " + ppr.OrderNumber},
//			errors.New("shipment canceled not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ShipmentCanceledActoin(ppr, req)
//	if err != nil {
//		logger.Err("shipment canceled failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ReturnShipmentCanceled(ctx context.Context, req *pb.ReturnShipmentCanceledRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ReturnShipmentCanceled[ppr.Status.Current]; !ok {
//		logger.Err("return shipment canceled not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "return shipment canceled not allowed for this order: " + ppr.OrderNumber},
//			errors.New("return shipment canceled not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ReturnShipmentCanceledActoin(ppr, req)
//	if err != nil {
//		logger.Err("return shipment canceled failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ShipmentDeliveryProblem(ctx context.Context, req *pb.ShipmentDeliveryProblemRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ShipmentDeliveryProblem[ppr.Status.Current]; !ok {
//		logger.Err("shipment delivery problem not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "shipment delivery problem not allowed for this order: " + ppr.OrderNumber},
//			errors.New("shipment delivery problem not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ShipmentDeliveryProblemAction(ppr, req)
//	if err != nil {
//		logger.Err("shipment delivery problem failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ReturnShipmentDeliveryProblem(ctx context.Context, req *pb.ReturnShipmentDeliveryProblemRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ReturnShipmentDeliveryProblem[ppr.Status.Current]; !ok {
//		logger.Err("return shipment delivery problem not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "return shipment delivery problem not allowed for this order: " + ppr.OrderNumber},
//			errors.New("return shipment delivery problem not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ReturnShipmentDeliveryProblemAction(ppr, req)
//	if err != nil {
//		logger.Err("return shipment delivery problem failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ShipmentSuccess(ctx context.Context, req *pb.ShipmentSuccessRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ShipmentSuccess[ppr.Status.Current]; !ok {
//		logger.Err("shipment success not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "shipment success not allowed for this order: " + ppr.OrderNumber},
//			errors.New("shipment success not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ShipmentSuccessAction(ppr, req)
//	if err != nil {
//		logger.Err("shipment success failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ReturnShipmentPending(ctx context.Context, req *pb.ReturnShipmentPendingRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ReturnShipmentPending[ppr.Status.Current]; !ok {
//		logger.Err("return shipment pending not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "return shipment pending not allowed for this order: " + ppr.OrderNumber},
//			errors.New("return shipment pending not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ReturnShipmentPendingAction(ppr, req)
//	if err != nil {
//		logger.Err("return shipment pending failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ReturnShipmentDetail(ctx context.Context, req *pb.ReturnShipmentDetailRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ReturnShipmentDetail[ppr.Status.Current]; !ok {
//		logger.Err("return shipment detail not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "return shipment detail not allowed for this order: " + ppr.OrderNumber},
//			errors.New("return shipment detail not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	if req.GetShipmentProvider() == "" || req.GetShipmentTrackingNumber() == "" {
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "shipment provider not defined"}, errors.New("shipment provider not defined")
//	}
//
//	err = ReturnShipmentDetailAction(ppr, req)
//	if err != nil {
//		logger.Err("return shipment detail failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ReturnShipmentDelivered(ctx context.Context, req *pb.ReturnShipmentDeliveredRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ReturnShipmentDelivered[ppr.Status.Current]; !ok {
//		logger.Err("return shipment delivered not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "return shipment delivered not allowed for this order: " + ppr.OrderNumber},
//			errors.New("return shipment delivered not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ReturnShipmentDeliveredAction(ppr, req)
//	if err != nil {
//		logger.Err("return shipment delivered failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) ReturnShipmentSuccess(ctx context.Context, req *pb.ReturnShipmentSuccessRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.ReturnShipmentSuccess[ppr.Status.Current]; !ok {
//		logger.Err("return shipment success not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "return shipment success not allowed for this order: " + ppr.OrderNumber},
//			errors.New("return shipment success not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = ReturnShipmentDeliveredGrpcAction(ppr, req)
//	if err != nil {
//		logger.Err("return shipment success failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//
//func (PaymentServer *PaymentServer) PayToBuyerSuccess(ctx context.Context, req *pb.PayToBuyerSuccessRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.PayToBuyerSuccess[ppr.Status.Current]; !ok {
//		logger.Err("pay to buyer not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "pay to buyer not allowed for this order: " + ppr.OrderNumber},
//			errors.New("pay to buyer not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = PayToBuyerSuccessAction(ppr, req)
//	if err != nil {
//		logger.Err("pay to buyer failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) PayToSellerSuccess(ctx context.Context, req *pb.PayToSellerSuccessRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.PayToSellerSuccess[ppr.Status.Current]; !ok {
//		logger.Err("pay to seller not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "pay to seller not allowed for this order: " + ppr.OrderNumber},
//			errors.New("pay to seller not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = PayToSellerSuccessAction(ppr, req)
//	if err != nil {
//		logger.Err("pay to seller failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
//func (PaymentServer *PaymentServer) PayToMarketSuccess(ctx context.Context, req *pb.PayToMarketSuccessRequest) (*pb.Response, error) {
//	ppr, err := GetOrder(req.GetOrderNumber())
//	if err != nil {
//		logger.Err("can't get order: %v", err)
//		return &pb.Response{OrderNumber: "", Status: string(http.StatusNotAcceptable),
//			Message: "can't get order: " + err.Error()}, err
//	}
//	// check grpc status with state machine rules
//	if _, ok := GrpcStatesRules.PayToMarketSuccess[ppr.Status.Current]; !ok {
//		logger.Err("pay to market not allowed for this order: %v", ppr.OrderNumber)
//		return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusNotAcceptable),
//				Message: "pay to market not allowed for this order: " + ppr.OrderNumber},
//			errors.New("pay to market not allowed for this order: " + ppr.OrderNumber)
//	}
//
//	err = PayToMarketSuccessAction(ppr, req)
//	if err != nil {
//		logger.Err("pay to market failed: %v", err)
//		if err.Error() == StateMachineNextStateNotAvailable {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusBadRequest),
//				Message: err.Error()}, err
//		} else {
//			return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusInternalServerError),
//				Message: err.Error()}, err
//		}
//	}
//
//	return &pb.Response{OrderNumber: req.OrderNumber, Status: string(http.StatusOK)}, nil
//}
