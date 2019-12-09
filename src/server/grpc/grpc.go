package grpc_server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/buyer"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
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

type RequestADT string
type RequestType string
type RequestName string
type UserType string
type SortDirection string
type FilterType string
type FilterValue string

//type ActionType string
type Action string

const (
	OrderStateFilter       FilterType = "OrderState"
	OrderReturnStateFilter FilterType = "OrderReturnState"
)

const (
	ApprovalPendingFilter       FilterValue = "ApprovalPending"
	ShipmentPendingFilter       FilterValue = "ShipmentPending"
	DeliveredFilter             FilterValue = "Delivered"
	DeliveryFailedFilter        FilterValue = "DeliveryFailed"
	ReturnRequestPendingFilter  FilterValue = "ReturnRequestPending"
	ReturnShipmentPendingFilter FilterValue = "ReturnShipmentPending"
	ReturnDeliveredFilter       FilterValue = "ReturnDelivered"
	ReturnDeliveryFailedFilter  FilterValue = "ReturnDeliveryFailed"
)

//const (
//	ApprovalPendingActionState       ActionType = "ApprovalPending"
//	ShipmentPendingActionState       ActionType = "ShipmentPending"
//	ShippedActionState               ActionType = "Shipped"
//	DeliveredActionState             ActionType = "Delivered"
//	ReturnRequestPendingActionState  ActionType = "ReturnRequestPending"
//	ReturnShipmentPendingActionState ActionType = "ReturnShipmentPending"
//	ReturnShippedActionState         ActionType = "ReturnShipped"
//	ReturnDeliveredActionState       ActionType = "ReturnDelivered"
//)

const (
	DeliverAction             Action = "Deliver"
	DeliveryFailAction        Action = "DeliveryFail"
	DeliveryDelayAction       Action = "DeliveryDelay"
	DeliveryPendingAction     Action = "DeliveryPending"
	SubmitReturnRequestAction Action = "SubmitReturnRequest"
	EnterShipmentDetailAction Action = "EnterShipmentDetail"
	ApproveAction             Action = "Approve"
	RejectAction              Action = "Reject"
	CancelAction              Action = "Cancel"
	AcceptAction              Action = "Accept"
	CloseAction               Action = "Close"
)

const (
	DataReqType   RequestType = "Data"
	ActionReqType RequestType = "Action"
)

const (
	ListType   RequestADT = "List"
	SingleType RequestADT = "Single"
)

const (
	OperatorUser  UserType = "Operator"
	SellerUser    UserType = "Seller"
	BuyerUser     UserType = "Buyer"
	SchedulerUser UserType = "Scheduler"
)

const (
	SellerOrderList         RequestName = "SellerOrderList"
	SellerOrderDetail       RequestName = "SellerOrderDetail"
	SellerOrderReturnList   RequestName = "SellerOrderReturnList"
	SellerOrderReturnDetail RequestName = "SellerOrderReturnDetail"

	BuyerOrderList          RequestName = "BuyerOrderList"
	BuyerOrderDetail        RequestName = "BuyerOrderDetail"
	BuyerReturnOrderReports RequestName = "BuyerReturnOrderReports"
	BuyerReturnOrderList    RequestName = "BuyerReturnOrderList"
	BuyerReturnOrderDetail  RequestName = "BuyerReturnOrderDetail"
)

const (
	ASC  SortDirection = "ASC"
	DESC SortDirection = "DESC"
)

const (
	// ISO8601 standard time format
	ISO8601 = "2006-01-02T15:04:05-0700"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	pg.UnimplementedBankResultHookServer
	flowManager  domain.IFlowManager
	address      string
	port         uint16
	filterStates map[FilterValue][]states.IEnumState
	actionStates map[UserType][]actions.IAction
}

