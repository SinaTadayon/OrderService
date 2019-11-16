package user_service

import (
	"context"
	"gitlab.faza.io/go-framework/acl"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type IUserService interface {
	UserLogin(ctx context.Context, username, password string) promise.IPromise
	AuthenticateContextToken(ctx context.Context) (*acl.Acl, error)
	GetSellerProfile(ctx context.Context, sellerId string) promise.IPromise
}

type LoginTokens struct {
	AccessToken  string
	RefreshToken string
}
