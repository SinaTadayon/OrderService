package states

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	NewStatus        = "NEW"
	InProgressStatus = "IN_PROGRESS"
	ClosedStatus     = "CLOSED"
)

type IState interface {
	Name() string
	Index() int
	Childes() []IState
	Parents() []IState
	ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise
	ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise
}
