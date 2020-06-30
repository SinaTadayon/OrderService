package state_56

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"time"
)

const (
	stepName  string = "Return_Delivery_Failed"
	stepIndex int    = 56
)

type returnDeliveryFailedState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnDeliveryFailedState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnDeliveryFailedState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &returnDeliveryFailedState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state returnDeliveryFailedState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderSIds)) {
		sids, ok := iFrame.Header().Value(string(frame.HeaderSIds)).([]uint64)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("iFrame.Header() doesn't have a sids header",
				"fn", "Process",
				"state", state.Name(),
				"iframe", iFrame)
			return
		}

		if iFrame.Body().Content() == nil {
			app.Globals.Logger.FromContext(ctx).Error("content of iFrame.Body() is nil",
				"fn", "Process",
				"state", state.Name(),
				"iframe", iFrame)
			return
		}

		pkgItem, ok := iFrame.Body().Content().(*entities.PackageItem)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("content of iFrame.Body() is not PackageItem",
				"fn", "Process",
				"state", state.Name(),
				"sids", sids,
				"iframe", iFrame)
			return
		}

		nextToAction := &entities.Action{
			Name:      system_action.Close.ActionName(),
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

		for i := 0; i < len(sids); i++ {
			for j := 0; j < len(pkgItem.Subpackages); j++ {
				if pkgItem.Subpackages[j].SId == sids[i] {
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], nextToAction)
				}
			}
		}

		app.Globals.Logger.FromContext(ctx).Debug("set status of subpackages success",
			"fn", "Process",
			"state", state.Name(),
			"oid", pkgItem.OrderId,
			"pid", pkgItem.PId,
			"sids", sids)
		state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(pkgItem).Build())

	} else {
		app.Globals.Logger.FromContext(ctx).Error("Frame Header/Body Invalid",
			"fn", "Process",
			"state", state.Name(),
			"iframe", iFrame)

		if iFrame.Header().KeyExists(string(frame.HeaderFuture)) {
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.BadRequest, "Request Invalid", errors.New("Request Invalid")).Send()
		}
	}
}
