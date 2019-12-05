package state_01

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"time"
)

const (
	stepName  string = "New_Order"
	stepIndex int    = 1
)

type newOrderState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &newOrderState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &newOrderState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &newOrderState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state newOrderState) Process(ctx context.Context, iFrame frame.IFrame) {
	var errStr string
	logger.Audit("New Order Received . . .")

	order := iFrame.Header().Value(string(frame.HeaderOrder)).(entities.Order)
	action := &entities.Action{
		Name:      state.Actions()[0].ActionEnum().ActionName(),
		Type:      state.Actions()[0].ActionType().ActionName(),
		Data:      nil,
		Result:    string(states.ActionSuccess),
		Reasons:   nil,
		CreatedAt: time.Now().UTC(),
	}
	state.UpdateOrderAllStatus(ctx, &order, action, states.OrderNewStatus, states.PackageNewStatus)
	newOrder, err := global.Singletons.OrderRepository.Save(ctx, order)
	if err != nil {
		errStr = fmt.Sprintf("OrderRepository.Save in %s state failed, order: %v, error: %s", state.Name(), order, err.Error())
		logger.Err(errStr)
		state.releasedStock(ctx, newOrder)
		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetError(future.InternalError, errStr, errors.Wrap(err, "")).
			Send()
	}

	state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetOrder(*newOrder).Build())
}

func (state newOrderState) releasedStock(ctx context.Context, order *entities.Order) {

	var inventories = make(map[string]int, 32)
	for i := 0; i < len(order.Packages); i++ {
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			for z := 0; z < len(order.Packages[i].Subpackages[j].Items); z++ {
				item := order.Packages[i].Subpackages[j].Items[z]
				inventories[item.InventoryId] = int(item.Quantity)
			}
		}
	}

	iFuture := global.Singletons.StockService.BatchStockActions(ctx, inventories,
		stock_action.New(stock_action.Release))
	futureData := iFuture.Get()
	if futureData == nil {
		logger.Err("StockService future channel has been closed, state: %s, order: %v", state.Name(), order)
		return
	}

	if futureData.Error() != nil {
		logger.Err("Reserved stock from stockService failed, state: %s, order: %v, error: %s", state.Name(), order, futureData.Error())
		return
	}

	logger.Audit("Release stock success, state: %s, order: %v", state.Name(), order)
}
