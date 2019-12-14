package state_36

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"time"
)

const (
	stepName  string = "Delivery_Failed"
	stepIndex int    = 36
)

type deliveryFailedState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &deliveryFailedState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &deliveryFailedState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &deliveryFailedState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state deliveryFailedState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderSubpackages)) {
		subpackages, ok := iFrame.Header().Value(string(frame.HeaderSubpackages)).([]*entities.Subpackage)
		if !ok {
			logger.Err("iFrame.Header() not a subpackages, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		sids, ok := iFrame.Header().Value(string(frame.HeaderSIds)).([]uint64)
		if !ok {
			logger.Err("iFrame.Header() not a sids, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		nextToAction := &entities.Action{
			Name:      system_action.NextToState.ActionName(),
			Type:      actions.System.ActionName(),
			Result:    string(states.ActionSuccess),
			Reasons:   nil,
			CreatedAt: time.Now().UTC(),
		}

		var updatedSubpackages = make([]*entities.Subpackage, 0, len(subpackages))
		for _, subpackage := range subpackages {
			state.UpdateSubPackage(ctx, subpackage, nextToAction)
			subPkgUpdated, err := app.Globals.SubPkgRepository.Update(ctx, *subpackage)
			if err != nil {
				logger.Err("SubPkgRepository.Update in %s state failed, orderId: %d, pid: %d, sid: %d, error: %s",
					state.Name(), subpackage.OrderId, subpackage.PId, subpackage.SId, err.Error())
				return
			}
			updatedSubpackages = append(updatedSubpackages, subPkgUpdated)
		}

		logger.Audit("%s state success, orderId: %d, pid: %d, sid: %v", state.Name(), updatedSubpackages[0].OrderId, updatedSubpackages[0].PId, sids)
		state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetSubpackages(updatedSubpackages).Build())

	} else {
		logger.Err("iFrame.Header() not a subpackage , state: %s iframe: %v", state.Name(), iFrame)
	}
}
