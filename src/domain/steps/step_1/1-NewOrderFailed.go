package new_order_failed_step

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
	"time"
)

const (
	stepName       string = "New_Order_Failed"
	stepIndex      int    = 1
	NewOrderFailed        = "NewOrderFailed"
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

	//finalizeState, ok := newOrderProcessingFailed.StatesMap()[0].(launcher_state.ILauncherState)
	//if ok != true || finalizeState.ActiveType() != actives.FinalizeAction {
	//	logger.Err("finalize state doesn't exist in index 0 of statesMap, order: %v", order)
	//	returnChannel := make(chan promise.FutureData, 1)
	//	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	//	defer close(returnChannel)
	//	return promise.NewPromise(returnChannel, 1, 1)
	//}
	//
	//newOrderProcessingFailed.UpdateAllOrderStatus(ctx, &order, itemsId, steps.ClosedStatus, true)
	//return finalizeState.ActionLauncher(ctx, order, nil, finalize_action.OrderFailedFinalizeAction)

	newOrderProcessingFailed.UpdateAllOrderStatus(ctx, &order, itemsId, steps.ClosedStatus, false)
	newOrderProcessingFailed.updateOrderItemsProgress(ctx, &order, itemsId, NewOrderFailed, true, steps.ClosedStatus)
	if err := newOrderProcessingFailed.persistOrder(ctx, &order); err != nil {
	}
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotAccepted, Reason: "Order Payment Failed"}}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (newOrderProcessingFailed newOrderProcessingFailedStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", newOrderProcessingFailed.Name(), order, err.Error())
	}

	return err
}

func (newOrderProcessingFailed newOrderProcessingFailedStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string,
	action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					newOrderProcessingFailed.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
					findFlag = true
				}
			}

			if findFlag == false {
				logger.Err("%s received itemId %s not exist in order, orderId: %v", newOrderProcessingFailed.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			newOrderProcessingFailed.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
		}
	}
}

func (newOrderProcessingFailed newOrderProcessingFailedStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
}
