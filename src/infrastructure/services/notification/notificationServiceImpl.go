package notify_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/protos/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"time"
)

type iNotificationServiceImpl struct {
	notifyService  NotificationService.NotificationServiceClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
	notifySeller   bool
	notifyBuyer    bool
	timeout        int
}

func (notification *iNotificationServiceImpl) ConnectToNotifyService() error {
	if notification.grpcConnection == nil || notification.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		notification.grpcConnection, err = grpc.DialContext(ctx, notification.serverAddress+":"+fmt.Sprint(notification.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			logger.Err("ConnectToNotifyService() => GRPC connect dial to notification service failed, err: %s", err.Error())
			return err
		}
		notification.notifyService = NotificationService.NewNotificationServiceClient(notification.grpcConnection)
	}
	return nil
}

func (notification *iNotificationServiceImpl) CloseConnection() {
	if err := notification.grpcConnection.Close(); err != nil {
		logger.Err("notification CloseConnection failed, error: %s", err)
	}
}

func NewNotificationService(address string, port int, notifySeller, notifyBuyer bool, timeout int) INotificationService {
	return &iNotificationServiceImpl{serverAddress: address,
		serverPort:   port,
		timeout:      timeout,
		notifySeller: notifySeller,
		notifyBuyer:  notifyBuyer}
}

func (notification iNotificationServiceImpl) NotifyBySMS(ctx context.Context, request SMSRequest) future.IFuture {
	if err := notification.ConnectToNotifyService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "Connect to NotifyService Failed")).
			BuildAndSend()
	}

	if (notification.notifySeller && request.User == SellerUser) ||
		(notification.notifyBuyer && request.User == BuyerUser) {

		var outCtx context.Context
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			outCtx = metadata.NewOutgoingContext(ctx, md)
		} else {
			outCtx = metadata.NewOutgoingContext(ctx, metadata.New(nil))
		}

		timeoutTimer := time.NewTimer(time.Duration(notification.timeout) * time.Second)

		req := &NotificationService.Sms{
			To:   request.Phone,
			Body: request.Body,
		}

		notifyFn := func() <-chan interface{} {
			notifyChan := make(chan interface{}, 0)
			go func() {
				result, err := notification.notifyService.SendSms(outCtx, req)
				if err != nil {
					notifyChan <- err
				} else {
					notifyChan <- result
				}
			}()
			return notifyChan
		}

		var obj interface{} = nil
		select {
		case obj = <-notifyFn():
			timeoutTimer.Stop()
			break
		case <-timeoutTimer.C:
			logger.Err("NotifyBySMS() => notifyService.SendSms timeout, request: %v", request)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "UnknownError", errors.New("NotifyBySMS Timeout")).
				BuildAndSend()
		}

		if err, ok := obj.(error); ok {
			if err != nil {
				logger.Err("NotifyBySMS() => notifyService.SendSms failed, request: %v, error: %s ", request, err.Error())
				return future.Factory().SetCapacity(1).
					SetError(future.InternalError, "UnknownError", errors.Wrap(err, "NotifyBySMS Failed")).
					BuildAndSend()
			}
		} else if result, ok := obj.(*NotificationService.Result); ok {
			if result.Status != 200 {
				logger.Err("NotifyBySMS() => notifyService.SendSms failed, request: %v, status: %d, error: %s", request, result.Status, result.Message)
				return future.Factory().SetCapacity(1).
					SetError(future.ErrorCode(result.Status), result.Message, errors.Wrap(err, "NotifyBySMS Failed")).
					BuildAndSend()
			}
		} else {
			logger.Err("NotifyBySMS() => notifyService.SendSms failed, result invalid, request: %v, result: %v", request, obj)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "UnknownError", errors.Wrap(err, "NotifyBySMS Failed")).
				BuildAndSend()
		}

		return future.Factory().SetCapacity(1).BuildAndSend()
	}

	return future.Factory().SetCapacity(1).
		SetError(future.NotAccepted, "Notification Not Enabled", errors.New("Notification Not Enabled")).BuildAndSend()
}

func (notification iNotificationServiceImpl) NotifyByMail(ctx context.Context, request EmailRequest) future.IFuture {
	if err := notification.ConnectToNotifyService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "Connect to NotifyService Failed")).
			BuildAndSend()
	}

	req := &NotificationService.Email{
		From:       request.From,
		To:         request.To,
		Subject:    request.Subject,
		Body:       request.Body,
		Attachment: request.Attachment,
	}

	result, err := notification.notifyService.SendEmail(ctx, req)
	if err != nil {
		logger.Err("NotifyByMail() =>=> failed, request: %v, error: %s ", request, err.Error())
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "NotifyByMail => Failed")).
			BuildAndSend()
	}
	if result.Status != 200 {
		logger.Err("NotifyByMail() => failed, request: %v, status: %d, error: %s", request, result.Status, result.Message)
		return future.Factory().SetCapacity(1).
			SetError(future.ErrorCode(result.Status), result.Message, errors.Wrap(err, "NotifyByMail Failed")).
			BuildAndSend()
	}

	return future.Factory().SetCapacity(1).
		SetError(future.ErrorCode(result.Status), result.Message, errors.Wrap(err, "NotifyBySMS Failed")).
		BuildAndSend()
}
