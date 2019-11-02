package grpc_server

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"

	//"errors"
	"net"
	//"net/http"
	//"time"

	_ "github.com/devfeel/mapper"
	"gitlab.faza.io/go-framework/logger"
	"google.golang.org/grpc"
)

type Server struct{
	pb.UnimplementedOrderServiceServer
	pg.UnimplementedBankResultHookServer
	flowManager 	domain.IFlowManager
	address 		string
	port			uint16
}

func NewServer(address string, port uint16, flowManager domain.IFlowManager) Server {
	return Server{flowManager:flowManager, address:address, port:port}
}

// TODO error handling
// TODO mongo query for id
// TODO mapping from order request to order model
// TODO Test Response
func (server *Server) OrderRequestsHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {

	flowManagerCtx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	promiseHandler := server.flowManager.MessageHandler(flowManagerCtx, req)
	futureData := <- promiseHandler.Channel()
	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	response, ok := futureData.Data.(pb.MessageResponse)
	if ok != true {
		logger.Err("received data of futureData invalid, type: %T, value, %v", futureData.Data, futureData.Data)
		return nil, status.Error(500, "Unknown Error")
	}

	return &response, nil

	//logger.Audit("Req Order Id: %v", req.GetOrderId())
	//logger.Audit("Req Item Id: %v", req.GetItemId())
	//logger.Audit("Req timestamp: %v", req.GetTime())
	//logger.Audit("Req Meta Page: %v", req.GetMeta().GetPage())
	//logger.Audit("Req Meta PerPage %v", req.GetMeta().GetPerPage())
	//logger.Audit("Req Meta Sorts: ")
	//for index, sort := range req.GetMeta().GetSorts() {
	//	logger.Audit("Req Meta Sort[%d].name: %v", index, sort.GetName())
	//	logger.Audit("Req Meta Sort[%d].name: %v", index, sort.GetDirection())
	//}
	//logger.Audit("Req Meta Filters: ")
	//for index, filter := range req.GetMeta().GetFilters() {
	//	logger.Audit("Req Meta Filter.filters[%d].name = %v", index, filter.GetName())
	//	logger.Audit("Req Meta Filter.filters[%d].opt = %v", index, filter.GetOpt())
	//	logger.Audit("Req Meta Filter.filters[%d].values = %v", index, filter.GetValue())
	//}
	//
	//var RequestNewOrder pb.RequestNewOrder
	//if err := ptypes.UnmarshalAny(req.Data, &RequestNewOrder); err != nil {
	//	logger.Err("Could not unmarshal OrderRequest from anything field: %s", err)
	//	return &message.Response{}, err
	//}
	//
	//logger.Audit("Req RequestNewOrder Buyer.firstName: %v", RequestNewOrder.GetBuyer().GetFirstName())
	//logger.Audit("Req RequestNewOrder Buyer.lastName: %v", RequestNewOrder.GetBuyer().GetLastName())
	//logger.Audit("Req RequestNewOrder Buyer.finance: %v", RequestNewOrder.GetBuyer().GetFinance())
	//logger.Audit("Req RequestNewOrder Buyer.Address: %v", RequestNewOrder.GetBuyer().GetShippingAddress())
	//
	//res1 , err1 := json.Marshal(RequestNewOrder)
	//if err1 != nil {
	//	logger.Err("json.Marshal failed, %s", err1)
	//}
	//
	//logger.Audit("json request: %s	", res1)
	//
	////status.Error(codes.NotFound, "Product Not found")
	//
	//st := status.New(codes.InvalidArgument, "invalid username")
	////desc := "The username must only contain alphanumeric characters"
	//
	//validations := []*message.ValidationErr {
	//	{
	//		Field: "username",
	//		Desc: "value2",
	//	},
	//	{
	//		Field: "password",
	//		Desc: "value2",
	//	},
	//}
	//
	//errDetails := message.ErrorDetails {
	//	Validation: validations,
	//}
	//
	////serializedOrder, err := proto.Marshal(&errDetails)
	////if err != nil {
	////	logger.Err("could not serialize timestamp")
	////}
	////
	////errors.Data = &any.Any{
	////TypeUrl: "baman.io/" + proto.MessageName(&errDetails),
	////Value:   serializedOrder,
	////}
	//
	//st, err := st.WithDetails(&errDetails)
	//if err != nil {
	//	// If this errored, it will always error
	//	// here, so better panic so we can figure
	//	// out why than have this silently passing.
	//	panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
	//}
	//
	//return &message.Response{}, st.Err()
}

