package checkout_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	checkout_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/checkout"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	launcher_state "gitlab.faza.io/order-project/order-service/domain/states/launcher"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	pb "gitlab.faza.io/protos/order"
	"time"

	//message "gitlab.faza.io/protos/order/general"
)

const (
	actorType = actors.CheckoutActor
	stateName string = "Checkout_Action_State"
)

type checkoutActionListener struct {
	*listener_state.BaseListenerImpl
}

func New(index int, childes, parents []states.IState, actions actions.IAction) listener_state.IListenerState {
	return &checkoutActionListener{listener_state.NewBaseListener(stateName, index, childes, parents,
		actions, actorType)}
}

func NewOf(name string, index int, childes, parents []states.IState, actions actions.IAction) listener_state.IListenerState {
	return &checkoutActionListener{listener_state.NewBaseListener(name, index, childes, parents,
		actions, actorType)}
}

func NewFrom(base *listener_state.BaseListenerImpl) listener_state.IListenerState {
	return &checkoutActionListener{base}
}

func NewValueOf(base *listener_state.BaseListenerImpl, params ...interface{}) listener_state.IListenerState {
	panic("implementation required")
}

// TODO context handling
func (checkoutAction checkoutActionListener) ActionListener(ctx context.Context, event events.IEvent, param interface{}) promise.IPromise {

	if event == nil {
		logger.Err("Received Event is nil")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	stockState, ok := checkoutAction.Childes()[0].(launcher_state.ILauncherState)
	if ok != true {
		logger.Err("StockState isn't child of CheckoutState, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if event.ActorType() != actors.CheckoutActor {
		logger.Err("Received actorType of event is not CheckoutActor, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if event.ActorAction().ActionEnums()[0] != checkout_action.NewOrderAction {
		logger.Err("Received actorAction of event is not NewOrderAction, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrderRequest, ok := event.Data().(pb.NewOrderRequest)
	if ok != true {
		logger.Err("Received data of event is not NewOrderRequest, event: %v", event)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	value, err := global.Singletons.Converter.Map(newOrderRequest, entities.Order{})
	if err != nil {
		logger.Err("Received NewOrderRequest invalid, error: %s, newOrderRequest: %v", err, newOrderRequest)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.BadRequest, Reason:"Received NewOrderRequest invalid"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrder := value.(entities.Order)
	checkoutAction.postProcessNewOrder(ctx, &newOrder, event.Timestamp())

	order, err := global.Singletons.OrderRepository.Save(newOrder)
	if err != nil {
		logger.Err("Save NewOrder Failed, error: %s, newOrder: %v", err, newOrder)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	return stockState.ActionLauncher(ctx, *order, nil)
}

func (checkoutAction checkoutActionListener) postProcessNewOrder(ctx context.Context, newOrder *entities.Order, timestamp time.Time) {
	for i := 0; i < len(newOrder.Items); i++ {
		newOrder.Items[i].OrderStep.CreatedAt = ctx.Value(global.CtxStepTimestamp).(time.Time)
		newOrder.Items[i].OrderStep.CurrentName = ctx.Value(global.CtxStepName).(string)
		newOrder.Items[i].OrderStep.CurrentIndex = ctx.Value(global.CtxStepIndex).(int)
		newOrder.Items[i].OrderStep.CurrentState.Name = checkoutAction.Name()
		newOrder.Items[i].OrderStep.CurrentState.Index = checkoutAction.Index()
		newOrder.Items[i].OrderStep.CurrentState.CreatedAt = time.Now().UTC()
		newOrder.Items[i].OrderStep.CurrentState.ActionResult = true
		newOrder.Items[i].OrderStep.CurrentState.Reason = ""

		newOrder.Items[i].OrderStep.CurrentState.Action.Name = checkout_action.NewOrderAction.String()
		newOrder.Items[i].OrderStep.CurrentState.Action.Type = actors.CheckoutActor.String()
		newOrder.Items[i].OrderStep.CurrentState.Action.Base = actions.ActorAction.String()
		newOrder.Items[i].OrderStep.CurrentState.Action.Data = ""
		newOrder.Items[i].OrderStep.CurrentState.Action.DispatchedTime = &timestamp

		newOrder.Items[i].OrderStep.StepsHistory = []entities.StepHistory{{
			Name: newOrder.Items[i].OrderStep.CurrentState.Name,
			Index: newOrder.Items[i].OrderStep.CurrentState.Index,
			CreatedAt: newOrder.Items[i].OrderStep.CurrentState.CreatedAt,
			StatesHistory: make([]entities.State, 0, 5),
		}}

		newOrder.Items[i].OrderStep.StepsHistory[0].StatesHistory = append(newOrder.Items[i].OrderStep.StepsHistory[0].StatesHistory, newOrder.Items[i].OrderStep.CurrentState)
	}
}