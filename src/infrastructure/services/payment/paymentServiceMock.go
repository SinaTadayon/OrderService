package payment_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type iPaymentServiceMock struct {
}

func NewPaymentServiceMock() IPaymentService {
	return &iPaymentServiceMock{}
}

func (payment iPaymentServiceMock) OrderPayment(ctx context.Context, request PaymentRequest) future.IFuture {
	paymentResponse := PaymentResponse{
		CallbackUrl: "http://staging.faza.io/callback-success",
		InvoiceId:   43464645465345,
		PaymentId:   "12345667788",
	}

	return future.Factory().SetCapacity(1).SetData(paymentResponse).BuildAndSend()
}