func NewServer(address string, port uint16, flowManager domain.IFlowManager) Server {
	filterStatesMap := make(map[FilterValue][]states.IEnumState, 8)
	filterStatesMap[ApprovalPendingFilter] = []states.IEnumState{states.ApprovalPending}
	filterStatesMap[ShipmentPendingFilter] = []states.IEnumState{states.ShipmentPending}
	filterStatesMap[DeliveredFilter] = []states.IEnumState{states.Shipped, states.DeliveryPending,
		states.DeliveryDelayed, states.Delivered}
	filterStatesMap[DeliveryFailedFilter] = []states.IEnumState{states.DeliveryFailed}
	filterStatesMap[ReturnRequestPendingFilter] = []states.IEnumState{states.ReturnRequestPending}
	filterStatesMap[ReturnShipmentPendingFilter] = []states.IEnumState{states.ReturnShipmentPending}
	filterStatesMap[ReturnDeliveredFilter] = []states.IEnumState{states.ReturnShipped, states.ReturnDeliveryPending,
		states.ReturnDeliveryDelayed, states.ReturnDelivered}
	filterStatesMap[ReturnDeliveryFailedFilter] = []states.IEnumState{states.ReturnDeliveryFailed}

	actionStateMap := make(map[UserType][]actions.IAction, 8)
	actionStateMap[SellerUser] = []actions.IAction{
		seller_action.New(seller_action.Approve),
		seller_action.New(seller_action.Reject),
		seller_action.New(seller_action.Cancel),
		seller_action.New(seller_action.Accept),
		seller_action.New(seller_action.Deliver),
		seller_action.New(seller_action.DeliveryFail),
		seller_action.New(seller_action.EnterShipmentDetail),
	}
	actionStateMap[BuyerUser] = []actions.IAction{
		buyer_action.New(buyer_action.DeliveryDelay),
		buyer_action.New(buyer_action.Cancel),
		buyer_action.New(buyer_action.SubmitReturnRequest),
		buyer_action.New(buyer_action.EnterShipmentDetails),
	}
	actionStateMap[OperatorUser] = []actions.IAction{
		operator_action.New(operator_action.DeliveryDelay),
		operator_action.New(operator_action.Deliver),
		operator_action.New(operator_action.DeliveryFail),
		operator_action.New(operator_action.Accept),
		operator_action.New(operator_action.Reject),
		operator_action.New(operator_action.Deliver),
	}
	actionStateMap[SchedulerUser] = []actions.IAction{
		scheduler_action.New(scheduler_action.Cancel),
		scheduler_action.New(scheduler_action.Close),
		scheduler_action.New(scheduler_action.DeliveryDelay),
		scheduler_action.New(scheduler_action.Deliver),
		scheduler_action.New(scheduler_action.DeliveryPending),
		scheduler_action.New(scheduler_action.Reject),
		scheduler_action.New(scheduler_action.Accept),
		scheduler_action.New(scheduler_action.Notification),
	}

	return Server{flowManager: flowManager, address: address, port: port, filterStates: filterStatesMap, actionStates: actionStateMap}
}

func (server *Server) RequestHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("RequestHandler() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	}

	// TODO check acl
	if uint64(userAcl.User().UserID) != req.Meta.UId {
		logger.Err("RequestHandler() => UserId mismatch with token userId, error: %s ", err)
		return nil, status.Error(codes.Code(future.InternalError), "User Not Authorized")
	}

	reqType := RequestType(req.Type)
	if reqType == DataReqType {
		return server.requestDataHandler(ctx, req)
	} else {
		return server.requestActionHandler(ctx, req)
	}
}

