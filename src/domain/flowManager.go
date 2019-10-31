package domain

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

type IFlowManager interface {
	MessageHandler(ctx context.Context, req *message.MessageRequest) promise.IPromise
}
