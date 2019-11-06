package domain

import (
	"context"
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
}
