package stock_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type iStockServiceImpl struct {
	serverAddress 	string
	serverPort		int
}

func NewStockService(address string, port int) IStockService {
	return &iStockServiceImpl{address, port}
}

func (stockService iStockServiceImpl) SingleStockAction(context context.Context, inventoryId string, count int, action string) promise.IPromise {
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


func (stockService iStockServiceImpl) BatchStockActions(context context.Context,itemsStock map[string]int, action string) promise.IPromise {
	//if action == stock_action.ReservedAction {
	//	panic("must be implement")
	//} else if action == stock_action.ReleasedAction {
	//	panic("must be implement")
	//} else if action == stock_action.SettlementAction {
	//	panic("must be implement")
	//}
	//return nil

		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:nil, Ex:nil}
		return promise.NewPromise(returnChannel, 1, 1)


	//if action == "StockReserved" {
	//	panic("must be implement")
	//} else if action == "StockReleased" {
	//	panic("must be implement")
	//} else if action == "StockSettlement" {
	//	panic("must be implement")
	//}
	//return nil

}


