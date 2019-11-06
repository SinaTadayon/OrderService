package stock_service

import (
	"context"
	"fmt"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"time"
)

type iStockServiceImpl struct {
	stockService 	stockProto.StockClient
	grpcConnection 	*grpc.ClientConn
	serverAddress 	string
	serverPort		int
}

func NewStockService(address string, port int) IStockService {
	return &iStockServiceImpl{nil, nil, address, port}
}

func (stock iStockServiceImpl) connectToStockService() error {
	if stock.grpcConnection == nil || stock.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		stock.grpcConnection, err = grpc.DialContext(ctx, stock.serverAddress+":"+fmt.Sprint(stock.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			logger.Err("GRPC connect dial to stock service failed, err: %s", err.Error())
			return err
		}
		stock.stockService = stockProto.NewStockClient(stock.grpcConnection)
	}
	return nil
}

func (stock iStockServiceImpl) SingleStockAction(context context.Context, inventoryId string, count int, action string) promise.IPromise {
	//if action == stock_action.ReservedAction {
	//	panic("must be implement")
	//} else if action == stock_action.ReleasedAction {
	//	panic("must be implement")
	//} else if action == stock_action.SettlementAction {
	//	panic("must be implement")
	//}
	//return nil

	if action == "StockReserved" {
		panic("must be implement")
	} else if action == "StockReleased" {
		panic("must be implement")
	} else if action == "StockSettlement" {
		panic("must be implement")
	}
	return nil
}


func (stock iStockServiceImpl) BatchStockActions(ctx context.Context,itemsStock map[string]int, action string) promise.IPromise {
	//if action == stock_action.ReservedAction {
	//	panic("must be implement")
	//} else if action == stock_action.ReleasedAction {
	//	panic("must be implement")
	//} else if action == stock_action.SettlementAction {
	//	panic("must be implement")
	//}
	//return nil

		//returnChannel := make(chan promise.FutureData, 1)
		//defer close(returnChannel)
		//returnChannel <- promise.FutureData{Data:nil, Ex:nil}
		//return promise.NewPromise(returnChannel, 1, 1)

	if err := stock.connectToStockService(); err != nil {
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{
			Code:   promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if action == "StockReserved" {
		var err error
		reservedStock := make(map[string]int, len(itemsStock))
		for inventoryId, quantity := range itemsStock {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}
			if _ ,err = stock.stockService.StockReserve(ctx, request); err != nil {
				stock.rollbackReservedStocks(reservedStock)
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
					Code: promise.NotAccepted, Reason: fmt.Sprintf("Stock reserved for inventoryId %s with quantity %d failed", inventoryId, quantity)}}
				return promise.NewPromise(returnChannel, 1, 1)
			} else {
				reservedStock[inventoryId] = quantity
			}
		}
	} else if action == "StockReleased" {
		panic("must be implement")

	} else if action == "StockSettlement" {
		var err error
		settlementStock := make(map[string]int, len(itemsStock))
		for inventoryId, quantity := range itemsStock {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}
			if _ ,err = stock.stockService.StockSettle(ctx, request); err != nil {
				logger.Err("stockService.StockSettle with inventoryId: %s, quantity: %d failed",
					request.InventoryId, request.Quantity)

				stock.rollbackSettlementStocks(settlementStock)
				returnChannel := make(chan promise.FutureData, 1)
				defer close(returnChannel)
				returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
					Code: promise.InternalError, Reason: "Unknown Error"}}
				return promise.NewPromise(returnChannel, 1, 1)
			} else {
				settlementStock[inventoryId] = quantity
			}
		}
	}
	return nil

}

func (stock iStockServiceImpl) rollbackReservedStocks(itemsStock map[string]int) {
	panic("must be implement")
}

func (stock iStockServiceImpl) rollbackSettlementStocks(itemsStock map[string]int) {
	panic("must be implement")
}

