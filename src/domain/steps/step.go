package steps

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	NewStatus		 = "NEW"
	InProgressStatus = "IN_PROGRESS"
	ClosedStatus	 = "CLOSED"
)

type IStep interface {
	Name() 		string
	Index()		int
	Childes()	[]IStep
	Parents()	[]IStep
	States() 	[]states.IState
	ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise
	ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise
}