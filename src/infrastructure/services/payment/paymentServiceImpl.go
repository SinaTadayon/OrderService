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
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

type iPaymentServiceImpl struct {
	paymentService payment_gateway.PaymentGatewayClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
	timeout        int
}

func NewPaymentService(address string, port int, timeout int) IPaymentService {
	return &iPaymentServiceImpl{nil, nil, address,
		port, timeout,
	}
}

func (payment *iPaymentServiceImpl) ConnectToPaymentService() error {
	if payment.grpcConnection == nil || payment.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		payment.grpcConnection, err = grpc.DialContext(ctx, payment.serverAddress+":"+fmt.Sprint(payment.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			logger.Err("ConnectToPaymentService() => GRPC connect dial to payment service failed, err: %s", err.Error())
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

	var outCtx context.Context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		outCtx = metadata.NewOutgoingContext(ctx, md)
	} else {
		outCtx = metadata.NewOutgoingContext(ctx, metadata.New(nil))
	}

	timeoutTimer := time.NewTimer(time.Duration(payment.timeout) * time.Second)

	gatewayRequest := &payment_gateway.GenerateRedirRequest{
		Gateway:  request.Gateway,
		Amount:   request.Amount,
		Currency: request.Currency,
		OrderID:  strconv.Itoa(int(request.OrderId)),
		Mobile:   request.Mobile,
	}

	paymentFn := func() <-chan interface{} {
		paymentChan := make(chan interface{}, 0)
		go func() {
			result, err := payment.paymentService.GenerateRedirectURL(outCtx, gatewayRequest)
			if err != nil {
				paymentChan <- err
			} else {
				paymentChan <- result
			}
		}()
		return paymentChan
	}

	var obj interface{} = nil
	select {
	case obj = <-paymentFn():
		timeoutTimer.Stop()
		break
	case <-timeoutTimer.C:
		logger.Err("OrderPayment() => request to payment gateway grpc timeout, orderId: %d, amount: %d, gateway: %s, currency: %s",
			request.OrderId, request.Amount, request.Gateway, request.Currency)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.New("GenerateRedirectURL Timeout")).
			BuildAndSend()
	}

	if err, ok := obj.(error); ok {
		if err != nil {
			logger.Err("OrderPayment() => request to payment gateway grpc failed, orderId: %d, amount: %d, gateway: %s, currency: %s, error: %s",
				request.OrderId, request.Amount, request.Gateway, request.Currency, err)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GenerateRedirectURL failed")).
				BuildAndSend()
		}
	} else if response, ok := obj.(*payment_gateway.GenerateRedirResponse); ok {
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

	logger.Err("OrderPayment() => request to payment gateway grpc failed, response invalid, orderId: %d, amount: %d, gateway: %s, currency: %s, response: %v",
		request.OrderId, request.Amount, request.Gateway, request.Currency, obj)
	return future.Factory().SetCapacity(1).
		SetError(future.InternalError, "Unknown Error", errors.New("GenerateRedirectURL failed")).
		BuildAndSend()
}

func (payment iPaymentServiceImpl) GetPaymentResult(ctx context.Context, orderId uint64) future.IFuture {
	if err := payment.ConnectToPaymentService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "ConnectToPaymentService failed")).
			BuildAndSend()
	}

	var outCtx context.Context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		outCtx = metadata.NewOutgoingContext(ctx, md)
	} else {
		outCtx = metadata.NewOutgoingContext(ctx, metadata.New(nil))
	}

	timeoutTimer := time.NewTimer(time.Duration(payment.timeout) * time.Second)

	payRequest := &payment_gateway.GetPaymentResultByOrderIdRequest{
		OrderID: strconv.Itoa(int(orderId)),
	}

	paymentFn := func() <-chan interface{} {
		paymentChan := make(chan interface{}, 0)
		go func() {
			result, err := payment.paymentService.GetPaymentResultByOrderID(outCtx, payRequest)
			if err != nil {
				paymentChan <- err
			} else {
				paymentChan <- result
			}
		}()
		return paymentChan
	}

	var obj interface{} = nil
	select {
	case obj = <-paymentFn():
		timeoutTimer.Stop()
		break
	case <-timeoutTimer.C:
		logger.Err("GetPaymentResult() => request to GetPaymentResultByOrderID grpc timeout, orderId: %d", orderId)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.New("GetPaymentResultByOrderID Timeout")).
			BuildAndSend()
	}

	// TODO decode err code
	if err, ok := obj.(error); ok {
		if err != nil {
			logger.Err("GetPaymentResult() => request to GetPaymentResultByOrderID grpc failed, orderId: %d, error: %s",
				orderId, err)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GetPaymentResultByOrderID failed")).
				BuildAndSend()
		}
	} else if response, ok := obj.(*payment_gateway.PaymentRequest); ok {
		logger.Audit("GetPaymentResult() => orderId: %d, response: %v", orderId, response)

		oid, err := strconv.Atoi(response.OrderID)
		if err != nil {
			logger.Err("GetPaymentResult() => request to GetPaymentResultByOrderID grpc failed, orderId: %d, error: %s",
				orderId, err)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GetPaymentResultByOrderID failed")).
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

	logger.Err("GetPaymentResult() => request to payment gateway grpc failed, response invalid, orderId: %d, response: %v", orderId, obj)
	return future.Factory().SetCapacity(1).
		SetError(future.InternalError, "Unknown Error", errors.New("GetPaymentResultByOrderID failed")).
		BuildAndSend()
}
