package stock_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
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
			logger.Err("ConnectToStockService() => GRPC connect dial to stock service failed, address: %s, port: %d, err: %s", stock.serverAddress, stock.serverPort, err.Error())
			return err
		}
		stock.stockService = stockProto.NewStockClient(stock.grpcConnection)
	}
	return nil
}

func (stock *iStockServiceImpl) CloseConnection() {
	if err := stock.grpcConnection.Close(); err != nil {
		logger.Err("CloseConnection() => stock CloseConnection failed, error: %s", err)
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
			logger.Err("SingleStockAction() => request to stock service grpc timeout, orderId: %d, request: %v", orderId, request)
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
				logger.Err("SingleStockAction() => stock reserved failed, orderId: %d, inventoryId %s with quantity %d, error: %s", orderId, request.InventoryId, request.Quantity, err)
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
			logger.Audit("SingleStockAction() => Stock Reserved success, orderId: %d, inventoryId: %s,  available: %d, reserved: %d", orderId, request.InventoryId, response.Available, response.Reserved)
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
			logger.Err("SingleStockAction() => request to stock service release grpc timeout, orderId: %d, request: %v", orderId, request)
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
				logger.Err("SingleStockAction() => Stock Release failed, orderId: %d, inventoryId %s with quantity %d, error: %s", orderId, request.InventoryId, request.Quantity, err)
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
			logger.Audit("SingleStockAction() => Stock Release success, orderId: %d, inventoryId: %s,  available: %d, reserved: %d", orderId, request.InventoryId, response.Available, response.Reserved)
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
			logger.Err("SingleStockAction() => request to stock service settlement grpc timeout, orderId: %d, request: %v", orderId, request)
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
				logger.Err("SingleStockAction() => stockService.StockSettle failed, orderId: %d ,inventoryId: %s, quantity: %d, error: %s",
					orderId, request.InventoryId, request.Quantity, err)
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
			logger.Audit("BatchStockActions() => Stock Settlement success, orderId: %d, inventoryId: %s,  available: %d, reserved: %d", orderId, request.InventoryId, response.Available, response.Reserved)
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
				logger.Err("BatchStockActions() => request to stock service grpc timeout, orderId: %d, request: %v", orderId, request)
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
					logger.Err("BatchStockActions() => stock reserved failed, orderId: %d, inventoryId %s with quantity %d, error: %s", orderId, request.InventoryId, request.Quantity, err)
					response := ResponseStock{
						InventoryId: requestStock.InventoryId,
						Count:       requestStock.Count,
						Result:      false,
					}
					responses = append(responses, response)
				}
			} else if response, ok := obj.(*stockProto.StockResponse); ok {
				logger.Audit("BatchStockActions() => Stock Reserved success, orderId: %d, inventoryId: %s,  available: %d, reserved: %d", orderId, request.InventoryId, response.Available, response.Reserved)
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
				logger.Err("BatchStockActions() => request to stock service release grpc timeout, orderId: %d, request: %v", orderId, request)
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
					logger.Err("BatchStockActions() => Stock Release failed, orderId: %d, inventoryId %s with quantity %d, error: %s", orderId, request.InventoryId, request.Quantity, err)
					response := ResponseStock{
						InventoryId: requestStock.InventoryId,
						Count:       requestStock.Count,
						Result:      false,
					}
					responses = append(responses, response)
				}
			} else if response, ok := obj.(*stockProto.StockResponse); ok {
				logger.Audit("BatchStockActions() => Stock Release success, orderId: %d, inventoryId: %s,  available: %d, reserved: %d", orderId, request.InventoryId, response.Available, response.Reserved)
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
				logger.Err("BatchStockActions() => request to stock service settlement grpc timeout, orderId: %d, request: %v", orderId, request)
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
					logger.Err("BatchStockActions() => stockService.StockSettle failed, orderId: %d ,inventoryId: %s, quantity: %d, error: %s",
						orderId, request.InventoryId, request.Quantity, err)
					response := ResponseStock{
						InventoryId: requestStock.InventoryId,
						Count:       requestStock.Count,
						Result:      false,
					}
					responses = append(responses, response)
				}
			} else if response, ok := obj.(*stockProto.StockResponse); ok {
				logger.Audit("BatchStockActions() => Stock Settlement success, orderId: %d, inventoryId: %s,  available: %d, reserved: %d", orderId, request.InventoryId, response.Available, response.Reserved)
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
