package stock_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"time"
)

type iStockServiceImpl struct {
	stockService   stockProto.StockClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
}

func NewStockService(address string, port int) IStockService {
	return &iStockServiceImpl{nil, nil, address, port}
}

func (stock *iStockServiceImpl) ConnectToStockService() error {
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

func (stock *iStockServiceImpl) CloseConnection() {
	if err := stock.grpcConnection.Close(); err != nil {
		logger.Err("stock CloseConnection failed, error: %s", err)
	}
}

func (stock *iStockServiceImpl) GetStockClient() stockProto.StockClient {
	return stock.stockService
}

func (stock *iStockServiceImpl) SingleStockAction(ctx context.Context, inventoryId string, count int, action actions.IAction) future.IFuture {
	return nil
}

func (stock *iStockServiceImpl) BatchStockActions(ctx context.Context, inventories map[string]int, action actions.IAction) future.IFuture {
	if err := stock.ConnectToStockService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "ConnectToPaymentService failed")).
			BuildAndSend()
	}

	if action.ActionEnum() == stock_action.Reserve {
		//var err error
		reservedStock := make(map[string]int, len(inventories))
		for inventoryId, quantity := range inventories {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}
			if response, err := stock.stockService.StockReserve(ctx, request); err != nil {
				logger.Err("Stock reserved failed, inventoryId %s with quantity %d, error: %s", inventoryId, quantity, err)
				return stock.rollbackReservedStocks(ctx, reservedStock, err)
			} else {
				logger.Audit("StockReserved success, inventoryId: %s,  available: %d, reserved: %d",
					inventoryId, response.Available, response.Reserved)
				reservedStock[inventoryId] = int(quantity)
			}
		}
	} else if action.ActionEnum() == stock_action.Release {
		var err error
		for inventoryId, quantity := range inventories {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}
			if _, err = stock.stockService.StockRelease(ctx, request); err != nil {
				logger.Err("Stock release failed, inventoryId %s with quantity %d, error: %s", inventoryId, quantity, err)
			}
		}

		if err != nil {
			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Release Failed", errors.Wrap(err, "")).
				BuildAndSend()
		}

	} else if action.ActionEnum() == stock_action.Settlement {
		var err error
		for inventoryId, quantity := range inventories {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			request := &stockProto.StockRequest{
				Quantity:    int32(quantity),
				InventoryId: inventoryId,
			}
			if _, err = stock.stockService.StockSettle(ctx, request); err != nil {
				logger.Err("stockService.StockSettle failed, inventoryId: %s, quantity: %d, error: %s",
					request.InventoryId, request.Quantity, err)
			}
		}

		if err != nil {
			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Settlement Failed", errors.Wrap(err, "")).
				BuildAndSend()
		}
	}

	return future.Factory().SetCapacity(1).BuildAndSend()
}

func (stock *iStockServiceImpl) rollbackReservedStocks(ctx context.Context, reservedStock map[string]int, err error) future.IFuture {
	logger.Audit("rollbackReservedStock . . .")
	for inventoryId, quantity := range reservedStock {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		request := &stockProto.StockRequest{
			Quantity:    int32(quantity),
			InventoryId: inventoryId,
		}

		if _, err := stock.stockService.StockRelease(ctx, request); err != nil {
			logger.Err("stockService.StockRelease failed, inventoryId %s with quantity %d", inventoryId, quantity)
		} else {
			reservedStock[inventoryId] = quantity
		}
	}

	return future.Factory().SetCapacity(1).
		SetError(future.NotAccepted, "Stock Reserved Failed", errors.Wrap(err, "")).
		BuildAndSend()
}
