package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type IStockService interface {
	SingleStockAction(ctx context.Context, inventoryId string, count int, action string) promise.IPromise
	BatchStockActions(ctx context.Context, order entities.Order, action string) promise.IPromise
}
