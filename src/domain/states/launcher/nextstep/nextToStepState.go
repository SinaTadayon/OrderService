package next_to_step_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	"time"
)

const (
	stateName  string = "Next_To_Step_Action_State"
	activeType        = actives.NextToStepAction
)

type nextToStepActionLauncher struct {
	*launcher_state.BaseLauncherImpl
	actionStepMap map[actions.IEnumAction]steps.IStep
}

func New(index int, childes, parents []states.IState, actions actions.IAction, actionStepMap map[actions.IEnumAction]steps.IStep) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType), actionStepMap}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction, actionStepMap map[actions.IEnumAction]steps.IStep) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType), actionStepMap}
}

func NewFrom(base *launcher_state.BaseLauncherImpl, actionStepMap map[actions.IEnumAction]steps.IStep) launcher_state.ILauncherState {
	return &nextToStepActionLauncher{base, actionStepMap}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

func (nextStep nextToStepActionLauncher) ActionStepMap() map[actions.IEnumAction]steps.IStep {
	return nextStep.actionStepMap
}

func (nextStep nextToStepActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	if param == nil {
		logger.Err("received param in NextToStepState is nil, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	actionEnum, ok := param.(actions.IEnumAction)
	if ok != true {
		logger.Err("param in NextToStepState is not actions.IEnumAction type, order: %v", order)
		nextStep.persistOrderState(ctx, &order, itemsId, nil, false, "received param type is not a actions.IEnumAction")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if step, ok := nextStep.ActionStepMap()[actionEnum]; ok {
		return step.ProcessOrder(ctx, order, nil, nil)
	} else {
		logger.Err("Received action not exist in nextStep.ActionStepMap(), order: %v", order)
		nextStep.persistOrderState(ctx, &order, itemsId, actionEnum, false, "received actions.IEnumAction is not valid for this nextToStep")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}
}

func (nextStep nextToStepActionLauncher) persistOrderState(ctx context.Context, order *entities.Order, itemsId []string,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					nextStep.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
				} else {
					logger.Err("nextToStep received itemId %s not exist in order, order: %v", id, order)
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
		logger.Err("Save NextToStep State Failed, error: %s, order: %v", err, orderChecked)
	}
}

func (nextStep nextToStepActionLauncher) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	//order.Items[index].Progress.CreatedAt = ctx.Value(global.CtxStepTimestamp).(time.Time)
	//order.Items[index].Progress.CurrentStepName = ctx.Value(global.CtxStepName).(string)
	//order.Items[index].Progress.CurrentStepIndex = ctx.Value(global.CtxStepIndex).(int)
	//
	//order.Items[index].Progress.CurrentState.Name = nextStep.Name()
	//order.Items[index].Progress.CurrentState.Index = nextStep.Index()
	//order.Items[index].Progress.CurrentState.Type = nextStep.Actions().ActionType().Name()
	//order.Items[index].Progress.CurrentState.CreatedAt = time.Now().UTC()
	//order.Items[index].Progress.CurrentState.Result = result
	//order.Items[index].Progress.CurrentState.Reason = reason
	//
	//if acceptedAction != nil {
	//	order.Items[index].Progress.CurrentState.AcceptedAction.Name = acceptedAction.Name()
	//} else {
	//	order.Items[index].Progress.CurrentState.AcceptedAction.Name = ""
	//}
	//
	//order.Items[index].Progress.CurrentState.AcceptedAction.Type = actives.NextToStepAction.String()
	//order.Items[index].Progress.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	//order.Items[index].Progress.CurrentState.AcceptedAction.Data = nil
	//order.Items[index].Progress.CurrentState.AcceptedAction.Time = &order.Items[index].Progress.CurrentState.CreatedAt
	//
	//order.Items[index].Progress.CurrentState.Actions = []entities.Action{order.Items[index].Progress.CurrentState.AcceptedAction}
	//
	//stateHistory := entities.StateHistory {
	//	Name: order.Items[index].Progress.CurrentState.Name,
	//	Index: order.Items[index].Progress.CurrentState.Index,
	//	Type: order.Items[index].Progress.CurrentState.Type,
	//	Action: order.Items[index].Progress.CurrentState.AcceptedAction,
	//	Result: order.Items[index].Progress.CurrentState.Result,
	//	Reason: order.Items[index].Progress.CurrentState.Reason,
	//	CreatedAt:order.Items[index].Progress.CurrentState.CreatedAt,
	//}
	//
	//order.Items[index].Progress.StepsHistory[len(order.Items[index].Progress.StepsHistory)].StatesHistory =
	//	append(order.Items[index].Progress.StepsHistory[len(order.Items[index].Progress.StepsHistory)].StatesHistory, stateHistory)
}
