package grpc_server

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/buyer"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"path"
	"runtime/debug"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.faza.io/go-framework/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"net"
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
	ShipmentDelayedFilter          FilterValue = "ShipmentDelayed"
	ShippedFilter                  FilterValue = "Shipped"
	DeliveryPendingFilter          FilterValue = "DeliveryPending"
	DeliveryDelayedFilter          FilterValue = "DeliveryDelayed"
	DeliveredFilter                FilterValue = "Delivered"
	DeliveryFailedFilter           FilterValue = "DeliveryFailed"
	ReturnRequestPendingFilter     FilterValue = "ReturnRequestPending"
	ReturnRequestRejectedFilter    FilterValue = "ReturnRequestRejected"
	ReturnCanceledFilter           FilterValue = "ReturnCanceled"
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
	SchedulerUser UserType = "Schedulers"
)

const (
	//SellerAllOrders             RequestName = "SellerAllOrders"
	SellerOrderList             RequestName = "SellerOrderList"
	SellerOrderDetail           RequestName = "SellerOrderDetail"
	SellerReturnOrderDetailList RequestName = "SellerReturnOrderDetailList"
	SellerOrderDashboardReports RequestName = "SellerOrderDashboardReports"
	SellerOrderShipmentReports  RequestName = "SellerOrderShipmentReports"
	SellerOrderDeliveredReports RequestName = "SellerOrderDeliveredReports"
	SellerOrderReturnReports    RequestName = "SellerOrderReturnReports"
	SellerOrderCancelReports    RequestName = "SellerOrderCancelReports"

	//BuyerAllOrders			   RequestName = "BuyerAllOrders"
	BuyerAllReturnOrders       RequestName = "BuyerAllReturnOrders"
	BuyerOrderDetailList       RequestName = "BuyerOrderDetailList"
	BuyerReturnOrderReports    RequestName = "BuyerReturnOrderReports"
	BuyerReturnOrderDetailList RequestName = "BuyerReturnOrderDetailList"

	//OperatorAllOrders	RequestName = "OperatorAllOrders"
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

type stackTraceDisabler struct{}

func (s stackTraceDisabler) Enabled(zapcore.Level) bool {
	return false
}

type FilterState struct {
	expectedState states.IEnumState
	actualState   states.IEnumState
}

type FilterQueryState struct {
	state     states.IEnumState
	queryPath string
}

type Server struct {
	pb.UnimplementedOrderServiceServer
	pg.UnimplementedBankResultHookServer
	flowManager          domain.IFlowManager
	address              string
	port                 uint16
	requestFilters       map[RequestName][]FilterValue
	buyerFilterStates    map[FilterValue][]FilterState
	sellerFilterStates   map[FilterValue][]FilterState
	operatorFilterStates map[FilterValue][]FilterState
	queryPathStates      map[FilterValue]FilterQueryState
	actionStates         map[UserType][]actions.IAction
}

func NewServer(address string, port uint16, flowManager domain.IFlowManager) Server {
	buyerStatesMap := make(map[FilterValue][]FilterState, 8)
	buyerStatesMap[ApprovalPendingFilter] = []FilterState{{states.ApprovalPending, states.ApprovalPending}}
	buyerStatesMap[ShipmentPendingFilter] = []FilterState{{states.ShipmentPending, states.ShipmentDelayed}}
	buyerStatesMap[ShippedFilter] = []FilterState{{states.Shipped, states.Shipped}}
	buyerStatesMap[DeliveredFilter] = []FilterState{{states.DeliveryPending, states.DeliveryPending}, {states.DeliveryDelayed, states.DeliveryDelayed}, {states.Delivered, states.Delivered}}
	buyerStatesMap[DeliveryFailedFilter] = []FilterState{{states.DeliveryFailed, states.PayToBuyer}}
	buyerStatesMap[ReturnRequestPendingFilter] = []FilterState{{states.ReturnRequestPending, states.ReturnRequestPending}, {states.ReturnRequestRejected, states.ReturnRequestRejected}}
	buyerStatesMap[ReturnShipmentPendingFilter] = []FilterState{{states.ReturnShipmentPending, states.ReturnShipmentPending}}
	buyerStatesMap[ReturnShippedFilter] = []FilterState{{states.ReturnShipped, states.ReturnShipped}}
	buyerStatesMap[ReturnDeliveredFilter] = []FilterState{{states.ReturnDeliveryPending, states.ReturnDeliveryPending}, {states.ReturnDeliveryDelayed, states.ReturnDeliveryDelayed}, {states.ReturnDelivered, states.ReturnDelivered}}
	buyerStatesMap[ReturnDeliveryFailedFilter] = []FilterState{{states.ReturnDeliveryFailed, states.PayToSeller}}

	operatorFilterStatesMap := make(map[FilterValue][]FilterState, 30)
	operatorFilterStatesMap[NewOrderFilter] = []FilterState{{states.NewOrder, states.NewOrder}}
	operatorFilterStatesMap[PaymentPendingFilter] = []FilterState{{states.PaymentPending, states.PaymentPending}}
	operatorFilterStatesMap[PaymentSuccessFilter] = []FilterState{{states.PaymentSuccess, states.ApprovalPending}}
	operatorFilterStatesMap[PaymentFailedFilter] = []FilterState{{states.PaymentFailed, states.PaymentFailed}}
	operatorFilterStatesMap[OrderVerificationPendingFilter] = []FilterState{{states.OrderVerificationPending, states.ApprovalPending}}
	operatorFilterStatesMap[OrderVerificationSuccessFilter] = []FilterState{{states.OrderVerificationSuccess, states.ApprovalPending}}
	operatorFilterStatesMap[OrderVerificationFailedFilter] = []FilterState{{states.OrderVerificationFailed, states.PayToBuyer}}
	operatorFilterStatesMap[ApprovalPendingFilter] = []FilterState{{states.ApprovalPending, states.ApprovalPending}}
	operatorFilterStatesMap[CanceledBySellerFilter] = []FilterState{{states.CanceledBySeller, states.PayToBuyer}}
	operatorFilterStatesMap[CanceledByBuyerFilter] = []FilterState{{states.CanceledByBuyer, states.PayToBuyer}}
	operatorFilterStatesMap[ShipmentPendingFilter] = []FilterState{{states.ShipmentPending, states.ShipmentPending}}
	operatorFilterStatesMap[ShipmentDelayedFilter] = []FilterState{{states.ShipmentDelayed, states.ShipmentDelayed}}
	operatorFilterStatesMap[ShippedFilter] = []FilterState{{states.Shipped, states.Shipped}}
	operatorFilterStatesMap[DeliveryPendingFilter] = []FilterState{{states.DeliveryPending, states.DeliveryPending}}
	operatorFilterStatesMap[DeliveryDelayedFilter] = []FilterState{{states.DeliveryDelayed, states.DeliveryDelayed}}
	operatorFilterStatesMap[DeliveredFilter] = []FilterState{{states.Delivered, states.Delivered}}
	operatorFilterStatesMap[DeliveryFailedFilter] = []FilterState{{states.DeliveryFailed, states.PayToBuyer}}
	operatorFilterStatesMap[ReturnRequestPendingFilter] = []FilterState{{states.ReturnRequestPending, states.ReturnRequestPending}}
	operatorFilterStatesMap[ReturnRequestRejectedFilter] = []FilterState{{states.ReturnRequestRejected, states.ReturnRequestRejected}}
	operatorFilterStatesMap[ReturnCanceledFilter] = []FilterState{{states.ReturnCanceled, states.PayToSeller}}
	operatorFilterStatesMap[ReturnShipmentPendingFilter] = []FilterState{{states.ReturnShipmentPending, states.ReturnShipmentPending}}
	operatorFilterStatesMap[ReturnShippedFilter] = []FilterState{{states.ReturnShipped, states.ReturnShipped}}
	operatorFilterStatesMap[ReturnDeliveryPendingFilter] = []FilterState{{states.ReturnDeliveryPending, states.ReturnDeliveryPending}}
	operatorFilterStatesMap[ReturnDeliveryDelayedFilter] = []FilterState{{states.ReturnDeliveryDelayed, states.ReturnDeliveryDelayed}}
	operatorFilterStatesMap[ReturnDeliveredFilter] = []FilterState{{states.ReturnDelivered, states.ReturnDelivered}}
	operatorFilterStatesMap[ReturnDeliveryFailedFilter] = []FilterState{{states.ReturnDeliveryFailed, states.PayToSeller}}
	operatorFilterStatesMap[ReturnRejectedFilter] = []FilterState{{states.ReturnRejected, states.ReturnRejected}}
	operatorFilterStatesMap[PayToBuyerFilter] = []FilterState{{states.PayToBuyer, states.PayToBuyer}}
	operatorFilterStatesMap[PayToSellerFilter] = []FilterState{{states.PayToSeller, states.PayToSeller}}

	sellerFilterStatesMap := make(map[FilterValue][]FilterState, 30)
	sellerFilterStatesMap[ApprovalPendingFilter] = []FilterState{{states.ApprovalPending, states.ApprovalPending}}
	sellerFilterStatesMap[CanceledBySellerFilter] = []FilterState{{states.CanceledBySeller, states.PayToBuyer}}
	sellerFilterStatesMap[CanceledByBuyerFilter] = []FilterState{{states.CanceledByBuyer, states.PayToBuyer}}
	sellerFilterStatesMap[ShipmentPendingFilter] = []FilterState{{states.ShipmentPending, states.ShipmentPending}}
	sellerFilterStatesMap[ShipmentDelayedFilter] = []FilterState{{states.ShipmentDelayed, states.ShipmentDelayed}}
	sellerFilterStatesMap[ShippedFilter] = []FilterState{{states.Shipped, states.Shipped}}
	sellerFilterStatesMap[DeliveryPendingFilter] = []FilterState{{states.DeliveryPending, states.DeliveryPending}, {states.DeliveryDelayed, states.DeliveryDelayed}}
	sellerFilterStatesMap[DeliveredFilter] = []FilterState{{states.Delivered, states.Delivered}}
	sellerFilterStatesMap[DeliveryFailedFilter] = []FilterState{{states.DeliveryFailed, states.PayToBuyer}}
	sellerFilterStatesMap[ReturnRequestPendingFilter] = []FilterState{{states.ReturnRequestPending, states.ReturnRequestPending}}
	sellerFilterStatesMap[ReturnRequestRejectedFilter] = []FilterState{{states.ReturnRequestRejected, states.ReturnRequestRejected}}
	sellerFilterStatesMap[ReturnCanceledFilter] = []FilterState{{states.ReturnCanceled, states.PayToSeller}}
	sellerFilterStatesMap[ReturnShipmentPendingFilter] = []FilterState{{states.ReturnShipmentPending, states.ReturnShipmentPending}}
	sellerFilterStatesMap[ReturnShippedFilter] = []FilterState{{states.ReturnShipped, states.ReturnShipped}}
	sellerFilterStatesMap[ReturnDeliveryPendingFilter] = []FilterState{{states.ReturnDeliveryPending, states.ReturnDeliveryPending}}
	sellerFilterStatesMap[ReturnDeliveryDelayedFilter] = []FilterState{{states.ReturnDeliveryDelayed, states.ReturnDeliveryDelayed}}
	sellerFilterStatesMap[ReturnDeliveredFilter] = []FilterState{{states.ReturnDelivered, states.ReturnDelivered}}
	sellerFilterStatesMap[ReturnDeliveryFailedFilter] = []FilterState{{states.ReturnDeliveryFailed, states.PayToSeller}}
	sellerFilterStatesMap[ReturnRejectedFilter] = []FilterState{{states.ReturnRejected, states.ReturnRejected}}
	sellerFilterStatesMap[PayToSellerFilter] = []FilterState{{states.PayToSeller, states.PayToSeller}}

	queryPathStatesMap := make(map[FilterValue]FilterQueryState, 30)
	queryPathStatesMap[NewOrderFilter] = FilterQueryState{states.NewOrder, "packages.subpackages.status"}
	queryPathStatesMap[PaymentPendingFilter] = FilterQueryState{states.PaymentPending, "packages.subpackages.status"}
	queryPathStatesMap[PaymentSuccessFilter] = FilterQueryState{states.PaymentSuccess, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[PaymentFailedFilter] = FilterQueryState{states.PaymentFailed, "packages.subpackages.status"}
	queryPathStatesMap[OrderVerificationPendingFilter] = FilterQueryState{states.OrderVerificationPending, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[OrderVerificationSuccessFilter] = FilterQueryState{states.OrderVerificationSuccess, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[OrderVerificationFailedFilter] = FilterQueryState{states.OrderVerificationFailed, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[ApprovalPendingFilter] = FilterQueryState{states.ApprovalPending, "packages.subpackages.status"}
	queryPathStatesMap[CanceledBySellerFilter] = FilterQueryState{states.CanceledBySeller, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[CanceledByBuyerFilter] = FilterQueryState{states.CanceledByBuyer, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[ShipmentPendingFilter] = FilterQueryState{states.ShipmentPending, "packages.subpackages.status"}
	queryPathStatesMap[ShipmentDelayedFilter] = FilterQueryState{states.ShipmentDelayed, "packages.subpackages.status"}
	queryPathStatesMap[ShippedFilter] = FilterQueryState{states.Shipped, "packages.subpackages.status"}
	queryPathStatesMap[DeliveryPendingFilter] = FilterQueryState{states.DeliveryPending, "packages.subpackages.status"}
	queryPathStatesMap[DeliveryDelayedFilter] = FilterQueryState{states.DeliveryDelayed, "packages.subpackages.status"}
	queryPathStatesMap[DeliveredFilter] = FilterQueryState{states.Delivered, "packages.subpackages.status"}
	queryPathStatesMap[DeliveryFailedFilter] = FilterQueryState{states.DeliveryFailed, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[ReturnRequestPendingFilter] = FilterQueryState{states.ReturnRequestPending, "packages.subpackages.status"}
	queryPathStatesMap[ReturnRequestRejectedFilter] = FilterQueryState{states.ReturnRequestRejected, "packages.subpackages.status"}
	queryPathStatesMap[ReturnCanceledFilter] = FilterQueryState{states.ReturnCanceled, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[ReturnShipmentPendingFilter] = FilterQueryState{states.ReturnShipmentPending, "packages.subpackages.status"}
	queryPathStatesMap[ReturnShippedFilter] = FilterQueryState{states.ReturnShipped, "packages.subpackages.status"}
	queryPathStatesMap[ReturnDeliveryPendingFilter] = FilterQueryState{states.ReturnDeliveryPending, "packages.subpackages.status"}
	queryPathStatesMap[ReturnDeliveryDelayedFilter] = FilterQueryState{states.ReturnDeliveryDelayed, "packages.subpackages.status"}
	queryPathStatesMap[ReturnDeliveredFilter] = FilterQueryState{states.ReturnDelivered, "packages.subpackages.status"}
	queryPathStatesMap[ReturnDeliveryFailedFilter] = FilterQueryState{states.ReturnDeliveryFailed, "packages.subpackages.tracking.history.name"}
	queryPathStatesMap[ReturnRejectedFilter] = FilterQueryState{states.ReturnRejected, "packages.subpackages.status"}
	queryPathStatesMap[PayToBuyerFilter] = FilterQueryState{states.PayToBuyer, "packages.subpackages.status"}
	queryPathStatesMap[PayToSellerFilter] = FilterQueryState{states.PayToSeller, "packages.subpackages.status"}

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
		scheduler_action.New(scheduler_action.PaymentFail),
		scheduler_action.New(scheduler_action.DeliveryDelay),
		scheduler_action.New(scheduler_action.Deliver),
		scheduler_action.New(scheduler_action.DeliveryPending),
		scheduler_action.New(scheduler_action.Reject),
		scheduler_action.New(scheduler_action.Accept),
		scheduler_action.New(scheduler_action.Notification),
	}

	reqFilters := make(map[RequestName][]FilterValue, 8)
	//reqFilters[SellerAllOrders] = []FilterValue{
	//	ApprovalPendingFilter,
	//	CanceledBySellerFilter,
	//	CanceledByBuyerFilter,
	//	ShipmentPendingFilter,
	//	ShipmentDelayedFilter,
	//	ShippedFilter,
	//	DeliveryPendingFilter,
	//	DeliveredFilter,
	//	DeliveryFailedFilter,
	//	ReturnRequestPendingFilter,
	//	ReturnRequestRejectedFilter,
	//	ReturnShipmentPendingFilter,
	//	ReturnShippedFilter,
	//	ReturnDeliveryPendingFilter,
	//	ReturnDeliveryDelayedFilter,
	//	ReturnDeliveredFilter,
	//	ReturnDeliveryFailedFilter,
	//	PayToSellerFilter,
	//}

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
		ReturnCanceledFilter,
		ReturnShipmentPendingFilter,
		ReturnShippedFilter,
		ReturnDeliveryPendingFilter,
		ReturnDeliveryDelayedFilter,
		ReturnDeliveredFilter,
		ReturnDeliveryFailedFilter,
		ReturnRejectedFilter,
	}

	reqFilters[SellerOrderDashboardReports] = []FilterValue{}
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
		ReturnCanceledFilter,
		ReturnShipmentPendingFilter,
		ReturnShippedFilter,
		ReturnDeliveryPendingFilter,
		ReturnDeliveryDelayedFilter,
		ReturnDeliveredFilter,
		ReturnDeliveryFailedFilter,
		ReturnRejectedFilter,
	}

	return Server{
		flowManager: flowManager, address: address, port: port,
		requestFilters:       reqFilters,
		buyerFilterStates:    buyerStatesMap,
		sellerFilterStates:   sellerFilterStatesMap,
		operatorFilterStates: operatorFilterStatesMap,
		queryPathStates:      queryPathStatesMap,
		actionStates:         actionStateMap,
	}
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

func (server *Server) SchedulerMessageHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {

	if ctx.Value(string(utils.CtxUserID)) == nil {
		ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(0))
	}

	userType := SchedulerUser
	var userAction actions.IAction

	var schedulerActionRequest pb.SchedulerActionRequest
	if err := ptypes.UnmarshalAny(req.Data, &schedulerActionRequest); err != nil {
		logger.Err("Could not unmarshal schedulerActionRequest from request anything field, request: %v, error %s", req, err)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	for _, orderReq := range schedulerActionRequest.Orders {
		userActions, ok := server.actionStates[userType]
		if !ok {
			logger.Err("SchedulerMessageHandler() => action %s user not supported, request: %v", userType, req)
			return nil, status.Error(codes.Code(future.BadRequest), "User Action Invalid")
		}

		for _, action := range userActions {
			if action.ActionEnum().ActionName() == orderReq.ActionState {
				userAction = action
				break
			}
		}

		if userAction == nil {
			logger.Err("SchedulerMessageHandler() => %s action invalid, request: %v", req.Meta.Action.ActionState, req)
			return nil, status.Error(codes.Code(future.BadRequest), "Action Invalid")
		}

		if userAction.ActionEnum() == scheduler_action.PaymentFail {
			event := events.New(events.Action, orderReq.OID, 0, 0,
				orderReq.StateIndex, userAction,
				time.Unix(req.Time.GetSeconds(), int64(req.Time.GetNanos())), nil)

			iFuture := future.Factory().SetCapacity(1).Build()
			iFrame := frame.Factory().SetFuture(iFuture).SetEvent(event).Build()
			server.flowManager.MessageHandler(ctx, iFrame)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				logger.Err("SchedulerMessageHandler() => flowManager.MessageHandler failed, event: %v, error: %s", event, futureData.Error().Reason())
			}

			response := &pb.MessageResponse{
				Entity: "ActionResponse",
				Meta:   nil,
				Data:   nil,
			}
			return response, nil
		}

		for _, pkgReq := range orderReq.Packages {
			subpackages := make([]events.ActionSubpackage, 0, len(pkgReq.Subpackages))
			for _, subPkgReq := range pkgReq.Subpackages {

				subpackage := events.ActionSubpackage{
					SId:   subPkgReq.SID,
					Items: nil,
				}
				subpackage.Items = make([]events.ActionItem, 0, len(subPkgReq.Items))
				for _, item := range subPkgReq.Items {
					actionItem := events.ActionItem{
						InventoryId: item.InventoryId,
						Quantity:    item.Quantity,
					}
					subpackage.Items = append(subpackage.Items, actionItem)
				}
				subpackages = append(subpackages, subpackage)
			}

			actionData := events.ActionData{
				SubPackages:    subpackages,
				Carrier:        "",
				TrackingNumber: "",
			}

			event := events.New(events.Action, orderReq.OID, pkgReq.PID, 0,
				orderReq.StateIndex, userAction,
				time.Unix(req.Time.GetSeconds(), int64(req.Time.GetNanos())), actionData)

			iFuture := future.Factory().SetCapacity(1).Build()
			iFrame := frame.Factory().SetFuture(iFuture).SetEvent(event).Build()
			server.flowManager.MessageHandler(ctx, iFrame)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				logger.Err("SchedulerMessageHandler() => flowManager.MessageHandler failed, event: %v, error: %s", event, futureData.Error().Reason())
			}
		}
	}

	response := &pb.MessageResponse{
		Entity: "ActionResponse",
		Meta:   nil,
		Data:   nil,
	}
	return response, nil
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
		reqName != SellerOrderDashboardReports &&
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
		reqName != SellerOrderDashboardReports &&
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

	case SellerOrderDashboardReports:
		return server.sellerOrderDashboardReportsHandler(ctx, req.Meta.UID)
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

	logger.Audit("requestActionHandler() => received request action: %v", req)

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
		logger.Err("requestActionHandler() => Could not unmarshal reqActionData from request anything field, request: %v, error %s", req, err)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	subpackages := make([]events.ActionSubpackage, 0, len(reqActionData.Subpackages))
	for _, reqSubpackage := range reqActionData.Subpackages {
		subpackage := events.ActionSubpackage{
			SId: reqSubpackage.SID,
		}
		subpackage.Items = make([]events.ActionItem, 0, len(reqSubpackage.Items))
		for _, item := range reqSubpackage.Items {

			if item.Quantity <= 0 {
				logger.Err("requestActionHandler() => %s action invalid, request: %v", req.Meta.Action.ActionState, req)
				return nil, status.Error(codes.Code(future.BadRequest), "Action Quantity Invalid")
			}

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

func (server *Server) PaymentGatewayHook(ctx context.Context, req *pg.PaygateHookRequest) (*pg.PaygateHookResponse, error) {

	logger.Audit("PaymentGatewayHook() => received payment response: orderId: %s, PaymentId: %s, InvoiceId: %d, result: %v",
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
		logger.Err("NewOrder() => UserService.AuthenticateContextToken failed, error: %s ", err)
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

	//if futureData.Error() != nil {
	//	futureErr := futureData.Error()
	//	return nil, status.Error(codes.Code(futureErr.Code()), futureErr.Message())
	//}

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
//				System: &pb.System{
//					Amount: order.Invoice.System.Amount,
//					Code:   order.Invoice.System.Code,
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
	port := strconv.Itoa(int(server.port))
	lis, err := net.Listen("tcp", server.address+":"+port)
	if err != nil {
		logger.Err("Failed to listen to TCP on port " + port + err.Error())
	}
	logger.Audit("app started at %s:%s", server.address, port)

	customFunc := func(p interface{}) (err error) {
		logger.Err("rpc panic recovered, panic: %v, stacktrace: %v", p, string(debug.Stack()))
		return grpc.Errorf(codes.Unknown, "panic triggered: %v", p)
	}

	//zapLogger, _ := zap.NewProduction()
	//stackDisableOpt := zap.AddStacktrace(stackTraceDisabler{})
	//noStackLogger := app.Globals.ZapLogger.WithOptions(stackDisableOpt)

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}

	uIntOpt := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_prometheus.UnaryServerInterceptor,
		grpc_recovery.UnaryServerInterceptor(opts...),
		myUnaryLogger(app.Globals.Logger),
		//grpc_zap.UnaryServerInterceptor(zapLogger),
	))

	sIntOpt := grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		grpc_prometheus.StreamServerInterceptor,
		grpc_recovery.StreamServerInterceptor(opts...),
		//grpc_zap.StreamServerInterceptor(app.Globals.ZapLogger),
	))

	// enable grpc prometheus interceptors to log timing info for grpc APIs
	grpc_prometheus.EnableHandlingTimeHistogram()

	//Start GRPC server and register the server
	grpcServer := grpc.NewServer(uIntOpt, sIntOpt)
	pb.RegisterOrderServiceServer(grpcServer, &server)
	pg.RegisterBankResultHookServer(grpcServer, &server)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Err("GRPC server start field " + err.Error())
		panic("GRPC server start field")
	}
}

func (server Server) StartTest() {
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
}

func myUnaryLogger(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		resp, err = handler(ctx, req)
		dur := time.Since(startTime)
		lg := log.FromContext(ctx)
		lg = lg.With(
			zap.Duration("took_sec", dur),
			zap.String("grpc.Method", path.Base(info.FullMethod)),
			zap.String("grpc.Service", path.Dir(info.FullMethod)[1:]),
			zap.String("grpc.Code", grpc.Code(err).String()),
		)
		lg.Debug("finished unary call")
		return
	}
}
