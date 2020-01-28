package stock_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	stockProto "gitlab.faza.io/protos/stock-proto.git"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"time"
)

type iStockServiceImpl struct {
	stockService   stockProto.StockClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
	timeout        int
}

func NewStockService(address string, port int, timeout int) IStockService {
	return &iStockServiceImpl{nil, nil, address, port, timeout}
}

func (stock *iStockServiceImpl) ConnectToStockService() error {
	if stock.grpcConnection == nil || stock.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		stock.grpcConnection, err = grpc.DialContext(ctx, stock.serverAddress+":"+fmt.Sprint(stock.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			applog.GLog.Logger.Error("GRPC connect dial to stock service failed",
				"fn", "ConnectToStockService",
				"address", stock.serverAddress,
				"port", stock.serverPort,
				"err", err)
			return err
		}
		stock.stockService = stockProto.NewStockClient(stock.grpcConnection)
	}
	return nil
}

func (stock *iStockServiceImpl) CloseConnection() {
	if err := stock.grpcConnection.Close(); err != nil {
		applog.GLog.Logger.Error("stock CloseConnection failed",
			"error", err)
	}
}

func (stock *iStockServiceImpl) GetStockClient() stockProto.StockClient {
	return stock.stockService
}

func (stock *iStockServiceImpl) SingleStockAction(ctx context.Context, requestStock RequestStock, orderId uint64, action actions.IAction) future.IFuture {
	if err := stock.ConnectToStockService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "ConnectToPaymentService failed")).
			BuildAndSend()
	}

	var outCtx context.Context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		outCtx = metadata.NewOutgoingContext(ctx, md)
	} else {
		outCtx = metadata.NewOutgoingContext(ctx, metadata.New(nil))
	}

	timeoutTimer := time.NewTimer(time.Duration(stock.timeout) * time.Second)

	if action.ActionEnum() == system_action.StockReserve {
		var err error
		request := &stockProto.StockRequest{
			Quantity:    int32(requestStock.Count),
			InventoryId: requestStock.InventoryId,
		}

		stockFn := func() <-chan interface{} {
			stockChan := make(chan interface{}, 0)
			go func() {
				result, err := stock.stockService.StockReserve(outCtx, request)
				if err != nil {
					stockChan <- err
				} else {
					stockChan <- result
				}
			}()
			return stockChan
		}

		var obj interface{} = nil
		select {
		case obj = <-stockFn():
			timeoutTimer.Stop()
			break
		case <-timeoutTimer.C:
			applog.GLog.Logger.FromContext(ctx).Error("request to stock service grpc timeout",
				"fn", "SingleStockAction",
				"oid", orderId, "request", request)
			response := ResponseStock{
				InventoryId: requestStock.InventoryId,
				Count:       requestStock.Count,
				Result:      false,
			}
			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Reserve Failed", errors.New("Stock Reserve Timeout")).
				SetData(response).
				BuildAndSend()
		}

		if e, ok := obj.(error); ok {
			if e != nil {
				err = e
				applog.GLog.Logger.FromContext(ctx).Error("stock reserved failed",
					"fn", "SingleStockAction",
					"oid", orderId,
					"inventoryId", request.InventoryId,
					"quantity", request.Quantity,
					"error", err)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      false,
				}
				return future.Factory().SetCapacity(1).
					SetError(future.NotAccepted, "Stock Reserve Failed", errors.Wrap(err, "Stock Reserve Timeout")).
					SetData(response).
					BuildAndSend()
			}
		} else if response, ok := obj.(*stockProto.StockResponse); ok {
			applog.GLog.Logger.FromContext(ctx).Debug("Stock Reserved success",
				"fn", "SingleStockAction",
				"orderId", orderId,
				"inventoryId", request.InventoryId,
				"available", response.Available,
				"reserved", response.Reserved)
			response := ResponseStock{
				InventoryId: requestStock.InventoryId,
				Count:       requestStock.Count,
				Result:      true,
			}
			return future.Factory().SetCapacity(1).
				SetData(response).
				BuildAndSend()
		}

	} else if action.ActionEnum() == system_action.StockRelease {
		var err error

		request := &stockProto.StockRequest{
			Quantity:    int32(requestStock.Count),
			InventoryId: requestStock.InventoryId,
		}

		stockFn := func() <-chan interface{} {
			stockChan := make(chan interface{}, 0)
			go func() {
				result, err := stock.stockService.StockRelease(outCtx, request)
				if err != nil {
					stockChan <- err
				} else {
					stockChan <- result
				}
			}()
			return stockChan
		}

		var obj interface{} = nil
		select {
		case obj = <-stockFn():
			timeoutTimer.Stop()
			break
		case <-timeoutTimer.C:
			applog.GLog.Logger.FromContext(ctx).Error("request to stock service release grpc timeout",
				"fn", "SingleStockAction",
				"oid", orderId,
				"request", request)
			response := ResponseStock{
				InventoryId: requestStock.InventoryId,
				Count:       requestStock.Count,
				Result:      false,
			}

			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Release Failed", errors.New("Stock Release Timeout")).
				SetData(response).
				BuildAndSend()
		}

		if e, ok := obj.(error); ok {
			if e != nil {
				err = e
				applog.GLog.Logger.FromContext(ctx).Error("Stock Release failed",
					"fn", "SingleStockAction",
					"oid", orderId,
					"inventoryId", request.InventoryId,
					"quantity", request.Quantity,
					"error", err)

				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      false,
				}
				return future.Factory().SetCapacity(1).
					SetError(future.NotAccepted, "Stock Release Failed", errors.Wrap(err, "Stock Release Failed")).
					SetData(response).
					BuildAndSend()
			}
		} else if response, ok := obj.(*stockProto.StockResponse); ok {
			applog.GLog.Logger.FromContext(ctx).Debug("Stock Release success",
				"fn", "SingleStockAction",
				"orderId", orderId,
				"inventoryId", request.InventoryId,
				"available", response.Available,
				"reserved", response.Reserved)
			response := ResponseStock{
				InventoryId: requestStock.InventoryId,
				Count:       requestStock.Count,
				Result:      true,
			}
			return future.Factory().SetCapacity(1).
				SetData(response).
				BuildAndSend()
		}

	} else if action.ActionEnum() == system_action.StockSettlement {
		var err error
		request := &stockProto.StockRequest{
			Quantity:    int32(requestStock.Count),
			InventoryId: requestStock.InventoryId,
		}

		stockFn := func() <-chan interface{} {
			stockChan := make(chan interface{}, 0)
			go func() {
				result, err := stock.stockService.StockSettle(ctx, request)
				if err != nil {
					stockChan <- err
				} else {
					stockChan <- result
				}
			}()
			return stockChan
		}

		var obj interface{} = nil
		select {
		case obj = <-stockFn():
			timeoutTimer.Stop()
			break
		case <-timeoutTimer.C:
			applog.GLog.Logger.FromContext(ctx).Error("request to stock service settlement grpc timeout",
				"fn", "SingleStockAction",
				"orderId", orderId, "request", request)
			response := ResponseStock{
				InventoryId: requestStock.InventoryId,
				Count:       requestStock.Count,
				Result:      false,
			}
			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Settlement Failed", errors.New("Stock Settlement Timeout")).
				SetData(response).
				BuildAndSend()
		}

		if e, ok := obj.(error); ok {
			if e != nil {
				err = e
				applog.GLog.Logger.FromContext(ctx).Error("stockService.StockSettle failed",
					"fn", "SingleStockAction",
					"orderId", orderId,
					"inventoryId", request.InventoryId,
					"quantity", request.Quantity,
					"error", err)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      false,
				}
				return future.Factory().SetCapacity(1).
					SetError(future.NotAccepted, "Stock Settlement Failed", errors.Wrap(err, "Stock Settlement Failed")).
					SetData(response).
					BuildAndSend()
			}
		} else if response, ok := obj.(*stockProto.StockResponse); ok {
			applog.GLog.Logger.FromContext(ctx).Debug("Stock Settlement success",
				"fn", "SingleStockAction",
				"orderId", orderId,
				"inventoryId", request.InventoryId,
				"available", response.Available,
				"reserved", response.Reserved)
			response := ResponseStock{
				InventoryId: requestStock.InventoryId,
				Count:       requestStock.Count,
				Result:      true,
			}
			return future.Factory().SetCapacity(1).
				SetData(response).
				BuildAndSend()
		}
	}

	response := ResponseStock{
		InventoryId: requestStock.InventoryId,
		Count:       requestStock.Count,
		Result:      false,
	}
	return future.Factory().SetCapacity(1).
		SetError(future.NotAccepted, "Stock Action Failed", errors.New("Stock Action Failed")).
		SetData(response).
		BuildAndSend()

}

