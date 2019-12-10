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
	SellerOrderList             RequestName = "SellerOrderList"
	SellerOrderDetail           RequestName = "SellerOrderDetail"
	SellerReturnOrderDetailList RequestName = "SellerReturnOrderDetailList"
	//SellerOrderReturnDetail     RequestName = "SellerOrderReturnDetail"

	//BuyerOrderList             RequestName = "BuyerOrderList"
	BuyerOrderDetailList       RequestName = "BuyerOrderDetailList"
	BuyerReturnOrderReports    RequestName = "BuyerReturnOrderReports"
	BuyerReturnOrderDetailList RequestName = "BuyerReturnOrderDetailList"
	//BuyerReturnOrderDetail  RequestName = "BuyerReturnOrderDetail"
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
	if uint64(userAcl.User().UserID) != req.Meta.UID {
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
	userType := UserType(req.Meta.UTP)
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

	if (reqName == SellerReturnOrderDetailList || reqName == BuyerReturnOrderDetailList) && filterType != OrderReturnStateFilter {
		logger.Err("requestDataHandler() => request name %s mismatch with %s filterType, request: %v", reqName, filterType, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with filterType")
	}

	if userType == SellerUser && (reqName != SellerOrderList || reqName != SellerOrderDetail ||
		reqName != SellerReturnOrderDetailList) {
		logger.Err("requestDataHandler() => userType %s mismatch with %s requestName, request: %v", userType, reqName, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestName")
	}

	if userType == BuyerUser && (reqName != BuyerOrderDetailList ||
		reqName != BuyerReturnOrderReports || reqName != BuyerReturnOrderDetailList) {
		logger.Err("requestDataHandler() => userType %s mismatch with %s requestName, request: %v", userType, reqName, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestName")
	}

	if req.Meta.OID > 0 && reqADT != SingleType {
		logger.Err("requestDataHandler() => %s orderId mismatch with %s requestADT, request: %v", userType, reqADT, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestADT")
	}

	switch reqName {
	case SellerOrderList:
		return server.sellerOrderListHandler(ctx, req.Meta.PID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case SellerOrderDetail:
		return server.sellerOrderDetailHandler(ctx, req.Meta.PID, req.Meta.OID, filterValue)
	case SellerReturnOrderDetailList:
		return server.sellerOrderReturnDetailListHandler(ctx, req.Meta.PID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case BuyerOrderDetailList:
		return server.buyerOrderDetailHandler(ctx, req.Meta.UID, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case BuyerReturnOrderReports:
		return server.buyerReturnOrderReportsHandler(ctx, req.Meta.UID)
	case BuyerReturnOrderDetailList:
		return server.buyerReturnOrderDetailListHandler(ctx, req.Meta.UID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	}

	return nil, status.Error(codes.Code(future.BadRequest), "Invalid Request")
}

func (server *Server) requestActionHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	userType := UserType(req.Meta.UTP)
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
			SId: reqSubpackage.SID,
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

	event := events.New(events.Action, req.Meta.OID, req.Meta.PID, req.Meta.UID,
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
		OID:  eventResponse.OrderId,
		SIDs: eventResponse.SIds,
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

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	pkgFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil, "packages.subpackages.status": filter}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
			{"$project": bson.M{"_id": 0, "packages": 1}},
			{"$sort": bson.M{"packages.subpackages." + sortName: sortDirect}},
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
			OID:       pkgItem.OrderId,
			RequestAt: pkgItem.CreatedAt.Format(ISO8601),
			Amount:    pkgItem.Invoice.Subtotal,
		}
		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		PID:   pid,
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

func (server *Server) sellerOrderDetailHandler(ctx context.Context, pid, orderId uint64, filter FilterValue) (*pb.MessageResponse, error) {
	order, err := global.Singletons.OrderRepository.FindById(ctx, orderId)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => PkgItemRepository.FindById failed, orderId: %d, pid: %d, filter:%d , error: %s", orderId, pid, filter, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	var pkgItem entities.PackageItem
	for i := 0; i < len(order.Packages); i++ {
		if order.Packages[i].PId == pid {
			pkgItem = order.Packages[i]
			break
		}
	}

	sellerOrderDetailItems := make([]*pb.SellerOrderDetail_ItemDetail, 0, 32)
	for i := 0; i < len(pkgItem.Subpackages); i++ {
		if pkgItem.Subpackages[i].Status == string(filter) {
			for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
				itemDetail := &pb.SellerOrderDetail_ItemDetail{
					SID:         pkgItem.Subpackages[i].SId,
					Sku:         pkgItem.Subpackages[i].Items[j].SKU,
					InventoryId: pkgItem.Subpackages[i].Items[j].InventoryId,
					Title:       pkgItem.Subpackages[i].Items[j].Title,
					Brand:       pkgItem.Subpackages[i].Items[j].Brand,
					Category:    pkgItem.Subpackages[i].Items[j].Category,
					Guaranty:    pkgItem.Subpackages[i].Items[j].Guaranty,
					Image:       pkgItem.Subpackages[i].Items[j].Image,
					Returnable:  pkgItem.Subpackages[i].Items[j].Returnable,
					Quantity:    pkgItem.Subpackages[i].Items[j].Quantity,
					Attributes:  pkgItem.Subpackages[i].Items[j].Attributes,
					Invoice: &pb.SellerOrderDetail_ItemDetail_Invoice{
						Unit:             pkgItem.Subpackages[i].Items[j].Invoice.Unit,
						Total:            pkgItem.Subpackages[i].Items[j].Invoice.Total,
						Original:         pkgItem.Subpackages[i].Items[j].Invoice.Original,
						Special:          pkgItem.Subpackages[i].Items[j].Invoice.Special,
						Discount:         pkgItem.Subpackages[i].Items[j].Invoice.Discount,
						SellerCommission: pkgItem.Subpackages[i].Items[j].Invoice.SellerCommission,
					},
				}
				sellerOrderDetailItems = append(sellerOrderDetailItems, itemDetail)
			}
		}
	}

	sellerOrderDetail := &pb.SellerOrderDetail{
		OID:       orderId,
		PID:       pid,
		Amount:    pkgItem.Invoice.Subtotal,
		Status:    string(filter),
		RequestAt: order.CreatedAt.Format(ISO8601),
		Address: &pb.SellerOrderDetail_ShipmentAddress{
			FirstName:     order.BuyerInfo.ShippingAddress.FirstName,
			LastName:      order.BuyerInfo.ShippingAddress.LastName,
			Address:       order.BuyerInfo.ShippingAddress.Address,
			Phone:         order.BuyerInfo.ShippingAddress.Phone,
			Mobile:        order.BuyerInfo.ShippingAddress.Mobile,
			Country:       order.BuyerInfo.ShippingAddress.Country,
			City:          order.BuyerInfo.ShippingAddress.City,
			Province:      order.BuyerInfo.ShippingAddress.Province,
			Neighbourhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
			Lat:           "",
			Long:          "",
			ZipCode:       order.BuyerInfo.ShippingAddress.ZipCode,
		},
		Items: sellerOrderDetailItems,
	}

	if order.BuyerInfo.ShippingAddress.Location != nil {
		sellerOrderDetail.Address.Lat = strconv.Itoa(int(order.BuyerInfo.ShippingAddress.Location.Coordinates[0]))
		sellerOrderDetail.Address.Long = strconv.Itoa(int(order.BuyerInfo.ShippingAddress.Location.Coordinates[1]))
	}

	serializedData, err := proto.Marshal(sellerOrderDetail)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => could not serialize sellerOrderDetail, sellerOrderDetail: %v, error:%s", sellerOrderDetail, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderDetail",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderDetail),
			Value:   serializedData,
		},
	}

	return response, nil
}

// TODO should be refactor and projection pkg
func (server *Server) sellerOrderReturnDetailListHandler(ctx context.Context, pid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page < 0 || perPage <= 0 {
		logger.Err("sellerOrderReturnDetailListHandler() => page or perPage invalid, pid: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}
	//
	//countFilter := func() interface{} {
	//	return []bson.M{
	//		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil, "packages.subpackages.status": filter}},
	//		{"$project": bson.M{"subSize": 1}},
	//		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
	//		{"$project": bson.M{"_id": 0, "count": 1}},
	//	}
	//}
	//
	//var totalCount, err = global.Singletons.PkgItemRepository.CountWithFilter(ctx, countFilter)
	//if err != nil {
	//	logger.Err("sellerOrderReturnDetailListHandler() => CountWithFilter failed,  sellerId: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
	//	return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	//}
	//
	//if totalCount == 0 {
	//	logger.Err("sellerOrderReturnDetailListHandler() => total count is zero,  sellerId: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
	//	return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	//}
	//
	//// total 160 page=6 perPage=30
	//var availablePages int64
	//
	//if totalCount%int64(perPage) != 0 {
	//	availablePages = (totalCount / int64(perPage)) + 1
	//} else {
	//	availablePages = totalCount / int64(perPage)
	//}
	//
	//if totalCount < int64(perPage) {
	//	availablePages = 1
	//}
	//
	//if availablePages < int64(page) {
	//	logger.Err("sellerOrderListHandler() => availablePages less than page, availablePages: %d, sellerId: %d, filterValue: %s, page: %d, perPage: %d", availablePages, pid, filter, page, perPage)
	//	return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	//}
	//
	//var offset = (page - 1) * perPage
	//if int64(offset) >= totalCount {
	//	logger.Err("sellerOrderListHandler() => offset invalid, offset: %d, sellerId: %d, filterValue: %s, page: %d, perPage: %d", offset, pid, filter, page, perPage)
	//	return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	//}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	//pkgFilter := func() interface{} {
	//	return []bson.M{
	//		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil, "packages.subpackages.status": filter}},
	//		{"$unwind": "$packages"},
	//		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
	//		{"$project": bson.M{"_id": 0, "packages": 1}},
	//		{"$sort": bson.M{"packages.subpackages." + sortName: sortDirect}},
	//		{"$skip": offset},
	//		{"$limit": perPage},
	//		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
	//	}
	//}

	pkgFilter := func() (interface{}, string, int) {
		return []bson.M{
				{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil, "packages.subpackages.status": filter}}},
			sortName, sortDirect
	}

	orderList, total, err := global.Singletons.OrderRepository.FindByFilterWithPageAndSort(ctx, pkgFilter, int64(page), int64(perPage))
	if err != nil {
		logger.Err("sellerOrderReturnDetailListHandler() => FindByFilter failed, pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	//pb.SellerReturnOrderDetailList{
	//	Items: nil,
	//}

	sellerReturnOrderList := make([]*pb.SellerReturnOrderDetailList_ReturnOrderDetail, 0, len(orderList))

	for i := 0; i < len(orderList); i++ {
		for j := 0; j < len(orderList[i].Packages); j++ {
			if orderList[i].Packages[j].PId == pid {
				for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
					if orderList[i].Packages[j].Subpackages[z].Status == string(filter) {
						itemDetailList := make([]*pb.SellerReturnOrderDetailList_ReturnOrderDetail_ItemDetail, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
						for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
							itemOrderDetail := &pb.SellerReturnOrderDetailList_ReturnOrderDetail_ItemDetail{
								SID:             orderList[i].Packages[j].Subpackages[z].SId,
								Sku:             orderList[i].Packages[j].Subpackages[z].Items[t].SKU,
								InventoryId:     orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
								Title:           orderList[i].Packages[j].Subpackages[z].Items[t].Title,
								Brand:           orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
								Category:        orderList[i].Packages[j].Subpackages[z].Items[t].Category,
								Guaranty:        orderList[i].Packages[j].Subpackages[z].Items[t].Guaranty,
								Image:           orderList[i].Packages[j].Subpackages[z].Items[t].Image,
								Returnable:      orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
								Quantity:        orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
								Attributes:      orderList[i].Packages[j].Subpackages[z].Items[t].Attributes,
								ReturnRequestAt: "",
								ReturnShippedAt: "",
								Invoice: &pb.SellerReturnOrderDetailList_ReturnOrderDetail_ItemDetail_Invoice{
									Unit:             orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit,
									Total:            orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total,
									Original:         orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original,
									Special:          orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special,
									Discount:         orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount,
									SellerCommission: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.SellerCommission,
									Currency:         orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Currency,
								},
							}

							if orderList[i].Packages[j].Subpackages[z].Shipments != nil &&
								orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail != nil {
								if orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt != nil {
									itemOrderDetail.ReturnRequestAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
								}
								if orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.ShippedAt != nil {
									itemOrderDetail.ReturnShippedAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.ShippedAt.Format(ISO8601)
								}
							}

							itemDetailList = append(itemDetailList, itemOrderDetail)
						}

					}
				}
				returnOrderDetail := &pb.SellerReturnOrderDetailList_ReturnOrderDetail{
					OID:       orderList[i].Packages[j].OrderId,
					PID:       orderList[i].Packages[j].PId,
					Amount:    orderList[i].Packages[j].Invoice.Subtotal,
					Status:    string(filter),
					RequestAt: orderList[i].Packages[j].CreatedAt.Format(ISO8601),
					Address: &pb.SellerReturnOrderDetailList_ReturnOrderDetail_ShipmentAddress{
						FirstName:     orderList[i].BuyerInfo.ShippingAddress.FirstName,
						LastName:      orderList[i].BuyerInfo.ShippingAddress.LastName,
						Address:       orderList[i].BuyerInfo.ShippingAddress.Address,
						Phone:         orderList[i].BuyerInfo.ShippingAddress.Phone,
						Mobile:        orderList[i].BuyerInfo.ShippingAddress.Mobile,
						Country:       orderList[i].BuyerInfo.ShippingAddress.Country,
						City:          orderList[i].BuyerInfo.ShippingAddress.City,
						Province:      orderList[i].BuyerInfo.ShippingAddress.Province,
						Neighbourhood: orderList[i].BuyerInfo.ShippingAddress.Neighbourhood,
						Lat:           "",
						Long:          "",
						ZipCode:       orderList[i].BuyerInfo.ShippingAddress.ZipCode,
					},
				}
				if orderList[i].BuyerInfo.ShippingAddress.Location != nil {
					returnOrderDetail.Address.Lat = strconv.Itoa(int(orderList[i].BuyerInfo.ShippingAddress.Location.Coordinates[0]))
					returnOrderDetail.Address.Long = strconv.Itoa(int(orderList[i].BuyerInfo.ShippingAddress.Location.Coordinates[1]))
				}
				sellerReturnOrderList = append(sellerReturnOrderList, returnOrderDetail)
			}
		}
	}

	sellerReturnOrderDetailList := &pb.SellerReturnOrderDetailList{
		ReturnOrderDetail: sellerReturnOrderList,
	}

	serializedData, err := proto.Marshal(sellerReturnOrderDetailList)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => could not serialize sellerReturnOrderDetailList, sellerReturnOrderDetailList: %v, error:%s", sellerReturnOrderDetailList, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerReturnOrderDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(total),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerReturnOrderDetailList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerOrderDetailHandler(ctx context.Context, userId uint64, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		logger.Err("buyerOrderDetailHandler() => page or perPage invalid, userId: %d, page: %d, perPage: %d", userId, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	orderFilter := func() (interface{}, string, int) {
		return []bson.M{
				{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "order.status": bson.M{"$or": bson.A{states.OrderInProgressStatus, states.NewOrder}}}}},
			sortName, sortDirect
	}

	orderList, total, err := global.Singletons.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
	if err != nil {
		logger.Err("sellerOrderReturnDetailListHandler() => FindByFilter failed, userId: %d, page: %d, perPage: %d, error: %s", userId, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	orderDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		packageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package, 0, len(orderList[i].Packages))
		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				itemPackageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package_Item, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
				for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
					itemPackageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package_Item{
						SID:                orderList[i].Packages[j].Subpackages[z].SId,
						Status:             orderList[i].Packages[j].Subpackages[z].Status,
						IsCancelable:       false,
						IsReturnable:       false,
						IsReturnCancelable: false,
						ProductId:          orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId[0:7],
						InventoryId:        orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
						Title:              orderList[i].Packages[j].Subpackages[z].Items[t].Title,
						Brand:              orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
						Image:              orderList[i].Packages[j].Subpackages[z].Items[t].Image,
						Returnable:         orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
						Quantity:           orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
						Attributes:         orderList[i].Packages[j].Subpackages[z].Items[t].Attributes,
						Invoice: &pb.BuyerOrderDetailList_OrderDetail_Package_Item_Invoice{
							Unit:     orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit,
							Total:    orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total,
							Original: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original,
							Special:  orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special,
							Discount: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount,
							Currency: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Currency,
						},
					}
					if itemPackageDetail.Status == states.ApprovalPending.StateName() ||
						itemPackageDetail.Status == states.ShipmentPending.StateName() ||
						itemPackageDetail.Status == states.ShipmentDelayed.StateName() {
						itemPackageDetail.IsCancelable = true

					} else if itemPackageDetail.Status == states.Delivered.StateName() {
						itemPackageDetail.IsReturnable = true

					} else if itemPackageDetail.Status == states.ReturnRequestPending.StateName() {
						itemPackageDetail.IsReturnCancelable = true
					}

					itemPackageDetailList = append(itemPackageDetailList, itemPackageDetail)
				}

				packageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package{
					PID:          orderList[i].Packages[j].PId,
					ShopName:     orderList[i].Packages[j].ShopName,
					Items:        itemPackageDetailList,
					ShipmentInfo: nil,
				}

				if orderList[i].Packages[j].Subpackages[z].Shipments != nil &&
					orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail != nil {
					packageDetail.ShipmentInfo = &pb.BuyerOrderDetailList_OrderDetail_Package_Shipment{
						DeliveryAt:     nil,
						ShippedAt:      orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601),
						ShipmentAmount: orderList[i].Packages[j].ShipmentSpec.ShippingCost,
						CarrierName:    orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.CarrierName,
						TrackingNumber: orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.TrackingNumber,
					}

					packageDetail.ShipmentInfo.DeliveryAt = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.
						Add(time.Duration(orderList[i].Packages[j].ShipmentSpec.ShippingTime) * time.Hour).Format(ISO8601)
				}

				packageDetailList = append(packageDetailList, packageDetail)
			}
		}

		orderDetail := &pb.BuyerOrderDetailList_OrderDetail{
			Address: &pb.BuyerOrderDetailList_OrderDetail_BuyerAddress{
				FirstName:     orderList[i].BuyerInfo.ShippingAddress.FirstName,
				LastName:      orderList[i].BuyerInfo.ShippingAddress.LastName,
				Address:       orderList[i].BuyerInfo.ShippingAddress.Address,
				Phone:         orderList[i].BuyerInfo.ShippingAddress.Phone,
				Mobile:        orderList[i].BuyerInfo.ShippingAddress.Mobile,
				Country:       orderList[i].BuyerInfo.ShippingAddress.Country,
				City:          orderList[i].BuyerInfo.ShippingAddress.City,
				Province:      orderList[i].BuyerInfo.ShippingAddress.Province,
				Neighbourhood: orderList[i].BuyerInfo.ShippingAddress.Neighbourhood,
				Lat:           "",
				Long:          "",
				ZipCode:       orderList[i].BuyerInfo.ShippingAddress.ZipCode,
			},
			PackageCount:     int32(len(orderList[i].Packages)),
			TotalAmount:      orderList[i].Invoice.Subtotal,
			PayableAmount:    orderList[i].Invoice.GrandTotal,
			Discounts:        orderList[i].Invoice.Discount,
			ShipmentAmount:   orderList[i].Invoice.ShipmentTotal,
			IsPaymentSuccess: false,
			RequestAt:        orderList[i].CreatedAt.Format(ISO8601),
			Packages:         packageDetailList,
		}
		if orderList[i].BuyerInfo.ShippingAddress.Location != nil {
			orderDetail.Address.Lat = strconv.Itoa(int(orderList[i].BuyerInfo.ShippingAddress.Location.Coordinates[0]))
			orderDetail.Address.Long = strconv.Itoa(int(orderList[i].BuyerInfo.ShippingAddress.Location.Coordinates[1]))
		}

		if orderList[i].PaymentService != nil && orderList[i].PaymentService[0].PaymentResult != nil {
			orderDetail.IsPaymentSuccess = orderList[i].PaymentService[0].PaymentResult.Result
		}

		orderDetailList = append(orderDetailList, orderDetail)
	}

	buyerOrderDetailList := &pb.BuyerOrderDetailList{
		OrderDetails: orderDetailList,
	}

	serializedData, err := proto.Marshal(buyerOrderDetailList)
	if err != nil {
		logger.Err("buyerOrderDetailHandler() => could not serialize buyerOrderDetailList, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "buyerOrderDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(total),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerOrderDetailList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerReturnOrderReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	returnRequestPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnRequestPending.StateName()}},
			{"$project": bson.M{"subSize": 1}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnShipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipmentPending.StateName()}},
			{"$project": bson.M{"subSize": 1}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnShippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipped.StateName()}},
			{"$project": bson.M{"subSize": 1}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	// TODO check correct result
	returnDeliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnDelivered.StateName()}},
			{"$project": bson.M{"subSize": 1}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	// TODO check correct result
	returnDeliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnDeliveryFailed.StateName()}},
			{"$project": bson.M{"subSize": 1}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnRequestPendingCount, err := global.Singletons.SubPkgRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnRequestPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShipmentPendingCount, err := global.Singletons.SubPkgRepository.CountWithFilter(ctx, returnShipmentPendingFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnShipmentPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShippedCount, err := global.Singletons.SubPkgRepository.CountWithFilter(ctx, returnShippedFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnShippedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveredCount, err := global.Singletons.SubPkgRepository.CountWithFilter(ctx, returnDeliveredFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnDeliveredFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryFailedCount, err := global.Singletons.SubPkgRepository.CountWithFilter(ctx, returnDeliveryFailedFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnDeliveryFailedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	buyerReturnOrderReports := &pb.BuyerReturnOrderReports{
		ReturnRequestPending:  int32(returnRequestPendingCount),
		ReturnShipmentPending: int32(returnShipmentPendingCount),
		ReturnShipped:         int32(returnShippedCount),
		ReturnDelivered:       int32(returnDeliveredCount),
		ReturnDeliveryFailed:  int32(returnDeliveryFailedCount),
	}

	serializedData, err := proto.Marshal(buyerReturnOrderReports)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => could not serialize buyerReturnOrderReports, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}

	response := &pb.MessageResponse{
		Entity: "BuyerReturnOrderReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerReturnOrderReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerReturnOrderDetailListHandler(ctx context.Context, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		logger.Err("buyerReturnOrderDetailListHandler() => page or perPage invalid, userId: %d, filter: %s, page: %d, perPage: %d", userId, filter, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	orderFilter := func() (interface{}, string, int) {
		return []bson.M{
				{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipped.StateName()}}},
			sortName, sortDirect
	}

	orderList, total, err := global.Singletons.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
	if err != nil {
		logger.Err("buyerReturnOrderDetailListHandler() => FindByFilter failed, userId: %d, filter: %s, page: %d, perPage: %d, error: %s", userId, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnOrderDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		returnPackageDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail, 0, len(orderList[i].Packages))
		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				returnItemPackageDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
				for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
					returnItemPackageDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item{
						SID:          orderList[i].Packages[j].Subpackages[z].SId,
						Status:       orderList[i].Packages[j].Subpackages[z].Status,
						IsCancelable: false,
						IsAccepted:   false,
						//ShopName:           orderList[i].Packages[j].Subpackages[z].Items[t].ShopName,
						ProductId:       orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId[0:7],
						InventoryId:     orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
						Title:           orderList[i].Packages[j].Subpackages[z].Items[t].Title,
						Brand:           orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
						Image:           orderList[i].Packages[j].Subpackages[z].Items[t].Image,
						Returnable:      orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
						Quantity:        orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
						Attributes:      orderList[i].Packages[j].Subpackages[z].Items[t].Attributes,
						Reason:          "",
						ReturnRequestAt: orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601),
						Invoice: &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item_Invoice{
							Unit:     orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit,
							Total:    orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total,
							Original: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original,
							Special:  orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special,
							Discount: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount,
							Currency: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Currency,
						},
					}

					if orderList[i].Packages[j].Subpackages[z].Items[t].Reasons != nil {
						returnItemPackageDetail.Reason = orderList[i].Packages[j].Subpackages[z].Items[t].Reasons[0]
					}

					if returnItemPackageDetail.Status == states.ReturnRequestPending.StateName() {
						returnItemPackageDetail.IsCancelable = true

					} else if returnItemPackageDetail.Status == states.ReturnShipmentPending.StateName() {
						returnItemPackageDetail.IsAccepted = true

					}

					returnItemPackageDetailList = append(returnItemPackageDetailList, returnItemPackageDetail)
				}

				returnPackageDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail{
					PID:      orderList[i].Packages[j].PId,
					ShopName: orderList[i].Packages[j].ShopName,
					Mobile:   "",
					Phone:    "",
					Shipment: nil,
					Items:    returnItemPackageDetailList,
				}

				if orderList[i].Packages[j].SellerInfo != nil {
					if orderList[i].Packages[j].SellerInfo.ReturnInfo != nil {
						returnPackageDetail.Shipment = &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_SellerReturnShipment{
							Country:      orderList[i].Packages[j].SellerInfo.ReturnInfo.Country,
							Province:     orderList[i].Packages[j].SellerInfo.ReturnInfo.Province,
							City:         orderList[i].Packages[j].SellerInfo.ReturnInfo.City,
							Neighborhood: orderList[i].Packages[j].SellerInfo.ReturnInfo.Neighborhood,
							Address:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalAddress,
							ZipCode:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalCode}
					}
					if orderList[i].Packages[j].SellerInfo.GeneralInfo != nil {
						returnPackageDetail.Mobile = orderList[i].Packages[j].SellerInfo.GeneralInfo.MobilePhone
						returnPackageDetail.Phone = orderList[i].Packages[j].SellerInfo.GeneralInfo.LandPhone
					}
				}

				returnPackageDetailList = append(returnPackageDetailList, returnPackageDetail)
			}
		}

		returnOrderDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail{
			OID:                 orderList[i].OrderId,
			CreatedAt:           orderList[i].CreatedAt.Format(ISO8601),
			TotalAmount:         orderList[i].Invoice.Subtotal,
			ReturnPackageDetail: returnPackageDetailList,
		}

		returnOrderDetailList = append(returnOrderDetailList, returnOrderDetail)
	}

	buyerReturnOrderDetailList := &pb.BuyerReturnOrderDetailList{
		ReturnOrderDetail: returnOrderDetailList,
	}

	serializedData, err := proto.Marshal(buyerReturnOrderDetailList)
	if err != nil {
		logger.Err("buyerReturnOrderDetailListHandler() => could not serialize buyerReturnOrderDetailList, userId: %d, filter: %s, error:%s", userId, filter, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "buyerReturnOrderDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(total),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerReturnOrderDetailList),
			Value:   serializedData,
		},
	}

	return response, nil

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
		return nil, status.Error(codes.Code(future.BadRequest), "PID Invalid")
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
