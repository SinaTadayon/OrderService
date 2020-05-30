package state_80

import (
	"bytes"
	"context"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"text/template"
	"time"
)

const (
	stepName  string = "Pay_To_Buyer"
	stepIndex int    = 80
)

type payToBuyerState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &payToBuyerState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &payToBuyerState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &payToBuyerState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state payToBuyerState) Process(ctx context.Context, iFrame frame.IFrame) {
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

		event, ok := iFrame.Header().Value(string(frame.HeaderEvent)).(events.IEvent)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("received frame doesn't have a event",
				"fn", "Process",
				"state", state.Name(),
				"sids", sids,
				"iframe", iFrame)
			return
		}

		if states.FromIndex(event.StateIndex()) == states.ReturnRejected ||
			states.FromIndex(event.StateIndex()) == states.ReturnDelivered {

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
				Note:      "",
				Data:      nil,
				CreatedAt: time.Now().UTC(),
				Extended:  nil,
			}

			var message string
			if states.FromIndex(event.StateIndex()) == states.ReturnDelivered {
				message = app.Globals.SMSTemplate.OrderNotifyBuyerReturnDeliveredToPayToBuyerState
			} else if states.FromIndex(event.StateIndex()) == states.ReturnRejected {
				message = app.Globals.SMSTemplate.OrderNotifyBuyerReturnRejectedToPayToBuyerState
			}

			var templateData struct {
				OrderId  uint64
				ShopName string
			}

			templateData.OrderId = pkgItem.OrderId
			templateData.ShopName = pkgItem.ShopName

			smsTemplate, err := template.New("SMS").Parse(message)
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("smsTemplate.Parse failed",
					"fn", "Process",
					"state", state.Name(),
					"oid", pkgItem.OrderId,
					"pid", pkgItem.PId,
					"sids", sids,
					"message", message,
					"error", err)
			} else {
				var buf bytes.Buffer
				err = smsTemplate.Execute(&buf, templateData)
				newBuf := bytes.NewBuffer(bytes.Replace(buf.Bytes(), []byte("\\n"), []byte{10}, -1))
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("smsTemplate.Execute failed",
						"fn", "Process",
						"state", state.Name(),
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sids", sids,
						"message", message,
						"error", err)
				} else {
					buyerNotify := notify_service.SMSRequest{
						Phone: pkgItem.ShippingAddress.Mobile,
						Body:  newBuf.String(),
						User:  notify_service.BuyerUser,
					}

					buyerFutureData := app.Globals.NotifyService.NotifyBySMS(ctx, buyerNotify).Get()
					if buyerFutureData.Error() != nil {
						app.Globals.Logger.FromContext(ctx).Error("NotifyService.NotifyBySMS failed",
							"fn", "Process",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sids", sids,
							"request", buyerNotify,
							"error", buyerFutureData.Error().Reason())
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
							Note:      "",
							Data:      nil,
							CreatedAt: time.Now().UTC(),
							Extended:  nil,
						}
					} else {
						app.Globals.Logger.FromContext(ctx).Debug("NotifyService.NotifyBySMS success",
							"fn", "Process",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"request", buyerNotify,
							"sids", sids)
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
							Note:      "",
							Data:      nil,
							CreatedAt: time.Now().UTC(),
							Extended:  nil,
						}
					}
				}
			}

			for i := 0; i < len(sids); i++ {
				for j := 0; j < len(pkgItem.Subpackages); j++ {
					if pkgItem.Subpackages[j].SId == sids[i] {
						state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], buyerNotificationAction)
					}
				}
			}
		}

		stockAction := state.releasedStock(ctx, sids, pkgItem)
		for i := 0; i < len(sids); i++ {
			for j := 0; j < len(pkgItem.Subpackages); j++ {
				if pkgItem.Subpackages[j].SId == sids[i] {
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], stockAction)
				}
			}
		}

		order, e := app.Globals.OrderRepository.FindById(ctx, pkgItem.OrderId)
		if e != nil {
			app.Globals.Logger.FromContext(ctx).Error("OrderRepository.FindById failed",
				"fn", "Process",
				"state", state.Name(),
				"oid", pkgItem.OrderId,
				"pid", pkgItem.PId,
				"sids", sids,
				"error", e)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.ErrorCode(e.Code()), e.Message(), e.Reason()).
				Send()
			return
		}

		var findFlag = true
		for i := 0; i < len(order.Packages); i++ {
			findFlag = true
			if order.Packages[i].PId == pkgItem.PId {
				order.Packages[i] = pkgItem
			}

			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if order.Packages[i].Subpackages[j].Status != states.PayToBuyer.StateName() &&
					order.Packages[i].Subpackages[j].Status != states.PayToSeller.StateName() {
					findFlag = false
					break
				}
			}

			if findFlag {
				state.SetPkgStatus(ctx, order.Packages[i], states.PackageClosedStatus)
				app.Globals.Logger.FromContext(ctx).Debug("set pkgItem status to closed",
					"fn", "Process",
					"state", state.Name(),
					"oid", order.Packages[i].OrderId,
					"pid", order.Packages[i].PId)
			}
		}

		findFlag = true
		for i := 0; i < len(order.Packages); i++ {
			if order.Packages[i].Status != string(states.PackageClosedStatus) {
				findFlag = false
				break
			}
		}

		if findFlag {
			state.SetOrderStatus(ctx, order, states.OrderClosedStatus)
		}

		if states.OrderStatus(order.Status) == states.OrderClosedStatus {
			_, err := app.Globals.OrderRepository.Save(ctx, *order)
			if err != nil {
				state.rollbackStock(ctx, sids, pkgItem)
				app.Globals.Logger.FromContext(ctx).Error("update order after seller finance recalculation failed",
					"fn", "Process",
					"state", state.Name(),
					"oid", order.OrderId,
					"error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).
					Send()
				return
			}

			app.Globals.Logger.FromContext(ctx).Debug("update order after seller finance recalculation success",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId)
		} else {
			_, e := app.Globals.PkgItemRepository.Update(ctx, *pkgItem)
			if e != nil {
				state.rollbackStock(ctx, sids, pkgItem)
				app.Globals.Logger.FromContext(ctx).Error("update package after seller finance recalculation success",
					"fn", "Process",
					"state", state.Name(),
					"oid", pkgItem.OrderId,
					"pid", pkgItem.PId,
					"sids", sids,
					"error", e)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.ErrorCode(e.Code()), e.Message(), e.Reason()).
					Send()
				return
			}

			app.Globals.Logger.FromContext(ctx).Debug("update order after finance recalculation success",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId)
		}

		response := events.ActionResponse{
			OrderId: pkgItem.OrderId,
			SIds:    sids,
		}

		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()
		app.Globals.Logger.FromContext(ctx).Debug("action success",
			"fn", "Process",
			"state", state.Name(),
			"oid", pkgItem.OrderId,
			"pid", pkgItem.PId,
			"sids", sids,
			"event", event.Action())
	} else {
		app.Globals.Logger.FromContext(ctx).Error("Frame Header/Body Invalid",
			"fn", "Process",
			"state", state.Name(),
			"iframe", iFrame)
	}
}

