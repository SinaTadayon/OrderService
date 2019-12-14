package payment_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
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

func (payment *iPaymentServiceImpl) ConnectToPaymentService() error {
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

	if err := payment.ConnectToPaymentService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "ConnectToPaymentService failed")).
			BuildAndSend()
	}

	ctx, _ = context.WithTimeout(ctx, 30*time.Second)
	gatewayRequest := &payment_gateway.GenerateRedirRequest{
		Gateway:  request.Gateway,
		Amount:   request.Amount,
		Currency: request.Currency,
		OrderID:  strconv.Itoa(int(request.OrderId)),
	}

	// TODO decode err code
	response, err := payment.paymentService.GenerateRedirectURL(ctx, gatewayRequest)
	if err != nil {
		logger.Err("request to payment gateway grpc failed, orderId: %d, amount: %d, gateway: %s, currency: %s, error: %s",
			request.OrderId, request.Amount, request.Gateway, request.Currency, err)

		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GenerateRedirectURL failed")).
			BuildAndSend()
	}

	paymentResponse := PaymentResponse{
		CallbackUrl: response.CallbackUrl,
		InvoiceId:   response.InvoiceId,
		PaymentId:   response.PaymentId,
	}

	return future.Factory().SetCapacity(1).
		SetData(paymentResponse).
		BuildAndSend()
}
