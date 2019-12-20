package grpc_server

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/buyer"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
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
type RequestMethod string
type UserType string
type SortDirection string
type FilterType string
type FilterValue string
type ActionType string
type Action string

const (
	PostMethod RequestMethod = "POST"
	GetMethod  RequestMethod = "GET"
)

const (
	OrderStateFilterType FilterType = "OrderState"
)

//const (
//	ApprovalPendingFilter       FilterValue = "ApprovalPending"
//	ShipmentPendingFilter       FilterValue = "ShipmentPending"
//	ShippedFilter               FilterValue = "Shipped"
//	DeliveredFilter             FilterValue = "Delivered"
//	DeliveryFailedFilter        FilterValue = "DeliveryFailed"
//	ReturnRequestPendingFilter  FilterValue = "ReturnRequestPending"
//	ReturnShipmentPendingFilter FilterValue = "ReturnShipmentPending"
//	ReturnShippedFilter         FilterValue = "ReturnShipped"
//	ReturnDeliveredFilter       FilterValue = "ReturnDelivered"
//	ReturnDeliveryFailedFilter  FilterValue = "ReturnDeliveryFailed"
//)

const (
	NewOrderFilter                 FilterValue = "NewOrder"
	PaymentPendingFilter           FilterValue = "PaymentPending"
	PaymentSuccessFilter           FilterValue = "PaymentSuccess"
	PaymentFailedFilter            FilterValue = "PaymentFailed"
	OrderVerificationPendingFilter FilterValue = "OrderVerificationPending"
	OrderVerificationSuccessFilter FilterValue = "OrderVerificationSuccess"
	OrderVerificationFailedFilter  FilterValue = "OrderVerificationFailed"
	ApprovalPendingFilter          FilterValue = "ApprovalPending"
	CanceledBySellerFilter         FilterValue = "CanceledBySeller"
	CanceledByBuyerFilter          FilterValue = "CanceledByBuyer"
	ShipmentPendingFilter          FilterValue = "ShipmentPending"
	ShipmentDelayedFilter          FilterValue = "ShipmentDelayedFilter"
	ShippedFilter                  FilterValue = "Shipped"
	DeliveryPendingFilter          FilterValue = "DeliveryPending"
	DeliveryDelayedFilter          FilterValue = "DeliveryDelayed"
	DeliveredFilter                FilterValue = "Delivered"
	DeliveryFailedFilter           FilterValue = "DeliveryFailed"
	ReturnRequestPendingFilter     FilterValue = "ReturnRequestPending"
	ReturnRequestRejectedFilter    FilterValue = "ReturnRequestRejected"
	ReturnShipmentPendingFilter    FilterValue = "ReturnShipmentPending"
	ReturnShippedFilter            FilterValue = "ReturnShipped"
	ReturnDeliveryPendingFilter    FilterValue = "ReturnDeliveryPending"
	ReturnDeliveryDelayedFilter    FilterValue = "ReturnDeliveryDelayed"
	ReturnDeliveredFilter          FilterValue = "ReturnDelivered"
	ReturnDeliveryFailedFilter     FilterValue = "ReturnDeliveryFailed"
	ReturnRejectedFilter           FilterValue = "ReturnRejected"
	PayToBuyerFilter               FilterValue = "PayToBuyer"
	PayToSellerFilter              FilterValue = "PayToSeller"

	AllOrdersFilter       FilterValue = "AllOrders"
	AllCanceledFilter     FilterValue = "AllCanceled"
	ShipmentReportFilter  FilterValue = "ShipmentReport"
	ReturnReportFilter    FilterValue = "ReturnReport"
	DeliveredReportFilter FilterValue = "DeliveredReport"
	CanceledReportFilter  FilterValue = "CanceledReport"
)

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
	SellerAllOrders             RequestName = "SellerAllOrders"
	SellerOrderList             RequestName = "SellerOrderList"
	SellerOrderDetail           RequestName = "SellerOrderDetail"
	SellerReturnOrderDetailList RequestName = "SellerReturnOrderDetailList"
	SellerOrderShipmentReports  RequestName = "SellerOrderShipmentReports"
	SellerOrderDeliveredReports RequestName = "SellerOrderDeliveredReports"
	SellerOrderReturnReports    RequestName = "SellerOrderReturnReports"
	SellerOrderCancelReports    RequestName = "SellerOrderCancelReports"

	//SellerOrderReturnDetail     RequestName = "SellerOrderReturnDetail"
	//BuyerOrderList             RequestName = "BuyerOrderList"

	BuyerAllReturnOrders       RequestName = "BuyerAllReturnOrders"
	BuyerOrderDetailList       RequestName = "BuyerOrderDetailList"
	BuyerReturnOrderReports    RequestName = "BuyerReturnOrderReports"
	BuyerReturnOrderDetailList RequestName = "BuyerReturnOrderDetailList"
	//BuyerReturnOrderDetail  RequestName = "BuyerReturnOrderDetail"

	OperatorOrderList   RequestName = "OperatorOrderList"
	OperatorOrderDetail RequestName = "OperatorOrderDetail"
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
	flowManager         domain.IFlowManager
	address             string
	port                uint16
	requestFilters      map[RequestName][]FilterValue
	buyerFilterStates   map[FilterValue][]states.IEnumState
	sellerFilterStates  map[FilterValue][]states.IEnumState
	operatorFilterState map[FilterValue]states.IEnumState
	actionStates        map[UserType][]actions.IAction
}

func NewServer(address string, port uint16, flowManager domain.IFlowManager) Server {
	buyerStatesMap := make(map[FilterValue][]states.IEnumState, 8)
	buyerStatesMap[ApprovalPendingFilter] = []states.IEnumState{states.ApprovalPending}
	buyerStatesMap[ShipmentPendingFilter] = []states.IEnumState{states.ShipmentPending, states.ShipmentDelayed}
	buyerStatesMap[ShippedFilter] = []states.IEnumState{states.Shipped}
	buyerStatesMap[DeliveredFilter] = []states.IEnumState{states.DeliveryPending, states.DeliveryDelayed, states.Delivered}
	buyerStatesMap[DeliveryFailedFilter] = []states.IEnumState{states.DeliveryFailed}
	buyerStatesMap[ReturnRequestPendingFilter] = []states.IEnumState{states.ReturnRequestPending, states.ReturnRequestRejected}
	buyerStatesMap[ReturnShipmentPendingFilter] = []states.IEnumState{states.ReturnShipmentPending}
	buyerStatesMap[ReturnShippedFilter] = []states.IEnumState{states.ReturnShipped}
	buyerStatesMap[ReturnDeliveredFilter] = []states.IEnumState{states.ReturnDeliveryPending, states.ReturnDeliveryDelayed, states.ReturnDelivered}
	buyerStatesMap[ReturnDeliveryFailedFilter] = []states.IEnumState{states.ReturnDeliveryFailed}

	operatorFilterStatesMap := make(map[FilterValue]states.IEnumState, 30)
	operatorFilterStatesMap[NewOrderFilter] = states.NewOrder
	operatorFilterStatesMap[PaymentPendingFilter] = states.PaymentPending
	operatorFilterStatesMap[PaymentSuccessFilter] = states.PaymentSuccess
	operatorFilterStatesMap[PaymentFailedFilter] = states.PaymentFailed
	operatorFilterStatesMap[OrderVerificationPendingFilter] = states.OrderVerificationPending
	operatorFilterStatesMap[OrderVerificationSuccessFilter] = states.OrderVerificationSuccess
	operatorFilterStatesMap[OrderVerificationFailedFilter] = states.OrderVerificationFailed
	operatorFilterStatesMap[ApprovalPendingFilter] = states.ApprovalPending
	operatorFilterStatesMap[CanceledBySellerFilter] = states.CanceledBySeller
	operatorFilterStatesMap[CanceledByBuyerFilter] = states.CanceledByBuyer
	operatorFilterStatesMap[ShipmentPendingFilter] = states.ShipmentPending
	operatorFilterStatesMap[ShipmentDelayedFilter] = states.ShipmentDelayed
	operatorFilterStatesMap[ShippedFilter] = states.Shipped
	operatorFilterStatesMap[DeliveryPendingFilter] = states.DeliveryPending
	operatorFilterStatesMap[DeliveryDelayedFilter] = states.DeliveryDelayed
	operatorFilterStatesMap[DeliveredFilter] = states.Delivered
	operatorFilterStatesMap[DeliveryFailedFilter] = states.DeliveryFailed
	operatorFilterStatesMap[ReturnRequestPendingFilter] = states.ReturnRequestPending
	operatorFilterStatesMap[ReturnRequestRejectedFilter] = states.ReturnRequestRejected
	operatorFilterStatesMap[ReturnShipmentPendingFilter] = states.ReturnShipmentPending
	operatorFilterStatesMap[ReturnShippedFilter] = states.ReturnShipped
	operatorFilterStatesMap[ReturnDeliveryPendingFilter] = states.ReturnDeliveryPending
	operatorFilterStatesMap[ReturnDeliveryDelayedFilter] = states.ReturnDeliveryDelayed
	operatorFilterStatesMap[ReturnDeliveredFilter] = states.ReturnDelivered
	operatorFilterStatesMap[ReturnDeliveryFailedFilter] = states.ReturnDeliveryFailed
	operatorFilterStatesMap[ReturnRejectedFilter] = states.ReturnRejected
	operatorFilterStatesMap[PayToBuyerFilter] = states.PayToBuyer
	operatorFilterStatesMap[PayToSellerFilter] = states.PayToSeller

	sellerFilterStatesMap := make(map[FilterValue][]states.IEnumState, 30)
	sellerFilterStatesMap[ApprovalPendingFilter] = []states.IEnumState{states.ApprovalPending}
	sellerFilterStatesMap[CanceledBySellerFilter] = []states.IEnumState{states.CanceledBySeller}
	sellerFilterStatesMap[CanceledByBuyerFilter] = []states.IEnumState{states.CanceledByBuyer}
	sellerFilterStatesMap[ShipmentPendingFilter] = []states.IEnumState{states.ShipmentPending}
	sellerFilterStatesMap[ShipmentDelayedFilter] = []states.IEnumState{states.ShipmentDelayed}
	sellerFilterStatesMap[ShippedFilter] = []states.IEnumState{states.Shipped}
	sellerFilterStatesMap[DeliveryPendingFilter] = []states.IEnumState{states.DeliveryPending, states.DeliveryDelayed}
	sellerFilterStatesMap[DeliveredFilter] = []states.IEnumState{states.Delivered}
	sellerFilterStatesMap[DeliveryFailedFilter] = []states.IEnumState{states.DeliveryFailed}
	sellerFilterStatesMap[ReturnRequestPendingFilter] = []states.IEnumState{states.ReturnRequestPending}
	sellerFilterStatesMap[ReturnRequestRejectedFilter] = []states.IEnumState{states.ReturnRequestRejected}
	sellerFilterStatesMap[ReturnShipmentPendingFilter] = []states.IEnumState{states.ReturnShipmentPending}
	sellerFilterStatesMap[ReturnShippedFilter] = []states.IEnumState{states.ReturnShipped}
	sellerFilterStatesMap[ReturnDeliveryPendingFilter] = []states.IEnumState{states.ReturnDeliveryPending}
	sellerFilterStatesMap[ReturnDeliveryDelayedFilter] = []states.IEnumState{states.ReturnDeliveryDelayed}
	sellerFilterStatesMap[ReturnDeliveredFilter] = []states.IEnumState{states.ReturnDelivered}
	sellerFilterStatesMap[ReturnDeliveryFailedFilter] = []states.IEnumState{states.ReturnDeliveryFailed}
	sellerFilterStatesMap[ReturnRejectedFilter] = []states.IEnumState{states.ReturnRejected}
	sellerFilterStatesMap[PayToSellerFilter] = []states.IEnumState{states.PayToSeller}

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
		buyer_action.New(buyer_action.EnterShipmentDetail),
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

	reqFilters := make(map[RequestName][]FilterValue, 8)
	reqFilters[SellerAllOrders] = []FilterValue{
		ApprovalPendingFilter,
		CanceledBySellerFilter,
		CanceledByBuyerFilter,
		ShipmentPendingFilter,
		ShipmentDelayedFilter,
		ShippedFilter,
		DeliveryPendingFilter,
		DeliveredFilter,
		DeliveryFailedFilter,
		ReturnRequestPendingFilter,
		ReturnRequestRejectedFilter,
		ReturnShipmentPendingFilter,
		ReturnShippedFilter,
		ReturnDeliveryPendingFilter,
		ReturnDeliveryDelayedFilter,
		ReturnDeliveredFilter,
		ReturnDeliveryFailedFilter,
		PayToSellerFilter,
	}

	reqFilters[SellerOrderList] = []FilterValue{
		ApprovalPendingFilter,
		CanceledBySellerFilter,
		CanceledByBuyerFilter,
		ShipmentPendingFilter,
		ShipmentDelayedFilter,
		ShippedFilter,
		DeliveryPendingFilter,
		DeliveredFilter,
		DeliveryFailedFilter,
		AllCanceledFilter,
		AllOrdersFilter,
	}

	reqFilters[SellerOrderDetail] = []FilterValue{
		ApprovalPendingFilter,
		CanceledBySellerFilter,
		CanceledByBuyerFilter,
		ShipmentPendingFilter,
		ShipmentDelayedFilter,
		ShippedFilter,
		DeliveryPendingFilter,
		DeliveredFilter,
		DeliveryFailedFilter,
		AllCanceledFilter,
	}

	reqFilters[SellerReturnOrderDetailList] = []FilterValue{
		ReturnRequestPendingFilter,
		ReturnRequestRejectedFilter,
		ReturnShipmentPendingFilter,
		ReturnShippedFilter,
		ReturnDeliveryPendingFilter,
		ReturnDeliveryDelayedFilter,
		ReturnDeliveredFilter,
		ReturnDeliveryFailedFilter,
		ReturnRejectedFilter,
	}

	reqFilters[SellerOrderShipmentReports] = []FilterValue{}
	reqFilters[SellerOrderDeliveredReports] = []FilterValue{}
	reqFilters[SellerOrderReturnReports] = []FilterValue{}
	reqFilters[SellerOrderCancelReports] = []FilterValue{}

	reqFilters[BuyerOrderDetailList] = []FilterValue{
		PaymentSuccessFilter,
		PaymentFailedFilter,
		OrderVerificationPendingFilter,
		OrderVerificationSuccessFilter,
		OrderVerificationFailedFilter,
		ApprovalPendingFilter,
		CanceledBySellerFilter,
		CanceledByBuyerFilter,
		ShipmentPendingFilter,
		ShipmentDelayedFilter,
		ShippedFilter,
		DeliveryPendingFilter,
		DeliveryDelayedFilter,
		DeliveredFilter,
		DeliveryFailedFilter,
		PayToBuyerFilter,
		PayToSellerFilter,
	}

	reqFilters[BuyerReturnOrderReports] = []FilterValue{}

	reqFilters[BuyerReturnOrderDetailList] = []FilterValue{
		ReturnRequestPendingFilter,
		ReturnShipmentPendingFilter,
		ReturnShippedFilter,
		ReturnDeliveredFilter,
		ReturnDeliveryFailedFilter,
		AllOrdersFilter,
	}

	reqFilters[BuyerAllReturnOrders] = []FilterValue{
		ReturnRequestPendingFilter,
		ReturnRequestRejectedFilter,
		ReturnShipmentPendingFilter,
		ReturnShippedFilter,
		ReturnDeliveryPendingFilter,
		ReturnDeliveryDelayedFilter,
		ReturnDeliveredFilter,
		ReturnDeliveryFailedFilter,
		ReturnRejectedFilter,
	}

	return Server{flowManager: flowManager, address: address, port: port,
		requestFilters:      reqFilters,
		buyerFilterStates:   buyerStatesMap,
		sellerFilterStates:  sellerFilterStatesMap,
		operatorFilterState: operatorFilterStatesMap,
		actionStates:        actionStateMap}
}

