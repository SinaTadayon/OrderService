package domain

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	message "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
)

type IFlowManager interface {
	MessageHandler(ctx context.Context, req *message.MessageRequest) future.IFuture
	SellerApprovalPending(ctx context.Context, req *message.RequestSellerOrderAction) future.IFuture
	BuyerApprovalPending(ctx context.Context, req *message.RequestBuyerOrderAction) future.IFuture
	PaymentGatewayResult(ctx context.Context, req *pg.PaygateHookRequest) future.IFuture
	OperatorActionPending(ctx context.Context, req *message.RequestBackOfficeOrderAction) future.IFuture

	BackOfficeOrdersListView(ctx context.Context, req *message.RequestBackOfficeOrdersList) future.IFuture
	BackOfficeOrderDetailView(ctx context.Context, req *message.RequestIdentifier) future.IFuture
	SellerReportOrders(req *message.RequestSellerReportOrders, srv message.OrderService_SellerReportOrdersServer) future.IFuture
	BackOfficeReportOrderItems(req *message.RequestBackOfficeReportOrderItems, srv message.OrderService_BackOfficeReportOrderItemsServer) future.IFuture
	SchedulerEvents(event events.ISchedulerEvent)
}
