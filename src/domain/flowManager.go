package domain

import (
	"context"
	message "gitlab.faza.io/protos/order/general"
)

type IFlowManager interface {
	MessageHandler(ctx context.Context, req *message.Request) (*message.Response, error)
}
