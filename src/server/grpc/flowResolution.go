//
// The flow resolution file contains mapping of states,
// which will be used for accurate resolution of responses
// and requests base on inputs
package grpc_server

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/buyer"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	"gitlab.faza.io/order-project/order-service/domain/states"
)

func initialOperatorFilterStatesMap() map[FilterValue][]FilterState {
	operatorFilterStatesMap := make(map[FilterValue][]FilterState, 30)
	operatorFilterStatesMap[NewOrderFilter] = []FilterState{{[]states.IEnumState{states.NewOrder}, states.NewOrder}}
	operatorFilterStatesMap[PaymentPendingFilter] = []FilterState{{[]states.IEnumState{states.PaymentPending}, states.PaymentPending}}
	operatorFilterStatesMap[PaymentSuccessFilter] = []FilterState{{[]states.IEnumState{states.PaymentSuccess}, states.ApprovalPending}}
	operatorFilterStatesMap[PaymentFailedFilter] = []FilterState{{[]states.IEnumState{states.PaymentFailed}, states.PaymentFailed}}
	operatorFilterStatesMap[OrderVerificationPendingFilter] = []FilterState{{[]states.IEnumState{states.OrderVerificationPending}, states.ApprovalPending}}
	operatorFilterStatesMap[OrderVerificationSuccessFilter] = []FilterState{{[]states.IEnumState{states.OrderVerificationSuccess}, states.ApprovalPending}}
	operatorFilterStatesMap[OrderVerificationFailedFilter] = []FilterState{{[]states.IEnumState{states.OrderVerificationFailed}, states.PayToBuyer}}
	operatorFilterStatesMap[ApprovalPendingFilter] = []FilterState{{[]states.IEnumState{states.ApprovalPending}, states.ApprovalPending}}
	operatorFilterStatesMap[CanceledBySellerFilter] = []FilterState{{[]states.IEnumState{states.CanceledBySeller}, states.PayToBuyer}}
	operatorFilterStatesMap[CanceledByBuyerFilter] = []FilterState{{[]states.IEnumState{states.CanceledByBuyer}, states.PayToBuyer}}
	operatorFilterStatesMap[ShipmentPendingFilter] = []FilterState{{[]states.IEnumState{states.ShipmentPending}, states.ShipmentPending}}
	operatorFilterStatesMap[ShipmentDelayedFilter] = []FilterState{{[]states.IEnumState{states.ShipmentDelayed}, states.ShipmentDelayed}}
	operatorFilterStatesMap[ShippedFilter] = []FilterState{{[]states.IEnumState{states.Shipped}, states.Shipped}}
	operatorFilterStatesMap[DeliveryPendingFilter] = []FilterState{{[]states.IEnumState{states.DeliveryPending}, states.DeliveryPending}}
	operatorFilterStatesMap[DeliveryDelayedFilter] = []FilterState{{[]states.IEnumState{states.DeliveryDelayed}, states.DeliveryDelayed}}
	operatorFilterStatesMap[DeliveredFilter] = []FilterState{{[]states.IEnumState{states.Delivered}, states.Delivered}}
	operatorFilterStatesMap[DeliveryFailedFilter] = []FilterState{{[]states.IEnumState{states.DeliveryFailed}, states.PayToBuyer}}
	operatorFilterStatesMap[ReturnRequestPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnRequestPending}, states.ReturnRequestPending}}
	operatorFilterStatesMap[ReturnRequestRejectedFilter] = []FilterState{{[]states.IEnumState{states.ReturnRequestRejected}, states.ReturnRequestRejected}}
	operatorFilterStatesMap[ReturnCanceledFilter] = []FilterState{{[]states.IEnumState{states.ReturnCanceled}, states.PayToSeller}}
	operatorFilterStatesMap[ReturnShipmentPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnShipmentPending}, states.ReturnShipmentPending}}
	operatorFilterStatesMap[ReturnShippedFilter] = []FilterState{{[]states.IEnumState{states.ReturnShipped}, states.ReturnShipped}}
	operatorFilterStatesMap[ReturnDeliveryPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryPending}, states.ReturnDeliveryPending}}
	operatorFilterStatesMap[ReturnDeliveryDelayedFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryDelayed}, states.ReturnDeliveryDelayed}}
	operatorFilterStatesMap[ReturnDeliveredFilter] = []FilterState{{[]states.IEnumState{states.ReturnDelivered}, states.ReturnDelivered}}
	operatorFilterStatesMap[ReturnDeliveryFailedFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryFailed}, states.PayToSeller}}
	operatorFilterStatesMap[ReturnRejectedFilter] = []FilterState{{[]states.IEnumState{states.ReturnRejected}, states.ReturnRejected}}
	operatorFilterStatesMap[PayToBuyerFilter] = []FilterState{{[]states.IEnumState{states.PayToBuyer}, states.PayToBuyer}}
	operatorFilterStatesMap[PayToSellerFilter] = []FilterState{{[]states.IEnumState{states.PayToSeller}, states.PayToSeller}}

	return operatorFilterStatesMap
}

