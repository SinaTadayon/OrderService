package payment_service

import (
	"context"
	"fmt"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	payment_gateway "gitlab.faza.io/protos/payment-gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"strconv"
	"time"
)

type iPaymentServiceImpl struct {
	paymentService payment_gateway.PaymentGatewayClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
}

func NewPaymentService(address string, port int) IPaymentService {
	return &iPaymentServiceImpl{nil, nil, address, port}
}

func (payment *iPaymentServiceImpl) ConnectToStockService() error {
	if payment.grpcConnection == nil || payment.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		payment.grpcConnection, err = grpc.DialContext(ctx, payment.serverAddress+":"+fmt.Sprint(payment.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			logger.Err("GRPC connect dial to payment service failed, err: %s", err.Error())
			return err
		}
		payment.paymentService = payment_gateway.NewPaymentGatewayClient(payment.grpcConnection)
	}
	return nil
}

// TODO checking return error of payment
func (payment iPaymentServiceImpl) OrderPayment(ctx context.Context, request PaymentRequest) future.IFuture {

	if err := payment.ConnectToStockService(); err != nil {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{
			Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}

	ctx1, _ := context.WithCancel(context.Background())
	gatewayRequest := &payment_gateway.GenerateRedirRequest{
		Gateway:  request.Gateway,
		Amount:   request.Amount,
		Currency: request.Currency,
		OrderID:  strconv.Itoa(int(request.OrderId)),
	}

	response, err := payment.paymentService.GenerateRedirectURL(ctx1, gatewayRequest)
	if err != nil {
		logger.Err("request to payment gateway grpc failed, orderId: %d, amount: %d, gateway: %s, currency: %s, error: %s",
			request.OrderId, request.Amount, request.Gateway, request.Currency, err)
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}

	paymentResponse := PaymentResponse{
		CallbackUrl: response.CallbackUrl,
		InvoiceId:   response.InvoiceId,
		PaymentId:   response.PaymentId,
	}

	returnChannel := make(chan future.IDataFuture, 1)
	defer close(returnChannel)
	returnChannel <- future.IDataFuture{Data: paymentResponse, Ex: nil}
	return future.NewFuture(returnChannel, 1, 1)
}
