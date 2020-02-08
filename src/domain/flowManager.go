package domain

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
)

type IFlowManager interface {
	MessageHandler(ctx context.Context, iFrame frame.IFrame)
	PaymentGatewayResult(ctx context.Context, req *pg.PaygateHookRequest) future.IFuture
	GetState(state states.IEnumState) states.IState
	ReportOrderItems(ctx context.Context, req *pb.RequestReportOrderItems, srv pb.OrderService_ReportOrderItemsServer) future.IFuture
}
