package grpc_server

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	//"github.com/pkg/errors"
	"net"
	//"net/http"
	//"time"

	_ "github.com/devfeel/mapper"
	"gitlab.faza.io/go-framework/logger"
	"google.golang.org/grpc"
)

const (
	// ISO8601 standard time format
	ISO8601 = "2006-01-02T15:04:05-0700"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	pg.UnimplementedBankResultHookServer
	flowManager domain.IFlowManager
	address     string
	port        uint16
}

func NewServer(address string, port uint16, flowManager domain.IFlowManager) Server {
	return Server{flowManager: flowManager, address: address, port: port}
}

func (server *Server) OrderRequestsHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {

	flowManagerCtx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	promiseHandler := server.flowManager.MessageHandler(flowManagerCtx, req)
	futureData := <-promiseHandler.Channel()
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

	callbackUrl, ok := futureData.Data.(string)
	if ok != true {
		logger.Err("NewOrder received data of futureData invalid, type: %T, value, %v", futureData.Data, futureData.Data)
		return nil, status.Error(500, "Unknown Error")
	}

	responseNewOrder := pb.ResponseNewOrder{
		CallbackUrl: callbackUrl,
	}

	return &responseNewOrder, nil

}

// TODO Add checking acl
func (server Server) SellerFindAllItems(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseSellerFindAllItems, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("SellerFindAllItems() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if strconv.Itoa(int(userAcl.User().UserID)) != req.Id {
		logger.Err(" SellerFindAllItems() => token userId %d not authorized for sellerId %s", userAcl.User().UserID, req.Id)
		return nil, status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

	sellerId, err := strconv.Atoi(req.Id)
	if err != nil {
		logger.Err(" SellerFindAllItems() => sellerId invalid: %s", req.Id)
		return nil, status.Error(codes.Code(promise.BadRequest), "SellerId Invalid")
	}

	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
		return bson.D{{"items.sellerInfo.sellerId", uint64(sellerId)}}
	})

	if err != nil {
		logger.Err("SellerFindAllItems failed, sellerId: %s, error: %s", req.Id, err.Error())
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	sellerItemMap := make(map[string]*pb.SellerFindAllItems, 16)

	for _, order := range orders {
		for _, orderItem := range order.Items {
			if strconv.Itoa(int(orderItem.SellerInfo.SellerId)) == req.Id {
				if _, ok := sellerItemMap[orderItem.InventoryId]; !ok {
					newResponseItem := &pb.SellerFindAllItems{
						OrderId:     order.OrderId,
						ItemId:      orderItem.ItemId,
						InventoryId: orderItem.InventoryId,
						Title:       orderItem.Title,
						Image:       orderItem.Image,
						Returnable:  orderItem.Returnable,
						Status: &pb.Status{
							OrderStatus: order.Status,
							ItemStatus:  orderItem.Status,
							StepStatus:  "none",
						},
						CreatedAt:  orderItem.CreatedAt.Format(ISO8601),
						UpdatedAt:  orderItem.UpdatedAt.Format(ISO8601),
						Quantity:   orderItem.Quantity,
						Attributes: orderItem.Attributes,
						Price: &pb.SellerFindAllItems_Price{
							Unit:             orderItem.Price.Unit,
							Total:            orderItem.Price.Original,
							SellerCommission: orderItem.Price.SellerCommission,
							Currency:         orderItem.Price.Currency,
						},
						DeliveryAddress: &pb.Address{
							FirstName:     order.BuyerInfo.ShippingAddress.FirstName,
							LastName:      order.BuyerInfo.ShippingAddress.LastName,
							Address:       order.BuyerInfo.ShippingAddress.Address,
							Phone:         order.BuyerInfo.ShippingAddress.Phone,
							Mobile:        order.BuyerInfo.ShippingAddress.Mobile,
							Country:       order.BuyerInfo.ShippingAddress.Country,
							City:          order.BuyerInfo.ShippingAddress.City,
							Province:      order.BuyerInfo.ShippingAddress.Province,
							Neighbourhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
							ZipCode:       order.BuyerInfo.ShippingAddress.ZipCode,
						},
					}

					if order.BuyerInfo.ShippingAddress.Location != nil {
						newResponseItem.DeliveryAddress.Lat = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[1])
						newResponseItem.DeliveryAddress.Long = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[0])
					}

					lastStep := orderItem.Progress.StepsHistory[len(orderItem.Progress.StepsHistory)-1]
					if lastStep.ActionHistory != nil {
						lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
						newResponseItem.Status.StepStatus = lastAction.Name
					} else {
						newResponseItem.Status.StepStatus = "none"
						logger.Audit("SellerFindAllItems() => Action History is nil, orderId: %d, itemId: %d", order.OrderId, orderItem.ItemId)
					}

					sellerItemMap[orderItem.InventoryId] = newResponseItem
				}
			}
		}
	}

	var response = pb.ResponseSellerFindAllItems{}
	response.Items = make([]*pb.SellerFindAllItems, 0, len(sellerItemMap))

	for _, item := range sellerItemMap {
		response.Items = append(response.Items, item)
	}

	return &response, nil
}

