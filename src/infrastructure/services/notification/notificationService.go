package notify_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type INotificationService interface {
	NotifyBySMS(ctx context.Context, request SMSRequest) promise.IPromise
	NotifyByMail(ctx context.Context, request EmailRequest) promise.IPromise
}

type EmailRequest struct {
	From       string
	To         string
	Subject    string
	Body       string
	Attachment []string
}

type SMSRequest struct {
	Phone  		string
	Body 		string
}