func (server *Server) PaymentGatewayHook(ctx context.Context, req *pg.PaygateHookRequest) (*pg.PaygateHookResponse, error) {
	promiseHandler := server.flowManager.PaymentGatewayResult(ctx, req)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pg.PaygateHookResponse{Ok: true}, nil
}

func (server Server) NewOrder(ctx context.Context, req *pb.RequestNewOrder) (*pb.ResponseNewOrder, error) {

	//var request *pb.MessageRequest
	//var response *pb.MessageResponse

	messageRequest := server.convertNewOrderRequestToMessage(req)

	//ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	promiseHandler := server.flowManager.MessageHandler(ctx, messageRequest)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	paymentResponse, ok := futureData.Data.(payment_service.PaymentResponse)
	if ok != true {
		logger.Err("NewOrder received data of futureData invalid, type: %T, value, %v", futureData.Data, futureData.Data)
		return nil, status.Error(500, "Unknown Error")
	}

	responseNewOrder := pb.ResponseNewOrder{
		CallbackUrl: paymentResponse.CallbackUrl,
	}

	return &responseNewOrder, nil

}

func (server Server) SellerFindAllItems(context.Context, *pb.RequestSellerFindAllItems) (*pb.ResponseSellerFindAllItems, error) {
	panic("must be implement")
}

func (server Server) BuyerOrderAction(ctx context.Context, req *pb.RequestBuyerOrderAction) (*pb.ResponseBuyerOrderAction, error) {
	promiseHandler := server.flowManager.BuyerApprovalPending(ctx, req)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pb.ResponseBuyerOrderAction{Result: true}, nil
}

func (server Server) SellerOrderAction(ctx context.Context, req *pb.RequestSellerOrderAction) (*pb.ResponseSellerOrderAction, error) {
	promiseHandler := server.flowManager.SellerApprovalPending(ctx, req)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pb.ResponseSellerOrderAction{Result: true}, nil
}

func (server Server) BuyerFindAllOrders(context.Context, *pb.RequestBuyerFindAllOrders) (*pb.ResponseBuyerFindAllOrders, error) {
	panic("must be implement")
}

func (server Server) convertNewOrderRequestToMessage(req *pb.RequestNewOrder) *pb.MessageRequest {

	serializedOrder, err := proto.Marshal(req)
	if err != nil {
		logger.Err("could not serialize timestamp")
	}

	request := pb.MessageRequest {
		OrderId: "",
		//ItemId: orderId + strconv.Itoa(int(entities.GenerateRandomNumber())),
		Time: ptypes.TimestampNow(),
		Meta: nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(req),
			Value:   serializedOrder,
		},
	}

	return &request
}

func (server Server) Start() {
	//addGrpcStateRule()

	port := strconv.Itoa(int(server.port))
	lis, err := net.Listen("tcp", server.address + ":" + port)
	if err != nil {
		logger.Err("Failed to listen to TCP on port " + port + err.Error())
	}
	logger.Audit("app started at %s:%s", server.address, port)

	// Start GRPC server and register the server
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, &server)
	pg.RegisterBankResultHookServer(grpcServer, &server)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Err("GRPC server start field " + err.Error())
		panic("GRPC server start field")
	}

	//logger.Audit("GRPC server is running . . . ")
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
//			i.Guarantee = item.Guarantee
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
