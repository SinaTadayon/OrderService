package stock_service

import (
	"context"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/stock"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type iStockServiceImpl struct {
	serverAddress 	string
	serverPort		int
}

func NewStockService(address string, port int) IStockService {
	return &iStockServiceImpl{address, port}
}

func (stockService iStockServiceImpl) SingleStockAction(context context.Context, inventoryId string, count int, action stock_action.ActionEnums) promise.IPromise {
	if action == stock_action.ReservedAction {
		panic("must be implement")
	} else if action == stock_action.ReleasedAction {
		panic("must be implement")
	} else if action == stock_action.SettlementAction {
		panic("must be implement")
	}
	return nil
}


func (stockService iStockServiceImpl) BatchStockActions(context context.Context,itemsStock map[string]int, action stock_action.ActionEnums) promise.IPromise {
	if action == stock_action.ReservedAction {
		panic("must be implement")
	} else if action == stock_action.ReleasedAction {
		panic("must be implement")
	} else if action == stock_action.SettlementAction {
		panic("must be implement")
	}
	return nil
}


