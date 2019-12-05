package notify_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type INotificationService interface {
	NotifyBySMS(ctx context.Context, request SMSRequest) future.IFuture
	NotifyByMail(ctx context.Context, request EmailRequest) future.IFuture
}

type EmailRequest struct {
	From       string
	To         string
	Subject    string
	Body       string
	Attachment []string
}

type SMSRequest struct {
	Phone string
	Body  string
}