func (server *Server) RequestHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {

	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("RequestHandler() => UserService.AuthenticateContextToken failed, error: %s ", err)
		return nil, status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	}

	// TODO check acl
	if uint64(userAcl.User().UserID) != req.Meta.UID {
		logger.Err("RequestHandler() => request userId %d mismatch with token userId: %d", req.Meta.UID, userAcl.User().UserID)
		return nil, status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	}

	if ctx.Value(string(utils.CtxUserID)) == nil {
		ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(req.Meta.UID))
		ctx = context.WithValue(ctx, string(utils.CtxUserACL), userAcl)
	}

	reqType := RequestType(req.Type)
	if reqType == DataReqType {
		return server.requestDataHandler(ctx, req)
	} else {
		return server.requestActionHandler(ctx, req)
	}
}

func (server *Server) buyerGeneratePipelineFilter(ctx context.Context, filter FilterValue) []interface{} {

	newFilter := make([]interface{}, 2)

	if filter == ApprovalPendingFilter {
		newFilter[0] = "packages.subpackages.status"
		newFilter[1] = server.buyerFilterStates[filter][0].StateName()
	} else if filter == ShipmentPendingFilter {
		newFilter[0] = "$or"
		newFilter[1] = bson.A{
			bson.M{"packages.subpackages.status": states.ShipmentPending.StateName()},
			bson.M{"packages.subpackages.status": states.ShipmentDelayed.StateName()}}
	} else if filter == ShippedFilter {
		newFilter[0] = "packages.subpackages.status"
		newFilter[1] = server.buyerFilterStates[filter][0].StateName()
	} else if filter == DeliveredFilter {
		newFilter[0] = "$or"
		newFilter[1] = bson.A{
			bson.M{"packages.subpackages.status": states.DeliveryPending.StateName()},
			bson.M{"packages.subpackages.status": states.DeliveryDelayed.StateName()},
			bson.M{"packages.subpackages.status": states.Delivered.StateName()}}
	} else if filter == DeliveryFailedFilter {
		newFilter[0] = "packages.subpackages.tracking.history.name"
		newFilter[1] = server.buyerFilterStates[filter][0].StateName()
	} else if filter == ReturnRequestPendingFilter {
		newFilter[0] = "$or"
		newFilter[1] = bson.A{
			bson.M{"packages.subpackages.status": states.ReturnRequestPending.StateName()},
			bson.M{"packages.subpackages.status": states.ReturnRequestRejected.StateName()}}
	} else if filter == ReturnShipmentPendingFilter {
		newFilter[0] = "packages.subpackages.status"
		newFilter[1] = server.buyerFilterStates[filter][0].StateName()
	} else if filter == ReturnShippedFilter {
		newFilter[0] = "packages.subpackages.status"
		newFilter[1] = server.buyerFilterStates[filter][0].StateName()
	} else if filter == ReturnDeliveredFilter {
		newFilter[0] = "$or"
		newFilter[1] = bson.A{
			bson.M{"packages.subpackages.status": states.ReturnDeliveryPending.StateName()},
			bson.M{"packages.subpackages.status": states.ReturnDeliveryDelayed.StateName()},
			bson.M{"packages.subpackages.status": states.ReturnDelivered.StateName()}}
	} else if filter == DeliveryFailedFilter {
		newFilter[0] = "packages.subpackages.tracking.history.name"
		newFilter[1] = server.buyerFilterStates[filter][0].StateName()
	}
	return newFilter
}

func (server *Server) sellerGeneratePipelineFilter(ctx context.Context, filter FilterValue) []interface{} {

	newFilter := make([]interface{}, 2)
	if filter == CanceledBySellerFilter {
		newFilter[0] = "packages.subpackages.tracking.history.name"
		newFilter[1] = server.sellerFilterStates[filter][0].StateName()

	} else if filter == CanceledByBuyerFilter {
		newFilter[0] = "packages.subpackages.tracking.history.name"
		newFilter[1] = server.sellerFilterStates[filter][0].StateName()

	} else if filter == DeliveryFailedFilter {
		newFilter[0] = "packages.subpackages.tracking.history.name"
		newFilter[1] = server.sellerFilterStates[filter][0].StateName()

	} else if filter == ReturnDeliveryFailedFilter {
		newFilter[0] = "packages.subpackages.tracking.history.name"
		newFilter[1] = server.sellerFilterStates[filter][0].StateName()

	} else if filter == DeliveryPendingFilter {
		newFilter[0] = "$or"
		newFilter[1] = bson.A{
			bson.M{"packages.subpackages.status": states.DeliveryPending.StateName()},
			bson.M{"packages.subpackages.status": states.DeliveryDelayed.StateName()}}

	} else if filter == AllCanceledFilter {
		newFilter[0] = "$or"
		newFilter[1] = bson.A{
			bson.M{"packages.subpackages.tracking.history.name": states.CanceledBySeller.StateName()},
			bson.M{"packages.subpackages.tracking.history.name": states.CanceledByBuyer.StateName()}}

	} else {
		newFilter[0] = "packages.subpackages.status"
		newFilter[1] = server.sellerFilterStates[filter][0].StateName()

	}

	return newFilter
}