func (state payToBuyerState) releasedStock(ctx context.Context, sids []uint64, pkgItem *entities.PackageItem) *entities.Action {

	var stockAction *entities.Action
	for _, sid := range sids {
		for i := 0; i < len(pkgItem.Subpackages); i++ {
			if sid != pkgItem.Subpackages[i].SId {
				continue
			}

			result := true
			stockActionDataList := make([]entities.StockActionData, 0, 32)
			for z := 0; z < len(pkgItem.Subpackages[i].Items); z++ {
				item := pkgItem.Subpackages[i].Items[z]
				requestStock := stock_service.RequestStock{
					InventoryId: item.InventoryId,
					Count:       int(item.Quantity),
				}

				iFuture := app.Globals.StockService.SingleStockAction(ctx, requestStock, pkgItem.Subpackages[i].OrderId,
					system_action.New(system_action.StockRelease))

				futureData := iFuture.Get()
				if futureData.Error() != nil {
					result = false
					if futureData.Data() != nil {
						response := futureData.Data().(stock_service.ResponseStock)
						actionData := entities.StockActionData{
							InventoryId: response.InventoryId,
							Quantity:    response.Count,
							Result:      response.Result,
						}
						stockActionDataList = append(stockActionDataList, actionData)
						app.Globals.Logger.FromContext(ctx).Error("Release stock from stockService failed",
							"fn", "releasedStock",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", sid,
							"actionData", actionData,
							"error", futureData.Error())

					} else {
						actionData := entities.StockActionData{
							InventoryId: requestStock.InventoryId,
							Quantity:    requestStock.Count,
							Result:      false,
						}
						stockActionDataList = append(stockActionDataList, actionData)
						app.Globals.Logger.FromContext(ctx).Error("Release stock from stockService failed",
							"fn", "releasedStock",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", sid,
							"stockAction", actionData,
							"error", futureData.Error())
					}
				} else {
					response := futureData.Data().(stock_service.ResponseStock)
					actionData := entities.StockActionData{
						InventoryId: response.InventoryId,
						Quantity:    response.Count,
						Result:      response.Result,
					}
					stockActionDataList = append(stockActionDataList, actionData)
					app.Globals.Logger.FromContext(ctx).Info("Release stock success",
						"fn", "releasedStock",
						"state", state.Name(),
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", sid,
						"stockAction", actionData)
				}
			}

			if !result {
				stockAction = &entities.Action{
					Name:      system_action.StockRelease.ActionName(),
					Type:      "",
					UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
					UTP:       actions.System.ActionName(),
					Perm:      "",
					Priv:      "",
					Policy:    "",
					Result:    string(states.ActionFail),
					Reasons:   nil,
					Data:      map[string]interface{}{"stockActionData": stockActionDataList},
					CreatedAt: time.Now().UTC(),
					Extended:  nil,
				}
			} else {
				stockAction = &entities.Action{
					Name:      system_action.StockRelease.ActionName(),
					Type:      "",
					UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
					UTP:       actions.System.ActionName(),
					Perm:      "",
					Priv:      "",
					Policy:    "",
					Result:    string(states.ActionSuccess),
					Reasons:   nil,
					Data:      map[string]interface{}{"stockActionData": stockActionDataList},
					CreatedAt: time.Now().UTC(),
					Extended:  nil,
				}
			}
		}
	}

	return stockAction
}

