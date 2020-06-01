package grpc_server

import (
	"gitlab.faza.io/order-project/order-service/domain/states"
	"go.uber.org/zap/zapcore"
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

	AllOrdersFilter             FilterValue = "AllOrders"
	AllCanceledFilter           FilterValue = "AllCanceled"
	DashboardReportFilter       FilterValue = "DashboardReport"
	ShipmentReportFilter        FilterValue = "ShipmentReport"
	ReturnReportFilter          FilterValue = "ReturnReport"
	DeliveredReportFilter       FilterValue = "DeliveredReport"
	CanceledReportFilter        FilterValue = "CanceledReport"
	AllReportFilter             FilterValue = "AllReport"
	ApprovalPendingReportFilter FilterValue = "ApprovalPendingReport"
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
	//SellerAllOrders             		RequestName = "SellerAllOrders"
	SellerOrderList                   RequestName = "SellerOrderList"
	SellerOrderDetail                 RequestName = "SellerOrderDetail"
	SellerReturnOrderList             RequestName = "SellerReturnOrderList"
	SellerReturnOrderDetail           RequestName = "SellerReturnOrderDetail"
	SellerOrderDashboardReports       RequestName = "SellerOrderDashboardReports"
	SellerOrderShipmentReports        RequestName = "SellerOrderShipmentReports"
	SellerOrderDeliveredReports       RequestName = "SellerOrderDeliveredReports"
	SellerOrderReturnReports          RequestName = "SellerOrderReturnReports"
	SellerOrderCancelReports          RequestName = "SellerOrderCancelReports"
	SellerAllOrderReports             RequestName = "SellerAllOrderReports"
	SellerApprovalPendingOrderReports RequestName = "SellerApprovalPendingOrderReports"

	//BuyerAllOrders			   RequestName = "BuyerAllOrders"
	//BuyerAllReturnOrders       RequestName = "BuyerAllReturnOrders"
	BuyerOrderDetailList       RequestName = "BuyerOrderDetailList"
	BuyerAllOrderReports       RequestName = "BuyerAllOrderReports"
	BuyerReturnOrderReports    RequestName = "BuyerReturnOrderReports"
	BuyerReturnOrderDetailList RequestName = "BuyerReturnOrderDetailList"

	//OperatorAllOrders	RequestName = "OperatorAllOrders"
	OperatorOrderList          RequestName = "OperatorOrderList"
	OperatorOrderDetail        RequestName = "OperatorOrderDetail"
	OperatorOrderInvoiceDetail RequestName = "OperatorOrderInvoiceDetail"
)

const (
	ASC  SortDirection = "ASC"
	DESC SortDirection = "DESC"
)

type stackTraceDisabler struct{}

func (s stackTraceDisabler) Enabled(zapcore.Level) bool {
	return false
}

type FilterState struct {
	expectedState []states.IEnumState
	actualState   states.IEnumState
}

type FilterQueryState struct {
	state     states.IEnumState
	queryPath string
}
