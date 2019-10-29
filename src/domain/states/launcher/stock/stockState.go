package stock_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/stock"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	"time"
)

const (
	stateName string = "Stock_Action_State"
	activeType = actives.StockAction
)

type stockActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &stockActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction,
	launcherType actives.ActiveType) launcher_state.ILauncherState {
	return &stockActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, launcherType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &stockActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}


// TODO sencetive checking for save stock state and stock action
func (stockState stockActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, param interface{}) promise.IPromise {

	nextToStepState, ok := stockState.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("NextToStepState isn't child of StockState, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	itemStocks := make(map[string]int, len(order.Items))
	for i:= 0; i < len(order.Items); i++ {
		if value, ok := itemStocks[order.Items[i].InventoryId]; ok {
			itemStocks[order.Items[i].InventoryId] = value + 1
		} else {
			itemStocks[order.Items[i].InventoryId] = 1
		}
	}

	iPromise := global.Singletons.StockService.BatchStockActions(ctx, itemStocks, stock_action.ReservedAction)
	futureData, err := iPromise.Data()
	if err != nil {
		stockState.persistOrderState(ctx, &order, stock_action.ReservedAction, false)
		logger.Err("StockService promise channel has been closed, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		go func() {
			nextToStepState.ActionLauncher(ctx, order, stock_action.FailedAction)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if futureData.Ex != nil {
		logger.Err("Reserved stock from stockService failed, error: %s, order: %v", futureData.Ex.Error(), order)
		stockState.persistOrderState(ctx, &order, stock_action.ReservedAction, false)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		go func() {
			nextToStepState.ActionLauncher(ctx, order, stock_action.FailedAction)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	stockState.persistOrderState(ctx, &order, stock_action.ReservedAction, true)
	return nextToStepState.ActionLauncher(ctx, order, stock_action.ReservedAction)
}

func (stockState stockActionLauncher) persistOrderState(ctx context.Context, order *entities.Order,
	action actions.IEnumAction, result bool) {
	order.UpdatedAt = time.Now().UTC()
	for i := 0; i < len(order.Items); i++ {
		order.Items[i].BuyerInfo = order.BuyerInfo
		order.Items[i].OrderStep.CreatedAt = time.Now().UTC()

		order.Items[i].OrderStep.CurrentName = ctx.Value(global.CtxStepName).(string)
		order.Items[i].OrderStep.CurrentIndex = ctx.Value(global.CtxStepIndex).(int)
		order.Items[i].OrderStep.CurrentState.Name = stockState.Name()
		order.Items[i].OrderStep.CurrentState.Index = stockState.Index()
		order.Items[i].OrderStep.CurrentState.CreatedAt = time.Now().UTC()
		order.Items[i].OrderStep.CurrentState.ActionResult = result
		order.Items[i].OrderStep.CurrentState.Reason = ""

		order.Items[i].OrderStep.CurrentState.Action.Name = action.Name()
		order.Items[i].OrderStep.CurrentState.Action.Type = actives.StockAction.String()
		order.Items[i].OrderStep.CurrentState.Action.Base = actions.ActiveAction.String()
		order.Items[i].OrderStep.CurrentState.Action.Data = ""
		order.Items[i].OrderStep.CurrentState.Action.DispatchedTime = nil

		order.Items[i].OrderStep.StepsHistory[len(order.Items[i].OrderStep.StepsHistory)].StatesHistory =
			append(order.Items[i].OrderStep.StepsHistory[0].StatesHistory, order.Items[i].OrderStep.CurrentState)
	}

	orderChecked, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("Save Stock State Failed, error: %s, order: %v", err, orderChecked)
	}
}
