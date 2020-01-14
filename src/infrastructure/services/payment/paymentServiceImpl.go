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
		Mobile:   request.Mobile,
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

	logger.Audit("OrderPayment() => request: %v, response: %v", request, response)

	paymentResponse := PaymentResponse{
		CallbackUrl: response.CallbackUrl,
		InvoiceId:   response.InvoiceId,
		PaymentId:   response.PaymentId,
	}

	return future.Factory().SetCapacity(1).
		SetData(paymentResponse).
		BuildAndSend()
}

func (payment iPaymentServiceImpl) GetPaymentResult(ctx context.Context, orderId uint64) future.IFuture {
	if err := payment.ConnectToPaymentService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "ConnectToPaymentService failed")).
			BuildAndSend()
	}

	ctx, _ = context.WithTimeout(ctx, 5*time.Second)
	payRequest := &payment_gateway.GetPaymentResultByOrderIdRequest{
		OrderID: strconv.Itoa(int(orderId)),
	}

	// TODO decode err code
	response, err := payment.paymentService.GetPaymentResultByOrderID(ctx, payRequest)
	if err != nil {
		logger.Err("request to GetPaymentResultByOrderID grpc failed, orderId: %d, error: %s",
			orderId, err)

		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GenerateRedirectURL failed")).
			BuildAndSend()
	}

	logger.Audit("GetPaymentResult() => orderId: %d, response: %v", orderId, response)

	oid, err := strconv.Atoi(response.OrderID)
	if err != nil {
		logger.Err("request to GetPaymentResultByOrderID grpc failed, orderId: %d, error: %s",
			orderId, err)

		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GenerateRedirectURL failed")).
			BuildAndSend()
	}

	payQueryResult := PaymentQueryResult{
		OrderId:   uint64(oid),
		PaymentId: response.PaymentId,
		InvoiceId: response.InvoiceId,
		Amount:    response.Amount,
		CardMask:  response.CardMask,
		Status:    PaymentRequestStatus(response.Status),
	}

	return future.Factory().SetCapacity(1).
		SetData(payQueryResult).
		BuildAndSend()
}
