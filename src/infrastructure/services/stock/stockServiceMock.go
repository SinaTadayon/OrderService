package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
)

type iStockServiceMock struct {

}

func NewStockServiceMock() IStockService {
	return &iStockServiceMock{}
}

func (stock iStockServiceMock) SingleStockAction(ctx context.Context, inventoryId string, count int, action string) promise.IPromise {
	if action == "StockReserved" || action == "StockReleased" || action == "StockSettlement" {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: nil}
		return promise.NewPromise(returnChannel, 1, 1)
	} else {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Action Invalid"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}
}

func (stock iStockServiceMock) BatchStockActions(ctx context.Context, order entities.Order, itemsId []string, action string) promise.IPromise {
	if action == "StockReserved" || action == "StockReleased" || action == "StockSettlement" {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: nil}
		return promise.NewPromise(returnChannel, 1, 1)
	} else {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
			Code: promise.InternalError, Reason: "Action Invalid"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}
}

func (stock iStockServiceMock) GetStockClient() stockProto.StockClient {
	return nil
}

func (stock iStockServiceMock) ConnectToStockService() error {
	return nil
}