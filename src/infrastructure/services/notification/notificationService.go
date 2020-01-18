package notify_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type SMSUserType string

const (
	BuyerUser  SMSUserType = "Buyer"
	SellerUser SMSUserType = "Seller"
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
	User  SMSUserType
}
