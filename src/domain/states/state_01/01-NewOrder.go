package state_01

import (
	"context"
	"fmt"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
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
	//logger.Audit("New Order Received . . .")

	order := iFrame.Header().Value(string(frame.HeaderOrder)).(*entities.Order)
	action := &entities.Action{
		Name:      state.Actions()[0].ActionEnum().ActionName(),
		Type:      "",
		UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
		UTP:       state.Actions()[0].ActionType().ActionName(),
		Perm:      "",
		Priv:      "",
		Policy:    "",
		Result:    string(states.ActionSuccess),
		Reasons:   nil,
		Data:      nil,
		CreatedAt: time.Now().UTC(),
		Extended:  nil,
	}
	state.UpdateOrderAllStatus(ctx, order, states.OrderNewStatus, states.PackageNewStatus, action)
	newOrder, err := app.Globals.OrderRepository.Save(ctx, *order)
	if err != nil {
		errStr = fmt.Sprintf("OrderRepository.Save in %s state failed, order: %v, error: %s", state.Name(), order, err.Error())
		logger.Err(errStr)
		_ = state.releasedStock(ctx, newOrder)
		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetError(future.InternalError, errStr, err).
			Send()

	} else {
		newFrame := frame.Factory().
			SetFuture(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetOrderId(newOrder.OrderId).SetBody(newOrder).Build()

		state.StatesMap()[state.Actions()[0]].Process(ctx, newFrame)
	}
}

func (state newOrderState) releasedStock(ctx context.Context, order *entities.Order) error {

	var inventories = make(map[string]int, 32)
	for i := 0; i < len(order.Packages); i++ {
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			for z := 0; z < len(order.Packages[i].Subpackages[j].Items); z++ {
				item := order.Packages[i].Subpackages[j].Items[z]
				inventories[item.InventoryId] = int(item.Quantity)
			}
		}
	}

	iFuture := app.Globals.StockService.BatchStockActions(ctx, inventories,
		system_action.New(system_action.StockRelease))
	futureData := iFuture.Get()

	if futureData.Error() != nil {
		logger.Err("Reserved stock from stockService failed, state: %s, order: %v, error: %s", state.Name(), order, futureData.Error())
		return futureData.Error().Reason()
	}

	logger.Audit("Release stock success, state: %s, order: %v", state.Name(), order)
	return nil
}
