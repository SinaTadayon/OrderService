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
	"time"
)

type iNotificationServiceImpl struct {
	notifyService  NotificationService.NotificationServiceClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
}

func (notification *iNotificationServiceImpl) ConnectToNotifyService() error {
	if notification.grpcConnection == nil || notification.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		notification.grpcConnection, err = grpc.DialContext(ctx, notification.serverAddress+":"+fmt.Sprint(notification.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			logger.Err("GRPC connect dial to notification service failed, err: %s", err.Error())
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

func NewNotificationService(address string, port int) INotificationService {
	return &iNotificationServiceImpl{serverAddress: address, serverPort: port}
}

func (notification iNotificationServiceImpl) NotifyBySMS(ctx context.Context, request SMSRequest) future.IFuture {
	if err := notification.ConnectToNotifyService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "Connect to NotifyService Failed")).
			BuildAndSend()
	}

	req := &NotificationService.Sms{
		To:   request.Phone,
		Body: request.Body,
	}

	result, err := notification.notifyService.SendSms(ctx, req)
	if err != nil {
		logger.Err("NotifyBySMS() => failed, request: %v, error: %s ", request, err.Error())
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "NotifyBySMS Failed")).
			BuildAndSend()
	}

	if result.Status != 200 {
		logger.Err("NotifyBySMS() => failed, request: %v, status: %d, error: %s", request, result.Status, result.Message)
		return future.Factory().SetCapacity(1).
			SetError(future.ErrorCode(result.Status), result.Message, errors.Wrap(err, "NotifyBySMS Failed")).
			BuildAndSend()
	}

	return future.Factory().SetCapacity(1).
		SetError(future.ErrorCode(result.Status), result.Message, errors.Wrap(err, "NotifyBySMS Failed")).
		BuildAndSend()
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
