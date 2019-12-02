package notify_service

import (
	"context"
	"fmt"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
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

func (notification iNotificationServiceImpl) NotifyBySMS(ctx context.Context, request SMSRequest) promise.IPromise {
	if err := notification.ConnectToNotifyService(); err != nil {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	req := &NotificationService.Sms{
		To:   request.Phone,
		Body: request.Body,
	}

	result, err := notification.notifyService.SendSms(ctx, req)
	if err != nil {
		logger.Err("NotifyBySMS() => failed, request: %v, error: %s ", request, err.Error())
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if result.Status != 200 {
		logger.Err("NotifyBySMS() => failed, request: %v, status: %d, error: %s", request, result.Status, result.Message)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex: nil}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (notification iNotificationServiceImpl) NotifyByMail(ctx context.Context, request EmailRequest) promise.IPromise {
	if err := notification.ConnectToNotifyService(); err != nil {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
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
		logger.Err("NotifyByMail() => failed, request: %v, error: %s ", request, err.Error())
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if result.Status != 200 {
		logger.Err("NotifyByMail() => failed, request: %v, status: %d, error: %s", request, result.Status, result.Message)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex: nil}
	return promise.NewPromise(returnChannel, 1, 1)
}
