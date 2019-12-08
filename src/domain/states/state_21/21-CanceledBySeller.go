package state_21

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"time"
)

const (
	stepName  string = "Canceled_By_Seller"
	stepIndex int    = 21
)

type canceledBySellerState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledBySellerState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledBySellerState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &canceledBySellerState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state canceledBySellerState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderSubpackage)) {
		subpkg, ok := iFrame.Header().Value(string(frame.HeaderSubpackage)).(*entities.Subpackage)
		if !ok {
			logger.Err("iFrame.Header() not a subpackage, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		var releaseStockAction *entities.Action
		if err := state.releasedStock(ctx, subpkg); err != nil {
			releaseStockAction = &entities.Action{
				Name:      stock_action.Release.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionFail),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		} else {
			releaseStockAction = &entities.Action{
				Name:      stock_action.Release.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionSuccess),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		}

		nextToAction := &entities.Action{
			Name:      system_action.NextToState.ActionName(),
			Type:      actions.System.ActionName(),
			Result:    string(states.ActionSuccess),
			Reasons:   nil,
			CreatedAt: time.Now().UTC(),
		}

		state.UpdateSubPackage(ctx, subpkg, releaseStockAction)
		state.UpdateSubPackage(ctx, subpkg, nextToAction)
		subPkgUpdated, err := global.Singletons.SubPkgRepository.Update(ctx, *subpkg)
		if err != nil {
			logger.Err("SubPkgRepository.Update in %s state failed, orderId: %d, sellerId: %d, itemId: %d, error: %s",
				state.Name(), subpkg.OrderId, subpkg.SellerId, subpkg.ItemId, err.Error())
		} else {
			logger.Audit("Cancel by seller success, orderId: %d, sellerId: %d, itemId: %d", subpkg.OrderId, subpkg.SellerId, subpkg.ItemId)
			state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(subPkgUpdated).Build())
		}
	} else if iFrame.Header().KeyExists(string(frame.HeaderPackage)) {
		pkgItem, ok := iFrame.Header().Value(string(frame.HeaderPackage)).(*entities.PackageItem)
		if !ok {
			logger.Err("iFrame.Header() not a sellerId, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		var releaseStockAction *entities.Action
		if err := state.releasedStockPackage(ctx, pkgItem); err != nil {
			releaseStockAction = &entities.Action{
				Name:      stock_action.Release.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionFail),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		} else {
			releaseStockAction = &entities.Action{
				Name:      stock_action.Release.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionSuccess),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		}

		nextToAction := &entities.Action{
			Name:      system_action.NextToState.ActionName(),
			Type:      actions.System.ActionName(),
			Result:    string(states.ActionSuccess),
			Reasons:   nil,
			CreatedAt: time.Now().UTC(),
		}

		for i := 0; i < len(pkgItem.Subpackages); i++ {
			state.UpdateSubPackage(ctx, &pkgItem.Subpackages[i], releaseStockAction)
			state.UpdateSubPackage(ctx, &pkgItem.Subpackages[i], nextToAction)
		}
		pkgItemUpdated, err := global.Singletons.PkgItemRepository.Update(ctx, *pkgItem)
		if err != nil {
			logger.Err("PkgItemRepository.Update in %s state failed, orderId: %d, sellerId: %d, error: %s",
				state.Name(), pkgItem.OrderId, pkgItem.SellerId, err.Error())
		} else {
			logger.Audit("Cancel by seller success, orderId: %d, sellerId: %d, itemId: %d", pkgItemUpdated.OrderId, pkgItemUpdated.SellerId)
			state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(pkgItemUpdated).Build())
		}
	} else {
		logger.Err("iFrame.Header() not a subpackage or sellerId not found, state: %s iframe: %v", state.Name(), iFrame)
	}
}

func (state canceledBySellerState) releasedStock(ctx context.Context, subpackage *entities.Subpackage) error {

	var inventories = make(map[string]int, 32)
	for z := 0; z < len(subpackage.Items); z++ {
		item := subpackage.Items[z]
		inventories[item.InventoryId] = int(item.Quantity)
	}

	iFuture := global.Singletons.StockService.BatchStockActions(ctx, inventories,
		stock_action.New(stock_action.Release))
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		logger.Err("Reserved stock from stockService failed, state: %s, orderId: %d, sellerId: %d, itemId: %d, error: %s",
			state.Name(), subpackage.OrderId, subpackage.SellerId, subpackage.ItemId, futureData.Error())
		return futureData.Error().Reason()
	}

	logger.Audit("Release stock success, state: %s, orderId: %d, sellerId: %d, itemId: %d",
		state.Name(), subpackage.OrderId, subpackage.SellerId, subpackage.ItemId)
	return nil
}

func (state canceledBySellerState) releasedStockPackage(ctx context.Context, pkgItem *entities.PackageItem) error {

	var inventories = make(map[string]int, 32)

	for j := 0; j < len(pkgItem.Subpackages); j++ {
		for z := 0; z < len(pkgItem.Subpackages[j].Items); z++ {
			item := pkgItem.Subpackages[j].Items[z]
			inventories[item.InventoryId] = int(item.Quantity)
		}
	}

	iFuture := global.Singletons.StockService.BatchStockActions(ctx, inventories,
		stock_action.New(stock_action.Release))
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		logger.Err("Reserved stock from stockService failed, state: %s, orderId: %d, sellerId: %d, error: %s",
			state.Name(), pkgItem.OrderId, pkgItem.SellerId, futureData.Error())
		return futureData.Error().Reason()
	}

	logger.Audit("Release stock success, state: %s, orderId: %d, sellerId: %d",
		state.Name(), pkgItem.OrderId, pkgItem.SellerId)

	return nil
}