func initialSellerStatesMapping() map[string][]states.IEnumState {
	sellerStatesMapping := make(map[string][]states.IEnumState, 30)
	sellerStatesMapping[states.ApprovalPending.StateName()] = []states.IEnumState{states.ApprovalPending}
	sellerStatesMapping[states.ShipmentPending.StateName()] = []states.IEnumState{states.ShipmentPending}
	sellerStatesMapping[states.ShipmentDelayed.StateName()] = []states.IEnumState{states.ShipmentDelayed}
	sellerStatesMapping[states.Shipped.StateName()] = []states.IEnumState{states.Shipped}
	sellerStatesMapping[states.DeliveryPending.StateName()] = []states.IEnumState{states.DeliveryPending}
	sellerStatesMapping[states.DeliveryDelayed.StateName()] = []states.IEnumState{states.DeliveryDelayed}
	sellerStatesMapping[states.Delivered.StateName()] = []states.IEnumState{states.Delivered}
	sellerStatesMapping[states.ReturnRequestPending.StateName()] = []states.IEnumState{states.ReturnRequestPending}
	sellerStatesMapping[states.ReturnRequestRejected.StateName()] = []states.IEnumState{states.ReturnRequestRejected}
	sellerStatesMapping[states.ReturnShipmentPending.StateName()] = []states.IEnumState{states.ReturnShipmentPending}
	sellerStatesMapping[states.ReturnShipped.StateName()] = []states.IEnumState{states.ReturnShipped}
	sellerStatesMapping[states.ReturnDeliveryPending.StateName()] = []states.IEnumState{states.ReturnDeliveryPending}
	sellerStatesMapping[states.ReturnDeliveryDelayed.StateName()] = []states.IEnumState{states.ReturnDeliveryDelayed}
	sellerStatesMapping[states.ReturnDelivered.StateName()] = []states.IEnumState{states.ReturnDelivered}
	sellerStatesMapping[states.ReturnRejected.StateName()] = []states.IEnumState{states.ReturnRejected}
	sellerStatesMapping[states.PayToSeller.StateName()] = []states.IEnumState{states.ReturnCanceled, states.ReturnDeliveryFailed, states.ReturnShipmentPending, states.ReturnRequestRejected, states.ReturnRejected, states.Delivered}
	sellerStatesMapping[states.PayToBuyer.StateName()] = []states.IEnumState{states.CanceledBySeller, states.CanceledByBuyer, states.DeliveryFailed, states.ReturnRejected, states.ReturnDelivered}

	return sellerStatesMapping
}