func (state payToBuyerState) rollbackStock(ctx context.Context, sids []uint64, pkgItem *entities.PackageItem) {

	for _, sid := range sids {
		for i := 0; i < len(pkgItem.Subpackages); i++ {
			if sid != pkgItem.Subpackages[i].SId {
				continue
			}

			//result := true
			stockActionDataList := make([]entities.StockActionData, 0, 32)
			for z := 0; z < len(pkgItem.Subpackages[i].Items); z++ {
				item := pkgItem.Subpackages[i].Items[z]
				requestStock := stock_service.RequestStock{
					InventoryId: item.InventoryId,
					Count:       int(item.Quantity),
				}

				iFuture := app.Globals.StockService.SingleStockAction(ctx, requestStock, pkgItem.Subpackages[i].OrderId,
					system_action.New(system_action.StockReserve))

				futureData := iFuture.Get()
				if futureData.Error() != nil {
					//result = false
					if futureData.Data() != nil {
						response := futureData.Data().(stock_service.ResponseStock)
						actionData := entities.StockActionData{
							InventoryId: response.InventoryId,
							Quantity:    response.Count,
							Result:      response.Result,
						}
						stockActionDataList = append(stockActionDataList, actionData)
						app.Globals.Logger.FromContext(ctx).Error("rollback released to reserved stock from stockService failed",
							"fn", "rollbackStock",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", sid,
							"actionData", actionData,
							"error", futureData.Error())

					} else {
						actionData := entities.StockActionData{
							InventoryId: requestStock.InventoryId,
							Quantity:    requestStock.Count,
							Result:      false,
						}
						stockActionDataList = append(stockActionDataList, actionData)
						app.Globals.Logger.FromContext(ctx).Error("rollback released to reserved stock from stockService failed",
							"fn", "rollbackStock",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", sid,
							"stockAction", actionData,
							"error", futureData.Error())
					}
				} else {
					response := futureData.Data().(stock_service.ResponseStock)
					actionData := entities.StockActionData{
						InventoryId: response.InventoryId,
						Quantity:    response.Count,
						Result:      response.Result,
					}
					stockActionDataList = append(stockActionDataList, actionData)
					app.Globals.Logger.FromContext(ctx).Info("rollback released to reserved stock success",
						"fn", "rollbackStock",
						"state", state.Name(),
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", sid,
						"stockAction", actionData)
				}
			}
		}
	}
}
