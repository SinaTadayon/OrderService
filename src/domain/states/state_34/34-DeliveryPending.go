package state_34

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
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
	stepName  string = "Delivery_Pending"
	stepIndex int    = 34
)

type DeliveryPendingState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &DeliveryPendingState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &DeliveryPendingState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &DeliveryPendingState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state DeliveryPendingState) Process(ctx context.Context, iFrame frame.IFrame) {
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

		var deliveredAt time.Time
		timeUnit := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig].(string)
		if timeUnit == app.DurationTimeUnit {
			value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveryPendingStateConfig].(time.Duration)
			deliveredAt = time.Now().UTC().Add(value)
		} else {
			value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerDeliveryPendingStateConfig].(int)
			if timeUnit == string(app.HourTimeUnit) {
				deliveredAt = time.Now().UTC().Add(
					time.Hour*time.Duration(value) +
						time.Minute*time.Duration(0) +
						time.Second*time.Duration(0))
			} else {
				deliveredAt = time.Now().UTC().Add(
					time.Hour*time.Duration(0) +
						time.Minute*time.Duration(value) +
						time.Second*time.Duration(0))
			}
		}

		var notifyAt time.Time
		if timeUnit == app.DurationTimeUnit {
			value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerNotifyDeliveryPendingStateConfig].(time.Duration)
			notifyAt = time.Now().UTC().Add(value)
		} else {
			notifyValue := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerNotifyDeliveryPendingStateConfig].(int)
			if timeUnit == string(app.HourTimeUnit) {
				notifyAt = time.Now().UTC().Add(
					time.Hour*time.Duration(notifyValue) +
						time.Minute*time.Duration(0) +
						time.Second*time.Duration(0))
			} else {
				notifyAt = time.Now().UTC().Add(
					time.Hour*time.Duration(0) +
						time.Minute*time.Duration(notifyValue) +
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
			"notifyAt", notifyAt.UTC().String(),
			"deliveredAt", deliveredAt.UTC().String())

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
							scheduler_action.Notification.ActionName(),
							0,
							0,
							"",
							nil,
							nil,
							string(states.SchedulerSubpackageStateNotify),
							"",
							nil,
							true,
							notifyAt,
							time.Now().UTC(),
							time.Now().UTC(),
							nil,
							nil,
						},
						{
							pkgItem.Subpackages[j].OrderId,
							pkgItem.Subpackages[j].PId,
							pkgItem.Subpackages[j].SId,
							state.Name(),
							state.Index(),
							states.SchedulerJobName,
							states.SchedulerGroupName,
							scheduler_action.Deliver.ActionName(),
							1,
							0,
							"",
							nil,
							nil,
							string(states.SchedulerSubpackageStateExpire),
							"",
							nil,
							true,
							deliveredAt,
							time.Now().UTC(),
							time.Now().UTC(),
							nil,
							nil,
						},
					}
					state.UpdateSubPackageWithScheduler(ctx, pkgItem.Subpackages[j], schedulers, nil)
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

			if event.Action().ActionEnum() == seller_action.EnterShipmentDetail {
				if actionData.Carrier == "" {
					app.Globals.Logger.FromContext(ctx).Error("Carrier is empty",
						"fn", "Process",
						"state", state.Name(),
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"event", event)
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.BadRequest, "Carrier is empty", errors.New("Event Action Data Invalid")).Send()
					return
				}

				enterShipmentDetailAction := &entities.Action{
					Name:      seller_action.EnterShipmentDetail.ActionName(),
					Type:      "",
					UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
					UTP:       actions.Seller.ActionName(),
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

				var sids = make([]uint64, 0, 32)
				shipmentTime := time.Now().UTC()
				var findSubPkg = false
				var isPkgUpdated = false
				for _, eventSubPkg := range actionData.SubPackages {
					findSubPkg = false
					for i := 0; i < len(pkgItem.Subpackages); i++ {
						if eventSubPkg.SId == pkgItem.Subpackages[i].SId && pkgItem.Subpackages[i].Status == state.Name() {
							findSubPkg = true
							isPkgUpdated = true
							if pkgItem.Subpackages[i].Shipments == nil {
								app.Globals.Logger.FromContext(ctx).Warn("Shipment is nil",
									"fn", "Process",
									"state", state.Name(),
									"oid", pkgItem.OrderId,
									"pid", pkgItem.PId,
									"sid", pkgItem.Subpackages[i].SId,
									"event", event)

								pkgItem.Subpackages[i].Shipments = &entities.Shipment{
									ShipmentDetail: &entities.ShippingDetail{
										CourierName:    actionData.Carrier,
										ShippingMethod: "",
										TrackingNumber: actionData.TrackingNumber,
										Image:          "",
										Description:    "",
										ShippedAt:      &shipmentTime,
										CreatedAt:      shipmentTime,
										UpdatedAt:      &shipmentTime,
										Extended:       nil,
									},
									ReturnShipmentDetail: nil,
								}
							} else {
								pkgItem.Subpackages[i].Shipments.ShipmentDetail.CourierName = actionData.Carrier
								pkgItem.Subpackages[i].Shipments.ShipmentDetail.TrackingNumber = actionData.TrackingNumber
								pkgItem.Subpackages[i].Shipments.ShipmentDetail.UpdatedAt = &shipmentTime
							}

							sids = append(sids, eventSubPkg.SId)
							state.UpdateSubPackage(ctx, pkgItem.Subpackages[i], enterShipmentDetailAction)
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

				if isPkgUpdated {
					_, err := app.Globals.CQRSRepository.CmdR().PkgCR().Update(ctx, *pkgItem, false)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("PkgItemRepository.Update failed",
							"fn", "Process",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sids", sids,
							"error", err)
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).Send()
						return
					}

					response := events.ActionResponse{
						OrderId: pkgItem.OrderId,
						SIds:    sids,
					}

					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetData(response).Send()
				} else {
					app.Globals.Logger.FromContext(ctx).Info("Action could not applied to Pkg",
						"fn", "Process",
						"state", state.Name(),
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"event", event)
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.BadRequest, "Action Data Invalid", errors.New("Action Data Invalid")).Send()
				}

				return
			}

			if event.Action().ActionType() == actions.Scheduler &&
				(event.Action().ActionEnum() == scheduler_action.Notification ||
					event.Action().ActionEnum() == scheduler_action.Deliver) {

				schedulerAction := &entities.Action{
					Name:      scheduler_action.Notification.ActionName(),
					Type:      "",
					UId:       0,
					UTP:       actions.Scheduler.ActionName(),
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

				var buyerNotificationAction = &entities.Action{
					Name:      system_action.BuyerNotification.ActionName(),
					Type:      "",
					UId:       0,
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

				var sids = make([]uint64, 0, 32)
				var isSendSMSEnabled = false
				var findSubPkg = false
				var isPkgUpdated = false
				for _, eventSubPkg := range actionData.SubPackages {
					findSubPkg = false
					for i := 0; i < len(pkgItem.Subpackages); i++ {
						if eventSubPkg.SId == pkgItem.Subpackages[i].SId && pkgItem.Subpackages[i].Status == state.Name() {
							findSubPkg = true
							if pkgItem.Subpackages[i].Tracking.State.Schedulers != nil {
								for _, schedulerData := range pkgItem.Subpackages[i].Tracking.State.Schedulers {
									if schedulerData.Type == string(states.SchedulerSubpackageStateNotify) && schedulerData.Enabled {
										isPkgUpdated = true
										schedulerData.Enabled = false
										sids = append(sids, pkgItem.Subpackages[i].SId)
										var buyerNotify notify_service.SMSRequest

										if !isSendSMSEnabled {
											isSendSMSEnabled = true
											smsTemplate, err := template.New("SMS").Parse(app.Globals.SMSTemplate.OrderNotifyBuyerDeliveryPendingState)
											if err != nil {
												app.Globals.Logger.FromContext(ctx).Error("smsTemplate.Parse failed",
													"fn", "Process",
													"state", state.Name(),
													"oid", pkgItem.OrderId,
													"pid", pkgItem.PId,
													"sids", sids,
													"message", app.Globals.SMSTemplate.OrderNotifyBuyerDeliveryPendingState,
													"error", err)
											} else {
												var buf bytes.Buffer
												err = smsTemplate.Execute(&buf, pkgItem.OrderId)
												newBuf := bytes.NewBuffer(bytes.Replace(buf.Bytes(), []byte("\\n"), []byte{10}, -1))
												if err != nil {
													app.Globals.Logger.FromContext(ctx).Error("smsTemplate.Execute failed",
														"fn", "Process",
														"state", state.Name(),
														"oid", pkgItem.OrderId,
														"pid", pkgItem.PId,
														"sids", sids,
														"message", app.Globals.SMSTemplate.OrderNotifyBuyerDeliveryPendingState,
														"error", err)
												} else {
													buyerNotify = notify_service.SMSRequest{
														Phone: pkgItem.ShippingAddress.Mobile,
														Body:  newBuf.String(),
														User:  notify_service.BuyerUser,
													}

													futureData := app.Globals.NotifyService.NotifyBySMS(ctx, buyerNotify).Get()
													if futureData.Error() != nil {
														app.Globals.Logger.FromContext(ctx).Error("NotifyService.NotifyBySMS failed",
															"fn", "Process",
															"state", state.Name(),
															"oid", pkgItem.OrderId,
															"pid", pkgItem.PId,
															"sids", sids,
															"request", buyerNotify,
															"error", futureData.Error().Reason())
														buyerNotificationAction = &entities.Action{
															Name:      system_action.BuyerNotification.ActionName(),
															Type:      "",
															UId:       0,
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
															UId:       0,
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

										state.UpdateSubPackage(ctx, pkgItem.Subpackages[i], schedulerAction)
										state.UpdateSubPackage(ctx, pkgItem.Subpackages[i], buyerNotificationAction)
									}
								}
							} else {
								app.Globals.Logger.FromContext(ctx).Error("Tracking.State.Data is nil",
									"fn", "Process",
									"state", state.Name(),
									"oid", pkgItem.Subpackages[i].OrderId,
									"pid", pkgItem.Subpackages[i].PId,
									"sid", pkgItem.Subpackages[i].SId,
									"event", event)
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

				if event.Action().ActionEnum() == scheduler_action.Notification {
					if isPkgUpdated {
						_, err := app.Globals.CQRSRepository.CmdR().PkgCR().Update(ctx, *pkgItem, false)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("PkgItemRepository.Update failed",
								"fn", "Process",
								"state", state.Name(),
								"oid", pkgItem.OrderId,
								"pid", pkgItem.PId,
								"sids", sids,
								"error", err)
							future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
								SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).Send()
							return
						}

						response := events.ActionResponse{
							OrderId: pkgItem.OrderId,
							SIds:    sids,
						}

						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetData(response).Send()
					} else {
						app.Globals.Logger.FromContext(ctx).Info("Action could not applied to Pkg",
							"fn", "Process",
							"state", state.Name(),
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"event", event)
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetError(future.BadRequest, "Action Data Invalid", errors.New("Action Data Invalid")).Send()
					}

					return
				}
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
