package payment_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type iPaymentServiceMock struct {
}

func NewPaymentServiceMock() IPaymentService {
	return &iPaymentServiceMock{}
}

func (payment iPaymentServiceMock) OrderPayment(ctx context.Context, request PaymentRequest) promise.IPromise {
	paymentResponse := PaymentResponse {
		CallbackUrl: "http://assanpardakht.com/",
		InvoiceId: 43464645465345,
		PaymentId: "12345667788",
	}

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data:paymentResponse, Ex:nil}
	return promise.NewPromise(returnChannel, 1, 1)
}