func initialSellerFilterStatesMap() map[FilterValue][]FilterState {
	sellerFilterStatesMap := make(map[FilterValue][]FilterState, 30)
	sellerFilterStatesMap[ApprovalPendingFilter] = []FilterState{{[]states.IEnumState{states.ApprovalPending}, states.ApprovalPending}}
	sellerFilterStatesMap[CanceledBySellerFilter] = []FilterState{{[]states.IEnumState{states.CanceledBySeller}, states.PayToBuyer}}
	sellerFilterStatesMap[CanceledByBuyerFilter] = []FilterState{{[]states.IEnumState{states.CanceledByBuyer}, states.PayToBuyer}}
	sellerFilterStatesMap[ShipmentPendingFilter] = []FilterState{{[]states.IEnumState{states.ShipmentPending}, states.ShipmentPending}}
	sellerFilterStatesMap[ShipmentDelayedFilter] = []FilterState{{[]states.IEnumState{states.ShipmentDelayed}, states.ShipmentDelayed}}
	sellerFilterStatesMap[ShippedFilter] = []FilterState{{[]states.IEnumState{states.Shipped}, states.Shipped}}
	sellerFilterStatesMap[DeliveryPendingFilter] = []FilterState{{[]states.IEnumState{states.DeliveryPending}, states.DeliveryPending}}
	sellerFilterStatesMap[DeliveryDelayedFilter] = []FilterState{{[]states.IEnumState{states.DeliveryDelayed}, states.DeliveryDelayed}}
	sellerFilterStatesMap[DeliveredFilter] = []FilterState{{[]states.IEnumState{states.Delivered}, states.Delivered}}
	sellerFilterStatesMap[DeliveryFailedFilter] = []FilterState{{[]states.IEnumState{states.DeliveryFailed}, states.PayToBuyer}}
	sellerFilterStatesMap[ReturnRequestPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnRequestPending}, states.ReturnRequestPending}}
	sellerFilterStatesMap[ReturnRequestRejectedFilter] = []FilterState{{[]states.IEnumState{states.ReturnRequestRejected}, states.ReturnRequestRejected}}
	sellerFilterStatesMap[ReturnCanceledFilter] = []FilterState{{[]states.IEnumState{states.ReturnCanceled}, states.PayToSeller}}
	sellerFilterStatesMap[ReturnShipmentPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnShipmentPending}, states.ReturnShipmentPending}}
	sellerFilterStatesMap[ReturnShippedFilter] = []FilterState{{[]states.IEnumState{states.ReturnShipped}, states.ReturnShipped}}
	sellerFilterStatesMap[ReturnDeliveryPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryPending}, states.ReturnDeliveryPending}}
	sellerFilterStatesMap[ReturnDeliveryDelayedFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryDelayed}, states.ReturnDeliveryDelayed}}
	sellerFilterStatesMap[ReturnDeliveredFilter] = []FilterState{{[]states.IEnumState{states.ReturnDelivered}, states.ReturnDelivered}}
	sellerFilterStatesMap[ReturnDeliveryFailedFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryFailed}, states.PayToSeller}}
	sellerFilterStatesMap[ReturnRejectedFilter] = []FilterState{{[]states.IEnumState{states.ReturnRejected}, states.ReturnRejected}}
	sellerFilterStatesMap[PayToSellerFilter] = []FilterState{{[]states.IEnumState{states.ReturnCanceled, states.ReturnDeliveryFailed, states.ReturnShipmentPending, states.ReturnRequestRejected, states.Delivered, states.ReturnRejected}, states.PayToSeller}}

	return sellerFilterStatesMap
}

func initialBuyerReturnStatesMapping() map[string][]states.IEnumState {
	buyerReturnStatesMapping := make(map[string][]states.IEnumState, 16)
	buyerReturnStatesMapping[states.ReturnRequestPending.StateName()] = []states.IEnumState{states.ReturnRequestPending}
	buyerReturnStatesMapping[states.ReturnRequestRejected.StateName()] = []states.IEnumState{states.ReturnRequestRejected}
	buyerReturnStatesMapping[states.ReturnShipmentPending.StateName()] = []states.IEnumState{states.ReturnShipmentPending}
	buyerReturnStatesMapping[states.ReturnShipped.StateName()] = []states.IEnumState{states.ReturnShipped}
	buyerReturnStatesMapping[states.ReturnDeliveryPending.StateName()] = []states.IEnumState{states.ReturnDeliveryPending}
	buyerReturnStatesMapping[states.ReturnDeliveryDelayed.StateName()] = []states.IEnumState{states.ReturnDeliveryDelayed}
	buyerReturnStatesMapping[states.ReturnDelivered.StateName()] = []states.IEnumState{states.ReturnDelivered}
	buyerReturnStatesMapping[states.ReturnRejected.StateName()] = []states.IEnumState{states.ReturnRejected}
	buyerReturnStatesMapping[states.PayToBuyer.StateName()] = []states.IEnumState{states.ReturnRejected, states.ReturnDelivered}

	return buyerReturnStatesMapping
}