// TODO Add checking acl
func (server Server) BuyerOrderAction(ctx context.Context, req *pb.RequestBuyerOrderAction) (*pb.ResponseBuyerOrderAction, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BuyerOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if userAcl.User().UserID != int64(req.BuyerId) {
		logger.Err(" BuyerOrderAction() => token userId %d not authorized for buyerId %d", userAcl.User().UserID, req.BuyerId)
		return nil, status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

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

// TODO Add checking acl
func (server Server) SellerOrderAction(ctx context.Context, req *pb.RequestSellerOrderAction) (*pb.ResponseSellerOrderAction, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("SellerOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if userAcl.User().UserID != int64(req.SellerId) {
		logger.Err("SellerOrderAction() => token userId %d not authorized for sellerId %d", userAcl.User().UserID, req.SellerId)
		return nil, status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

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

// TODO Add checking acl
func (server Server) BuyerFindAllOrders(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseBuyerFindAllOrders, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BuyerFindAllOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if strconv.Itoa(int(userAcl.User().UserID)) != req.Id {
		logger.Err(" BuyerFindAllOrders() => token userId %d not authorized of buyerId %s", userAcl.User().UserID, req.Id)
		return nil, status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

	buyerId, err := strconv.Atoi(req.Id)
	if err != nil {
		logger.Err(" SellerFindAllItems() => buyerId invalid: %s", req.Id)
		return nil, status.Error(codes.Code(promise.BadRequest), "BuyerId Invalid")
	}

	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
		return bson.D{{"buyerInfo.buyerId", uint64(buyerId)}}
	})

	if err != nil {
		logger.Err("SellerFindAllItems failed, buyerId: %s, error: %s", req.Id, err.Error())
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	var response pb.ResponseBuyerFindAllOrders
	responseOrders := make([]*pb.BuyerAllOrders, 0, len(orders))

	for _, order := range orders {
		responseOrder := &pb.BuyerAllOrders{
			OrderId:     order.OrderId,
			CreatedAt:   order.CreatedAt.Format(ISO8601),
			UpdatedAt:   order.UpdatedAt.Format(ISO8601),
			OrderStatus: order.Status,
			Amount: &pb.Amount{
				Total:         order.Amount.Total,
				Subtotal:      order.Amount.Subtotal,
				Discount:      order.Amount.Discount,
				Currency:      order.Amount.Currency,
				ShipmentTotal: order.Amount.ShipmentTotal,
				PaymentMethod: order.Amount.PaymentMethod,
				PaymentOption: order.Amount.PaymentOption,
				Voucher: &pb.Voucher{
					Amount: order.Amount.Voucher.Amount,
					Code:   order.Amount.Voucher.Code,
				},
			},
			ShippingAddress: &pb.Address{
				FirstName:     order.BuyerInfo.ShippingAddress.FirstName,
				LastName:      order.BuyerInfo.ShippingAddress.LastName,
				Address:       order.BuyerInfo.ShippingAddress.Address,
				Phone:         order.BuyerInfo.ShippingAddress.Phone,
				Mobile:        order.BuyerInfo.ShippingAddress.Mobile,
				Country:       order.BuyerInfo.ShippingAddress.Country,
				City:          order.BuyerInfo.ShippingAddress.City,
				Province:      order.BuyerInfo.ShippingAddress.Province,
				Neighbourhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
			},
			Items: make([]*pb.BuyerOrderItems, 0, len(order.Items)),
		}

		if order.BuyerInfo.ShippingAddress.Location != nil {
			responseOrder.ShippingAddress.Lat = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[1])
			responseOrder.ShippingAddress.Long = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[0])
		}

		orderItemMap := make(map[string]*pb.BuyerOrderItems, 16)

		for _, item := range order.Items {
			if _, ok := orderItemMap[item.InventoryId]; !ok {
				newResponseOrderItem := &pb.BuyerOrderItems{
					InventoryId: item.InventoryId,
					Title:       item.Title,
					Brand:       item.Brand,
					Category:    item.Category,
					Guaranty:    item.Guaranty,
					Image:       item.Image,
					Returnable:  item.Returnable,
					SellerId:    item.SellerInfo.SellerId,
					Quantity:    item.Quantity,
					Attributes:  item.Attributes,
					ItemStatus:  item.Status,
					Price: &pb.BuyerOrderItems_Price{
						Unit:     item.Price.Unit,
						Total:    item.Price.Total,
						Original: item.Price.Original,
						Special:  item.Price.Special,
						Currency: item.Price.Currency,
					},
					Shipment: &pb.BuyerOrderItems_ShipmentSpec{
						CarrierName:  item.ShipmentSpec.CarrierName,
						ShippingCost: item.ShipmentSpec.ShippingCost,
					},
				}

				lastStep := item.Progress.StepsHistory[len(item.Progress.StepsHistory)-1]

				if lastStep.ActionHistory != nil {
					lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
					newResponseOrderItem.StepStatus = lastAction.Name
				} else {
					newResponseOrderItem.StepStatus = "none"
					logger.Audit("BuyerFindAllOrders() => Action History is nil, orderId: %d, itemId: %d", order.OrderId, item.ItemId)
				}
				orderItemMap[item.InventoryId] = newResponseOrderItem
			}
		}

		for _, orderItem := range orderItemMap {
			responseOrder.Items = append(responseOrder.Items, orderItem)
		}

		responseOrders = append(responseOrders, responseOrder)
	}

	response.Orders = responseOrders
	return &response, nil
}

func (server Server) convertNewOrderRequestToMessage(req *pb.RequestNewOrder) *pb.MessageRequest {

	serializedOrder, err := proto.Marshal(req)
	if err != nil {
		logger.Err("could not serialize timestamp")
	}

	request := pb.MessageRequest{
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

// TODO Add checking acl and authenticate
func (server Server) BackOfficeOrdersListView(ctx context.Context, req *pb.RequestBackOfficeOrdersList) (*pb.ResponseBackOfficeOrdersList, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BackOfficeOrdersListView() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrdersListView() => token userId %d not authorized", userAcl.User().UserID)
		return nil, status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.BackOfficeOrdersListView(ctx, req)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return futureData.Data.(*pb.ResponseBackOfficeOrdersList), nil
}

// TODO Add checking acl and authenticate
func (server Server) BackOfficeOrderDetailView(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseOrderDetailView, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BackOfficeOrderDetailView() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrderDetailView() => token userId %d not authorized", userAcl.User().UserID)
		return nil, status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.BackOfficeOrderDetailView(ctx, req)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return futureData.Data.(*pb.ResponseOrderDetailView), nil
}

// TODO Add checking acl and authenticate
func (server Server) BackOfficeOrderAction(ctx context.Context, req *pb.RequestBackOfficeOrderAction) (*pb.ResponseBackOfficeOrderAction, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BackOfficeOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrderAction() => token userId %d not authorized", userAcl.User().UserID)
		return nil, status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.OperatorActionPending(ctx, req)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return nil, status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pb.ResponseBackOfficeOrderAction{Result: true}, nil

}

// TODO Add checking acl and authenticate
func (server Server) SellerReportOrders(req *pb.RequestSellerReportOrders, srv pb.OrderService_SellerReportOrdersServer) error {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(srv.Context())
	if err != nil {
		logger.Err("SellerReportOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if userAcl.User().UserID != int64(req.SellerId) {
		logger.Err(" SellerFindAllItems() => token userId %d not authorized for sellerId %d", userAcl.User().UserID, req.SellerId)
		return status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.SellerReportOrders(req, srv)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return nil
}

// TODO Add checking acl and authenticate
func (server Server) BackOfficeReportOrderItems(req *pb.RequestBackOfficeReportOrderItems, srv pb.OrderService_BackOfficeReportOrderItemsServer) error {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(srv.Context())
	if err != nil {
		logger.Err("SellerReportOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrderAction() => token userId %d not authorized", userAcl.User().UserID)
		return status.Error(codes.Code(promise.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.BackOfficeReportOrderItems(req, srv)
	futureData := promiseHandler.Data()
	if futureData == nil {
		return status.Error(codes.Code(promise.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(promise.FutureError)
		return status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return nil
}

func (server Server) Start() {
	//addGrpcStateRule()

	port := strconv.Itoa(int(server.port))
	lis, err := net.Listen("tcp", server.address+":"+port)
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
