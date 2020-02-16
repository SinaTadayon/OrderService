package payment_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	payment_gateway "gitlab.faza.io/protos/payment-gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
	"sync"
	"time"
)

type iPaymentServiceImpl struct {
	paymentService       payment_gateway.PaymentGatewayClient
	grpcConnection       *grpc.ClientConn
	serverAddress        string
	serverPort           int
	callbackTimeout      int
	paymentResultTimeout int
	mux                  sync.Mutex
}

func NewPaymentService(address string, port int, callbackTimeout, paymentResultTimeout int) IPaymentService {
	return &iPaymentServiceImpl{nil, nil, address,
		port, callbackTimeout, paymentResultTimeout, sync.Mutex{},
	}
}

func (payment *iPaymentServiceImpl) ConnectToPaymentService() error {
	if payment.grpcConnection == nil {
		payment.mux.Lock()
		defer payment.mux.Unlock()
		if payment.grpcConnection == nil {
			var err error
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			payment.grpcConnection, err = grpc.DialContext(ctx, payment.serverAddress+":"+fmt.Sprint(payment.serverPort),
				grpc.WithBlock(), grpc.WithInsecure())
			if err != nil {
				applog.GLog.Logger.Error("GRPC connect dial to payment service failed",
					"fn", "ConnectToPaymentService",
					"address", payment.serverAddress,
					"port", payment.serverPort,
					"error", err.Error())
				return err
			}
			payment.paymentService = payment_gateway.NewPaymentGatewayClient(payment.grpcConnection)
		}
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

	timeoutTimer := time.NewTimer(time.Duration(payment.callbackTimeout) * time.Second)
	var paymentFn func() <-chan interface{}

	if request.Method == IPG {
		gatewayRequest := &payment_gateway.GenerateRedirRequest{
			Gateway:  request.Gateway,
			Amount:   request.Amount,
			Currency: request.Currency,
			OrderID:  strconv.Itoa(int(request.OrderId)),
			Mobile:   request.Mobile,
		}

		paymentFn = func() <-chan interface{} {
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
	} else if request.Method == MPG {
		gatewayRequest := &payment_gateway.MPGStartRequest{
			Amount:   request.Amount,
			Currency: request.Currency,
			OrderID:  strconv.Itoa(int(request.OrderId)),
			Mobile:   request.Mobile,
		}

		paymentFn = func() <-chan interface{} {
			paymentChan := make(chan interface{}, 0)
			go func() {
				result, err := payment.paymentService.MPGStart(outCtx, gatewayRequest)
				if err != nil {
					paymentChan <- err
				} else {
					paymentChan <- result
				}
			}()
			return paymentChan
		}
	} else {
		applog.GLog.Logger.Error("Payment Request Invalid",
			"fn", "OrderPayment",
			"request", request)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.New("Payment Request Invalid")).
			BuildAndSend()
	}

	var obj interface{} = nil
	select {
	case obj = <-paymentFn():
		timeoutTimer.Stop()
		break
	case <-timeoutTimer.C:
		applog.GLog.Logger.FromContext(ctx).Error("request to payment gateway grpc timeout",
			"fn", "OrderPayment",
			"oid", request.OrderId,
			"amount", request.Amount,
			"gateway", request.Gateway,
			"currency", request.Currency)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.New("GenerateRedirectURL Timeout")).
			BuildAndSend()
	}

	if err, ok := obj.(error); ok {
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("request to payment gateway grpc failed",
				"fn", "OrderPayment",
				"oid", request.OrderId,
				"amount", request.Amount,
				"gateway", request.Gateway,
				"currency", request.Currency,
				"error", err)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GenerateRedirectURL failed")).
				BuildAndSend()
		}
	} else if response, ok := obj.(*payment_gateway.GenerateRedirResponse); ok {
		applog.GLog.Logger.FromContext(ctx).Debug("received payment service response",
			"fn", "OrderPayment",
			"request", request, "response", response)
		paymentResponse := IPGPaymentResponse{
			CallbackUrl: response.CallbackUrl,
			InvoiceId:   response.InvoiceId,
			PaymentId:   response.PaymentId,
		}

		return future.Factory().SetCapacity(1).
			SetData(paymentResponse).
			BuildAndSend()
	} else if response, ok := obj.(*payment_gateway.MPGStartResponse); ok {
		applog.GLog.Logger.FromContext(ctx).Debug("received payment service response",
			"fn", "OrderPayment",
			"request", request,
			"response", response)
		paymentResponse := MPGPaymentResponse{
			HostRequest:     response.HostRequest,
			HostRequestSign: response.HostRequestSign,
			PaymentId:       response.PaymentId,
		}
		return future.Factory().SetCapacity(1).
			SetData(paymentResponse).
			BuildAndSend()
	}

	applog.GLog.Logger.FromContext(ctx).Error("request to payment gateway grpc failed, response invalid",
		"fn", "OrderPayment",
		"oid", request.OrderId,
		"amount", request.Amount,
		"gateway", request.Gateway,
		"currency", request.Currency,
		"response", obj)
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

	timeoutTimer := time.NewTimer(time.Duration(payment.paymentResultTimeout) * time.Second)

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
		applog.GLog.Logger.FromContext(ctx).Error("request to GetPaymentResultByOrderID grpc timeout",
			"fn", "GetPaymentResult",
			"oid", orderId)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.New("GetPaymentResultByOrderID Timeout")).
			BuildAndSend()
	}

	// TODO decode err code
	if err, ok := obj.(error); ok {
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("request to GetPaymentResultByOrderID grpc failed",
				"fn", "GetPaymentResult",
				"oid", orderId,
				"error", err)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "GetPaymentResultByOrderID failed")).
				BuildAndSend()
		}
	} else if response, ok := obj.(*payment_gateway.PaymentRequest); ok {
		applog.GLog.Logger.FromContext(ctx).Debug("received payment gateway result",
			"fn", "GetPaymentResult",
			"oid", orderId, "response", response)

		oid, err := strconv.Atoi(response.OrderID)
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("request to GetPaymentResultByOrderID grpc failed",
				"fn", "GetPaymentResult",
				"oid", orderId, "error", err)
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

	applog.GLog.Logger.FromContext(ctx).Error("request to payment gateway grpc failed, response invalid",
		"fn", "GetPaymentResult",
		"oid", orderId, "response", obj)
	return future.Factory().SetCapacity(1).
		SetError(future.InternalError, "Unknown Error", errors.New("GetPaymentResultByOrderID failed")).
		BuildAndSend()
}
