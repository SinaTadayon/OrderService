package state_90

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
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
	if iFrame.Header().KeyExists(string(frame.HeaderSubpackages)) {
		subpackages, ok := iFrame.Header().Value(string(frame.HeaderSubpackages)).([]*entities.Subpackage)
		if !ok {
			logger.Err("iFrame.Header() not a subpackages, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		var settlementStockAction *entities.Action
		if err := state.settlementStock(ctx, subpackages); err != nil {
			settlementStockAction = &entities.Action{
				Name:      stock_action.Settlement.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionFail),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		} else {
			settlementStockAction = &entities.Action{
				Name:      stock_action.Settlement.ActionName(),
				Type:      actions.Stock.ActionName(),
				Result:    string(states.ActionSuccess),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}
		}

		for _, subpackage := range subpackages {
			state.UpdateSubPackage(ctx, subpackage, settlementStockAction)
			_, err := app.Globals.SubPkgRepository.Update(ctx, *subpackage)
			if err != nil {
				logger.Err("SubPkgRepository.Update in %s state failed, orderId: %d, pid: %d, sid: %d, error: %s",
					state.Name(), subpackage.OrderId, subpackage.PId, subpackage.SId, err.Error())
				return
			} else {
				logger.Audit("%s state success, orderId: %d, pid: %d, sid: %d", state.Name(), subpackage.OrderId, subpackage.PId, subpackage.SId)
			}
		}

		order, err := app.Globals.OrderRepository.FindById(ctx, subpackages[0].OrderId)
		if err != nil {
			logger.Err("OrderRepository.FindById in %s state failed, orderId: %d, pid: %d, sid: %d, error: %s",
				state.Name(), subpackages[0].OrderId, subpackages[0].PId, subpackages[0].SId, err.Error())
			return
		}

		var findFlag = true
		for i := 0; i < len(order.Packages); i++ {
			findFlag = true
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if order.Packages[i].Subpackages[j].Status != states.PayToBuyer.StateName() &&
					order.Packages[i].Subpackages[j].Status != states.PayToSeller.StateName() {
					findFlag = false
					break
				}
			}

			if findFlag {
				state.SetPkgStatus(ctx, &order.Packages[i], states.PackageClosedStatus)
				_, err := app.Globals.PkgItemRepository.Update(ctx, order.Packages[i])
				if err != nil {
					logger.Err("update pkgItem status to closed failed, state: %s, orderId: %d, pid: %d, error: %s",
						state.Name(), order.Packages[i].OrderId, order.Packages[i].PId, err.Error())
				} else {
					logger.Audit("update pkgItem status to closed success, state: %s, orderId: %d, pid: %d",
						state.Name(), order.Packages[i].OrderId, order.Packages[i].PId)
				}
			}
		}

		findFlag = true
		for i := 0; i < len(order.Packages); i++ {
			if order.Packages[i].Status != string(states.PackageClosedStatus) {
				findFlag = false
				break
			}
		}

		if findFlag {
			state.SetOrderStatus(ctx, order, states.OrderClosedStatus)
			_, err := app.Globals.OrderRepository.Save(ctx, *order)
			if err != nil {
				logger.Err("update order status to closed failed,state: %s, orderId: %d, error: %s",
					state.Name(), order.OrderId, err.Error())
			} else {
				logger.Audit("update order status to closed success, state: %s, orderId: %d", state.Name(), order.OrderId)
			}
		}
	} else {
		logger.Err("iFrame.Header() not a subpackage or pid not found, state: %s iframe: %v", state.Name(), iFrame)
	}
}

func (state payToSellerState) settlementStock(ctx context.Context, subpackages []*entities.Subpackage) error {

	var sids = make([]uint64, 0, len(subpackages))
	var inventories = make(map[string]int, 32)
	for _, subpackage := range subpackages {
		for z := 0; z < len(subpackage.Items); z++ {
			item := subpackage.Items[z]
			inventories[item.InventoryId] = int(item.Quantity)
		}
		sids = append(sids, subpackage.SId)
	}

	iFuture := app.Globals.StockService.BatchStockActions(ctx, inventories,
		stock_action.New(stock_action.Settlement))
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		logger.Err("Settlement stock from stockService failed, state: %s, orderId: %d, pid: %d, sids: %v, error: %s",
			state.Name(), subpackages[0].OrderId, subpackages[0].PId, sids, futureData.Error())
		return futureData.Error().Reason()
	}

	logger.Audit("Settlement stock success, state: %s, orderId: %d, pid: %d, sids: %v",
		state.Name(), subpackages[0].OrderId, subpackages[0].PId, sids)
	return nil
}
