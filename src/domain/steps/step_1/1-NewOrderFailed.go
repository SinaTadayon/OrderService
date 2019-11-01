package new_order_failed_step

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	finalize_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/finalize"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	launcher_state "gitlab.faza.io/order-project/order-service/domain/states/launcher"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName string 	= "New_Order_Failed"
	stepIndex int		= 1
)

type newOrderProcessingFailedStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingFailedStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingFailedStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &newOrderProcessingFailedStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (newOrderProcessingFailed newOrderProcessingFailedStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

// TODO must be dynamic check state
func (newOrderProcessingFailed newOrderProcessingFailedStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	//state := newOrderProcessingFailed.StatesMap()[0]
	//if state.Actions().ActionType() == actions.ActiveAction {
	//	activeState := state.(launcher_state.ILauncherState)
	//} else {
	//	listenerState := state.(listener_state.IListenerState)
	//}
	//

	finalizeState, ok := newOrderProcessingFailed.StatesMap()[0].(launcher_state.ILauncherState)
	if ok != true || finalizeState.ActiveType() != actives.FinalizeAction {
		logger.Err("finalize state doesn't exist in index 0 of statesMap, order: %v", order)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrderProcessingFailed.UpdateOrderStep(ctx, &order, itemsId, "CLOSED")
	return finalizeState.ActionLauncher(ctx, order, nil, finalize_action.OrderFailedFinalizeAction)
}