func (server *Server) requestDataHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	reqName := RequestName(req.Name)
	userType := UserType(req.Meta.UserType)
	reqADT := RequestADT(req.ADT)

	var filterType FilterType
	var filterValue FilterValue
	var sortName string
	var sortDirection SortDirection
	if req.Meta.Filters != nil {
		filterType = FilterType(req.Meta.Filters[0].Type)
		filterValue = FilterValue(req.Meta.Filters[0].Value)
	}

	if req.Meta.Sorts != nil {
		sortName = req.Meta.Sorts[0].Name
		sortDirection = SortDirection(req.Meta.Sorts[0].Direction)
	}

	if reqName == SellerOrderList && filterType != OrderStateFilter {
		logger.Err("requestDataHandler() => request name %s mismatch with %s filter, request: %v", reqName, filterType, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with filter")
	}

	if (reqName == SellerOrderReturnList || reqName == BuyerReturnOrderList) && filterType != OrderReturnStateFilter {
		logger.Err("requestDataHandler() => request name %s mismatch with %s filterType, request: %v", reqName, filterType, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with filterType")
	}

	if userType == SellerUser && (reqName != SellerOrderList || reqName != SellerOrderDetail ||
		reqName != SellerOrderReturnList || reqName != SellerOrderReturnDetail) {
		logger.Err("requestDataHandler() => userType %s mismatch with %s requestName, request: %v", userType, reqName, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestName")
	}

	if userType == BuyerUser && (reqName != BuyerOrderList || reqName != BuyerOrderDetail ||
		reqName != BuyerReturnOrderReports || reqName != BuyerReturnOrderList || reqName != BuyerReturnOrderDetail) {
		logger.Err("requestDataHandler() => userType %s mismatch with %s requestName, request: %v", userType, reqName, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestName")
	}

	if req.Meta.OId > 0 && reqADT != SingleType {
		logger.Err("requestDataHandler() => %s orderId mismatch with %s requestADT, request: %v", userType, reqADT, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestADT")
	}

	switch reqName {
	case SellerOrderList:
		return server.sellerOrderListHandler(ctx, req.Meta.PId, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case SellerOrderDetail:
		return server.sellerOrderDetailHandler(ctx, req.Meta.PId, req.Meta.OId, req.Meta.SId)
	case SellerOrderReturnList:
		return server.sellerOrderReturnListHandler(ctx, req.Meta.PId, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case SellerOrderReturnDetail:
		return server.sellerOrderReturnDetailHandler(ctx, req.Meta.PId, req.Meta.OId, req.Meta.SId)
	case BuyerOrderList:
		return server.buyerOrderListHandler(ctx, req.Meta.PId, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case BuyerOrderDetail:
		return server.buyerOrderDetailHandler(ctx, req.Meta.PId, req.Meta.OId, req.Meta.SId)
	case BuyerReturnOrderReports:
		return server.buyerReturnOrderReportsHandler(ctx, req.Meta.PId, req.Meta.OId, req.Meta.SId)
	case BuyerReturnOrderList:
		return server.buyerReturnOrderListHandler(ctx, req.Meta.PId, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case BuyerReturnOrderDetail:
		return server.buyerReturnOrderDetailHandler(ctx, req.Meta.PId, req.Meta.OId, req.Meta.SId)
	}

	return nil, status.Error(codes.Code(future.BadRequest), "Invalid Request")
}

func (server *Server) requestActionHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	userType := UserType(req.Meta.UserType)
	var userAction actions.IAction

	userActions, ok := server.actionStates[userType]
	if !ok {
		logger.Err("requestActionHandler() => action %s user not supported, request: %v", userType, req)
		return nil, status.Error(codes.Code(future.BadRequest), "User Action Invalid")
	}

	for _, action := range userActions {
		if action.ActionEnum().ActionName() == req.Meta.Action.ActionState {
			userAction = action
			break
		}
	}

	if userAction == nil {
		logger.Err("requestActionHandler() => %s action invalid, request: %v", req.Meta.Action.ActionState, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Action Invalid")
	}

	var reqActionData pb.ActionData
	if err := ptypes.UnmarshalAny(req.Data, &reqActionData); err != nil {
		logger.Err("Could not unmarshal reqActionData from request anything field, request: %v, error %s", req, err)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	subpackages := make([]events.ActionSubpackage, 0, len(reqActionData.Subpackages))
	for _, reqSubpackage := range reqActionData.Subpackages {
		subpackage := events.ActionSubpackage{
			SId: reqSubpackage.SId,
		}
		subpackage.Items = make([]events.ActionItem, 0, len(reqSubpackage.Items))
		for _, item := range reqSubpackage.Items {
			actionItem := events.ActionItem{
				InventoryId: item.InventoryId,
				Quantity:    item.Quantity,
			}
			if item.Reasons != nil {
				actionItem.Reasons = make([]string, 0, len(item.Reasons))
				for _, reason := range item.Reasons {
					actionItem.Reasons = append(actionItem.Reasons, reason)
				}
			}
			subpackage.Items = append(subpackage.Items, actionItem)
		}
		subpackages = append(subpackages, subpackage)
	}

	actionData := events.ActionData{
		SubPackages:    subpackages,
		Carrier:        reqActionData.Carrier,
		TrackingNumber: reqActionData.TrackingNumber,
	}

	event := events.New(events.Action, req.Meta.OId, req.Meta.PId, req.Meta.UId,
		req.Meta.Action.StateIndex, userAction,
		time.Unix(req.Time.GetSeconds(), int64(req.Time.GetNanos())), actionData)

	iFuture := future.Factory().SetCapacity(1).Build()
	iFrame := frame.Factory().SetFuture(iFuture).SetEvent(event).Build()
	server.flowManager.MessageHandler(ctx, iFrame)
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		return nil, status.Error(codes.Code(futureData.Error().Code()), futureData.Error().Message())
	}

	eventResponse := futureData.Data().(events.ActionResponse)

	actionResponse := &pb.ActionResponse{
		OId:  eventResponse.OrderId,
		SIds: eventResponse.SIds,
	}

	serializedResponse, err := proto.Marshal(actionResponse)
	if err != nil {
		logger.Err("could not serialize timestamp")
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "ActionResponse",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionResponse),
			Value:   serializedResponse,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderListHandler(ctx context.Context, pid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page < 0 || perPage <= 0 {
		logger.Err("sellerOrderListHandler() => page or perPage invalid, sellerId: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	countFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil, "packages.subpackages.status": filter}},
			{"$project": bson.M{"subSize": 1}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	var totalCount, err = global.Singletons.PkgItemRepository.CountWithFilter(ctx, countFilter)
	if err != nil {
		logger.Err("sellerOrderListHandler() => CountWithFilter failed,  sellerId: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if totalCount == 0 {
		logger.Err("sellerOrderListHandler() => total count is zero,  sellerId: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%int64(perPage) != 0 {
		availablePages = (totalCount / int64(perPage)) + 1
	} else {
		availablePages = totalCount / int64(perPage)
	}

	if totalCount < int64(perPage) {
		availablePages = 1
	}

	if availablePages < int64(page) {
		logger.Err("sellerOrderListHandler() => availablePages less than page, availablePages: %d, sellerId: %d, filterValue: %s, page: %d, perPage: %d", availablePages, pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		logger.Err("sellerOrderListHandler() => offset invalid, offset: %d, sellerId: %d, filterValue: %s, page: %d, perPage: %d", offset, pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	pkgFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil, "packages.subpackages.status": filter}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
			{"$project": bson.M{"_id": 0, "packages": 1}},
			{"$sort": bson.M{"packages.subpackages." + sortName: direction}},
			{"$skip": offset},
			{"$limit": perPage},
			{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		}
	}

	pkgList, err := global.Singletons.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		logger.Err("sellerOrderListHandler() => FindByFilter failed, sellerId: %d, filterValue: %s, page: %d, perPage: %d, error: %s", offset, pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	itemList := make([]*pb.SellerOrderList_ItemList, 0, perPage)

	for _, pkgItem := range pkgList {
		sellerOrderItem := &pb.SellerOrderList_ItemList{
			OId:                   pkgItem.OrderId,
			OrderRequestTimestamp: pkgItem.CreatedAt.Format(ISO8601),
			Amount:                pkgItem.Invoice.Subtotal,
		}
		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		Items: itemList,
	}

	serializedData, err := proto.Marshal(sellerOrderList)
	if err != nil {
		logger.Err("could not serialize timestamp")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(availablePages),
		Page:    page,
		PerPage: perPage,
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderList",
		Meta:   meta,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderDetailHandler(ctx context.Context, userId, orderId, sid uint64) (*pb.MessageResponse, error) {

}

func (server *Server) sellerOrderReturnListHandler(ctx context.Context, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

}

func (server *Server) sellerOrderReturnDetailHandler(ctx context.Context, userId, orderId, sid uint64) (*pb.MessageResponse, error) {

}

func (server *Server) buyerOrderListHandler(ctx context.Context, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

}

func (server *Server) buyerOrderDetailHandler(ctx context.Context, userId, orderId, sid uint64) (*pb.MessageResponse, error) {

}

func (server *Server) buyerReturnOrderReportsHandler(ctx context.Context, userId, orderId, sid uint64) (*pb.MessageResponse, error) {

}

func (server *Server) buyerReturnOrderListHandler(ctx context.Context, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

}

func (server *Server) buyerReturnOrderDetailHandler(ctx context.Context, userId, orderId, sid uint64) (*pb.MessageResponse, error) {

}

func (server *Server) PaymentGatewayHook(ctx context.Context, req *pg.PaygateHookRequest) (*pg.PaygateHookResponse, error) {
	promiseHandler := server.flowManager.PaymentGatewayResult(ctx, req)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pg.PaygateHookResponse{Ok: true}, nil
}

func (server Server) NewOrder(ctx context.Context, req *pb.RequestNewOrder) (*pb.ResponseNewOrder, error) {

	//ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)

	iFuture := future.Factory().SetCapacity(1).Build()
	iFrame := frame.Factory().SetDefaultHeader(frame.HeaderNewOrder, req).SetFuture(iFuture).Build()
	server.flowManager.MessageHandler(ctx, iFrame)
	futureData := iFuture.Get()
	if futureData == nil {
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Error() != nil {
		futureErr := futureData.Error()
		return nil, status.Error(codes.Code(futureErr.Code()), futureErr.Message())
	}

	callbackUrl, ok := futureData.Data().(string)
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
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if strconv.Itoa(int(userAcl.User().UserID)) != req.Id {
		logger.Err(" SellerFindAllItems() => token userId %d not authorized for sellerId %s", userAcl.User().UserID, req.Id)
		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	sellerId, err := strconv.Atoi(req.Id)
	if err != nil {
		logger.Err(" SellerFindAllItems() => sellerId invalid: %s", req.Id)
		return nil, status.Error(codes.Code(future.BadRequest), "PId Invalid")
	}

	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
		return bson.D{{"items.sellerInfo.sellerId", uint64(sellerId)}}
	})

	if err != nil {
		logger.Err("SellerFindAllItems failed, sellerId: %s, error: %s", req.Id, err.Error())
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
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
							Unit:             orderItem.Invoice.Unit,
							Total:            orderItem.Invoice.Original,
							SellerCommission: orderItem.Invoice.SellerCommission,
							Currency:         orderItem.Invoice.Currency,
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
						logger.Audit("SellerFindAllItems() => Actions History is nil, orderId: %d, sid: %d", order.OrderId, orderItem.ItemId)
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
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if userAcl.User().UserID != int64(req.BuyerId) {
		logger.Err(" BuyerOrderAction() => token userId %d not authorized for buyerId %d", userAcl.User().UserID, req.BuyerId)
		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.BuyerApprovalPending(ctx, req)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pb.ResponseBuyerOrderAction{Result: true}, nil
}

// TODO Add checking acl
func (server Server) SellerOrderAction(ctx context.Context, req *pb.RequestSellerOrderAction) (*pb.ResponseSellerOrderAction, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("SellerOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if userAcl.User().UserID != int64(req.SellerId) {
		logger.Err("SellerOrderAction() => token userId %d not authorized for sellerId %d", userAcl.User().UserID, req.SellerId)
		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.SellerApprovalPending(ctx, req)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pb.ResponseSellerOrderAction{Result: true}, nil
}

// TODO Add checking acl
func (server Server) BuyerFindAllOrders(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseBuyerFindAllOrders, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BuyerFindAllOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if strconv.Itoa(int(userAcl.User().UserID)) != req.Id {
		logger.Err(" BuyerFindAllOrders() => token userId %d not authorized of buyerId %s", userAcl.User().UserID, req.Id)
		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	buyerId, err := strconv.Atoi(req.Id)
	if err != nil {
		logger.Err(" SellerFindAllItems() => buyerId invalid: %s", req.Id)
		return nil, status.Error(codes.Code(future.BadRequest), "BuyerId Invalid")
	}

	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
		return bson.D{{"buyerInfo.buyerId", uint64(buyerId)}}
	})

	if err != nil {
		logger.Err("SellerFindAllItems failed, buyerId: %s, error: %s", req.Id, err.Error())
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
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
				Total:         order.Invoice.Total,
				Subtotal:      order.Invoice.Subtotal,
				Discount:      order.Invoice.Discount,
				Currency:      order.Invoice.Currency,
				ShipmentTotal: order.Invoice.ShipmentTotal,
				PaymentMethod: order.Invoice.PaymentMethod,
				PaymentOption: order.Invoice.PaymentGateway,
				Voucher: &pb.Voucher{
					Amount: order.Invoice.Voucher.Amount,
					Code:   order.Invoice.Voucher.Code,
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
						Unit:     item.Invoice.Unit,
						Total:    item.Invoice.Total,
						Original: item.Invoice.Original,
						Special:  item.Invoice.Special,
						Currency: item.Invoice.Currency,
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
					logger.Audit("BuyerFindAllOrders() => Actions History is nil, orderId: %d, sid: %d", order.OrderId, item.ItemId)
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

//func (server Server) convertNewOrderRequestToMessage(req *pb.RequestNewOrder) *pb.MessageRequest {
//
//	serializedOrder, err := proto.Marshal(req)
//	if err != nil {
//		logger.Err("could not serialize timestamp")
//	}
//
//	request := pb.MessageRequest{
//		Name:   "NewOrder",
//		Type:   string(DataReqType),
//		ADT:    "Single",
//		Method: "GRPC",
//		Time: ptypes.TimestampNow(),
//		Meta: nil,
//		Data: &any.Any{
//			TypeUrl: "baman.io/" + proto.MessageName(req),
//			Value:   serializedOrder,
//		},
//	}
//
//	return &request
//}

// TODO Add checking acl and authenticate
func (server Server) BackOfficeOrdersListView(ctx context.Context, req *pb.RequestBackOfficeOrdersList) (*pb.ResponseBackOfficeOrdersList, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BackOfficeOrdersListView() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrdersListView() => token userId %d not authorized", userAcl.User().UserID)
		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.BackOfficeOrdersListView(ctx, req)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return futureData.Data.(*pb.ResponseBackOfficeOrdersList), nil
}

// TODO Add checking acl and authenticate
func (server Server) BackOfficeOrderDetailView(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseOrderDetailView, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BackOfficeOrderDetailView() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrderDetailView() => token userId %d not authorized", userAcl.User().UserID)
		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.BackOfficeOrderDetailView(ctx, req)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return futureData.Data.(*pb.ResponseOrderDetailView), nil
}

// TODO Add checking acl and authenticate
func (server Server) BackOfficeOrderAction(ctx context.Context, req *pb.RequestBackOfficeOrderAction) (*pb.ResponseBackOfficeOrderAction, error) {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("BackOfficeOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrderAction() => token userId %d not authorized", userAcl.User().UserID)
		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.OperatorActionPending(ctx, req)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return &pb.ResponseBackOfficeOrderAction{Result: true}, nil

}

// TODO Add checking acl and authenticate
func (server Server) SellerReportOrders(req *pb.RequestSellerReportOrders, srv pb.OrderService_SellerReportOrdersServer) error {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(srv.Context())
	if err != nil {
		logger.Err("SellerReportOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if userAcl.User().UserID != int64(req.SellerId) {
		logger.Err(" SellerFindAllItems() => token userId %d not authorized for sellerId %d", userAcl.User().UserID, req.SellerId)
		return status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.SellerReportOrders(req, srv)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
		return status.Error(codes.Code(futureErr.Code), futureErr.Reason)
	}

	return nil
}

// TODO Add checking acl and authenticate
func (server Server) BackOfficeReportOrderItems(req *pb.RequestBackOfficeReportOrderItems, srv pb.OrderService_BackOfficeReportOrderItemsServer) error {

	userAcl, err := global.Singletons.UserService.AuthenticateContextToken(srv.Context())
	if err != nil {
		logger.Err("SellerReportOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	// TODO Must Be changed
	if userAcl.User().UserID <= 0 {
		logger.Err("BackOfficeOrderAction() => token userId %d not authorized", userAcl.User().UserID)
		return status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	promiseHandler := server.flowManager.BackOfficeReportOrderItems(req, srv)
	futureData := promiseHandler.Get()
	if futureData == nil {
		return status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if futureData.Ex != nil {
		futureErr := futureData.Ex.(future.FutureError)
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
