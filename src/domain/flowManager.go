package domain

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order/general"
)

type IFlowManager interface {
	MessageHandler(ctx context.Context, req *message.Request) promise.IPromise
}
