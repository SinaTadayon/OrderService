package steps

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order/general"
)

type IStep interface {
	Name() 		string
	Index()		int
	Childes()	[]IStep
	Parents()	[]IStep
	States() 	[]states.IState
	ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise
	ProcessOrder(ctx context.Context, order entities.Order, itemsId []string) promise.IPromise
}