func initialBuyerAllStatesMapping() map[string][]states.IEnumState {
	buyerAllStatesMapping := make(map[string][]states.IEnumState, 16)
	buyerAllStatesMapping[states.NewOrder.StateName()] = []states.IEnumState{states.NewOrder}
	buyerAllStatesMapping[states.PaymentPending.StateName()] = []states.IEnumState{states.PaymentPending}
	buyerAllStatesMapping[states.PaymentSuccess.StateName()] = []states.IEnumState{states.PaymentSuccess}
	buyerAllStatesMapping[states.PaymentFailed.StateName()] = []states.IEnumState{states.PaymentFailed}
	buyerAllStatesMapping[states.OrderVerificationPending.StateName()] = []states.IEnumState{states.OrderVerificationPending}
	buyerAllStatesMapping[states.OrderVerificationSuccess.StateName()] = []states.IEnumState{states.OrderVerificationSuccess}
	buyerAllStatesMapping[states.OrderVerificationFailed.StateName()] = []states.IEnumState{states.PayToBuyer}
	buyerAllStatesMapping[states.ApprovalPending.StateName()] = []states.IEnumState{states.ApprovalPending}
	buyerAllStatesMapping[states.ShipmentPending.StateName()] = []states.IEnumState{states.ShipmentPending}
	buyerAllStatesMapping[states.ShipmentDelayed.StateName()] = []states.IEnumState{states.ShipmentDelayed}
	buyerAllStatesMapping[states.Shipped.StateName()] = []states.IEnumState{states.Shipped}
	buyerAllStatesMapping[states.DeliveryPending.StateName()] = []states.IEnumState{states.DeliveryPending}
	buyerAllStatesMapping[states.DeliveryDelayed.StateName()] = []states.IEnumState{states.DeliveryDelayed}
	buyerAllStatesMapping[states.Delivered.StateName()] = []states.IEnumState{states.Delivered}
	buyerAllStatesMapping[states.PayToBuyer.StateName()] = []states.IEnumState{states.PayToBuyer}

	return buyerAllStatesMapping
}

func initialBuyerStatesMap() map[FilterValue][]FilterState {
	buyerStatesMap := make(map[FilterValue][]FilterState, 8)
	buyerStatesMap[NewOrderFilter] = []FilterState{{[]states.IEnumState{states.NewOrder}, states.NewOrder}}
	buyerStatesMap[PaymentPendingFilter] = []FilterState{{[]states.IEnumState{states.PaymentPending}, states.PaymentPending}}
	buyerStatesMap[PaymentSuccessFilter] = []FilterState{{[]states.IEnumState{states.PaymentSuccess}, states.ApprovalPending}}
	buyerStatesMap[PaymentFailedFilter] = []FilterState{{[]states.IEnumState{states.PaymentFailed}, states.PaymentFailed}}
	buyerStatesMap[OrderVerificationPendingFilter] = []FilterState{{[]states.IEnumState{states.OrderVerificationPending}, states.ApprovalPending}}
	buyerStatesMap[OrderVerificationSuccessFilter] = []FilterState{{[]states.IEnumState{states.OrderVerificationSuccess}, states.ApprovalPending}}
	buyerStatesMap[OrderVerificationFailedFilter] = []FilterState{{[]states.IEnumState{states.OrderVerificationFailed}, states.PayToBuyer}}
	buyerStatesMap[ApprovalPendingFilter] = []FilterState{{[]states.IEnumState{states.ApprovalPending}, states.ApprovalPending}}
	buyerStatesMap[CanceledBySellerFilter] = []FilterState{{[]states.IEnumState{states.CanceledBySeller}, states.PayToBuyer}}
	buyerStatesMap[CanceledByBuyerFilter] = []FilterState{{[]states.IEnumState{states.CanceledByBuyer}, states.PayToBuyer}}
	buyerStatesMap[ShipmentPendingFilter] = []FilterState{{[]states.IEnumState{states.ShipmentPending}, states.ShipmentPending}}
	buyerStatesMap[ShipmentDelayedFilter] = []FilterState{{[]states.IEnumState{states.ShipmentDelayed}, states.ShipmentDelayed}}
	buyerStatesMap[ShippedFilter] = []FilterState{{[]states.IEnumState{states.Shipped}, states.Shipped}}
	buyerStatesMap[DeliveredFilter] = []FilterState{{[]states.IEnumState{states.DeliveryPending}, states.DeliveryPending}, {[]states.IEnumState{states.DeliveryDelayed}, states.DeliveryDelayed}, {[]states.IEnumState{states.Delivered}, states.Delivered}}
	buyerStatesMap[DeliveryFailedFilter] = []FilterState{{[]states.IEnumState{states.DeliveryFailed}, states.PayToBuyer}}
	buyerStatesMap[ReturnRequestPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnRequestPending}, states.ReturnRequestPending}, {[]states.IEnumState{states.ReturnRequestRejected}, states.ReturnRequestRejected}}
	buyerStatesMap[ReturnShipmentPendingFilter] = []FilterState{{[]states.IEnumState{states.ReturnShipmentPending}, states.ReturnShipmentPending}}
	buyerStatesMap[ReturnShippedFilter] = []FilterState{{[]states.IEnumState{states.ReturnShipped}, states.ReturnShipped}}
	buyerStatesMap[ReturnDeliveredFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryPending}, states.ReturnDeliveryPending}, {[]states.IEnumState{states.ReturnDeliveryDelayed}, states.ReturnDeliveryDelayed}, {[]states.IEnumState{states.ReturnDelivered}, states.ReturnDelivered}}
	buyerStatesMap[ReturnDeliveryFailedFilter] = []FilterState{{[]states.IEnumState{states.ReturnDeliveryFailed}, states.PayToSeller}}

	return buyerStatesMap
}

