package domain

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
)

type IFlowManager interface {
	MessageHandler(ctx context.Context, req *message.MessageRequest) promise.IPromise
	SellerApprovalPending(ctx context.Context, req *message.RequestSellerOrderAction) promise.IPromise
	BuyerApprovalPending(ctx context.Context, req *message.RequestBuyerOrderAction) promise.IPromise
	PaymentGatewayResult(ctx context.Context, req *pg.PaygateHookRequest) promise.IPromise
	OperatorActionPending(ctx context.Context, req *message.RequestBackOfficeOrderAction) promise.IPromise

	BackOfficeOrdersListView(ctx context.Context, req *message.RequestBackOfficeOrdersList) promise.IPromise
	BackOfficeOrderDetailView(ctx context.Context, req *message.RequestIdentifier) promise.IPromise
	SellerReportOrders(req *message.RequestSellerReportOrders, srv message.OrderService_SellerReportOrdersServer) promise.IPromise
	BackOfficeReportOrderItems(req *message.RequestBackOfficeReportOrderItems, srv message.OrderService_BackOfficeReportOrderItemsServer) promise.IPromise
	SchedulerEvents(event events.SchedulerEvent)
}
