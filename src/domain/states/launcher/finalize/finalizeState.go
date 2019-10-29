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
func (finalize finalizeActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, param interface{}) promise.IPromise {
	if param == nil {
		logger.Err("received param in FinalizeState is nil, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	actionEnum, ok := param.(actions.IEnumAction)
	if ok != true {
		logger.Err("param in FinalizeState is not actions.IEnumAction type, order: %v", order)
		finalize.persistOrderState(ctx, &order, nil, false, "received param type is not a actions.IEnumAction")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	switch actionEnum {
	case finalize_action.OrderFailedFinalizeAction:
		finalize.persistOrderState(ctx, &order, actionEnum, true, "")
	case finalize_action.BuyerFinalizeAction:
		panic("must be implement")
		//break
	case finalize_action.PaymentFailedFinalizeAction:
		panic("must be implement")
		//break
	case finalize_action.MarketFinalizeAction:
		panic("must be implement")
		//break
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data:promise.FutureData{}, Ex:nil}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (finalize finalizeActionLauncher) persistOrderState(ctx context.Context, order *entities.Order,
	action actions.IEnumAction, result bool, reason string) {
	order.UpdatedAt = time.Now().UTC()
	for i := 0; i < len(order.Items); i++ {
		order.Items[i].OrderStep.CreatedAt = ctx.Value(global.CtxStepTimestamp).(time.Time)
		order.Items[i].OrderStep.CurrentName = ctx.Value(global.CtxStepName).(string)
		order.Items[i].OrderStep.CurrentIndex = ctx.Value(global.CtxStepIndex).(int)
		order.Items[i].OrderStep.CurrentState.Name = finalize.Name()
		order.Items[i].OrderStep.CurrentState.Index = finalize.Index()
		order.Items[i].OrderStep.CurrentState.CreatedAt = time.Now().UTC()
		order.Items[i].OrderStep.CurrentState.ActionResult = result
		order.Items[i].OrderStep.CurrentState.Reason = ""

		if action != nil {
			order.Items[i].OrderStep.CurrentState.Action.Name = action.Name()
		} else {
			order.Items[i].OrderStep.CurrentState.Action.Name = ""
		}

		order.Items[i].OrderStep.CurrentState.Action.Type = actives.FinalizeAction.String()
		order.Items[i].OrderStep.CurrentState.Action.Base = actions.ActiveAction.String()
		order.Items[i].OrderStep.CurrentState.Action.Data = ""
		order.Items[i].OrderStep.CurrentState.Action.DispatchedTime = nil

		stepsHistory := entities.StepHistory{
			Name: order.Items[i].OrderStep.CurrentState.Name,
			Index: order.Items[i].OrderStep.CurrentState.Index,
			CreatedAt: order.Items[i].OrderStep.CurrentState.CreatedAt,
			StatesHistory: make([]entities.State, 0, 1),
		}

		order.Items[i].OrderStep.StepsHistory = append(order.Items[i].OrderStep.StepsHistory, stepsHistory)

		order.Items[i].OrderStep.StepsHistory[len(order.Items[i].OrderStep.StepsHistory)].StatesHistory =
			append(order.Items[i].OrderStep.StepsHistory[0].StatesHistory, order.Items[i].OrderStep.CurrentState)
	}

	orderChecked, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("Save Stock State Failed, error: %s, order: %v", err, orderChecked)
	}
}

