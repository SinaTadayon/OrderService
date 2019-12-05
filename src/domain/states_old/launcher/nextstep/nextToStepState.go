package next_to_step_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"time"
)

const (
	stateName  string = "Next_To_Step_Action_State"
	activeType        = actives.NextToStepAction
)

type nextToStepActionLauncher struct {
	*launcher_state.BaseLauncherImpl
	actionStepMap map[actions.IEnumAction]states.IState
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction, actionStepMap map[actions.IEnumAction]states.IState) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType), actionStepMap}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction, actionStepMap map[actions.IEnumAction]states.IState) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType), actionStepMap}
}

func NewFrom(base *launcher_state.BaseLauncherImpl, actionStepMap map[actions.IEnumAction]states.IState) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{base, actionStepMap}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

func (nextStep nextToStepActionLauncher) ActionStepMap() map[actions.IEnumAction]states.IState {
	return nextStep.actionStepMap
}

func (nextStep nextToStepActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {

	if param == nil {
		logger.Err("received param in NextToStepState is nil, order: %v", order)
		returnChannel := make(chan future.IDataFuture, 1)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewFuture(returnChannel, 1, 1)
	}

	actionEnum, ok := param.(actions.IEnumAction)
	if ok != true {
		logger.Err("param in NextToStepState is not actions.IEnumAction type, order: %v", order)
		nextStep.persistOrderState(ctx, &order, itemsId, nil, false, "received param type is not a actions.IEnumAction")
		returnChannel := make(chan future.IDataFuture, 1)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewFuture(returnChannel, 1, 1)
	}

	if step, ok := nextStep.ActionStepMap()[actionEnum]; ok {
		return step.ProcessOrder(ctx, order, nil, nil)
	} else {
		logger.Err("Received action not exist in nextStep.ActionStepMap(), order: %v", order)
		nextStep.persistOrderState(ctx, &order, itemsId, actionEnum, false, "received actions.IEnumAction is not valid for this nextToStep")
		returnChannel := make(chan future.IDataFuture, 1)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewFuture(returnChannel, 1, 1)
	}
}

func (nextStep nextToStepActionLauncher) persistOrderState(ctx context.Context, order *entities.Order, itemsId []uint64,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					nextStep.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
				} else {
					logger.Err("nextToStep received itemId %d not exist in order, order: %v", id, order)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			nextStep.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
		}
	}

	orderChecked, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("Save NextToStep Status Failed, error: %s, order: %v", err, orderChecked)
	}
}

func (nextStep nextToStepActionLauncher) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	//order.Items[index].Tracking.CreatedAt = ctx.Value(global.CtxStepTimestamp).(time.Time)
	//order.Items[index].Tracking.StateName = ctx.Value(global.CtxStepName).(string)
	//order.Items[index].Tracking.StateIndex = ctx.Value(global.CtxStepIndex).(int)
	//
	//order.Items[index].Tracking.CurrentState.ActionName = nextStep.ActionName()
	//order.Items[index].Tracking.CurrentState.Index = nextStep.Index()
	//order.Items[index].Tracking.CurrentState.Type = nextStep.Actions().ActionType().ActionName()
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
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Type = actives.NextToStepAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Get = nil
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Time = &order.Items[index].Tracking.CurrentState.CreatedAt
	//
	//order.Items[index].Tracking.CurrentState.Actions = []entities.Actions{order.Items[index].Tracking.CurrentState.AcceptedAction}
	//
	//stateHistory := entities.StateHistory {
	//	ActionName: order.Items[index].Tracking.CurrentState.ActionName,
	//	Index: order.Items[index].Tracking.CurrentState.Index,
	//	Type: order.Items[index].Tracking.CurrentState.Type,
	//	Actions: order.Items[index].Tracking.CurrentState.AcceptedAction,
	//	Result: order.Items[index].Tracking.CurrentState.Result,
	//	Reason: order.Items[index].Tracking.CurrentState.Reason,
	//	CreatedAt:order.Items[index].Tracking.CurrentState.CreatedAt,
	//}
	//
	//order.Items[index].Tracking.States[len(order.Items[index].Tracking.States)].States =
	//	append(order.Items[index].Tracking.States[len(order.Items[index].Tracking.States)].States, stateHistory)
}
