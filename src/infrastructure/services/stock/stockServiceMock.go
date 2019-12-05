package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
)

type iStockServiceMock struct {
}

func NewStockServiceMock() IStockService {
	return &iStockServiceMock{}
}

func (stock iStockServiceMock) SingleStockAction(ctx context.Context, inventoryId string, count int, action string) future.IFuture {
	if action == "StockReserved" || action == "StockReleased" || action == "StockSettlement" {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
		return future.NewFuture(returnChannel, 1, 1)
	} else {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{
			Code: future.InternalError, Reason: "Action Invalid"}}
		return future.NewFuture(returnChannel, 1, 1)
	}
}

func (stock iStockServiceMock) BatchStockActions(ctx context.Context, order entities.Order, itemsId []uint64, action string) future.IFuture {
	if action == "StockReserved" || action == "StockReleased" || action == "StockSettlement" {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
		return future.NewFuture(returnChannel, 1, 1)
	} else {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{
			Code: future.InternalError, Reason: "Action Invalid"}}
		return future.NewFuture(returnChannel, 1, 1)
	}
}

func (stock iStockServiceMock) GetStockClient() stockProto.StockClient {
	return nil
}

func (stock iStockServiceMock) ConnectToStockService() error {
	return nil
}
