package state_22

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
	stepName  string = "Canceled_By_Buyer"
	stepIndex int    = 22
)

type canceledByBuyerState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledByBuyerState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledByBuyerState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &canceledByBuyerState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state canceledByBuyerState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderSubpackage)) {
		subpkg, ok := iFrame.Header().Value(string(frame.HeaderSubpackage)).(*entities.Subpackage)
		if !ok {
			logger.Err("iFrame.Header() not a subpackage, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		nextToAction := &entities.Action{
			Name:      system_action.NextToState.ActionName(),
			Type:      actions.System.ActionName(),
			Result:    string(states.ActionSuccess),
			Reasons:   nil,
			CreatedAt: time.Now().UTC(),
		}

		state.UpdateSubPackage(ctx, subpkg, nextToAction)
		subPkgUpdated, err := global.Singletons.SubPkgRepository.Update(ctx, *subpkg)
		if err != nil {
			logger.Err("SubPkgRepository.Update in %s state failed, orderId: %d, sellerId: %d, itemId: %d, error: %s",
				state.Name(), subpkg.OrderId, subpkg.SellerId, subpkg.ItemId, err.Error())
		} else {
			logger.Audit("Cancel by seller success, orderId: %d, sellerId: %d, itemId: %d", subpkg.OrderId, subpkg.SellerId, subpkg.ItemId)
			state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(subPkgUpdated).Build())
		}
	} else {
		logger.Err("iFrame.Header() not a subpackage or sellerId not found, state: %s iframe: %v", state.Name(), iFrame)
	}
}
