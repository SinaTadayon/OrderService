package finalize_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	finalize_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/finalize"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	"time"
)

const (
	stateName string = "Finalize_Action_State"
	activeType = actives.FinalizeAction
)

type finalizeActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &finalizeActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction) launcher_state.ILauncherState {
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
func (finalizeState finalizeActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	for _, action := range finalizeState.Actions().(actives.IActiveAction).ActionEnums() {
		if action == finalize_action.OrderFailedFinalizeAction {
			finalizeState.persistOrderState(ctx, &order, itemsId, action, true, "")
			returnChannel <- promise.FutureData{Data:promise.FutureData{}, Ex:nil}
			break
		} else if action == finalize_action.BuyerFinalizeAction {
			panic("must be implement")
		} else if action == finalize_action.PaymentFailedFinalizeAction {
			finalizeState.persistOrderState(ctx, &order, itemsId, action, true, "")
			returnChannel <- promise.FutureData{Data:promise.FutureData{}, Ex:nil}
		} else if action == finalize_action.MarketFinalizeAction {
			panic("must be implement")
		} else {
			logger.Err("actions in not valid for finalize, action: %v, order: %v", action, order)
			finalizeState.persistOrderState(ctx, &order, itemsId, action, false, "received param type is not a actions.IEnumAction")
			returnChannel <- promise.FutureData{Data:promise.FutureData{}, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
			break
		}
	}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (finalizeState finalizeActionLauncher) persistOrderState(ctx context.Context, order *entities.Order, itemsId []string,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					finalizeState.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
				} else {
					logger.Err("finalize received itemId %s not exist in order, order: %v", id, order)
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
	order.Items[index].OrderStep.CurrentState.Name = finalizeState.Name()
	order.Items[index].OrderStep.CurrentState.Index = finalizeState.Index()
	order.Items[index].OrderStep.CurrentState.Type = finalizeState.Actions().ActionType().Name()
	order.Items[index].OrderStep.CurrentState.CreatedAt = time.Now().UTC()
	order.Items[index].OrderStep.CurrentState.Result = result
	order.Items[index].OrderStep.CurrentState.Reason = reason

	if acceptedAction != nil {
		order.Items[index].OrderStep.CurrentState.AcceptedAction.Name = acceptedAction.Name()
	} else {
		order.Items[index].OrderStep.CurrentState.AcceptedAction.Name = ""
	}

	order.Items[index].OrderStep.CurrentState.AcceptedAction.Type = actives.FinalizeAction.String()
	order.Items[index].OrderStep.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	order.Items[index].OrderStep.CurrentState.AcceptedAction.Data = ""
	order.Items[index].OrderStep.CurrentState.AcceptedAction.Time = &order.Items[index].OrderStep.CurrentState.CreatedAt

	order.Items[index].OrderStep.CurrentState.Actions = []entities.Action{order.Items[index].OrderStep.CurrentState.AcceptedAction}

	stateHistory := entities.StateHistory {
		Name: order.Items[index].OrderStep.CurrentState.Name,
		Index: order.Items[index].OrderStep.CurrentState.Index,
		Type: order.Items[index].OrderStep.CurrentState.Type,
		Action: order.Items[index].OrderStep.CurrentState.AcceptedAction,
		Result: order.Items[index].OrderStep.CurrentState.Result,
		Reason: order.Items[index].OrderStep.CurrentState.Reason,
		CreatedAt:order.Items[index].OrderStep.CurrentState.CreatedAt,
	}

	order.Items[index].OrderStep.StepsHistory[len(order.Items[index].OrderStep.StepsHistory)].StatesHistory =
		append(order.Items[index].OrderStep.StepsHistory[len(order.Items[index].OrderStep.StepsHistory)].StatesHistory, stateHistory)
}

