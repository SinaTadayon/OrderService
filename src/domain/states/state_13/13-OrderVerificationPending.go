package state_13

import (
	"context"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"time"
)

const (
	stepName  string = "Order_Verification_Pending"
	stepIndex int    = 13
)

type orderPaymentVerificationState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &orderPaymentVerificationState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &orderPaymentVerificationState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &orderPaymentVerificationState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state orderPaymentVerificationState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) && iFrame.Body().Content() != nil {
		order, ok := iFrame.Body().Content().(*entities.Order)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("Content of frame body isn't an order",
				"fn", "Process",
				"state", state.Name(),
				"oid", iFrame.Header().Value(string(frame.HeaderOrderId)),
				"content", iFrame.Body().Content())
			return
		}

		orderVerifyAction := &entities.Action{
			Name:      system_action.Success.ActionName(),
			Type:      "",
			UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
			UTP:       actions.System.ActionName(),
			Perm:      "",
			Priv:      "",
			Policy:    "",
			Result:    string(states.ActionSuccess),
			Reasons:   nil,
			Note:      "",
			Data:      nil,
			CreatedAt: time.Now().UTC(),
			Extended:  nil,
		}

		state.UpdateOrderAllStatus(ctx, order, states.OrderInProgressStatus, states.PackageInProgressStatus, orderVerifyAction)
		//orderUpdated, err := app.Globals.OrderRepository.Save(ctx, *order)
		//if err != nil {
		//	logger.Err("OrderRepository.Save in %s state failed, orderId: %d, error: %v", state.Name(), order.OrderId, err)
		//} else {
		//	logger.Audit("Order Verification success, orderId: %d", order.OrderId)
		//	successAction := state.GetAction(system_action.Success.ActionName())
		//	state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(orderUpdated).Build())
		//}
		app.Globals.Logger.FromContext(ctx).Debug("Order state of all subpackages update",
			"fn", "Process",
			"state", state.Name(),
			"oid", order.OrderId)
		successAction := state.GetAction(system_action.Success.ActionName())
		state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
	} else {
		app.Globals.Logger.FromContext(ctx).Error("Frame Header/Body Invalid",
			"fn", "Process",
			"state", state.Name(),
			"iframe", iFrame)
	}
}
