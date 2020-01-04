package state_22

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
	"strconv"
	"text/template"
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
	if iFrame.Header().KeyExists(string(frame.HeaderSIds)) {
		sids, ok := iFrame.Header().Value(string(frame.HeaderSIds)).([]uint64)
		if !ok {
			logger.Err("Process() => iFrame.Header() not a sids, state: %s, frame: %v", state.Name(), iFrame)
			return
		}

		if iFrame.Body().Content() == nil {
			logger.Err("Process() => iFrame.Body().Content() is nil, state: %s, frame: %v", state.Name(), iFrame)
			return
		}

		pkgItem, ok := iFrame.Body().Content().(*entities.PackageItem)
		if !ok {
			logger.Err("Process() => pkgItem in iFrame.Body().Content() is not found, %s state, sids: %v, frame: %v",
				state.Name(), sids, iFrame)
			return
		}

		var buyerNotificationAction = &entities.Action{
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

		var sellerNotificationAction = &entities.Action{
			Name:      system_action.SellerNotification.ActionName(),
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

		smsTemplate, err := template.New("SMS").Parse(app.Globals.SMSTemplate.OrderNotifyBuyerCanceledByBuyerState)
		if err != nil {
			logger.Err("Process() => smsTemplate.Parse failed, state: %s, orderId: %d, message: %s, err: %s",
				state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifyBuyerCanceledByBuyerState, err)
		} else {
			var buf bytes.Buffer
			err = smsTemplate.Execute(&buf, pkgItem.OrderId)
			newBuf := bytes.NewBuffer(bytes.Replace(buf.Bytes(), []byte("\\n"), []byte{10}, -1))
			if err != nil {
				logger.Err("Process() => smsTemplate.Execute failed, state: %s, orderId: %d, message: %s, err: %s",
					state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifyBuyerCanceledByBuyerState, err)
			} else {
				buyerNotify := notify_service.SMSRequest{
					Phone: pkgItem.ShippingAddress.Mobile,
					Body:  newBuf.String(),
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

		futureData := app.Globals.UserService.GetSellerProfile(ctx, strconv.Itoa(int(pkgItem.PId))).Get()
		if futureData.Error() != nil {
			logger.Err("Process() => UserService.GetSellerProfile failed, send sms message failed, state: %s, orderId: %d, pid: %d, sids: %v, error: %s",
				state.Name(), pkgItem.OrderId, pkgItem.PId, sids, futureData.Error().Reason())
		} else {
			if futureData.Data() != nil {
				sellerProfile := futureData.Data().(*entities.SellerProfile)

				smsTemplate, err := template.New("SMS").Parse(app.Globals.SMSTemplate.OrderNotifySellerCanceledByBuyerState)
				if err != nil {
					logger.Err("Process() => smsTemplate.Parse failed, state: %s, orderId: %d, message: %s, err: %s",
						state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifySellerCanceledByBuyerState, err)
				} else {
					var buf bytes.Buffer
					err = smsTemplate.Execute(&buf, pkgItem.OrderId)
					newBuf := bytes.NewBuffer(bytes.Replace(buf.Bytes(), []byte("\\n"), []byte{10}, -1))
					if err != nil {
						logger.Err("Process() => smsTemplate.Execute failed, state: %s, orderId: %d, message: %s, err: %s",
							state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifySellerCanceledByBuyerState, err)
					} else {
						sellerNotify := notify_service.SMSRequest{
							Phone: sellerProfile.GeneralInfo.MobilePhone,
							Body:  newBuf.String(),
						}
						sellerFutureData := app.Globals.NotifyService.NotifyBySMS(ctx, sellerNotify).Get()
						if sellerFutureData.Error() != nil {
							logger.Err("Process() => NotifyService.NotifyBySMS failed, request: %v, state: %s, orderId: %d, pid: %d, sids: %v, error: %s",
								sellerNotify, state.Name(), pkgItem.OrderId, pkgItem.PId, sids, sellerFutureData.Error().Reason())
							sellerNotificationAction = &entities.Action{
								Name:      system_action.SellerNotification.ActionName(),
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
							logger.Audit("Process() => NotifyService.NotifyBySMS success, sellerNotify: %v, state: %s, orderId: %d, pid: %d, sids: %v",
								sellerNotify, state.Name(), pkgItem.OrderId, pkgItem.PId, sids)
							sellerNotificationAction = &entities.Action{
								Name:      system_action.SellerNotification.ActionName(),
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
			} else {
				logger.Err("Process() => UserService.GetSellerProfile futureData.Data() is nil, send sms message failed, state: %s, orderId: %d, pid: %d, sids: %v",
					state.Name(), pkgItem.OrderId, pkgItem.PId, sids)
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

		for i := 0; i < len(sids); i++ {
			for j := 0; j < len(pkgItem.Subpackages); j++ {
				if pkgItem.Subpackages[j].SId == sids[i] {
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], sellerNotificationAction)
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], buyerNotificationAction)
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], nextToAction)
				}
			}
		}

		pkgItemUpdated, err := app.Globals.PkgItemRepository.Update(ctx, *pkgItem)
		if err != nil {
			logger.Err("Process() => PkgItemRepository.Update failed, state: %s, orderId: %d, pid: %d, sids: %v, error: %s", state.Name(),
				pkgItem.OrderId, pkgItem.PId, sids, err.Error())
			return
		}

		logger.Audit("Process() => Status of subpackages update success, state: %s ,orderId: %d, pid: %d, sids: %v", state.Name(), pkgItem.OrderId, pkgItem.PId, sids)
		state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(pkgItemUpdated).Build())
	} else {
		logger.Err("iFrame.Header() not a subpackage or sellerId not found, state: %s iframe: %v", state.Name(), iFrame)
	}
}
