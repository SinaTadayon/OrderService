package payment_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type IPaymentService interface {
	OrderPayment(ctx context.Context, request PaymentRequest) future.IFuture
}

type PaymentRequest struct {
	Amount   int64
	Gateway  string
	Currency string
	OrderId  uint64
}

type PaymentResponse struct {
	CallbackUrl string
	InvoiceId   int64
	PaymentId   string
}

type PaymentResult struct {
	OrderId   string
	PaymentId string
	InvoiceId int64
	Amount    int64
	ReqBody   string
	ResBody   string
	CardMask  string
	Result    bool
}
