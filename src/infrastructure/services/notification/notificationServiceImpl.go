package notify_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	//_ notification_client "gitlab.faza.io/services/notification-client"
)

type iNotificationServiceImpl struct {

}

func NewNotificationService() INotificationService {
	return &iNotificationServiceImpl{}
}

func(notification iNotificationServiceImpl) NotifyBySMS(ctx context.Context, request SMSRequest) promise.IPromise {
	//notification_client.SendSms(ctx, request.Phone, request.Body)
	panic("must be implemented")
}

func(notification iNotificationServiceImpl) NotifyByMail(ctx context.Context, request EmailRequest) promise.IPromise {
	panic("must be implemented")
}