func (server *Server) requestDataHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	reqName := RequestName(req.Name)
	userType := UserType(req.Meta.UTP)
	//reqADT := RequestADT(req.ADT)

	//var filterType FilterType
	var filterValue FilterValue
	var sortName string
	var sortDirection SortDirection
	if req.Meta.Filters != nil {
		//filterType = FilterType(req.Meta.Filters[0].UTP)
		filterValue = FilterValue(req.Meta.Filters[0].Value)
	}

	if req.Meta.Sorts != nil {
		sortName = req.Meta.Sorts[0].Name
		sortDirection = SortDirection(req.Meta.Sorts[0].Direction)
	}

	//if reqName == SellerOrderList && filterType != OrderStateFilterType {
	//	logger.Err("requestDataHandler() => request name %s mismatch with %s filter, request: %v", reqName, filterType, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with filter")
	//}

	//if (reqName == SellerReturnOrderDetailList || reqName == BuyerReturnOrderDetailList) && filterType != OrderReturnStateFilter {
	//	logger.Err("requestDataHandler() => request name %s mismatch with %s filterType, request: %v", reqName, filterType, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with filterType")
	//}

	if userType == SellerUser &&
		reqName != SellerOrderList &&
		reqName != SellerOrderDetail &&
		reqName != SellerReturnOrderDetailList &&
		reqName != SellerOrderDeliveredReports &&
		reqName != SellerOrderReturnReports &&
		reqName != SellerOrderShipmentReports &&
		reqName != SellerOrderCancelReports {
		logger.Err("requestDataHandler() => userType %s mismatch with %s requestName, request: %v", userType, reqName, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestName")
	}

	if userType == BuyerUser &&
		reqName != BuyerOrderDetailList &&
		reqName != BuyerReturnOrderReports &&
		reqName != BuyerReturnOrderDetailList {
		logger.Err("requestDataHandler() => userType %s mismatch with %s requestName, request: %v", userType, reqName, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestName")
	}

	if userType == OperatorUser &&
		reqName != OperatorOrderList &&
		reqName != OperatorOrderDetail {
		logger.Err("requestDataHandler() => userType %s mismatch with %s requestName, request: %v", userType, reqName, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestName")
	}

	//if req.Meta.OID > 0 && reqADT == ListType {
	//	logger.Err("requestDataHandler() => %s orderId mismatch with %s requestADT, request: %v", userType, reqADT, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestADT")
	//}

	//if req.Meta.OID > 0 && reqName != SellerOrderList && reqName != OperatorOrderList {
	//	logger.Err("requestDataHandler() => %s orderId mismatch with %s requestName, request: %v", userType, reqName, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch OrderId with Request name")
	//}

	if userType == BuyerUser && reqName != BuyerReturnOrderReports {
		if reqName == BuyerOrderDetailList {
			if filterValue != "" {
				var findFlag = false
				for _, filter := range server.requestFilters[reqName] {
					if filter == filterValue {
						findFlag = true
						break
					}
				}

				if !findFlag && req.Meta.OID <= 0 {
					logger.Err("requestDataHandler() => %s requestName mismatch with %s Filter, request: %v", reqName, filterValue, req)
					return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with Filter")
				}
			}
		} else {
			var findFlag = false
			for _, filter := range server.requestFilters[reqName] {
				if filter == filterValue {
					findFlag = true
					break
				}
			}

			if !findFlag {
				logger.Err("requestDataHandler() => %s requestName mismatch with %s Filter, request: %v", reqName, filterValue, req)
				return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with Filter")
			}
		}
	} else if userType == SellerUser &&
		reqName != SellerOrderShipmentReports &&
		reqName != SellerOrderDeliveredReports &&
		reqName != SellerOrderReturnReports &&
		reqName != SellerOrderCancelReports {
		var findFlag = false
		for _, filter := range server.requestFilters[reqName] {
			if filter == filterValue {
				findFlag = true
				break
			}
		}

		if !findFlag {
			logger.Err("requestDataHandler() => %s requestName mismatch with %s Filter, request: %v", reqName, filterValue, req)
			return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with Filter")
		}
	}

	if reqName == OperatorOrderDetail && filterValue != "" {
		logger.Err("requestDataHandler() => %s requestName doesn't need anything filter, %s Filter, request: %v", reqName, filterValue, req)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Name Filter Invalid")
	}

	//if req.Meta.OID > 0 && reqName == SellerOrderList {
	//	return server.sellerGetOrderByIdHandler(ctx, , req.Meta.PID, filterValue)
	//}

	switch reqName {
	case SellerOrderList:
		return server.sellerOrderListHandler(ctx, req.Meta.OID, req.Meta.PID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case SellerOrderDetail:
		return server.sellerOrderDetailHandler(ctx, req.Meta.PID, req.Meta.OID, filterValue)
	case SellerReturnOrderDetailList:
		return server.sellerOrderReturnDetailListHandler(ctx, req.Meta.PID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)

	case SellerOrderShipmentReports:
		return server.sellerOrderShipmentReportsHandler(ctx, req.Meta.UID)
	case SellerOrderReturnReports:
		return server.sellerOrderReturnReportsHandler(ctx, req.Meta.UID)
	case SellerOrderDeliveredReports:
		return server.sellerOrderDeliveredReportsHandler(ctx, req.Meta.UID)
	case SellerOrderCancelReports:
		return server.sellerOrderCancelReportsHandler(ctx, req.Meta.UID)

	case BuyerOrderDetailList:
		return server.buyerOrderDetailListHandler(ctx, req.Meta.OID, req.Meta.UID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case BuyerReturnOrderReports:
		return server.buyerReturnOrderReportsHandler(ctx, req.Meta.UID)
	case BuyerReturnOrderDetailList:
		return server.buyerReturnOrderDetailListHandler(ctx, req.Meta.UID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)

	case OperatorOrderList:
		return server.operatorOrderListHandler(ctx, req.Meta.OID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case OperatorOrderDetail:
		return server.operatorOrderDetailHandler(ctx, req.Meta.OID)
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

// TODO refactor filter for optimization
func (server *Server) operatorOrderListHandler(ctx context.Context, oid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	var orderFilter func() (interface{}, string, int)
	if oid > 0 {
		return server.operatorGetOrderByIdHandler(ctx, oid, filter)
	} else {
		if filter != "" {
			orderFilter = func() (interface{}, string, int) {
				return bson.D{{"deletedAt", nil}, {"packages.subpackages.tracking.history.name", server.operatorFilterState[filter].StateName()}},
					sortName, sortDirect
			}
		} else {
			orderFilter = func() (interface{}, string, int) {
				return bson.D{{"deletedAt", nil}}, sortName, sortDirect
			}
		}
	}

	orderList, totalPage, err := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
	if err != nil {
		logger.Err("operatorOrderListHandler() => CountWithFilter failed,  oid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", oid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	operatorOrders := make([]*pb.OperatorOrderList_Order, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		order := &pb.OperatorOrderList_Order{
			OrderId:     orderList[i].OrderId,
			BuyerId:     orderList[i].BuyerInfo.BuyerId,
			PurchasedOn: orderList[i].CreatedAt.Format(ISO8601),
			BasketSize:  0,
			BillTo:      orderList[i].BuyerInfo.FirstName + " " + orderList[i].BuyerInfo.LastName,
			ShipTo:      orderList[i].BuyerInfo.ShippingAddress.FirstName + " " + orderList[i].BuyerInfo.ShippingAddress.LastName,
			Platform:    orderList[i].Platform,
			IP:          orderList[i].BuyerInfo.IP,
			Status:      orderList[i].Status,
			Invoice: &pb.OperatorOrderList_Order_Invoice{
				GrandTotal:     orderList[i].Invoice.GrandTotal,
				Subtotal:       orderList[i].Invoice.Subtotal,
				PaymentMethod:  orderList[i].Invoice.PaymentMethod,
				PaymentGateway: orderList[i].Invoice.PaymentGateway,
				Shipment:       orderList[i].Invoice.ShipmentTotal,
			},
		}

		if orderList[i].Invoice.Voucher != nil {
			order.Invoice.Voucher = float32(orderList[i].Invoice.Voucher.Amount)
		}

		if orderList[i].PaymentService != nil &&
			len(orderList[i].PaymentService) > 0 &&
			orderList[i].PaymentService[0].PaymentResult != nil {
			if orderList[i].PaymentService[0].PaymentResult.Result {
				order.Invoice.PaymentStatus = "success"
			} else {
				order.Invoice.PaymentStatus = "fail"
			}
		} else {
			order.Invoice.PaymentStatus = "pending"
		}

		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
					order.BasketSize += orderList[i].Packages[j].Subpackages[z].Items[t].Quantity
				}
			}
		}

		operatorOrders = append(operatorOrders, order)
	}

	operatorOrderList := &pb.OperatorOrderList{
		Orders: operatorOrders,
	}

	serializedData, err := proto.Marshal(operatorOrderList)
	if err != nil {
		logger.Err("operatorOrderListHandler() => could not serialize operatorOrderListHandler, operatorOrderList: %v, error:%s", operatorOrderList, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(totalPage),
		Page:    page,
		PerPage: perPage,
	}

	response := &pb.MessageResponse{
		Entity: "OperatorOrderList",
		Meta:   meta,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(operatorOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) operatorOrderDetailHandler(ctx context.Context, oid uint64) (*pb.MessageResponse, error) {

	order, err := app.Globals.OrderRepository.FindById(ctx, oid)
	if err != nil {
		logger.Err("operatorOrderDetailHandler() => FindById failed, oid: %d, error: %s", oid, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	orderDetail := &pb.OperatorOrderDetail{
		OrderId:     order.OrderId,
		PurchasedOn: order.CreatedAt.Format(ISO8601),
		IP:          order.BuyerInfo.IP,
		Invoice: &pb.OperatorOrderDetail_Invoice{
			GrandTotal:     order.Invoice.GrandTotal,
			Subtotal:       order.Invoice.Subtotal,
			PaymentMethod:  order.Invoice.PaymentMethod,
			PaymentGateway: order.Invoice.PaymentGateway,
			ShipmentTotal:  order.Invoice.ShipmentTotal,
		},
		Billing: &pb.OperatorOrderDetail_BillingInfo{
			BuyerId:    order.BuyerInfo.BuyerId,
			FirstName:  order.BuyerInfo.FirstName,
			LastName:   order.BuyerInfo.LastName,
			Phone:      order.BuyerInfo.Phone,
			Mobile:     order.BuyerInfo.Mobile,
			NationalId: order.BuyerInfo.NationalId,
		},
		ShippingInfo: &pb.OperatorOrderDetail_ShippingInfo{
			FirstName:    order.BuyerInfo.ShippingAddress.FirstName,
			LastName:     order.BuyerInfo.ShippingAddress.LastName,
			Country:      order.BuyerInfo.ShippingAddress.Country,
			City:         order.BuyerInfo.ShippingAddress.City,
			Province:     order.BuyerInfo.ShippingAddress.Province,
			Neighborhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
			Address:      order.BuyerInfo.ShippingAddress.Address,
			ZipCode:      order.BuyerInfo.ShippingAddress.ZipCode,
		},
		Subpackages: nil,
	}

	if order.Invoice.Voucher != nil {
		orderDetail.Invoice.VoucherAmount = float32(order.Invoice.Voucher.Amount)
	}

	if order.PaymentService != nil &&
		len(order.PaymentService) > 0 &&
		order.PaymentService[0].PaymentResult != nil {
		if order.PaymentService[0].PaymentResult.Result {
			orderDetail.Invoice.PaymentStatus = "success"
		} else {
			orderDetail.Invoice.PaymentStatus = "fail"
		}
	}

	orderDetail.Subpackages = make([]*pb.OperatorOrderDetail_Subpackage, 0, 32)
	for i := 0; i < len(order.Packages); i++ {
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			subpackage := &pb.OperatorOrderDetail_Subpackage{
				SID:                  order.Packages[i].Subpackages[j].SId,
				PID:                  order.Packages[i].Subpackages[j].PId,
				SellerId:             order.Packages[i].Subpackages[j].PId,
				ShopName:             order.Packages[i].ShopName,
				UpdatedAt:            order.Packages[i].Subpackages[j].UpdatedAt.Format(ISO8601),
				States:               nil,
				ShipmentDetail:       nil,
				ReturnShipmentDetail: nil,
				Items:                nil,
				Actions:              nil,
			}

			subpackage.States = make([]*pb.OperatorOrderDetail_Subpackage_StateHistory, 0, len(order.Packages[i].Subpackages[j].Tracking.History))
			for x := 0; x < len(order.Packages[i].Subpackages[j].Tracking.History); x++ {
				state := &pb.OperatorOrderDetail_Subpackage_StateHistory{
					Name:      order.Packages[i].Subpackages[j].Tracking.History[x].Name,
					Index:     int32(order.Packages[i].Subpackages[j].Tracking.History[x].Index),
					UTP:       "",
					CreatedAt: "",
				}

				if order.Packages[i].Subpackages[j].Tracking.History[x].Actions != nil {
					state.UTP = order.Packages[i].Subpackages[j].Tracking.History[x].Actions[len(order.Packages[i].Subpackages[j].Tracking.History[x].Actions)-1].UTP
					state.CreatedAt = order.Packages[i].Subpackages[j].Tracking.History[x].Actions[len(order.Packages[i].Subpackages[j].Tracking.History[x].Actions)-1].CreatedAt.Format(ISO8601)
				}
				subpackage.States = append(subpackage.States, state)
			}

			if order.Packages[i].Subpackages[j].Shipments != nil && order.Packages[i].Subpackages[j].Shipments.ShipmentDetail != nil {
				subpackage.ShipmentDetail = &pb.OperatorOrderDetail_Subpackage_ShipmentDetail{
					CarrierName:    order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.CarrierName,
					ShippingMethod: order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.ShippingMethod,
					TrackingNumber: order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.TrackingNumber,
					Image:          order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.Image,
					Description:    order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.Description,
					CreatedAt:      order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.CreatedAt.Format(ISO8601),
					ShippedAt:      "",
				}
				if order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.ShippedAt != nil {
					subpackage.ShipmentDetail.ShippedAt = order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601)
				}
			}

			if order.Packages[i].Subpackages[j].Shipments != nil && order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail != nil {
				subpackage.ReturnShipmentDetail = &pb.OperatorOrderDetail_Subpackage_ReturnShipmentDetail{
					CarrierName:    order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.CarrierName,
					ShippingMethod: order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippingMethod,
					TrackingNumber: order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.TrackingNumber,
					Image:          order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.Image,
					Description:    order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.Description,
					RequestedAt:    "",
					ShippedAt:      "",
					CreatedAt:      order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.CreatedAt.Format(ISO8601),
				}

				if order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.RequestedAt != nil {
					subpackage.ReturnShipmentDetail.RequestedAt = order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
				}

				if order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippedAt != nil {
					subpackage.ReturnShipmentDetail.ShippedAt = order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippedAt.Format(ISO8601)
				}
			}

			subpackage.Actions = make([]string, 0, 3)
			for _, action := range server.flowManager.GetState(states.FromString(order.Packages[i].Subpackages[j].Status)).Actions() {
				if action.ActionType() == actions.Operator {
					subpackage.Actions = append(subpackage.Actions, action.ActionEnum().ActionName())
				}
			}

			subpackage.Items = make([]*pb.OperatorOrderDetail_Subpackage_Item, 0, len(order.Packages[i].Subpackages[j].Items))
			for z := 0; z < len(order.Packages[i].Subpackages[j].Items); z++ {
				item := &pb.OperatorOrderDetail_Subpackage_Item{
					InventoryId: order.Packages[i].Subpackages[j].Items[z].InventoryId,
					Brand:       order.Packages[i].Subpackages[j].Items[z].Brand,
					Title:       order.Packages[i].Subpackages[j].Items[z].Title,
					Attributes:  order.Packages[i].Subpackages[j].Items[z].Attributes,
					Quantity:    order.Packages[i].Subpackages[j].Items[z].Quantity,
					Invoice: &pb.OperatorOrderDetail_Subpackage_Item_Invoice{
						Unit:     order.Packages[i].Subpackages[j].Items[z].Invoice.Unit,
						Total:    order.Packages[i].Subpackages[j].Items[z].Invoice.Total,
						Original: order.Packages[i].Subpackages[j].Items[z].Invoice.Original,
						Special:  order.Packages[i].Subpackages[j].Items[z].Invoice.Special,
						Discount: order.Packages[i].Subpackages[j].Items[z].Invoice.Discount,
						Currency: order.Packages[i].Subpackages[j].Items[z].Invoice.Currency,
					},
				}
				subpackage.Items = append(subpackage.Items, item)
			}
			orderDetail.Subpackages = append(orderDetail.Subpackages, subpackage)
		}
	}

	serializedData, err := proto.Marshal(orderDetail)
	if err != nil {
		logger.Err("operatorOrderDetailHandler() => could not serialize operatorOrderDetail, orderId: %d, error:%s", orderDetail.OrderId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "OperatorOrderDetail",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(orderDetail),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) operatorGetOrderByIdHandler(ctx context.Context, oid uint64, filter FilterValue) (*pb.MessageResponse, error) {

	var orderFilter func() interface{}
	if filter != "" {
		orderFilter = func() interface{} {
			return bson.D{{"orderId", oid}, {"deletedAt", nil}, {"packages.subpackages.tracking.history.name", server.operatorFilterState[filter].StateName()}}
		}
	} else {
		orderFilter = func() interface{} {
			return bson.D{{"orderId", oid}, {"deletedAt", nil}}
		}
	}

	orderList, err := app.Globals.OrderRepository.FindByFilter(ctx, orderFilter)
	if err != nil {
		logger.Err("operatorGetOrderByIdHandler() => CountWithFilter failed,  oid: %d, filterValue: %s, error: %s", oid, filter, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	operatorOrders := make([]*pb.OperatorOrderList_Order, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		order := &pb.OperatorOrderList_Order{
			OrderId:     orderList[i].OrderId,
			BuyerId:     orderList[i].BuyerInfo.BuyerId,
			PurchasedOn: orderList[i].CreatedAt.Format(ISO8601),
			BasketSize:  0,
			BillTo:      orderList[i].BuyerInfo.FirstName + " " + orderList[i].BuyerInfo.LastName,
			ShipTo:      orderList[i].BuyerInfo.ShippingAddress.FirstName + " " + orderList[i].BuyerInfo.ShippingAddress.LastName,
			Platform:    orderList[i].Platform,
			IP:          orderList[i].BuyerInfo.IP,
			Status:      orderList[i].Status,
			Invoice: &pb.OperatorOrderList_Order_Invoice{
				GrandTotal:     orderList[i].Invoice.GrandTotal,
				Subtotal:       orderList[i].Invoice.Subtotal,
				PaymentMethod:  orderList[i].Invoice.PaymentMethod,
				PaymentGateway: orderList[i].Invoice.PaymentGateway,
			},
		}

		if orderList[i].Invoice.Voucher != nil {
			order.Invoice.Voucher = float32(orderList[i].Invoice.Voucher.Amount)
		}

		if orderList[i].PaymentService != nil &&
			len(orderList[i].PaymentService) > 0 &&
			orderList[i].PaymentService[0].PaymentResult != nil {
			if orderList[i].PaymentService[0].PaymentResult.Result {
				order.Invoice.PaymentStatus = "success"
			} else {
				order.Invoice.PaymentStatus = "fail"
			}
		} else {
			order.Invoice.PaymentStatus = "pending"
		}

		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
					order.BasketSize += orderList[i].Packages[j].Subpackages[z].Items[t].Quantity
				}
			}
		}

		operatorOrders = append(operatorOrders, order)
	}

	operatorOrderList := &pb.OperatorOrderList{
		Orders: operatorOrders,
	}

	serializedData, err := proto.Marshal(operatorOrderList)
	if err != nil {
		logger.Err("operatorGetOrderByIdHandler() => could not serialize operatorGetOrderByIdHandler, operatorOrderList: %v, error:%s", operatorOrderList, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "OperatorOrderList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(1),
			Page:    1,
			PerPage: 1,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(operatorOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerGetOrderByIdHandler(ctx context.Context, oid uint64, pid uint64, filter FilterValue) (*pb.MessageResponse, error) {
	genFilter := server.sellerGeneratePipelineFilter(ctx, filter)
	filters := make(bson.M, 3)
	filters["packages.orderId"] = oid
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil
	filters[genFilter[0].(string)] = genFilter[1]
	findFilter := func() interface{} {
		return []bson.M{
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$project": bson.M{"_id": 0, "packages": 1}},
			{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		}
	}

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, findFilter)
	if err != nil {
		logger.Err("sellerGetOrderByIdHandler() => FindByFilter failed, pid: %d, filterValue: %s, error: %s", pid, filter, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	itemList := make([]*pb.SellerOrderList_ItemList, 0, 1)

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
		logger.Err("sellerGetOrderByIdHandler() => could not serialize sellerOrderList, sellerOrderList: %v, error:%s", sellerOrderList, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderList",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderListHandler(ctx context.Context, oid, pid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	if oid > 0 {
		return server.sellerGetOrderByIdHandler(ctx, oid, pid, filter)
	}

	if page <= 0 || perPage <= 0 {
		logger.Err("sellerOrderListHandler() => page or perPage invalid, pid: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	if filter == AllOrdersFilter {
		return server.sellerAllOrdersHandler(ctx, pid, page, perPage, sortName, direction)
	}

	genFilter := server.sellerGeneratePipelineFilter(ctx, filter)
	filters := make(bson.M, 3)
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil
	filters[genFilter[0].(string)] = genFilter[1]
	countFilter := func() interface{} {

		return []bson.M{
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	var totalCount, err = app.Globals.PkgItemRepository.CountWithFilter(ctx, countFilter)
	if err != nil {
		logger.Err("sellerOrderListHandler() => CountWithFilter failed,  pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if totalCount == 0 {
		logger.Err("sellerOrderListHandler() => total count is zero,  pid: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
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
		logger.Err("sellerOrderListHandler() => availablePages less than page, availablePages: %d, pid: %d, filterValue: %s, page: %d, perPage: %d", availablePages, pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		logger.Err("sellerOrderListHandler() => offset invalid, offset: %d, pid: %d, filterValue: %s, page: %d, perPage: %d", offset, pid, filter, page, perPage)
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
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$project": bson.M{"_id": 0, "packages": 1}},
			{"$sort": bson.M{"packages.subpackages." + sortName: sortDirect}},
			{"$skip": offset},
			{"$limit": perPage},
			{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		}
	}

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		logger.Err("sellerOrderListHandler() => FindByFilter failed, pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
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
		logger.Err("sellerOrderListHandler() => could not serialize sellerOrderList, sellerOrderList: %v, error:%s", sellerOrderList, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(totalCount),
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

func (server *Server) sellerAllOrdersHandler(ctx context.Context, pid uint64, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	if page <= 0 || perPage <= 0 {
		logger.Err("sellerAllOrdersHandler() => page or perPage invalid, pid: %d, page: %d, perPage: %d", pid, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	filters := make(bson.M, 3)
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil

	var criteria = make([]interface{}, 0, len(server.sellerFilterStates))
	for filter, _ := range server.sellerFilterStates {
		if filter == DeliveryPendingFilter {
			criteria = append(criteria, map[string]string{
				"packages.subpackages.status": states.DeliveryPending.StateName(),
			})
			criteria = append(criteria, map[string]string{
				"packages.subpackages.status": states.DeliveryDelayed.StateName(),
			})
		} else if filter != AllCanceledFilter {
			if filter == CanceledBySellerFilter || filter == CanceledByBuyerFilter ||
				filter == DeliveryFailedFilter || filter == ReturnDeliveryFailedFilter {
				criteria = append(criteria, map[string]string{
					"packages.subpackages.tracking.history.name": server.sellerFilterStates[filter][0].StateName(),
				})
			} else {
				criteria = append(criteria, map[string]string{
					"packages.subpackages.status": server.sellerFilterStates[filter][0].StateName(),
				})
			}
		}
	}
	filters["$or"] = bson.A(criteria)
	countFilter := func() interface{} {
		return []bson.M{
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	var totalCount, err = app.Globals.PkgItemRepository.CountWithFilter(ctx, countFilter)
	if err != nil {
		logger.Err("sellerAllOrdersHandler() => CountWithFilter failed,  pid: %d, page: %d, perPage: %d, error: %s", pid, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if totalCount == 0 {
		logger.Err("sellerAllOrdersHandler() => total count is zero,  pid: %d, page: %d, perPage: %d", pid, page, perPage)
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
		logger.Err("sellerAllOrdersHandler() => availablePages less than page, availablePages: %d, pid: %d, page: %d, perPage: %d", availablePages, pid, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		logger.Err("sellerAllOrdersHandler() => offset invalid, offset: %d, pid: %d, page: %d, perPage: %d", offset, pid, page, perPage)
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
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$project": bson.M{"_id": 0, "packages": 1}},
			{"$sort": bson.M{"packages.subpackages." + sortName: sortDirect}},
			{"$skip": offset},
			{"$limit": perPage},
			{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		}
	}

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		logger.Err("sellerAllOrdersHandler() => FindByFilter failed, pid: %d, page: %d, perPage: %d, error: %s", pid, page, perPage, err)
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
		logger.Err("sellerAllOrdersHandler() => could not serialize sellerOrderList, sellerOrderList: %v, error:%s", sellerOrderList, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(totalCount),
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

// todo refactor to support projection
// TODO add not found error for FindById
func (server *Server) sellerOrderDetailHandler(ctx context.Context, pid, orderId uint64, filter FilterValue) (*pb.MessageResponse, error) {
	order, err := app.Globals.OrderRepository.FindById(ctx, orderId)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => OrderRepository.FindById failed, orderId: %d, pid: %d, filter:%s , error: %s", orderId, pid, filter, err)
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
		for _, state := range server.sellerFilterStates[filter] {
			if pkgItem.Subpackages[i].Status == state.StateName() {
				for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
					itemDetail := &pb.SellerOrderDetail_ItemDetail{
						SID:         pkgItem.Subpackages[i].SId,
						Sku:         pkgItem.Subpackages[i].Items[j].SKU,
						Status:      pkgItem.Subpackages[i].Status,
						SIdx:        int32(states.FromString(pkgItem.Subpackages[i].Status).StateIndex()),
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
	}

	sellerOrderDetail := &pb.SellerOrderDetail{
		OID:       orderId,
		PID:       pid,
		Amount:    pkgItem.Invoice.Subtotal,
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
	if page <= 0 || perPage <= 0 {
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
	//var totalCount, err = app.Globals.PkgItemRepository.CountWithFilter(ctx, countFilter)
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

	var returnFilter bson.D

	switch filter {
	case ReturnRequestPendingFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnRequestPending.StateName()}}
		break
	case ReturnRequestRejectedFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnRequestRejected.StateName()}}
		break
	case ReturnShipmentPendingFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnShipmentPending.StateName()}}
		break
	case ReturnShippedFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnShipped.StateName()}}
		break
	case ReturnDeliveryPendingFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnDeliveryPending.StateName()}}
		break
	case ReturnDeliveryDelayedFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnDeliveryDelayed.StateName()}}
		break
	case ReturnDeliveredFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnDelivered.StateName()}}
		break
	case ReturnDeliveryFailedFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.tracking.history.name", states.ReturnDeliveryFailed.StateName()}}
		break
	case ReturnRejectedFilter:
		returnFilter = bson.D{{"packages.pid", pid}, {"packages.deletedAt", nil}, {"packages.subpackages.status", states.ReturnRejected.StateName()}}
		break
	}

	//if filter == ReturnDeliveredFilter {
	//	returnFilter = bson.D{{"packages.pid", pid}, {"deletedAt", nil}, {"$or", bson.A{
	//		bson.D{{"packages.subpackages.status", states.ReturnDeliveryPending.StateName()}},
	//		bson.D{{"packages.subpackages.status", states.ReturnDeliveryDelayed.StateName()}},
	//		bson.D{{"packages.subpackages.status", states.ReturnDelivered.StateName()}}}}}
	//} else if filter == DeliveryFailedFilter {
	//	returnFilter = bson.D{{"packages.pid", pid}, {"deletedAt", nil}, {"packages.subpackages.tracking.history.name", server.buyerFilterStates[filter][0].StateName()}}
	//} else if filter == ReturnRequestPendingFilter {
	//	returnFilter = bson.D{{"packages.pid", pid}, {"deletedAt", nil}, {"$or", bson.A{
	//		bson.D{{"packages.subpackages.status", states.ReturnRequestPending.StateName()}},
	//		bson.D{{"packages.subpackages.status", states.ReturnRequestRejected.StateName()}}}}}
	//} else {
	//	returnFilter = bson.D{{"packages.pid", pid}, {"deletedAt", nil}, {"packages.subpackages.status", server.buyerFilterStates[filter][0].StateName()}}
	//}

	//genFilter := server.buyerGeneratePipelineFilter(ctx, filter)
	//filters := make(bson.M, 3)
	//filters["packages.pid"] = pid
	//filters["packages.deletedAt"] = nil
	//filters[genFilter[0].(string)] = genFilter[1]
	//pkgFilter := func() (interface{}, string, int) {
	//	return []bson.M{{"$match": filters}}, sortName, sortDirect
	//}

	pkgFilter := func() (interface{}, string, int) {
		return returnFilter, sortName, sortDirect
	}

	orderList, total, err := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, pkgFilter, int64(page), int64(perPage))
	if err != nil {
		logger.Err("sellerOrderReturnDetailListHandler() => FindByFilter failed, pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	filterMap := make(map[string]string, 3)
	for _, enumState := range server.sellerFilterStates[filter] {
		filterMap[enumState.StateName()] = enumState.StateName()
	}

	sellerReturnOrderList := make([]*pb.SellerReturnOrderDetailList_ReturnOrderDetail, 0, len(orderList))
	var itemDetailList []*pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item

	for i := 0; i < len(orderList); i++ {
		itemDetailList = nil
		for j := 0; j < len(orderList[i].Packages); j++ {
			if orderList[i].Packages[j].PId == pid {
				for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
					if _, ok := filterMap[orderList[i].Packages[j].Subpackages[z].Status]; ok {
						itemDetailList = make([]*pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
						for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
							itemOrder := &pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item{
								SID:    orderList[i].Packages[j].Subpackages[z].SId,
								Sku:    orderList[i].Packages[j].Subpackages[z].Items[t].SKU,
								Status: orderList[i].Packages[j].Subpackages[z].Status,
								SIdx:   int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
								Detail: &pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item_Detail{
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
									Invoice: &pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item_Detail_Invoice{
										Unit:             orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit,
										Total:            orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total,
										Original:         orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original,
										Special:          orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special,
										Discount:         orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount,
										SellerCommission: orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.SellerCommission,
										Currency:         orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Currency,
									},
								},
							}

							if orderList[i].Packages[j].Subpackages[z].Shipments != nil &&
								orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail != nil {
								if orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt != nil {
									itemOrder.Detail.ReturnRequestAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
								}
								if orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.ShippedAt != nil {
									itemOrder.Detail.ReturnShippedAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.ShippedAt.Format(ISO8601)
								}
							}

							itemDetailList = append(itemDetailList, itemOrder)
						}
					}
				}

				if itemDetailList != nil {
					returnOrderDetail := &pb.SellerReturnOrderDetailList_ReturnOrderDetail{
						OID:       orderList[i].Packages[j].OrderId,
						Amount:    orderList[i].Packages[j].Invoice.Subtotal,
						RequestAt: orderList[i].Packages[j].CreatedAt.Format(ISO8601),
						Items:     itemDetailList,
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
				} else {
					logger.Err("sellerOrderReturnDetailListHandler() => get item from orderList failed, orderId: %d pid: %d, filterValue: %s, page: %d, perPage: %d", orderList[i].OrderId, pid, filter, page, perPage)
				}
			}
		}
	}

	sellerReturnOrderDetailList := &pb.SellerReturnOrderDetailList{
		PID:               pid,
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

func (server *Server) sellerOrderShipmentReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	shipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ShipmentPending.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ShipmentPending.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	shipmentDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ShipmentDelayed.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ShipmentDelayed.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	shippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.Shipped.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.Shipped.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	shipmentPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, shipmentPendingFilter)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => CountWithFilter for returnRequestPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	shipmentDelayedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, shipmentDelayedFilter)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => CountWithFilter for returnShipmentPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	shippedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, shippedFilter)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => CountWithFilter for returnShippedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderShipmentReports := &pb.SellerOrderShipmentReports{
		SellerId:        userId,
		ShipmentPending: uint32(shipmentPendingCount),
		ShipmentDelayed: uint32(shipmentDelayedCount),
		Shipped:         uint32(shippedCount),
	}

	serializedData, err := proto.Marshal(sellerOrderShipmentReports)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => could not serialize sellerOrderShipmentReports, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderShipmentReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderShipmentReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderReturnReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	returnRequestPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnRequestPending.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnRequestPending.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnRequestRejectedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnRequestRejected.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnRequestRejected.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnShipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipmentPending.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipmentPending.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnShippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipped.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipped.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnDeliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDelivered.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDelivered.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnDeliveryPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDeliveryPending.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDeliveryPending.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnDeliveryDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDeliveryDelayed.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDeliveryDelayed.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnDeliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnDeliveryFailed.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnDeliveryFailed.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnRequestPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnRequestPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShipmentPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnShipmentPendingFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnShipmentPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShippedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnShippedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnShippedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveredCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveredFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveredFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryFailedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveryFailedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveryFailedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnRequestRejectedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnRequestRejectedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnRequestRejectedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveryPendingFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveryPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryDelayedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveryDelayedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveryDelayedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderReturnReports := &pb.SellerOrderReturnReports{
		SellerId:              userId,
		ReturnRequestPending:  uint32(returnRequestPendingCount),
		ReturnShipmentPending: uint32(returnShipmentPendingCount),
		ReturnShipped:         uint32(returnShippedCount),
		ReturnDeliveryPending: uint32(returnDeliveryPendingCount),
		ReturnDeliveryDelayed: uint32(returnDeliveryDelayedCount),
		ReturnDelivered:       uint32(returnDeliveredCount),
		ReturnRequestRejected: uint32(returnRequestRejectedCount),
		ReturnDeliveryFailed:  uint32(returnDeliveryFailedCount),
	}

	serializedData, err := proto.Marshal(sellerOrderReturnReports)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => could not serialize sellerOrderReturnReports, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderReturnReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderReturnReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderDeliveredReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	deliveryPendingAndDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.DeliveryPending.StateName()},
				bson.M{"packages.subpackages.status": states.DeliveryDelayed.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.DeliveryPending.StateName()},
				bson.M{"packages.subpackages.status": states.DeliveryDelayed.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	deliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.Delivered.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.Delivered.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	deliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.DeliveryFailed.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "packages.subpackages.status": states.DeliveryFailed.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	deliveryPendingAndDelayedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, deliveryPendingAndDelayedFilter)
	if err != nil {
		logger.Err("sellerOrderDeliverReportsHandler() => CountWithFilter for deliveryPendingAndDelayedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	deliveredCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, deliveredFilter)
	if err != nil {
		logger.Err("sellerOrderDeliverReportsHandler() => CountWithFilter for deliveredFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	deliveryFailedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, deliveryFailedFilter)
	if err != nil {
		logger.Err("sellerOrderDeliverReportsHandler() => CountWithFilter for deliveryFailedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderDeliveredReports := &pb.SellerOrderDeliveredReports{
		SellerId:                  userId,
		DeliveryPendingAndDelayed: uint32(deliveryPendingAndDelayedCount),
		Delivered:                 uint32(deliveredCount),
		DeliveryFailed:            uint32(deliveryFailedCount),
	}

	serializedData, err := proto.Marshal(sellerOrderDeliveredReports)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => could not serialize sellerOrderDeliveredReports, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderDeliveredReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderDeliveredReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderCancelReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	cancelByBuyerFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.tracking.history.name": states.CanceledByBuyer.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.tracking.history.name": states.CanceledByBuyer.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	cancelBySellerFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.tracking.history.name": states.CanceledBySeller.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.tracking.history.name": states.CanceledBySeller.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	cancelByBuyerCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, cancelByBuyerFilter)
	if err != nil {
		logger.Err("sellerOrderCancelReportsHandler() => CountWithFilter for cancelByBuyerFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	cancelBySellerCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, cancelBySellerFilter)
	if err != nil {
		logger.Err("sellerOrderCancelReportsHandler() => CountWithFilter for cancelBySellerFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderCancelReports := &pb.SellerOrderCancelReports{
		SellerId:       userId,
		CancelBySeller: uint32(cancelBySellerCount),
		CancelByBuyer:  uint32(cancelByBuyerCount),
	}

	serializedData, err := proto.Marshal(sellerOrderCancelReports)
	if err != nil {
		logger.Err("sellerOrderCancelReportsHandler() => could not serialize sellerOrderCancelReports, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderCancelReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderCancelReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerOrderDetailListHandler(ctx context.Context, oid, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	if oid > 0 {
		return server.buyerGetOrderDetailByIdHandler(ctx, oid)
	}

	if page <= 0 || perPage <= 0 {
		logger.Err("buyerOrderDetailListHandler() => page or perPage invalid, userId: %d, page: %d, perPage: %d", userId, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	orderFilter := func() (interface{}, string, int) {
		return bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"$or", bson.A{
				bson.D{{"packages.subpackages.status", states.PaymentFailed.StateName()}},
				bson.D{{"packages.subpackages.status", states.ApprovalPending.StateName()}},
				bson.D{{"packages.subpackages.status", states.ShipmentPending.StateName()}},
				bson.D{{"packages.subpackages.status", states.ShipmentDelayed.StateName()}},
				bson.D{{"packages.subpackages.status", states.Shipped.StateName()}},
				bson.D{{"packages.subpackages.status", states.DeliveryPending.StateName()}},
				bson.D{{"packages.subpackages.status", states.DeliveryDelayed.StateName()}},
				bson.D{{"packages.subpackages.status", states.Delivered.StateName()}},
				bson.D{{"packages.subpackages.status", states.DeliveryFailed.StateName()}},
				bson.D{{"packages.subpackages.status", states.PayToBuyer.StateName()}},
				bson.D{{"packages.subpackages.status", states.PayToSeller.StateName()}}}}},
			sortName, sortDirect
	}

	orderList, total, err := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
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
						SIdx:               int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
						IsCancelable:       false,
						IsReturnable:       false,
						IsReturnCancelable: false,
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
						DeliveryAt:     "",
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
		BuyerId:      userId,
		OrderDetails: orderDetailList,
	}

	serializedData, err := proto.Marshal(buyerOrderDetailList)
	if err != nil {
		logger.Err("buyerOrderDetailListHandler() => could not serialize buyerOrderDetailList, userId: %d, error:%s", userId, err)
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

func (server *Server) buyerGetOrderDetailByIdHandler(ctx context.Context, oid uint64) (*pb.MessageResponse, error) {

	order, err := app.Globals.OrderRepository.FindById(ctx, oid)
	if err != nil {
		logger.Err("buyerGetOrderDetailByIdHandler() => FindByFilter failed, oid: %d, error: %s", oid, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	orderDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail, 0, 1)

	packageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package, 0, len(order.Packages))
	for j := 0; j < len(order.Packages); j++ {
		for z := 0; z < len(order.Packages[j].Subpackages); z++ {
			itemPackageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package_Item, 0, len(order.Packages[j].Subpackages[z].Items))
			for t := 0; t < len(order.Packages[j].Subpackages[z].Items); t++ {
				itemPackageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package_Item{
					SID:                order.Packages[j].Subpackages[z].SId,
					Status:             order.Packages[j].Subpackages[z].Status,
					SIdx:               int32(states.FromString(order.Packages[j].Subpackages[z].Status).StateIndex()),
					IsCancelable:       false,
					IsReturnable:       false,
					IsReturnCancelable: false,
					InventoryId:        order.Packages[j].Subpackages[z].Items[t].InventoryId,
					Title:              order.Packages[j].Subpackages[z].Items[t].Title,
					Brand:              order.Packages[j].Subpackages[z].Items[t].Brand,
					Image:              order.Packages[j].Subpackages[z].Items[t].Image,
					Returnable:         order.Packages[j].Subpackages[z].Items[t].Returnable,
					Quantity:           order.Packages[j].Subpackages[z].Items[t].Quantity,
					Attributes:         order.Packages[j].Subpackages[z].Items[t].Attributes,
					Invoice: &pb.BuyerOrderDetailList_OrderDetail_Package_Item_Invoice{
						Unit:     order.Packages[j].Subpackages[z].Items[t].Invoice.Unit,
						Total:    order.Packages[j].Subpackages[z].Items[t].Invoice.Total,
						Original: order.Packages[j].Subpackages[z].Items[t].Invoice.Original,
						Special:  order.Packages[j].Subpackages[z].Items[t].Invoice.Special,
						Discount: order.Packages[j].Subpackages[z].Items[t].Invoice.Discount,
						Currency: order.Packages[j].Subpackages[z].Items[t].Invoice.Currency,
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
				PID:          order.Packages[j].PId,
				ShopName:     order.Packages[j].ShopName,
				Items:        itemPackageDetailList,
				ShipmentInfo: nil,
			}

			if order.Packages[j].Subpackages[z].Shipments != nil &&
				order.Packages[j].Subpackages[z].Shipments.ShipmentDetail != nil {
				packageDetail.ShipmentInfo = &pb.BuyerOrderDetailList_OrderDetail_Package_Shipment{
					DeliveryAt:     "",
					ShippedAt:      order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601),
					ShipmentAmount: order.Packages[j].ShipmentSpec.ShippingCost,
					CarrierName:    order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.CarrierName,
					TrackingNumber: order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.TrackingNumber,
				}

				packageDetail.ShipmentInfo.DeliveryAt = order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.
					Add(time.Duration(order.Packages[j].ShipmentSpec.ShippingTime) * time.Hour).Format(ISO8601)
			}

			packageDetailList = append(packageDetailList, packageDetail)
		}
	}

	orderDetail := &pb.BuyerOrderDetailList_OrderDetail{
		Address: &pb.BuyerOrderDetailList_OrderDetail_BuyerAddress{
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
		PackageCount:     int32(len(order.Packages)),
		TotalAmount:      order.Invoice.Subtotal,
		PayableAmount:    order.Invoice.GrandTotal,
		Discounts:        order.Invoice.Discount,
		ShipmentAmount:   order.Invoice.ShipmentTotal,
		IsPaymentSuccess: false,
		RequestAt:        order.CreatedAt.Format(ISO8601),
		Packages:         packageDetailList,
	}
	if order.BuyerInfo.ShippingAddress.Location != nil {
		orderDetail.Address.Lat = strconv.Itoa(int(order.BuyerInfo.ShippingAddress.Location.Coordinates[0]))
		orderDetail.Address.Long = strconv.Itoa(int(order.BuyerInfo.ShippingAddress.Location.Coordinates[1]))
	}

	if order.PaymentService != nil && order.PaymentService[0].PaymentResult != nil {
		orderDetail.IsPaymentSuccess = order.PaymentService[0].PaymentResult.Result
	}

	orderDetailList = append(orderDetailList, orderDetail)

	buyerOrderDetailList := &pb.BuyerOrderDetailList{
		BuyerId:      order.BuyerInfo.BuyerId,
		OrderDetails: orderDetailList,
	}

	serializedData, err := proto.Marshal(buyerOrderDetailList)
	if err != nil {
		logger.Err("buyerGetOrderDetailByIdHandler() => could not serialize buyerOrderDetailList, orderId: %d, error:%s", oid, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "BuyerOrderDetailList",
		Meta:   nil,
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
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnRequestPending.StateName()},
				bson.M{"packages.subpackages.status": states.ReturnRequestRejected.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnRequestPending.StateName()},
				bson.M{"packages.subpackages.status": states.ReturnRequestRejected.StateName()}},
				"packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnShipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipmentPending.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.subpackages.status": states.ReturnShipmentPending.StateName(), "packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnShippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnShipped.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.subpackages.status": states.ReturnShipped.StateName(), "packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	// TODO check correct result
	returnDeliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDelivered.StateName()},
				bson.M{"packages.subpackages.status": states.ReturnDeliveryDelayed.StateName()},
				bson.M{"packages.subpackages.status": states.ReturnDeliveryPending.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"$or": bson.A{
				bson.M{"packages.subpackages.status": states.ReturnDelivered.StateName()},
				bson.M{"packages.subpackages.status": states.ReturnDeliveryDelayed.StateName()},
				bson.M{"packages.subpackages.status": states.ReturnDeliveryPending.StateName()}},
				"packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	// TODO check correct result
	returnDeliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "packages.subpackages.status": states.ReturnDeliveryFailed.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.subpackages.status": states.ReturnDeliveryFailed.StateName(), "packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnRequestPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnRequestPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShipmentPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnShipmentPendingFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnShipmentPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShippedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnShippedFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnShippedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveredCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveredFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnDeliveredFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryFailedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveryFailedFilter)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => CountWithFilter for returnDeliveryFailedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	buyerReturnOrderReports := &pb.BuyerReturnOrderReports{
		BuyerId:               userId,
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

func (server *Server) buyerAllReturnOrdersHandler(ctx context.Context, userId uint64, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		logger.Err("buyerAllReturnOrdersHandler() => page or perPage invalid, userId: %d, page: %d, perPage: %d", userId, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	var returnFilter bson.D
	returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"$or", bson.A{
		bson.D{{"packages.subpackages.status", states.ReturnRequestPending.StateName()}},
		bson.D{{"packages.subpackages.status", states.ReturnRequestRejected.StateName()}},
		bson.D{{"packages.subpackages.status", states.ReturnShipmentPending.StateName()}},
		bson.D{{"packages.subpackages.status", states.ReturnShipped.StateName()}},
		bson.D{{"packages.subpackages.status", states.ReturnDeliveryPending.StateName()}},
		bson.D{{"packages.subpackages.status", states.ReturnDeliveryDelayed.StateName()}},
		bson.D{{"packages.subpackages.status", states.ReturnDeliveryFailed.StateName()}},
		bson.D{{"packages.subpackages.status", states.ReturnDelivered.StateName()}}}}}

	//genFilter := server.buyerGeneratePipelineFilter(ctx, filter)
	//filters := make(bson.M, 3)
	//filters["buyerInfo.buyerId"] = userId
	//filters["deletedAt"] = nil
	//filters[genFilter[0].(string)] = genFilter[1]
	orderFilter := func() (interface{}, string, int) {
		return returnFilter, sortName, sortDirect
	}

	orderList, total, err := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
	if err != nil {
		logger.Err("buyerAllReturnOrdersHandler() => FindByFilter failed, userId: %d, page: %d, perPage: %d, error: %s", userId, page, perPage, err)
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
						SID:             orderList[i].Packages[j].Subpackages[z].SId,
						Status:          orderList[i].Packages[j].Subpackages[z].Status,
						SIdx:            int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
						IsCancelable:    false,
						IsAccepted:      false,
						InventoryId:     orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
						Title:           orderList[i].Packages[j].Subpackages[z].Items[t].Title,
						Brand:           orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
						Image:           orderList[i].Packages[j].Subpackages[z].Items[t].Image,
						Returnable:      orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
						Quantity:        orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
						Attributes:      orderList[i].Packages[j].Subpackages[z].Items[t].Attributes,
						Reason:          "",
						ReturnRequestAt: "",
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

					if orderList[i].Packages[j].Subpackages[z].Shipments != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt != nil {
						returnItemPackageDetail.ReturnRequestAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
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
		BuyerId:           userId,
		ReturnOrderDetail: returnOrderDetailList,
	}

	serializedData, err := proto.Marshal(buyerReturnOrderDetailList)
	if err != nil {
		logger.Err("buyerAllReturnOrdersHandler() => could not serialize buyerReturnOrderDetailList, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "BuyerReturnOrderDetailList",
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

func (server *Server) buyerReturnOrderDetailListHandler(ctx context.Context, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		logger.Err("buyerReturnOrderDetailListHandler() => page or perPage invalid, userId: %d, filter: %s, page: %d, perPage: %d", userId, filter, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	if filter == AllOrdersFilter {
		return server.buyerAllReturnOrdersHandler(ctx, userId, page, perPage, sortName, direction)
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	var returnFilter bson.D
	if filter == ReturnDeliveredFilter {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"$or", bson.A{
			bson.D{{"packages.subpackages.status", states.ReturnDeliveryPending.StateName()}},
			bson.D{{"packages.subpackages.status", states.ReturnDeliveryDelayed.StateName()}},
			bson.D{{"packages.subpackages.status", states.ReturnDelivered.StateName()}}}}}
	} else if filter == DeliveryFailedFilter {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"packages.subpackages.tracking.history.name", server.buyerFilterStates[filter][0].StateName()}}
	} else if filter == ReturnRequestPendingFilter {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"$or", bson.A{
			bson.D{{"packages.subpackages.status", states.ReturnRequestPending.StateName()}},
			bson.D{{"packages.subpackages.status", states.ReturnRequestRejected.StateName()}}}}}
	} else {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"packages.subpackages.status", server.buyerFilterStates[filter][0].StateName()}}
	}

	//genFilter := server.buyerGeneratePipelineFilter(ctx, filter)
	//filters := make(bson.M, 3)
	//filters["buyerInfo.buyerId"] = userId
	//filters["deletedAt"] = nil
	//filters[genFilter[0].(string)] = genFilter[1]
	orderFilter := func() (interface{}, string, int) {
		return returnFilter, sortName, sortDirect
	}

	orderList, total, err := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
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
						SID:             orderList[i].Packages[j].Subpackages[z].SId,
						Status:          orderList[i].Packages[j].Subpackages[z].Status,
						SIdx:            int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
						IsCancelable:    false,
						IsAccepted:      false,
						InventoryId:     orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
						Title:           orderList[i].Packages[j].Subpackages[z].Items[t].Title,
						Brand:           orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
						Image:           orderList[i].Packages[j].Subpackages[z].Items[t].Image,
						Returnable:      orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
						Quantity:        orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
						Attributes:      orderList[i].Packages[j].Subpackages[z].Items[t].Attributes,
						Reason:          "",
						ReturnRequestAt: "",
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

					if orderList[i].Packages[j].Subpackages[z].Shipments != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt != nil {
						returnItemPackageDetail.ReturnRequestAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
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
		BuyerId:           userId,
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

	logger.Audit("PaymentGatewayHook() => received payment response: orderId: %d, PaymentId: %s, InvoiceId: %d, result: %v",
		req.OrderID, req.PaymentId, req.InvoiceId, req.Result)
	futureData := server.flowManager.PaymentGatewayResult(ctx, req).Get()

	if futureData.Error() != nil {
		return nil, status.Error(codes.Code(futureData.Error().Code()), futureData.Error().Message())
	}

	return &pg.PaygateHookResponse{Ok: true}, nil
}

func (server Server) NewOrder(ctx context.Context, req *pb.RequestNewOrder) (*pb.ResponseNewOrder, error) {

	//ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)

	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
	if err != nil {
		logger.Err("RequestHandler() => UserService.AuthenticateContextToken failed, error: %s ", err)
		//return nil, status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	}

	//// TODO check acl
	//if uint64(userAcl.User().UserID) != req.Meta.UID {
	//	logger.Err("RequestHandler() => request userId %d mismatch with token userId: %d", req.Meta.UID, userAcl.User().UserID)
	//	return nil, status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	//}

	if ctx.Value(string(utils.CtxUserID)) == nil {
		if userAcl != nil {
			ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(userAcl.User().UserID))
		} else {
			ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(0))
		}
	}

	if ctx.Value(string(utils.CtxUserACL)) == nil {
		if userAcl != nil {
			ctx = context.WithValue(ctx, string(utils.CtxUserACL), userAcl)
		}
	}

	iFuture := future.Factory().SetCapacity(1).Build()
	iFrame := frame.Factory().SetDefaultHeader(frame.HeaderNewOrder, req).SetFuture(iFuture).Build()
	server.flowManager.MessageHandler(ctx, iFrame)
	futureData := iFuture.Get()

	if futureData.Error() != nil {
		futureErr := futureData.Error()
		return nil, status.Error(codes.Code(futureErr.Code()), futureErr.Message())
	}

	callbackUrl, ok := futureData.Data().(string)
	if ok != true {
		logger.Err("NewOrder received data of futureData invalid, type: %T, value, %v", futureData.Data(), futureData.Data())
		return nil, status.Error(500, "Unknown Error")
	}

	responseNewOrder := pb.ResponseNewOrder{
		CallbackUrl: callbackUrl,
	}

	return &responseNewOrder, nil
}

// TODO Add checking acl
//func (server Server) SellerFindAllItems(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseSellerFindAllItems, error) {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
//	if err != nil {
//		logger.Err("SellerFindAllItems() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if strconv.Itoa(int(userAcl.User().UserID)) != req.Id {
//		logger.Err(" SellerFindAllItems() => token userId %d not authorized for sellerId %s", userAcl.User().UserID, req.Id)
//		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	sellerId, err := strconv.Atoi(req.Id)
//	if err != nil {
//		logger.Err(" SellerFindAllItems() => sellerId invalid: %s", req.Id)
//		return nil, status.Error(codes.Code(future.BadRequest), "PID Invalid")
//	}
//
//	orders, err := app.Globals.OrderRepository.FindByFilter(func() interface{} {
//		return bson.D{{"items.sellerInfo.sellerId", uint64(sellerId)}}
//	})
//
//	if err != nil {
//		logger.Err("SellerFindAllItems failed, sellerId: %s, error: %s", req.Id, err.Error())
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	sellerItemMap := make(map[string]*pb.SellerFindAllItems, 16)
//
//	for _, order := range orders {
//		for _, orderItem := range order.Items {
//			if strconv.Itoa(int(orderItem.SellerInfo.PId)) == req.Id {
//				if _, ok := sellerItemMap[orderItem.InventoryId]; !ok {
//					newResponseItem := &pb.SellerFindAllItems{
//						OrderId:     order.OrderId,
//						ItemId:      orderItem.ItemId,
//						InventoryId: orderItem.InventoryId,
//						Title:       orderItem.Title,
//						Image:       orderItem.Image,
//						Returnable:  orderItem.Returnable,
//						Status: &pb.Status{
//							OrderStatus: order.Status,
//							ItemStatus:  orderItem.Status,
//							StepStatus:  "none",
//						},
//						CreatedAt:  orderItem.CreatedAt.Format(ISO8601),
//						UpdatedAt:  orderItem.UpdatedAt.Format(ISO8601),
//						Quantity:   orderItem.Quantity,
//						Attributes: orderItem.Attributes,
//						Price: &pb.SellerFindAllItems_Price{
//							Unit:             orderItem.Invoice.Unit,
//							Total:            orderItem.Invoice.Original,
//							SellerCommission: orderItem.Invoice.SellerCommission,
//							Currency:         orderItem.Invoice.Currency,
//						},
//						DeliveryAddress: &pb.Address{
//							FirstName:     order.BuyerInfo.ShippingAddress.FirstName,
//							LastName:      order.BuyerInfo.ShippingAddress.LastName,
//							Address:       order.BuyerInfo.ShippingAddress.Address,
//							Phone:         order.BuyerInfo.ShippingAddress.Phone,
//							Mobile:        order.BuyerInfo.ShippingAddress.Mobile,
//							Country:       order.BuyerInfo.ShippingAddress.Country,
//							City:          order.BuyerInfo.ShippingAddress.City,
//							Province:      order.BuyerInfo.ShippingAddress.Province,
//							Neighbourhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
//							ZipCode:       order.BuyerInfo.ShippingAddress.ZipCode,
//						},
//					}
//
//					if order.BuyerInfo.ShippingAddress.Location != nil {
//						newResponseItem.DeliveryAddress.Lat = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[1])
//						newResponseItem.DeliveryAddress.Long = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[0])
//					}
//
//					lastStep := orderItem.Progress.StepsHistory[len(orderItem.Progress.StepsHistory)-1]
//					if lastStep.ActionHistory != nil {
//						lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
//						newResponseItem.Status.StepStatus = lastAction.Name
//					} else {
//						newResponseItem.Status.StepStatus = "none"
//						logger.Audit("SellerFindAllItems() => Actions History is nil, orderId: %d, sid: %d", order.OrderId, orderItem.ItemId)
//					}
//
//					sellerItemMap[orderItem.InventoryId] = newResponseItem
//				}
//			}
//		}
//	}
//
//	var response = pb.ResponseSellerFindAllItems{}
//	response.Items = make([]*pb.SellerFindAllItems, 0, len(sellerItemMap))
//
//	for _, item := range sellerItemMap {
//		response.Items = append(response.Items, item)
//	}
//
//	return &response, nil
//}

// TODO Add checking acl
//func (server Server) BuyerOrderAction(ctx context.Context, req *pb.RequestBuyerOrderAction) (*pb.ResponseBuyerOrderAction, error) {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
//	if err != nil {
//		logger.Err("BuyerOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if userAcl.User().UserID != int64(req.BuyerId) {
//		logger.Err(" BuyerOrderAction() => token userId %d not authorized for buyerId %d", userAcl.User().UserID, req.BuyerId)
//		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	promiseHandler := server.flowManager.BuyerApprovalPending(ctx, req)
//	futureData := promiseHandler.Get()
//	if futureData == nil {
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if futureData.Ex != nil {
//		futureErr := futureData.Ex.(future.FutureError)
//		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
//	}
//
//	return &pb.ResponseBuyerOrderAction{Result: true}, nil
//}

// TODO Add checking acl
//func (server Server) SellerOrderAction(ctx context.Context, req *pb.RequestSellerOrderAction) (*pb.ResponseSellerOrderAction, error) {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
//	if err != nil {
//		logger.Err("SellerOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if userAcl.User().UserID != int64(req.PId) {
//		logger.Err("SellerOrderAction() => token userId %d not authorized for sellerId %d", userAcl.User().UserID, req.PId)
//		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	promiseHandler := server.flowManager.SellerApprovalPending(ctx, req)
//	futureData := promiseHandler.Get()
//	if futureData == nil {
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if futureData.Ex != nil {
//		futureErr := futureData.Ex.(future.FutureError)
//		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
//	}
//
//	return &pb.ResponseSellerOrderAction{Result: true}, nil
//}

// TODO Add checking acl
//func (server Server) BuyerFindAllOrders(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseBuyerFindAllOrders, error) {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
//	if err != nil {
//		logger.Err("BuyerFindAllOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if strconv.Itoa(int(userAcl.User().UserID)) != req.Id {
//		logger.Err(" BuyerFindAllOrders() => token userId %d not authorized of buyerId %s", userAcl.User().UserID, req.Id)
//		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	buyerId, err := strconv.Atoi(req.Id)
//	if err != nil {
//		logger.Err(" SellerFindAllItems() => buyerId invalid: %s", req.Id)
//		return nil, status.Error(codes.Code(future.BadRequest), "BuyerId Invalid")
//	}
//
//	orders, err := app.Globals.OrderRepository.FindByFilter(func() interface{} {
//		return bson.D{{"buyerInfo.buyerId", uint64(buyerId)}}
//	})
//
//	if err != nil {
//		logger.Err("SellerFindAllItems failed, buyerId: %s, error: %s", req.Id, err.Error())
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	var response pb.ResponseBuyerFindAllOrders
//	responseOrders := make([]*pb.BuyerAllOrders, 0, len(orders))
//
//	for _, order := range orders {
//		responseOrder := &pb.BuyerAllOrders{
//			OrderId:     order.OrderId,
//			CreatedAt:   order.CreatedAt.Format(ISO8601),
//			UpdatedAt:   order.UpdatedAt.Format(ISO8601),
//			OrderStatus: order.Status,
//			Amount: &pb.Amount{
//				Total:         order.Invoice.Total,
//				Subtotal:      order.Invoice.Subtotal,
//				Discount:      order.Invoice.Discount,
//				Currency:      order.Invoice.Currency,
//				ShipmentTotal: order.Invoice.ShipmentTotal,
//				PaymentMethod: order.Invoice.PaymentMethod,
//				PaymentOption: order.Invoice.PaymentGateway,
//				Voucher: &pb.Voucher{
//					Amount: order.Invoice.Voucher.Amount,
//					Code:   order.Invoice.Voucher.Code,
//				},
//			},
//			ShippingAddress: &pb.Address{
//				FirstName:     order.BuyerInfo.ShippingAddress.FirstName,
//				LastName:      order.BuyerInfo.ShippingAddress.LastName,
//				Address:       order.BuyerInfo.ShippingAddress.Address,
//				Phone:         order.BuyerInfo.ShippingAddress.Phone,
//				Mobile:        order.BuyerInfo.ShippingAddress.Mobile,
//				Country:       order.BuyerInfo.ShippingAddress.Country,
//				City:          order.BuyerInfo.ShippingAddress.City,
//				Province:      order.BuyerInfo.ShippingAddress.Province,
//				Neighbourhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
//			},
//			Items: make([]*pb.BuyerOrderItems, 0, len(order.Items)),
//		}
//
//		if order.BuyerInfo.ShippingAddress.Location != nil {
//			responseOrder.ShippingAddress.Lat = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[1])
//			responseOrder.ShippingAddress.Long = fmt.Sprintf("%f", order.BuyerInfo.ShippingAddress.Location.Coordinates[0])
//		}
//
//		orderItemMap := make(map[string]*pb.BuyerOrderItems, 16)
//
//		for _, item := range order.Items {
//			if _, ok := orderItemMap[item.InventoryId]; !ok {
//				newResponseOrderItem := &pb.BuyerOrderItems{
//					InventoryId: item.InventoryId,
//					Title:       item.Title,
//					Brand:       item.Brand,
//					Category:    item.Category,
//					Guaranty:    item.Guaranty,
//					Image:       item.Image,
//					Returnable:  item.Returnable,
//					PId:    item.SellerInfo.PId,
//					Quantity:    item.Quantity,
//					Attributes:  item.Attributes,
//					ItemStatus:  item.Status,
//					Price: &pb.BuyerOrderItems_Price{
//						Unit:     item.Invoice.Unit,
//						Total:    item.Invoice.Total,
//						Original: item.Invoice.Original,
//						Special:  item.Invoice.Special,
//						Currency: item.Invoice.Currency,
//					},
//					Shipment: &pb.BuyerOrderItems_ShipmentSpec{
//						CarrierName:  item.ShipmentSpec.CarrierName,
//						ShippingCost: item.ShipmentSpec.ShippingCost,
//					},
//				}
//
//				lastStep := item.Progress.StepsHistory[len(item.Progress.StepsHistory)-1]
//
//				if lastStep.ActionHistory != nil {
//					lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
//					newResponseOrderItem.StepStatus = lastAction.Name
//				} else {
//					newResponseOrderItem.StepStatus = "none"
//					logger.Audit("BuyerFindAllOrders() => Actions History is nil, orderId: %d, sid: %d", order.OrderId, item.ItemId)
//				}
//				orderItemMap[item.InventoryId] = newResponseOrderItem
//			}
//		}
//
//		for _, orderItem := range orderItemMap {
//			responseOrder.Items = append(responseOrder.Items, orderItem)
//		}
//
//		responseOrders = append(responseOrders, responseOrder)
//	}
//
//	response.Orders = responseOrders
//	return &response, nil
//}

//func (server Server) convertNewOrderRequestToMessage(req *pb.RequestNewOrder) *pb.MessageRequest {
//
//	serializedOrder, err := proto.Marshal(req)
//	if err != nil {
//		logger.Err("could not serialize timestamp")
//	}
//
//	request := pb.MessageRequest{
//		Name:   "NewOrder",
//		UTP:   string(DataReqType),
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
//func (server Server) BackOfficeOrdersListView(ctx context.Context, req *pb.RequestBackOfficeOrdersList) (*pb.ResponseBackOfficeOrdersList, error) {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
//	if err != nil {
//		logger.Err("BackOfficeOrdersListView() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	// TODO Must Be changed
//	if userAcl.User().UserID <= 0 {
//		logger.Err("BackOfficeOrdersListView() => token userId %d not authorized", userAcl.User().UserID)
//		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	promiseHandler := server.flowManager.BackOfficeOrdersListView(ctx, req)
//	futureData := promiseHandler.Get()
//	if futureData == nil {
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if futureData.Ex != nil {
//		futureErr := futureData.Ex.(future.FutureError)
//		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
//	}
//
//	return futureData.Data.(*pb.ResponseBackOfficeOrdersList), nil
//}

// TODO Add checking acl and authenticate
//func (server Server) BackOfficeOrderDetailView(ctx context.Context, req *pb.RequestIdentifier) (*pb.ResponseOrderDetailView, error) {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
//	if err != nil {
//		logger.Err("BackOfficeOrderDetailView() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	// TODO Must Be changed
//	if userAcl.User().UserID <= 0 {
//		logger.Err("BackOfficeOrderDetailView() => token userId %d not authorized", userAcl.User().UserID)
//		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	promiseHandler := server.flowManager.BackOfficeOrderDetailView(ctx, req)
//	futureData := promiseHandler.Get()
//	if futureData == nil {
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if futureData.Ex != nil {
//		futureErr := futureData.Ex.(future.FutureError)
//		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
//	}
//
//	return futureData.Data.(*pb.ResponseOrderDetailView), nil
//}

// TODO Add checking acl and authenticate
//func (server Server) BackOfficeOrderAction(ctx context.Context, req *pb.RequestBackOfficeOrderAction) (*pb.ResponseBackOfficeOrderAction, error) {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
//	if err != nil {
//		logger.Err("BackOfficeOrderAction() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	// TODO Must Be changed
//	if userAcl.User().UserID <= 0 {
//		logger.Err("BackOfficeOrderAction() => token userId %d not authorized", userAcl.User().UserID)
//		return nil, status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	promiseHandler := server.flowManager.OperatorActionPending(ctx, req)
//	futureData := promiseHandler.Get()
//	if futureData == nil {
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if futureData.Ex != nil {
//		futureErr := futureData.Ex.(future.FutureError)
//		return nil, status.Error(codes.Code(futureErr.Code), futureErr.Reason)
//	}
//
//	return &pb.ResponseBackOfficeOrderAction{Result: true}, nil
//
//}

// TODO Add checking acl and authenticate
//func (server Server) SellerReportOrders(req *pb.RequestSellerReportOrders, srv pb.OrderService_SellerReportOrdersServer) error {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(srv.Context())
//	if err != nil {
//		logger.Err("SellerReportOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if userAcl.User().UserID != int64(req.PId) {
//		logger.Err(" SellerFindAllItems() => token userId %d not authorized for sellerId %d", userAcl.User().UserID, req.PId)
//		return status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	promiseHandler := server.flowManager.SellerReportOrders(req, srv)
//	futureData := promiseHandler.Get()
//	if futureData == nil {
//		return status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if futureData.Ex != nil {
//		futureErr := futureData.Ex.(future.FutureError)
//		return status.Error(codes.Code(futureErr.Code), futureErr.Reason)
//	}
//
//	return nil
//}

// TODO Add checking acl and authenticate
//func (server Server) BackOfficeReportOrderItems(req *pb.RequestBackOfficeReportOrderItems, srv pb.OrderService_BackOfficeReportOrderItemsServer) error {
//
//	userAcl, err := app.Globals.UserService.AuthenticateContextToken(srv.Context())
//	if err != nil {
//		logger.Err("SellerReportOrders() => UserService.AuthenticateContextToken failed, error: %s ", err)
//		return status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	// TODO Must Be changed
//	if userAcl.User().UserID <= 0 {
//		logger.Err("BackOfficeOrderAction() => token userId %d not authorized", userAcl.User().UserID)
//		return status.Error(codes.Code(future.Forbidden), "User token not authorized")
//	}
//
//	promiseHandler := server.flowManager.BackOfficeReportOrderItems(req, srv)
//	futureData := promiseHandler.Get()
//	if futureData == nil {
//		return status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	if futureData.Ex != nil {
//		futureErr := futureData.Ex.(future.FutureError)
//		return status.Error(codes.Code(futureErr.Code), futureErr.Reason)
//	}
//
//	return nil
//}

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
