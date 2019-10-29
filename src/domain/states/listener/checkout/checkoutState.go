package checkout_action_state

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	checkout_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/checkout"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states/listener"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	pb "gitlab.faza.io/protos/order"
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
		returnChannel <- promise.FutureData{Data:nil, Error:promise.FutureError{Code:promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if event.ActorType() != actors.CheckoutActor {
		logger.Err("Received actorType of event is not CheckoutActor")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Error:promise.FutureError{Code:promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if event.ActorAction().ActionEnums()[0] != checkout_action.NewOrderAction {
		logger.Err("Received actorAction of event is not NewOrderAction")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Error:promise.FutureError{Code:promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrderRequest, ok := event.Data().(pb.NewOrderRequest)
	if ok != true {
		logger.Err("Received data of event is not NewOrderRequest")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Error:promise.FutureError{Code:promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	value, err := global.Singletons.Converter.Map(newOrderRequest, entities.Order{})
	if err != nil {
		logger.Err("Received NewOrderRequest invalid")
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Error:promise.FutureError{Code:promise.BadRequest, Reason:"Received NewOrderRequest invalid"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	order := value.(entities.Order)
	global.Singletons.OrderRepository.Save(order)
}