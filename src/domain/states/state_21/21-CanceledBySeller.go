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
	if iFrame.Header().KeyExists(string(frame.HeaderSIds)) {
		//subpackages, ok := iFrame.Header().Value(string(frame.HeaderSubpackages)).([]*entities.Subpackage)
		//if !ok {
		//	logger.Err("iFrame.Header() not a subpackages, frame: %v, %s state ", iFrame, state.Name())
		//	return
		//}

		//event, ok := iFrame.Header().Value(string(frame.HeaderEvent)).(events.IEvent)
		//if !ok {
		//	logger.Err("Process() => received frame doesn't have a event, state: %s, frame: %v", state.String(), iFrame)
		//	future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
		//		SetError(future.InternalError, "Unknown Err", nil).Send()
		//	return
		//}

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

		var templateData struct {
			OrderId  uint64
			ShopName string
		}

		templateData.OrderId = pkgItem.OrderId
		templateData.ShopName = pkgItem.ShopName

		smsTemplate, err := template.New("SMS").Parse(app.Globals.SMSTemplate.OrderNotifyBuyerCanceledBySellerState)
		if err != nil {
			logger.Err("Process() => smsTemplate.Parse failed, state: %s, orderId: %d, message: %s, err: %s",
				state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifyBuyerCanceledBySellerState, err)
		} else {
			var buf bytes.Buffer
			err = smsTemplate.Execute(&buf, templateData)
			newBuf := bytes.NewBuffer(bytes.Replace(buf.Bytes(), []byte("\\n"), []byte{10}, -1))
			if err != nil {
				logger.Err("Process() => smsTemplate.Execute failed, state: %s, orderId: %d, message: %s, err: %s",
					state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifyBuyerCanceledBySellerState, err)
			} else {
				buyerNotify := notify_service.SMSRequest{
					Phone: pkgItem.ShippingAddress.Mobile,
					Body:  newBuf.String(),
					User:  notify_service.BuyerUser,
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

		//var updatedSubpackages = make([]*entities.Subpackage, 0, len(subpackages))
		for i := 0; i < len(sids); i++ {
			for j := 0; j < len(pkgItem.Subpackages); j++ {
				if pkgItem.Subpackages[j].SId == sids[i] {
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], buyerNotificationAction)
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], nextToAction)
				}
			}
		}

		//pkgItemUpdated, err := app.Globals.PkgItemRepository.UpdateWithUpsert(ctx, *pkgItem)
		//if err != nil {
		//	logger.Err("Process() => PkgItemRepository.Update failed, state: %s, orderId: %d, pid: %d, sids: %v, error: %v", state.Name(),
		//		pkgItem.OrderId, pkgItem.PId, sids, err)
		//	return
		//}
		//
		//response := events.ActionResponse{
		//	OrderId: pkgItem.OrderId,
		//	SIds:    sids,
		//}
		//
		//future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()

		logger.Audit("Process() => set status of subpackages success, state: %s, orderId: %d, pid: %d, sids: %v",
			state.Name(), pkgItem.OrderId, pkgItem.PId, sids)
		state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(pkgItem).Build())

	} else {
		logger.Err("iFrame.Header() not a subpackage , state: %s iframe: %v", state.Name(), iFrame)
	}
}
