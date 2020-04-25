package user_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type IUserService interface {
	UserLogin(ctx context.Context, username, password string) future.IFuture
	AuthenticateContextToken(ctx context.Context) future.IFuture
	GetSellerProfile(ctx context.Context, sellerId string) future.IFuture
}

type LoginTokens struct {
	AccessToken  string
	RefreshToken string
}
