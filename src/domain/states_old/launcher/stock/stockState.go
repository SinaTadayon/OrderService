package stock_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	finalize_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/finalize"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/stock"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

const (
	stateName  string = "Stock_Action_State"
	activeType        = actives.StockAction
)

type stockActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &stockActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction,
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
func (stockState stockActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {

	var iPromise promise.IPromise
	for _, action := range stockState.Actions().(actives.IActiveAction).ActionEnums() {
		if action == stock_action.ReservedAction {
			iPromise = stockState.doReservedAction(ctx, &order, itemsId)
			break
		} else if action == stock_action.ReleasedAction {
			iPromise = stockState.doReleasedAction(ctx, &order, itemsId)
			break
		} else if action == stock_action.SettlementAction {
			iPromise = stockState.doSettlementAction(ctx, &order, itemsId)
			break
		} else if action == stock_action.FailedAction {

		} else {
			logger.Err("actions in not valid for StockState, order: %v", order)
			iPromise = stockState.createFailedPromise()
		}
	}

	return iPromise
}

func (stockState stockActionLauncher) createFailedPromise() promise.IPromise {
	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (stockState stockActionLauncher) doReservedAction(ctx context.Context, order *entities.Order, itemsId []uint64) promise.IPromise {
	nextToStepState, ok := stockState.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("NextToStepState isn't child of StockState, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	itemStocks := make(map[string]int, len(order.Items))
	for i := 0; i < len(order.Items); i++ {
		if value, ok := itemStocks[order.Items[i].InventoryId]; ok {
			itemStocks[order.Items[i].InventoryId] = value + 1
		} else {
			itemStocks[order.Items[i].InventoryId] = 1
		}
	}

	iPromise := global.Singletons.StockService.BatchStockActions(ctx, *order, itemsId, stock_action.ReservedAction.Name())
	futureData := iPromise.Data()
	if futureData == nil {
		stockState.persistOrderState(ctx, order, stock_action.ReservedAction, false)
		logger.Err("StockService promise channel has been closed, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		go func() {
			nextToStepState.ActionLauncher(ctx, *order, nil, stock_action.FailedAction)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if futureData.Ex != nil {
		logger.Err("Reserved stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
		stockState.persistOrderState(ctx, order, stock_action.ReservedAction, false)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		go func() {
			nextToStepState.ActionLauncher(ctx, *order, nil, stock_action.FailedAction)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	stockState.persistOrderState(ctx, order, stock_action.ReservedAction, true)
	return nextToStepState.ActionLauncher(ctx, *order, nil, stock_action.ReservedAction)
}

func (stockState stockActionLauncher) doReleasedAction(ctx context.Context, order *entities.Order, itemsId []uint64) promise.IPromise {
	notificationState, ok := stockState.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("notificationState isn't child of StockState, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	finalizedState, ok := stockState.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("finalizedState isn't child of StockState, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		return promise.NewPromise(returnChannel, 1, 1)
	}

	var itemStocks map[string]int
	if itemsId != nil && len(itemsId) > 0 {
		itemStocks = make(map[string]int, len(itemsId))
		for _, itemId := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == itemId {
					if value, ok := itemStocks[order.Items[i].InventoryId]; ok {
						itemStocks[order.Items[i].InventoryId] = value + 1
					} else {
						itemStocks[order.Items[i].InventoryId] = 1
					}
				}
			}
		}
	} else {
		itemStocks = make(map[string]int, len(order.Items))
		for i := 0; i < len(order.Items); i++ {
			if value, ok := itemStocks[order.Items[i].InventoryId]; ok {
				itemStocks[order.Items[i].InventoryId] = value + 1
			} else {
				itemStocks[order.Items[i].InventoryId] = 1
			}
		}
	}

	iPromise := global.Singletons.StockService.BatchStockActions(ctx, *order, itemsId, stock_action.ReleasedAction.Name())
	futureData := iPromise.Data()
	if futureData == nil {
		stockState.persistOrderState(ctx, order, stock_action.ReleasedAction, false)
		logger.Err("StockService promise channel has been closed, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		go func() {
			finalizedState.ActionLauncher(ctx, *order, nil, finalize_action.PaymentFailedFinalizeAction)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if futureData.Ex != nil {
		logger.Err("Reserved stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
		stockState.persistOrderState(ctx, order, stock_action.ReleasedAction, false)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		go func() {
			finalizedState.ActionLauncher(ctx, *order, nil, finalize_action.PaymentFailedFinalizeAction)
		}()
		return promise.NewPromise(returnChannel, 1, 1)
	}

	stockState.persistOrderState(ctx, order, stock_action.ReleasedAction, true)
	return notificationState.ActionLauncher(ctx, *order, itemsId, nil)
}

func (stockState stockActionLauncher) doSettlementAction(ctx context.Context, order *entities.Order, itemsId []uint64) promise.IPromise {
	panic("must be implement")
}

// TODO add reason
func (stockState stockActionLauncher) persistOrderState(ctx context.Context, order *entities.Order,
	action actions.IEnumAction, result bool) {
	//order.UpdatedAt = time.Now().UTC()
	//for i := 0; i < len(order.Items); i++ {
	//	order.Items[i].Tracking.CreatedAt = ctx.Value(global.CtxStepTimestamp).(time.Time)
	//	order.Items[i].Tracking.StateName = ctx.Value(global.CtxStepName).(string)
	//	order.Items[i].Tracking.StateIndex = ctx.Value(global.CtxStepIndex).(int)
	//	order.Items[i].Tracking.CurrentState.Name = stockState.Name()
	//	order.Items[i].Tracking.CurrentState.Index = stockState.Index()
	//	order.Items[i].Tracking.CurrentState.Type = stockState.Actions().ActionType().Name()
	//	order.Items[i].Tracking.CurrentState.CreatedAt = time.Now().UTC()
	//	order.Items[i].Tracking.CurrentState.Result = result
	//	order.Items[i].Tracking.CurrentState.Reason = ""
	//
	//	order.Items[i].Tracking.CurrentState.AcceptedAction.Name = action.Name()
	//	order.Items[i].Tracking.CurrentState.AcceptedAction.Type = actives.StockAction.String()
	//	order.Items[i].Tracking.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	//	order.Items[i].Tracking.CurrentState.AcceptedAction.Data = nil
	//	order.Items[i].Tracking.CurrentState.AcceptedAction.Time = &order.UpdatedAt
	//
	//	order.Items[i].Tracking.CurrentState.Actions = []entities.Action{order.Items[i].Tracking.CurrentState.AcceptedAction}
	//
	//	stateHistory := entities.StateHistory {
	//		Name: order.Items[i].Tracking.CurrentState.Name,
	//		Index: order.Items[i].Tracking.CurrentState.Index,
	//		Type: order.Items[i].Tracking.CurrentState.Type,
	//		Action: order.Items[i].Tracking.CurrentState.AcceptedAction,
	//		Result: order.Items[i].Tracking.CurrentState.Result,
	//		Reason: order.Items[i].Tracking.CurrentState.Reason,
	//		CreatedAt:order.Items[i].Tracking.CurrentState.CreatedAt,
	//	}
	//
	//	order.Items[i].Tracking.StatesHistory[len(order.Items[i].Tracking.StatesHistory)].StatesHistory =
	//		append(order.Items[i].Tracking.StatesHistory[len(order.Items[i].Tracking.StatesHistory)].StatesHistory, stateHistory)
	//}
	//
	//orderChecked, err := global.Singletons.OrderRepository.Save(*order)
	//if err != nil {
	//	logger.Err("Save Stock State Failed, error: %s, order: %v", err, orderChecked)
	//}
}
