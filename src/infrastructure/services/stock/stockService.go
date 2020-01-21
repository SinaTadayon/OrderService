package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
)

type RequestStock struct {
	InventoryId string
	Count       int
}

type ResponseStock struct {
	InventoryId string
	Count       int
	Result      bool
}

type IStockService interface {
	SingleStockAction(ctx context.Context, request RequestStock, orderId uint64, action actions.IAction) future.IFuture
	BatchStockActions(ctx context.Context, requests []RequestStock, orderId uint64, action actions.IAction) future.IFuture

	GetStockClient() stockProto.StockClient
	ConnectToStockService() error
}
