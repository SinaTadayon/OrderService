package stock_service

import (
	"context"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/stock"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type IStockService interface {
	SingleStockAction(context context.Context, inventoryId string, count int, action stock_action.ActionEnums) promise.IPromise
	BatchStockActions(context context.Context,itemsStock map[string]int, action stock_action.ActionEnums) promise.IPromise
}
