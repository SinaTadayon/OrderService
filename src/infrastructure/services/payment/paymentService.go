package payment_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type PaymentRequestStatus int32

const (
	PaymentRequestPending PaymentRequestStatus = 0
	PaymentRequestSuccess PaymentRequestStatus = 1
	PaymentRequestFail    PaymentRequestStatus = 2
)

type IPaymentService interface {
	OrderPayment(ctx context.Context, request PaymentRequest) future.IFuture
	GetPaymentResult(ctx context.Context, orderId uint64) future.IFuture
}

type PaymentRequest struct {
	Amount   int64
	Gateway  string
	Currency string
	OrderId  uint64
	Mobile   string
}

type PaymentResponse struct {
	CallbackUrl string
	InvoiceId   int64
	PaymentId   string
}

type PaymentResult struct {
	OrderId   uint64
	PaymentId string
	InvoiceId int64
	Amount    int64
	ReqBody   string
	ResBody   string
	CardMask  string
	Result    bool
}

type PaymentQueryResult struct {
	OrderId   uint64
	PaymentId string
	InvoiceId int64
	Amount    int64
	CardMask  string
	Status    PaymentRequestStatus
}
