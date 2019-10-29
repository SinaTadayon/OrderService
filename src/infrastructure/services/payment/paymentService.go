package payment

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type IPaymentService interface {
	OrderPayment(context context.Context, amount int64, gateway, currency, orderId string) promise.IPromise
}

