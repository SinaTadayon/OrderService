package state_90

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"time"
)

const (
	stepName  string = "Pay_To_Seller"
	stepIndex int    = 90
)

type payToSellerState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &payToSellerState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &payToSellerState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &payToSellerState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state payToSellerState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderSubpackage)) {
		subpkg, ok := iFrame.Header().Value(string(frame.HeaderSubpackage)).(*entities.Subpackage)
		if !ok {
			logger.Err("iFrame.Header() not a subpackage, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		var releaseStockAction *entities.Action
		if err := state.settlementStock(ctx, subpkg); err != nil {
			releaseStockAction = &entities.Action{
				Name:      stock_action.Settlement.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionFail),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		} else {
			releaseStockAction = &entities.Action{
				Name:      stock_action.Settlement.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionSuccess),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		}

		state.UpdateSubPackage(ctx, subpkg, releaseStockAction)
		_, err := global.Singletons.SubPkgRepository.Update(ctx, *subpkg)
		if err != nil {
			logger.Err("SubPkgRepository.Update in %s state failed, orderId: %d, sellerId: %d, itemId: %d, error: %s",
				state.Name(), subpkg.OrderId, subpkg.SellerId, subpkg.ItemId, err.Error())
		} else {
			logger.Audit("Cancel by seller success, orderId: %d, sellerId: %d, itemId: %d", subpkg.OrderId, subpkg.SellerId, subpkg.ItemId)
		}
	} else {
		logger.Err("iFrame.Header() not a subpackage or sellerId not found, state: %s iframe: %v", state.Name(), iFrame)
	}
}

func (state payToSellerState) settlementStock(ctx context.Context, subpackage *entities.Subpackage) error {

	var inventories = make(map[string]int, 32)
	for z := 0; z < len(subpackage.Items); z++ {
		item := subpackage.Items[z]
		inventories[item.InventoryId] = int(item.Quantity)
	}

	iFuture := global.Singletons.StockService.BatchStockActions(ctx, inventories,
		stock_action.New(stock_action.Settlement))
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		logger.Err("Settlement stock from stockService failed, state: %s, orderId: %d, sellerId: %d, itemId: %d, error: %s",
			state.Name(), subpackage.OrderId, subpackage.SellerId, subpackage.ItemId, futureData.Error())
		return futureData.Error().Reason()
	}

	logger.Audit("Settlement stock success, state: %s, orderId: %d, sellerId: %d, itemId: %d",
		state.Name(), subpackage.OrderId, subpackage.SellerId, subpackage.ItemId)
	return nil
}