func initialQueryStateMap() map[FilterValue]FilterQueryState {
	queryPathStatesMap := make(map[FilterValue]FilterQueryState, 30)
	queryPathStatesMap[NewOrderFilter] = FilterQueryState{states.NewOrder, "packages.subpackages.status"}
	queryPathStatesMap[PaymentPendingFilter] = FilterQueryState{states.PaymentPending, "packages.subpackages.status"}
	queryPathStatesMap[PaymentSuccessFilter] = FilterQueryState{states.PaymentSuccess, "packages.subpackages.status"}
	queryPathStatesMap[PaymentFailedFilter] = FilterQueryState{states.PaymentFailed, "packages.subpackages.status"}
	queryPathStatesMap[OrderVerificationPendingFilter] = FilterQueryState{states.OrderVerificationPending, "packages.subpackages.status"}
	queryPathStatesMap[OrderVerificationSuccessFilter] = FilterQueryState{states.OrderVerificationSuccess, "packages.subpackages.status"}
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

	return queryPathStatesMap
}

func initialActualStateMap() map[UserType][]actions.IAction {
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
		operator_action.New(operator_action.Cancel),
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

	return actionStateMap
}

func initialRequestFilters() map[RequestName][]FilterValue {
	reqFilters := make(map[RequestName][]FilterValue, 8)
	reqFilters[SellerOrderList] = []FilterValue{
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
		DeliveryDelayedFilter,
		DeliveredFilter,
		DeliveryFailedFilter,
		AllCanceledFilter,
		AllOrdersFilter,
	}

	reqFilters[SellerReturnOrderList] = []FilterValue{
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

	reqFilters[SellerReturnOrderDetail] = []FilterValue{
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
	reqFilters[SellerApprovalPendingOrderReports] = []FilterValue{}
	reqFilters[SellerAllOrderReports] = []FilterValue{}

	reqFilters[BuyerOrderDetailList] = []FilterValue{
		NewOrderFilter,
		PaymentPendingFilter,
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
		AllOrdersFilter,
	}

	//reqFilters[BuyerReturnOrderReports] = []FilterValue{}

	reqFilters[BuyerReturnOrderDetailList] = []FilterValue{
		ReturnRequestPendingFilter,
		ReturnShipmentPendingFilter,
		ReturnShippedFilter,
		ReturnDeliveredFilter,
		ReturnDeliveryFailedFilter,
		AllOrdersFilter,
	}

	//reqFilters[BuyerAllReturnOrders] = []FilterValue{
	//	ReturnRequestPendingFilter,
	//	ReturnRequestRejectedFilter,
	//	ReturnCanceledFilter,
	//	ReturnShipmentPendingFilter,
	//	ReturnShippedFilter,
	//	ReturnDeliveryPendingFilter,
	//	ReturnDeliveryDelayedFilter,
	//	ReturnDeliveredFilter,
	//	ReturnDeliveryFailedFilter,
	//	ReturnRejectedFilter,
	//}

	return reqFilters
}
