package stock_service

import (
	"context"
	"fmt"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
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

func (stock iStockServiceImpl) SingleStockAction(ctx context.Context, inventoryId string, count int, action string) promise.IPromise {
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


func (stock iStockServiceImpl) BatchStockActions(ctx context.Context, order entities.Order, action string) promise.IPromise {
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

	itemStocks := make(map[string]int, len(order.Items))
	for i:= 0; i < len(order.Items); i++ {
		if value, ok := itemStocks[order.Items[i].InventoryId]; ok {
			itemStocks[order.Items[i].InventoryId] = value + 1
		} else {
			itemStocks[order.Items[i].InventoryId] = 1
		}
	}

	if action == "StockReserved" {
		var err error
		reservedStock := make(map[string]int, len(itemStocks))
		for inventoryId, quantity := range itemStocks {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}
			if _ ,err = stock.stockService.StockReserve(ctx, request); err != nil {
				logger.Err("Stock reserved failed, orderId: %s, inventoryId %s with quantity %d", order.OrderId, inventoryId, quantity)
				return stock.rollbackReservedStocks(ctx, &order, reservedStock)
			} else {
				reservedStock[inventoryId] = quantity
			}
		}
	} else if action == "StockReleased" {
		var err error
		releaseStock := make(map[string]int, len(itemStocks))
		for inventoryId, quantity := range itemStocks {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}

			if _ ,err = stock.stockService.StockRelease(ctx, request); err != nil {
				logger.Err("Stock release failed, orderId: %s, inventoryId %s with quantity %d", order.OrderId, inventoryId, quantity)
				return stock.rollbackSettlementStocks(ctx, &order, releaseStock)
			} else {
				releaseStock[inventoryId] = quantity
			}
		}

	} else if action == "StockSettlement" {
		var err error
		settlementStock := make(map[string]int, len(itemStocks))
		for inventoryId, quantity := range itemStocks {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}
			if _ ,err = stock.stockService.StockSettle(ctx, request); err != nil {
				logger.Err("stockService.StockSettle failed, orderId: %s, inventoryId: %s, quantity: %d failed",
					order.OrderId, request.InventoryId, request.Quantity)

				return stock.rollbackSettlementStocks(ctx, &order, settlementStock)
			} else {
				settlementStock[inventoryId] = quantity
			}
		}
	}

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex:nil}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (stock iStockServiceImpl) rollbackReservedStocks(ctx context.Context, order *entities.Order, reservedStock map[string]int) promise.IPromise {
	logger.Audit("rollbackReservedStocks, orderId: %s", order.OrderId)
	for inventoryId, quantity := range reservedStock {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		request := &stockProto.StockRequest{
			Quantity:    int32(quantity),
			InventoryId: inventoryId,
		}

		if _ ,err := stock.stockService.StockRelease(ctx, request); err != nil {
			logger.Err("stockService.StockRelease failed, orderId: %s, inventoryId %s with quantity %d",
				order.OrderId, inventoryId, quantity)
		} else {
			reservedStock[inventoryId] = quantity
		}
	}

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
		Code: promise.NotAccepted, Reason: "Stock Reserved Failed"}}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (stock iStockServiceImpl) rollbackSettlementStocks(ctx context.Context, order *entities.Order, reservedStock map[string]int) promise.IPromise {

	//logger.Audit("rollbackSettlementStocks, orderId: %s", order.OrderId)
	//for inventoryId, quantity := range reservedStock {
	//	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	//	request := &stockProto.StockRequest{
	//		Quantity:    int32(quantity),
	//		InventoryId: inventoryId,
	//	}
	//
	//	if _ ,err := stock.stockService.(ctx, request); err != nil {
	//		logger.Err("stockService.StockRelease failed, orderId: %s, inventoryId %s with quantity %d",
	//			order.OrderId, inventoryId, quantity)
	//	} else {
	//		reservedStock[inventoryId] = quantity
	//	}
	//}

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{
		Code: promise.InternalError, Reason: "Unknown Error"}}
	return promise.NewPromise(returnChannel, 1, 1)
}

