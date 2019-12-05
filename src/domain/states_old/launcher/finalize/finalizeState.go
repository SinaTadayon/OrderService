package finalize_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	finalize_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/finalize"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"time"
)

const (
	stateName  string = "Finalize_Action_State"
	activeType        = actives.FinalizeAction
)

type finalizeActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &finalizeActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &finalizeActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &finalizeActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

// TODO must be dynamic
// TODO check actions and improve handling actions
func (finalizeState finalizeActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {

	returnChannel := make(chan future.IDataFuture, 1)
	defer close(returnChannel)
	for _, action := range finalizeState.Actions().(actives.IActiveAction).ActionEnums() {
		if action == finalize_action.OrderFailedFinalizeAction {
			finalizeState.persistOrderState(ctx, &order, itemsId, action, true, "")
			returnChannel <- future.IDataFuture{Data: future.IDataFuture{}, Ex: nil}
			break
		} else if action == finalize_action.BuyerFinalizeAction {
			panic("must be implement")
		} else if action == finalize_action.PaymentFailedFinalizeAction {
			finalizeState.persistOrderState(ctx, &order, itemsId, action, true, "")
			returnChannel <- future.IDataFuture{Data: future.IDataFuture{}, Ex: nil}
		} else if action == finalize_action.MarketFinalizeAction {
			panic("must be implement")
		} else {
			logger.Err("actions in not valid for finalize, action: %v, order: %v", action, order)
			finalizeState.persistOrderState(ctx, &order, itemsId, action, false, "received param type is not a actions.IEnumAction")
			returnChannel <- future.IDataFuture{Data: future.IDataFuture{}, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
			break
		}
	}
	return future.NewFuture(returnChannel, 1, 1)
}

func (finalizeState finalizeActionLauncher) persistOrderState(ctx context.Context, order *entities.Order, itemsId []uint64,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					finalizeState.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
				} else {
					logger.Err("finalize received itemId %d not exist in order, orderId: %d", id, order.OrderId)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			finalizeState.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
		}
	}

	orderChecked, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("Save finalize State Failed, error: %s, order: %v", err, orderChecked)
	}
}

func (finalizeState finalizeActionLauncher) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	//order.Items[index].Tracking.CurrentState.ActionName = finalizeState.ActionName()
	//order.Items[index].Tracking.CurrentState.Index = finalizeState.Index()
	//order.Items[index].Tracking.CurrentState.Type = finalizeState.Actions().ActionType().ActionName()
	//order.Items[index].Tracking.CurrentState.CreatedAt = time.Now().UTC()
	//order.Items[index].Tracking.CurrentState.Result = result
	//order.Items[index].Tracking.CurrentState.Reason = reason
	//
	//if acceptedAction != nil {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.ActionName = acceptedAction.ActionName()
	//} else {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.ActionName = ""
	//}
	//
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Type = actives.FinalizeAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Get = nil
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Time = &order.Items[index].Tracking.CurrentState.CreatedAt
	//
	//order.Items[index].Tracking.CurrentState.Actions = []entities.Action{order.Items[index].Tracking.CurrentState.AcceptedAction}
	//
	//stateHistory := entities.StateHistory {
	//	ActionName: order.Items[index].Tracking.CurrentState.ActionName,
	//	Index: order.Items[index].Tracking.CurrentState.Index,
	//	Type: order.Items[index].Tracking.CurrentState.Type,
	//	Action: order.Items[index].Tracking.CurrentState.AcceptedAction,
	//	Result: order.Items[index].Tracking.CurrentState.Result,
	//	Reason: order.Items[index].Tracking.CurrentState.Reason,
	//	CreatedAt:order.Items[index].Tracking.CurrentState.CreatedAt,
	//}
	//
	//order.Items[index].Tracking.StatesHistory[len(order.Items[index].Tracking.StatesHistory)].StatesHistory =
	//	append(order.Items[index].Tracking.StatesHistory[len(order.Items[index].Tracking.StatesHistory)].StatesHistory, stateHistory)
}
