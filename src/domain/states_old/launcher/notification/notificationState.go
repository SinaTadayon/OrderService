package notification_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	notification_action "gitlab.faza.io/order-project/order-service/domain/actions/notification"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	"time"
)

const (
	stateName  string = "Notification_Action_State"
	activeType        = actives.NotificationAction
)

type notificationActionLauncher struct {
	*launcher_state.BaseLauncherImpl
}

func New(index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &notificationActionLauncher{launcher_state.NewBaseLauncher(stateName, index, childes, parents,
		actions, activeType)}
}

func NewOf(name string, index int, childes, parents []states_old.IState, actions actions.IAction) launcher_state.ILauncherState {
	return &notificationActionLauncher{launcher_state.NewBaseLauncher(name, index, childes, parents,
		actions, activeType)}
}

func NewFrom(base *launcher_state.BaseLauncherImpl) launcher_state.ILauncherState {
	return &notificationActionLauncher{base}
}

func NewValueOf(base *launcher_state.BaseLauncherImpl, params ...interface{}) launcher_state.ILauncherState {
	panic("implementation required")
}

// TODO must be implement sms and email templates and related to states and actions
// TODO must decouple from child
func (notificationState notificationActionLauncher) ActionLauncher(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {

	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	for _, action := range notificationState.Actions().(actives.IActiveAction).ActionEnums() {
		if action == notification_action.OperatorNotification {
			notificationState.persistOrderState(ctx, &order, itemsId, action, true, "")
			//returnChannel <- promise.FutureData{Data:promise.FutureData{}, Ex:nil}
			break
		} else if action == notification_action.BuyerNotification {
			notificationState.persistOrderState(ctx, &order, itemsId, action, true, "")
			break
		} else if action == notification_action.SellerNotification {
			notificationState.persistOrderState(ctx, &order, itemsId, action, true, "")
			break
			//returnChannel <- promise.FutureData{Data:promise.FutureData{}, Ex:nil}
		} else if action == notification_action.MarketNotificationAction {
			notificationState.persistOrderState(ctx, &order, itemsId, action, true, "")
			break
		} else {
			logger.Err("actions in not valid for notification, action: %v, order: %v", action, order)
			notificationState.persistOrderState(ctx, &order, itemsId, action, false, "received param type is not a actions.IEnumAction")
			//returnChannel <- promise.FutureData{Data:promise.FutureData{}, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
			break
		}
	}

	if ctx.Value(global.CtxStepIndex).(int) == 12 {
		finalizeState, ok := notificationState.Childes()[0].(launcher_state.ILauncherState)
		if ok != true || finalizeState.ActiveType() != actives.FinalizeAction {
			logger.Err("finalize state doesn't exist in index 0 of notificationState, order: %v", order)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		return finalizeState.ActionLauncher(ctx, order, itemsId, nil)
	}

	return promise.NewPromise(returnChannel, 1, 1)
}

func (notificationState notificationActionLauncher) persistOrderState(ctx context.Context, order *entities.Order, itemsId []uint64,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	order.UpdatedAt = time.Now().UTC()

	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					notificationState.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
				} else {
					logger.Err("finalize received itemId %d not exist in order, orderId: %d", id, order.OrderId)
				}
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			notificationState.doUpdateOrderState(ctx, order, i, acceptedAction, result, reason)
		}
	}

	orderChecked, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("Save finalize State Failed, error: %s, order: %v", err, orderChecked)
	}
}

func (notificationState notificationActionLauncher) doUpdateOrderState(ctx context.Context, order *entities.Order, index int,
	acceptedAction actions.IEnumAction, result bool, reason string) {
	//order.Items[index].Tracking.CurrentState.Name = notificationState.Name()
	//order.Items[index].Tracking.CurrentState.Index = notificationState.Index()
	//order.Items[index].Tracking.CurrentState.Type = notificationState.Actions().ActionType().Name()
	//order.Items[index].Tracking.CurrentState.CreatedAt = time.Now().UTC()
	//order.Items[index].Tracking.CurrentState.Result = result
	//order.Items[index].Tracking.CurrentState.Reason = reason
	//
	//if acceptedAction != nil {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Name = acceptedAction.Name()
	//} else {
	//	order.Items[index].Tracking.CurrentState.AcceptedAction.Name = ""
	//}
	//
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Type = actives.FinalizeAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Base = actions.ActiveAction.String()
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Data = nil
	//order.Items[index].Tracking.CurrentState.AcceptedAction.Time = &order.Items[index].Tracking.CurrentState.CreatedAt
	//
	//order.Items[index].Tracking.CurrentState.Actions = []entities.Action{order.Items[index].Tracking.CurrentState.AcceptedAction}
	//
	//stateHistory := entities.StateHistory {
	//	Name: order.Items[index].Tracking.CurrentState.Name,
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
