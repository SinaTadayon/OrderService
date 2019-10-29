package payment

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type iPaymentService struct {

}

func NewPaymentService() IPaymentService {
	return &iPaymentService{}
}

func (payment iPaymentService) OrderPayment(context context.Context, amount int64, gateway, currency, orderId string) promise.IPromise {
	panic("must be implement")
}
