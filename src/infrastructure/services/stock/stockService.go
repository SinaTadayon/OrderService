package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
)

type IStockService interface {
	SingleStockAction(ctx context.Context, inventoryId string, count int, action string) future.IFuture
	BatchStockActions(ctx context.Context, order entities.Order, itemsId []uint64, action string) future.IFuture

	GetStockClient() stockProto.StockClient
	ConnectToStockService() error
}
