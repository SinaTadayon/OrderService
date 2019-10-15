package new_order_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "New_Order"
	stepIndex int		= 0
)

type newOrderProcessingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &newOrderProcessingStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &newOrderProcessingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (newOrderProcessing newOrderProcessingStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (newOrderProcessing newOrderProcessingStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}