func (stock *iStockServiceImpl) BatchStockActions(ctx context.Context, requests []RequestStock, orderId uint64, action actions.IAction) future.IFuture {
	if err := stock.ConnectToStockService(); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "ConnectToPaymentService failed")).
			BuildAndSend()
	}

	var outCtx context.Context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		outCtx = metadata.NewOutgoingContext(ctx, md)
	} else {
		outCtx = metadata.NewOutgoingContext(ctx, metadata.New(nil))
	}

	timeoutTimer := time.NewTimer(time.Duration(stock.timeout) * time.Second)
	responses := make([]ResponseStock, 0, len(requests))

	if action.ActionEnum() == system_action.StockReserve {
		var err error
		for _, requestStock := range requests {
			request := &stockProto.StockRequest{
				Quantity:    int32(requestStock.Count),
				InventoryId: requestStock.InventoryId,
			}

			stockFn := func() <-chan interface{} {
				stockChan := make(chan interface{}, 0)
				go func() {
					result, err := stock.stockService.StockReserve(outCtx, request)
					if err != nil {
						stockChan <- err
					} else {
						stockChan <- result
					}
				}()
				return stockChan
			}

			var obj interface{} = nil
			select {
			case obj = <-stockFn():
				timeoutTimer.Stop()
				break
			case <-timeoutTimer.C:
				applog.GLog.Logger.FromContext(ctx).Error("request to stock service grpc timeout",
					"fn", "BatchStockActions",
					"oid", orderId, "request", request)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      false,
				}
				responses = append(responses, response)
				timeoutTimer.Reset(time.Duration(stock.timeout) * time.Second)
				continue
			}

			if e, ok := obj.(error); ok {
				if e != nil {
					err = e
					applog.GLog.Logger.FromContext(ctx).Error("stock reserved failed",
						"fn", "BatchStockActions",
						"oid", orderId,
						"inventoryId", request.InventoryId,
						"quantity", request.Quantity,
						"error", err)
					response := ResponseStock{
						InventoryId: requestStock.InventoryId,
						Count:       requestStock.Count,
						Result:      false,
					}
					responses = append(responses, response)
				}
			} else if response, ok := obj.(*stockProto.StockResponse); ok {
				applog.GLog.Logger.FromContext(ctx).Debug("Stock Reserved success",
					"fn", "BatchStockActions",
					"oid", orderId,
					"inventoryId", request.InventoryId,
					"available", response.Available,
					"reserved", response.Reserved)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      true,
				}
				responses = append(responses, response)
			}

			timeoutTimer.Reset(time.Duration(stock.timeout) * time.Second)
		}

		if err != nil {
			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Reserve Failed", errors.Wrap(err, "Stock Reserve Failed")).
				SetData(responses).
				BuildAndSend()
		}

	} else if action.ActionEnum() == system_action.StockRelease {
		var err error
		for _, requestStock := range requests {
			request := &stockProto.StockRequest{
				Quantity:    int32(requestStock.Count),
				InventoryId: requestStock.InventoryId,
			}

			stockFn := func() <-chan interface{} {
				stockChan := make(chan interface{}, 0)
				go func() {
					result, err := stock.stockService.StockRelease(outCtx, request)
					if err != nil {
						stockChan <- err
					} else {
						stockChan <- result
					}
				}()
				return stockChan
			}

			var obj interface{} = nil
			select {
			case obj = <-stockFn():
				timeoutTimer.Stop()
				break
			case <-timeoutTimer.C:
				applog.GLog.Logger.FromContext(ctx).Error("request to stock service release grpc timeout",
					"fn", "BatchStockActions",
					"orderId", orderId, "request", request)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      false,
				}
				responses = append(responses, response)
				timeoutTimer.Reset(time.Duration(stock.timeout) * time.Second)
				continue
			}

			if e, ok := obj.(error); ok {
				if e != nil {
					err = e
					applog.GLog.Logger.FromContext(ctx).Error("Stock Release failed",
						"fn", "BatchStockActions",
						"oid", orderId,
						"inventoryId", request.InventoryId,
						"quantity", request.Quantity,
						"error", err)
					response := ResponseStock{
						InventoryId: requestStock.InventoryId,
						Count:       requestStock.Count,
						Result:      false,
					}
					responses = append(responses, response)
				}
			} else if response, ok := obj.(*stockProto.StockResponse); ok {
				applog.GLog.Logger.FromContext(ctx).Debug("Stock Release success",
					"fn", "BatchStockActions",
					"orderId", orderId,
					"inventoryId", request.InventoryId,
					"available", response.Available,
					"reserved", response.Reserved)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      true,
				}
				responses = append(responses, response)
			}

			timeoutTimer.Reset(time.Duration(stock.timeout) * time.Second)
		}

		if err != nil {
			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Release Failed", errors.Wrap(err, "Stock Release Failed")).
				SetData(responses).
				BuildAndSend()
		}

	} else if action.ActionEnum() == system_action.StockSettlement {
		var err error
		for _, requestStock := range requests {
			request := &stockProto.StockRequest{
				Quantity:    int32(requestStock.Count),
				InventoryId: requestStock.InventoryId,
			}

			stockFn := func() <-chan interface{} {
				stockChan := make(chan interface{}, 0)
				go func() {
					result, err := stock.stockService.StockSettle(ctx, request)
					if err != nil {
						stockChan <- err
					} else {
						stockChan <- result
					}
				}()
				return stockChan
			}

			var obj interface{} = nil
			select {
			case obj = <-stockFn():
				timeoutTimer.Stop()
				break
			case <-timeoutTimer.C:
				applog.GLog.Logger.FromContext(ctx).Error("request to stock service settlement grpc timeout",
					"fn", "BatchStockActions",
					"orderId", orderId,
					"request", request)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      false,
				}
				responses = append(responses, response)
				timeoutTimer.Reset(time.Duration(stock.timeout) * time.Second)
				continue
			}

			if e, ok := obj.(error); ok {
				if e != nil {
					err = e
					applog.GLog.Logger.FromContext(ctx).Error("stockService.StockSettle failed",
						"fn", "BatchStockActions",
						"orderId", orderId,
						"inventoryId", request.InventoryId,
						"quantity", request.Quantity,
						"error", err)
					response := ResponseStock{
						InventoryId: requestStock.InventoryId,
						Count:       requestStock.Count,
						Result:      false,
					}
					responses = append(responses, response)
				}
			} else if response, ok := obj.(*stockProto.StockResponse); ok {
				applog.GLog.Logger.FromContext(ctx).Debug("Stock Settlement success",
					"fn", "BatchStockActions",
					"oid", orderId,
					"inventoryId", request.InventoryId,
					"available", response.Available,
					"reserved", response.Reserved)
				response := ResponseStock{
					InventoryId: requestStock.InventoryId,
					Count:       requestStock.Count,
					Result:      true,
				}
				responses = append(responses, response)
			}
		}

		if err != nil {
			return future.Factory().SetCapacity(1).
				SetError(future.NotAccepted, "Stock Settlement Failed", errors.Wrap(err, "Stock Settlement Failed")).
				SetData(responses).
				BuildAndSend()
		}
	}

	return future.Factory().SetCapacity(1).SetData(responses).BuildAndSend()
}
