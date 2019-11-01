package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type IStockService interface {
	SingleStockAction(context context.Context, inventoryId string, count int, action string) promise.IPromise
	BatchStockActions(context context.Context,itemsStock map[string]int, action string) promise.IPromise
}
