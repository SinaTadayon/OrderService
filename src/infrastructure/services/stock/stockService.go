package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
)

type IStockService interface {
	SingleStockAction(ctx context.Context, inventoryId string, count int, action actions.IAction) future.IFuture
	BatchStockActions(ctx context.Context, inventories map[string]int, action actions.IAction) future.IFuture

	GetStockClient() stockProto.StockClient
	ConnectToStockService() error
}
