package payment_service

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	payment_gateway "gitlab.faza.io/protos/payment-gateway"
	"google.golang.org/grpc"
	"strconv"
	"time"
)

type iPaymentServiceImpl struct {
	serverAddress 	string
	serverPort		int
}

func NewPaymentService(address string, port int) IPaymentService {
	return &iPaymentServiceImpl{address, port}
}

// TODO checking return error of payment
func (payment iPaymentServiceImpl) OrderPayment(ctx context.Context, request PaymentRequest) promise.IPromise {
	ctx1 , _ := context.WithTimeout(context.Background(), 3 * time.Second)

	gatewayRequest := &payment_gateway.GenerateRedirRequest{
		Gateway:              request.Gateway,
		Amount:               request.Amount,
		Currency:             request.Currency,
		OrderID:              request.OrderId,
	}

	grpcConnPayment, err := grpc.DialContext(ctx1, payment.serverAddress + ":" +
		strconv.Itoa(int(payment.serverPort)), grpc.WithInsecure())

	if err != nil {
		logger.Err("connect to payment gateway grpc failed, request: %v, error: %s", request, err)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	ctx2 , _ := context.WithTimeout(ctx, 30 * time.Second)

	paymentService := payment_gateway.NewPaymentGatewayClient(grpcConnPayment)
	response, err := paymentService.GenerateRedirectURL(ctx2, gatewayRequest)
	if err != nil {
		logger.Err("request to payment gateway grpc failed, request: %v, error: %s", request, err)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	paymentResponse := PaymentResponse {
		CallbackUrl: response.CallbackUrl,
		InvoiceId: response.InvoiceId,
		PaymentId: response.PaymentId,
	}

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data:paymentResponse, Ex:nil}
	return promise.NewPromise(returnChannel, 1, 1)
}
