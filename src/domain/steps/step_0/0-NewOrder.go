package new_order_step

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	checkout_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/checkout"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	listener_state "gitlab.faza.io/order-project/order-service/domain/states/listener"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	pb "gitlab.faza.io/protos/order"
	message "gitlab.faza.io/protos/order/general"
	"time"
)

const (
	stepName string 	= "New_Order"
	stepIndex int		= 0
)

type newOrderProcessingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &newOrderProcessingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (newOrderProcessing newOrderProcessingStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	var newOrderRequest pb.NewOrderRequest

	if err := ptypes.UnmarshalAny(request.Data, &newOrderRequest); err != nil {
		logger.Err("Could not unmarshal NewOrderRequest from anything field, error: %s, request: %v", err, request)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.BadRequest, Reason:"Invalid NewOrderRequest"}}
		close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	timestamp, err := ptypes.Timestamp(request.Time)
	if err != nil {
		logger.Err("timestamp of NewOrderRequest invalid, error: %s, newOrderRequest: %v", err, newOrderRequest)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.BadRequest, Reason:"Invalid Request Timestamp"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	newOrderEvent := events.New(actors.CheckoutActor, checkout_action.NewOf(checkout_action.NewOrderAction),
		newOrderRequest, timestamp)

	checkoutState, ok := newOrderProcessing.StatesMap()[0].(listener_state.IListenerState)
	if ok != true || checkoutState.ActorType() != actors.CheckoutActor {
		logger.Err("checkout state doesn't exist in index 0 of statesMap, newOrderRequest: %v", newOrderRequest)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	ctx = context.WithValue(ctx, global.CtxStepName, newOrderProcessing.Name())
	ctx = context.WithValue(ctx, global.CtxStepIndex, newOrderProcessing.Index())
	ctx = context.WithValue(ctx, global.CtxStepTimestamp, time.Now().UTC())
	return checkoutState.ActionListener(ctx, newOrderEvent, nil)
}

func (newOrderProcessing newOrderProcessingStep) ProcessOrder(ctx context.Context, order entities.Order) promise.IPromise {
	panic("implementation required")
}

