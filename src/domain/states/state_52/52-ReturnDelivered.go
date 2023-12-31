package state_52

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"strconv"
	"text/template"
	"time"
)

const (
	stepName  string = "Return_Delivered"
	stepIndex int    = 52
)

type returnDeliveredState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnDeliveredState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &returnDeliveredState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &returnDeliveredState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state returnDeliveredState) Process(ctx context.Context, iFrame frame.IFrame) {
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

		if states.FromIndex(event.StateIndex()) == states.ReturnDeliveryPending ||
			states.FromIndex(event.StateIndex()) == states.ReturnDeliveryDelayed {

			var message string
			if states.FromIndex(event.StateIndex()) == states.ReturnDeliveryPending {
				message = app.Globals.SMSTemplate.OrderNotifyBuyerReturnDeliveryPendingToReturnDeliveredState
			} else if states.FromIndex(event.StateIndex()) == states.ReturnDeliveryDelayed {
				message = app.Globals.SMSTemplate.OrderNotifyBuyerReturnDeliveryDelayedToReturnDeliveredState
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
		}

		var expireTime time.Time
		timeUnit := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig].(string)
		if timeUnit == app.DurationTimeUnit {
			value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnDeliveredStateConfig].(time.Duration)
			expireTime = time.Now().UTC().Add(value)
		} else {
			value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerReturnDeliveredStateConfig].(int)
			if timeUnit == string(app.HourTimeUnit) {
				expireTime = time.Now().UTC().Add(
					time.Hour*time.Duration(value) +
						time.Minute*time.Duration(0) +
						time.Second*time.Duration(0))
			} else {
				expireTime = time.Now().UTC().Add(
					time.Hour*time.Duration(0) +
						time.Minute*time.Duration(value) +
						time.Second*time.Duration(0))
			}
		}

		app.Globals.Logger.FromContext(ctx).Debug("scheduler expireTime",
			"fn", "Process",
			"state", state.Name(),
			"oid", pkgItem.OrderId,
			"pid", pkgItem.PId,
			"sids", sids,
			"timeUnit", timeUnit,
			"expireTime", expireTime.UTC().String())

		for i := 0; i < len(sids); i++ {
			for j := 0; j < len(pkgItem.Subpackages); j++ {
				if pkgItem.Subpackages[j].SId == sids[i] {
					schedulers := []*entities.SchedulerData{
						{
							pkgItem.Subpackages[j].OrderId,
							pkgItem.Subpackages[j].PId,
							pkgItem.Subpackages[j].SId,
							state.Name(),
							state.Index(),
							states.SchedulerJobName,
							states.SchedulerGroupName,
							scheduler_action.Accept.ActionName(),
							0,
							0,
							"",
							nil,
							nil,
							string(states.SchedulerSubpackageStateExpire),
							"",
							nil,
							true,
							expireTime,
							time.Now().UTC(),
							time.Now().UTC(),
							nil,
							nil,
						},
					}
					state.UpdateSubPackageWithScheduler(ctx, pkgItem.Subpackages[j], schedulers, buyerNotificationAction)
				}
			}
		}

		pkgItemUpdated, e := app.Globals.CQRSRepository.CmdR().PkgCR().Update(ctx, *pkgItem, true)
		if e != nil {
			app.Globals.Logger.FromContext(ctx).Error("PkgItemRepository.Update failed",
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

		response := events.ActionResponse{
			OrderId: pkgItem.OrderId,
			SIds:    sids,
		}

		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()

		app.Globals.Logger.FromContext(ctx).Debug("set status of subpackages success",
			"fn", "Process",
			"state", state.Name(),
			"oid", pkgItemUpdated.OrderId,
			"pid", pkgItemUpdated.PId,
			"sids", sids)

	} else if iFrame.Header().KeyExists(string(frame.HeaderEvent)) {
		event, ok := iFrame.Header().Value(string(frame.HeaderEvent)).(events.IEvent)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("received frame doesn't have a event",
				"fn", "Process",
				"state", state.Name(),
				"iframe", iFrame)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", nil).Send()
			return
		}

		app.Globals.Logger.FromContext(ctx).Debug("received event",
			"fn", "Process",
			"state", state.Name(),
			"oid", event.OrderId(),
			"pid", event.PackageId(),
			"uid", event.UserId(),
			"sIdx", event.StateIndex(),
			"action", event.Action(),
			"data", event.Data(),
			"event", event)

		if event.EventType() == events.Action {
			pkgItem, ok := iFrame.Body().Content().(*entities.PackageItem)
			if !ok {
				app.Globals.Logger.FromContext(ctx).Error("content of frame body is not a PackageItem",
					"fn", "Process",
					"state", state.Name(),
					"event", event,
					"iframe", iFrame)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", errors.New("frame body invalid")).Send()
				return
			}

			actionData, ok := event.Data().(events.ActionData)
			if !ok {
				app.Globals.Logger.FromContext(ctx).Error("received action event data invalid",
					"fn", "Process",
					"state", state.Name(),
					"event", event)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", errors.New("Action Data event invalid")).Send()
				return
			}

			var newSubPackages []*entities.Subpackage
			var requestAction *entities.Action
			var newSubPkg *entities.Subpackage
			var fullItems []*entities.Item
			var nextActionState states.IState
			var actionState actions.IAction

			for action, nextState := range state.StatesMap() {
				if action.ActionType().ActionName() == event.Action().ActionType().ActionName() &&
					action.ActionEnum().ActionName() == event.Action().ActionEnum().ActionName() {
					nextActionState = nextState
					actionState = action
					break
				}
			}

			if nextActionState == nil || actionState == nil {
				app.Globals.Logger.FromContext(ctx).Error("received action not acceptable",
					"fn", "Process",
					"state", state.Name(),
					"oid", pkgItem.OrderId,
					"pid", pkgItem.PId,
					"event", event)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.NotAccepted, "Action Not Accepted", errors.New("Action Not Accepted")).Send()
				return
			}

		loop:
			// iterate subpackages
			for _, eventSubPkg := range actionData.SubPackages {
				findSubPkg := false
				for i := 0; i < len(pkgItem.Subpackages); i++ {
					if eventSubPkg.SId == pkgItem.Subpackages[i].SId && pkgItem.Subpackages[i].Status == state.Name() {
						findSubPkg = true
						newSubPkg = nil
						fullItems = nil
						var findItem = false

						// iterate items
						for _, actionItem := range eventSubPkg.Items {
							findItem = false
							for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
								if actionItem.InventoryId == pkgItem.Subpackages[i].Items[j].InventoryId {
									findItem = true

									if newSubPkg == nil {
										newSubPkg = pkgItem.Subpackages[i].DeepCopy()
										newSubPkg.SId = 0
										newSubPkg.Items = make([]*entities.Item, 0, len(eventSubPkg.Items))
										newSubPkg.CreatedAt = time.Now().UTC()
										newSubPkg.UpdatedAt = time.Now().UTC()

										requestAction = &entities.Action{
											Name:      actionState.ActionEnum().ActionName(),
											Type:      "",
											UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
											UTP:       actionState.ActionType().ActionName(),
											Perm:      "",
											Priv:      "",
											Policy:    "",
											Result:    string(states.ActionSuccess),
											Reasons:   actionItem.Reasons,
											Note:      "",
											Data:      nil,
											CreatedAt: time.Now().UTC(),
											Extended:  nil,
										}
									}

									// create new subpackages which contains new items along
									// with new quantity and recalculated related invoice
									if actionItem.Quantity < pkgItem.Subpackages[i].Items[j].Quantity {
										unit, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount)
										if err != nil {
											app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Unit.Amount invalid",
												"fn", "Process",
												"state", state.Name(),
												"oid", pkgItem.Subpackages[i].OrderId,
												"pid", pkgItem.Subpackages[i].PId,
												"sid", pkgItem.Subpackages[i].SId,
												"unit", pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount,
												"event", event)
											future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
												SetError(future.InternalError, "Unknown Err", errors.New("Subpackage Unit invalid")).Send()
											return
										}

										pkgItem.Subpackages[i].Items[j].Quantity -= actionItem.Quantity
										pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount = strconv.Itoa(int(unit.IntPart() * int64(pkgItem.Subpackages[i].Items[j].Quantity)))

										// create new item from requested action item
										newItem := pkgItem.Subpackages[i].Items[j].DeepCopy()
										newItem.Quantity = actionItem.Quantity
										newItem.Reasons = actionItem.Reasons
										newItem.Invoice.Total.Amount = strconv.Itoa(int(unit.IntPart() * int64(newItem.Quantity)))
										newSubPkg.Items = append(newSubPkg.Items, newItem)

									} else if actionItem.Quantity > pkgItem.Subpackages[i].Items[j].Quantity {
										app.Globals.Logger.FromContext(ctx).Error("received action not acceptable, Requested quantity greater than item quantity",
											"fn", "Process",
											"state", state.Name(),
											"oid", pkgItem.Subpackages[i].OrderId,
											"pid", pkgItem.Subpackages[i].PId,
											"sid", pkgItem.Subpackages[i].SId,
											"event", event)
										future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
											SetError(future.NotAccepted, "Requested quantity greater than item quantity", errors.New("Action Not Accepted")).Send()
										return

									} else {
										if fullItems == nil {
											fullItems = make([]*entities.Item, 0, len(pkgItem.Subpackages[i].Items))
										}
										fullItems = append(fullItems, pkgItem.Subpackages[i].Items[j])
										pkgItem.Subpackages[i].Items[len(pkgItem.Subpackages[i].Items)-1], pkgItem.Subpackages[i].Items[j] =
											pkgItem.Subpackages[i].Items[j], pkgItem.Subpackages[i].Items[len(pkgItem.Subpackages[i].Items)-1]
										pkgItem.Subpackages[i].Items = pkgItem.Subpackages[i].Items[:len(pkgItem.Subpackages[i].Items)-1]

										// calculate subpackages diff
										if len(pkgItem.Subpackages[i].Items) == 0 {
											if newSubPackages == nil {
												newSubPackages = make([]*entities.Subpackage, 0, len(actionData.SubPackages))
											}

											pkgItem.Subpackages[i].Items = fullItems
											newSubPackages = append(newSubPackages, pkgItem.Subpackages[i])
											continue loop
										}
									}
								}
							}
							if !findItem {
								app.Globals.Logger.FromContext(ctx).Error("received action item inventory not found, Requested action item inventory not found in requested subpackage",
									"fn", "Process",
									"state", state.Name(),
									"oid", pkgItem.Subpackages[i].OrderId,
									"pid", pkgItem.Subpackages[i].PId,
									"sid", pkgItem.Subpackages[i].SId,
									"inventoryId", actionItem.InventoryId,
									"event", event)
								future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
									SetError(future.NotFound, "Request action item not found", errors.New("Action Item Not Found")).Send()
								return
							}
						}

						if newSubPackages == nil {
							newSubPackages = make([]*entities.Subpackage, 0, len(actionData.SubPackages))
						}

						if newSubPkg != nil {
							if fullItems != nil {
								for z := 0; z < len(fullItems); z++ {
									newSubPkg.Items = append(newSubPkg.Items, fullItems[z])
								}
							}
							newSubPackages = append(newSubPackages, newSubPkg)
						}
					}
				}

				if !findSubPkg {
					app.Globals.Logger.FromContext(ctx).Warn("Action SId not found or subpackage status not equal with current state",
						"fn", "Process",
						"state", state.Name(),
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", eventSubPkg.SId,
						"event", event)
				}
			}

			if newSubPackages != nil {
				var sids = make([]uint64, 0, 32)
				for i := 0; i < len(newSubPackages); i++ {
					if newSubPackages[i].SId == 0 {
						newSid, err := app.Globals.CQRSRepository.CmdR().SubPkgCR().GenerateUniqSid(ctx, pkgItem.OrderId)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("SubPkgRepository.GenerateUniqSid failed",
								"fn", "Process",
								"state", state.Name(),
								"oid", newSubPackages[i].OrderId,
								"pid", newSubPackages[i].PId,
								"sid", newSubPackages[i].SId,
								"event", event)
							future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
								SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).Send()
							return
						}
						newSubPackages[i].SId = newSid
						pkgItem.Subpackages = append(pkgItem.Subpackages, newSubPackages[i])
					}
					sids = append(sids, newSubPackages[i].SId)
					state.UpdateSubPackage(ctx, newSubPackages[i], requestAction)
				}

				//if event.Action().ActionEnum() == seller_action.Accept {
				//	var rejectedSubtotal int64 = 0
				//	var rejectedDiscount int64 = 0
				//
				//	for i := 0; i < len(newSubPackages); i++ {
				//		for j := 0; j < len(newSubPackages[i].Items); j++ {
				//			amount, err := decimal.NewFromString(newSubPackages[i].Items[j].Invoice.Total.Amount)
				//			if err != nil {
				//				app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Total.Amount invalid",
				//					"fn", "Process",
				//					"state", state.Name(),
				//					"oid", newSubPackages[i].OrderId,
				//					"pid", newSubPackages[i].PId,
				//					"sid", newSubPackages[i].SId,
				//					"total", newSubPackages[i].Items[j].Invoice.Total.Amount,
				//					"event", event)
				//				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				//					SetError(future.InternalError, "Unknown Error", errors.New("Subpackage Total Invalid")).Send()
				//				return
				//			}
				//
				//			discount, err := decimal.NewFromString(newSubPackages[i].Items[j].Invoice.Discount.Amount)
				//			if err != nil {
				//				app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Invoice.Discount invalid",
				//					"fn", "Process",
				//					"state", state.Name(),
				//					"oid", newSubPackages[i].OrderId,
				//					"pid", newSubPackages[i].PId,
				//					"sid", newSubPackages[i].SId,
				//					"discount", newSubPackages[i].Items[j].Invoice.Discount,
				//					"event", event)
				//				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				//					SetError(future.InternalError, "Unknown Error", errors.New("Subpackage Discount Invalid")).Send()
				//				return
				//			}
				//
				//			rejectedSubtotal += amount.IntPart()
				//			rejectedDiscount += discount.IntPart()
				//		}
				//	}
				//
				//	subtotal, err := decimal.NewFromString(pkgItem.Invoice.Subtotal.Amount)
				//	if err != nil {
				//		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal.Amount invalid",
				//			"fn", "Process",
				//			"state", state.Name(),
				//			"oid", pkgItem.OrderId,
				//			"pid", pkgItem.PId,
				//			"subtotal", pkgItem.Invoice.Subtotal.Amount,
				//			"event", event)
				//		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				//			SetError(future.InternalError, "Unknown Error", errors.New("Package Invoice Invalid")).Send()
				//		return
				//	}
				//
				//	pkgDiscount, err := decimal.NewFromString(pkgItem.Invoice.Discount.Amount)
				//	if err != nil {
				//		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Pkg Discount.Amount invalid",
				//			"fn", "Process",
				//			"state", state.Name(),
				//			"oid", pkgItem.OrderId,
				//			"pid", pkgItem.PId,
				//			"pkg discount", pkgItem.Invoice.Discount.Amount,
				//			"event", event)
				//		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				//			SetError(future.InternalError, "Unknown Error", errors.New("Package Invoice Invalid")).Send()
				//		return
				//	}
				//
				//	if rejectedSubtotal < subtotal.IntPart() && rejectedDiscount < pkgDiscount.IntPart() {
				//		pkgItem.Invoice.Subtotal.Amount = strconv.Itoa(int(subtotal.IntPart() - rejectedSubtotal))
				//		pkgItem.Invoice.Discount.Amount = strconv.Itoa(int(pkgDiscount.IntPart() - rejectedDiscount))
				//		app.Globals.Logger.FromContext(ctx).Info("calculate package invoice success",
				//			"fn", "Process",
				//			"state", state.Name(),
				//			"oid", pkgItem.OrderId,
				//			"pid", pkgItem.PId,
				//			"action", event.Action().ActionEnum().ActionName(),
				//			"subtotal", pkgItem.Invoice.Subtotal.Amount,
				//			"discount", pkgItem.Invoice.Discount.Amount)
				//
				//	} else if rejectedSubtotal > subtotal.IntPart() || rejectedDiscount > pkgDiscount.IntPart() {
				//		app.Globals.Logger.FromContext(ctx).Error("calculate package invoice failed",
				//			"fn", "Process",
				//			"state", state.Name(),
				//			"oid", pkgItem.OrderId,
				//			"pid", pkgItem.PId,
				//			"action", event.Action().ActionEnum().ActionName(),
				//			"subtotal", pkgItem.Invoice.Subtotal.Amount,
				//			"discount", pkgItem.Invoice.Discount.Amount)
				//		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				//			SetError(future.InternalError, "Unknown Error", errors.New("Package Invoice Invalid")).Send()
				//		return
				//	}
				//}

				app.Globals.Logger.FromContext(ctx).Debug("set status of subpackages success",
					"fn", "Process",
					"state", state.Name(),
					"oid", pkgItem.OrderId,
					"pid", pkgItem.PId,
					"sids", sids,
					"action", event.Action().ActionEnum().ActionName())

				nextActionState.Process(ctx, frame.FactoryOf(iFrame).SetSIds(sids).SetBody(pkgItem).Build())
			} else {
				app.Globals.Logger.FromContext(ctx).Error("event action data invalid",
					"fn", "Process",
					"state", state.Name(),
					"event", event,
					"iframe", iFrame)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.BadRequest, "Event Action Data Invalid", errors.New("event action data invalid")).Send()
			}
		} else {
			app.Globals.Logger.FromContext(ctx).Error("event type not supported",
				"fn", "Process",
				"state", state.Name(),
				"event", event,
				"iframe", iFrame)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", errors.New("event type invalid")).Send()
			return
		}
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
