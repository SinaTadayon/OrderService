package state_15

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"time"
)

const (
	stepName  string = "Order_Verification_Failed"
	stepIndex int    = 15
)

type orderVerificationFailed struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &orderVerificationFailed{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &orderVerificationFailed{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &orderVerificationFailed{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state orderVerificationFailed) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) && iFrame.Body().Content() != nil {
		order, ok := iFrame.Body().Content().(*entities.Order)
		if !ok {
			logger.Err("iFrame.Body().Content() not a order, orderId: %d, %s state ",
				iFrame.Header().Value(string(frame.HeaderOrderId)), state.Name())
			return
		}

		orderVerifyAction := &entities.Action{
			Name:      system_action.Fail.ActionName(),
			Type:      actions.System.ActionName(),
			Data:      nil,
			Result:    string(states.ActionSuccess),
			Reasons:   nil,
			CreatedAt: time.Now().UTC(),
		}

		state.UpdateOrderAllSubPkg(ctx, order, orderVerifyAction)
		orderUpdated, err := global.Singletons.OrderRepository.Save(ctx, *order)
		if err != nil {
			logger.Err("OrderRepository.Save in %s state failed, orderId: %d, error: %s", state.Name(), order.OrderId, err.Error())
		} else {
			logger.Audit("Order Verification Fail, orderId: %d", order.OrderId)
			state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(orderUpdated).Build())
		}
	} else {
		logger.Err("HeaderOrderId of iFrame.Header not found and content of iFrame.Body() not set, state: %s iframe: %v", state.Name(), iFrame)
	}
}
