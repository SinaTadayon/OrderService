package state_21

import (
	"bytes"
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"text/template"
	"time"
)

const (
	stepName  string = "Canceled_By_Seller"
	stepIndex int    = 21
)

type canceledBySellerState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledBySellerState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &canceledBySellerState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &canceledBySellerState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state canceledBySellerState) Process(ctx context.Context, iFrame frame.IFrame) {
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

		pkgItem, ok := iFrame.Body().Content().(*entities.PackageItem)
		if !ok {
			logger.Err("Process() => iFrame.Body().Content() is nil, orderId: %d, pid: %d, sid: %d, %s state ",
				subpackages[0].OrderId, subpackages[0].PId, subpackages[0].SId, state.Name())
			return
		}

		var buyerNotificationAction *entities.Action
		var templateData struct {
			OrderId  uint64
			ShopName string
		}

		templateData.OrderId = pkgItem.OrderId
		templateData.ShopName = pkgItem.ShopName

		smsTemplate, err := template.New("SMS").Parse(app.Globals.Config.App.OrderNotifyBuyerCanceledBySellerState)
		if err != nil {
			logger.Err("Process() => smsTemplate.Parse failed, state: %s, orderId: %d, message: %s, err: %s",
				state.Name(), pkgItem.OrderId, app.Globals.Config.App.OrderNotifyBuyerCanceledBySellerState, err)
		} else {
			var buf bytes.Buffer
			err = smsTemplate.Execute(&buf, templateData)
			if err != nil {
				logger.Err("Process() => smsTemplate.Execute failed, state: %s, orderId: %d, message: %s, err: %s",
					state.Name(), app.Globals.Config.App.OrderNotifyBuyerCanceledBySellerState, pkgItem.OrderId, err)
			} else {
				buyerNotify := notify_service.SMSRequest{
					Phone: pkgItem.ShippingAddress.Mobile,
					Body:  buf.String(),
				}

				buyerFutureData := app.Globals.NotifyService.NotifyBySMS(ctx, buyerNotify).Get()
				if buyerFutureData.Error() != nil {
					logger.Err("Process() => NotifyService.NotifyBySMS failed, request: %v, state: %s, orderId: %d, pid: %d, sids: %v, error: %s",
						buyerNotify, state.Name(), pkgItem.OrderId, pkgItem.PId, sids, buyerFutureData.Error().Reason())
					buyerNotificationAction = &entities.Action{
						Name:      system_action.BuyerNotification.ActionName(),
						Type:      "",
						UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
						UTP:       actions.System.ActionName(),
						Perm:      "",
						Priv:      "",
						Policy:    "",
						Result:    string(states.ActionFail),
						Reasons:   nil,
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					}
				} else {
					logger.Audit("Process() => NotifyService.NotifyBySMS success, state: %s, orderId: %d, pid: %d, sids: %v",
						state.Name(), pkgItem.OrderId, pkgItem.PId, sids)
					buyerNotificationAction = &entities.Action{
						Name:      system_action.BuyerNotification.ActionName(),
						Type:      "",
						UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
						UTP:       actions.System.ActionName(),
						Perm:      "",
						Priv:      "",
						Policy:    "",
						Result:    string(states.ActionSuccess),
						Reasons:   nil,
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					}
				}
			}
		}

		nextToAction := &entities.Action{
			Name:      system_action.NextToState.ActionName(),
			Type:      "",
			UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
			UTP:       actions.System.ActionName(),
			Perm:      "",
			Priv:      "",
			Policy:    "",
			Result:    string(states.ActionSuccess),
			Reasons:   nil,
			Data:      nil,
			CreatedAt: time.Now().UTC(),
			Extended:  nil,
		}

		// TODO optimize it repository update
		var updatedSubpackages = make([]*entities.Subpackage, 0, len(subpackages))
		for _, subpackage := range subpackages {
			state.UpdateSubPackage(ctx, subpackage, buyerNotificationAction)
			state.UpdateSubPackage(ctx, subpackage, nextToAction)
			subPkgUpdated, err := app.Globals.SubPkgRepository.Update(ctx, *subpackage)
			if err != nil {
				logger.Err("Process() => SubPkgRepository.Update in %s state failed, orderId: %d, pid: %d, sid: %d, error: %s",
					state.Name(), subpackage.OrderId, subpackage.PId, subpackage.SId, err.Error())
			} else {
				//logger.Audit("Process() => Status of subpackages update to %s state, orderId: %d, pid: %d, sids: %v",
				//	state.Name(), subpackage.OrderId, subpackage.PId, subpackage.SId)
				updatedSubpackages = append(updatedSubpackages, subPkgUpdated)
			}
		}

		logger.Audit("%s state success, orderId: %d, pid: %d, sid: %v", state.Name(), updatedSubpackages[0].OrderId, updatedSubpackages[0].PId, sids)
		state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetSubpackages(updatedSubpackages).Build())

	} else {
		logger.Err("iFrame.Header() not a subpackage , state: %s iframe: %v", state.Name(), iFrame)
	}
}
