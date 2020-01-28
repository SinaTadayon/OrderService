package state_10

import (
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
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"strconv"
	"time"
)

const (
	stepName  string = "Payment_Pending"
	stepIndex int    = 10
)

type paymentPendingState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentPendingState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentPendingState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &paymentPendingState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state paymentPendingState) Process(ctx context.Context, iFrame frame.IFrame) {

	if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) && iFrame.Body().Content() != nil {
		order, ok := iFrame.Body().Content().(*entities.Order)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("Content of frame body isn't an order",
				"fn", "Process",
				"state", state.Name(),
				"oid", iFrame.Header().Value(string(frame.HeaderOrderId)),
				"content", iFrame.Body().Content())
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetData(app.Globals.Config.App.OrderPaymentCallbackUrlFail).
				Send()
			return
		}

		grandTotal, err := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("order.Invoice.GrandTotal.Amount invalid",
				"fn", "Process",
				"state", state.Name(),
				"amount", order.Invoice.GrandTotal.Amount,
				"oid", order.OrderId,
				"error", err)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetData(app.Globals.Config.App.OrderPaymentCallbackUrlFail + strconv.Itoa(int(order.OrderId))).
				Send()

			paymentAction := &entities.Action{
				Name:      system_action.PaymentFail.ActionName(),
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

			state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
			failAction := state.GetAction(system_action.PaymentFail.ActionName())
			state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
			return
		}

		var voucherAmount decimal.Decimal
		if order.Invoice.Voucher != nil && order.Invoice.Voucher.Price != nil {
			voucherAmount, err = decimal.NewFromString(order.Invoice.Voucher.Price.Amount)
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("order.Invoice.Voucher.Price.Amount invalid",
					"fn", "Process",
					"state", state.Name(),
					"price", order.Invoice.Voucher.Price.Amount,
					"oid", order.OrderId,
					"error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetData(app.Globals.Config.App.OrderPaymentCallbackUrlFail + strconv.Itoa(int(order.OrderId))).
					Send()

				paymentAction := &entities.Action{
					Name:      system_action.PaymentFail.ActionName(),
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

				state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
				failAction := state.GetAction(system_action.PaymentFail.ActionName())
				state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
				return
			}
		}

		if grandTotal.IsZero() && order.Invoice.Voucher != nil &&
			(order.Invoice.Voucher.Percent > 0 || !voucherAmount.IsZero()) {
			order.OrderPayment = []entities.PaymentService{
				{
					PaymentRequest: &entities.PaymentRequest{
						Price:     nil,
						Gateway:   "",
						CreatedAt: time.Now().UTC(),
						Mobile:    order.BuyerInfo.Mobile,
						Extended:  nil,
					},

					PaymentResult: &entities.PaymentResult{
						Result:      true,
						Reason:      "Invoice paid by voucher",
						PaymentId:   "",
						InvoiceId:   0,
						Price:       nil,
						CardNumMask: "",
						CreatedAt:   time.Now().UTC(),
					},

					PaymentResponse: &entities.PaymentResponse{
						Result:      true,
						CallBackUrl: app.Globals.Config.App.OrderPaymentCallbackUrlSuccess + strconv.Itoa(int(order.OrderId)),
						InvoiceId:   0,
						PaymentId:   "",
						CreatedAt:   time.Now().UTC(),
					},
				},
			}

			// TODO check it voucher amount and if voucherSettlement failed can be cancel order
			var voucherAction *entities.Action
			iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				app.Globals.Logger.FromContext(ctx).Error("VoucherService.VoucherSettlement failed",
					"fn", "Process",
					"state", state.Name(),
					"oid", order.OrderId,
					"voucherCode", order.Invoice.Voucher.Code,
					"error", futureData.Error().Reason())
				timestamp := time.Now().UTC()
				order.Invoice.Voucher.Settlement = string(states.ActionFail)
				order.Invoice.Voucher.SettlementAt = &timestamp
				voucherAction = &entities.Action{
					Name:      system_action.VoucherSettlement.ActionName(),
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
					CreatedAt: timestamp,
					Extended:  nil,
				}
			} else {
				if order.Invoice.Voucher.Percent > 0 {
					app.Globals.Logger.FromContext(ctx).Info("voucher applied in invoice of order success",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"voucher Percent", order.Invoice.Voucher.Percent,
						"voucher Code", order.Invoice.Voucher.Code)
				} else {
					app.Globals.Logger.FromContext(ctx).Info("voucher applied in invoice of order success",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"voucher Amount", voucherAmount,
						"voucher Code", order.Invoice.Voucher.Code)
				}
				timestamp := time.Now().UTC()
				order.Invoice.Voucher.Settlement = string(states.ActionSuccess)
				order.Invoice.Voucher.SettlementAt = &timestamp
				voucherAction = &entities.Action{
					Name:      system_action.VoucherSettlement.ActionName(),
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
					CreatedAt: timestamp,
					Extended:  nil,
				}
			}

			state.UpdateOrderAllSubPkg(ctx, order, voucherAction)
			//orderUpdated, err := app.Globals.OrderRepository.Save(ctx, *order)
			//if err != nil {
			//	logger.Err("Process() => OrderRepository.Save in %s state failed, order: %v, error: %v", state.Name(), order, err)
			//	future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			//		//SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).
			//		SetData(app.Globals.Config.App.OrderPaymentCallbackUrlFail + strconv.Itoa(int(order.OrderId))).
			//		Send()
			//} else {
			app.Globals.Logger.FromContext(ctx).Debug("Order state of all subpackages update",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetData(order.OrderPayment[0].PaymentResponse.CallBackUrl).
				Send()
			successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
			state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
			//}
		} else {
			app.Globals.Logger.FromContext(ctx).Info("invoice order without voucher applied",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId)
			paymentRequest := payment_service.PaymentRequest{
				Amount:   int64(grandTotal.IntPart()),
				Currency: order.Invoice.GrandTotal.Currency,
				Gateway:  order.Invoice.PaymentGateway,
				OrderId:  order.OrderId,
				Mobile:   order.BuyerInfo.Mobile,
			}

			order.OrderPayment = []entities.PaymentService{
				{
					PaymentRequest: &entities.PaymentRequest{
						Price: &entities.Money{
							Amount:   strconv.Itoa(int(paymentRequest.Amount)),
							Currency: order.Invoice.GrandTotal.Currency,
						},
						Gateway:   paymentRequest.Gateway,
						CreatedAt: time.Now().UTC(),
						Mobile:    order.BuyerInfo.Mobile,
						Extended:  nil,
					},
				},
			}

			iFuture := app.Globals.PaymentService.OrderPayment(ctx, paymentRequest)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
					Result:      false,
					CallBackUrl: app.Globals.Config.App.OrderPaymentCallbackUrlFail + strconv.Itoa(int(order.OrderId)),
					Reason:      strconv.Itoa(int(futureData.Error().Code())),
					CreatedAt:   time.Now().UTC(),
				}

				paymentAction := &entities.Action{
					Name:      system_action.PaymentFail.ActionName(),
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

				state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
				//orderUpdated, err := app.Globals.OrderRepository.Save(ctx, *order)
				//if err != nil {
				//	logger.Err("Process() => OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)
				//	future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				//		SetData(app.Globals.Config.App.OrderPaymentCallbackUrlFail + strconv.Itoa(int(order.OrderId))).
				//		//SetError(future.InternalError, "Unknown Error", err).
				//		Send()
				//	return
				//}

				app.Globals.Logger.FromContext(ctx).Error("PaymentService.OrderPayment failed",
					"fn", "Process",
					"state", state.Name(),
					"oid", order.OrderId,
					"error", futureData.Error().Reason())
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetData(order.OrderPayment[0].PaymentResponse.CallBackUrl).Send()

				failAction := state.GetAction(system_action.PaymentFail.ActionName())
				state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
				return
			} else {
				paymentResponse := futureData.Data().(payment_service.PaymentResponse)
				order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
					Result:      true,
					CallBackUrl: paymentResponse.CallbackUrl,
					InvoiceId:   paymentResponse.InvoiceId,
					PaymentId:   paymentResponse.PaymentId,
					CreatedAt:   time.Now().UTC(),
				}

				var expireTime time.Time
				timeUnit := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig].(string)
				if timeUnit == app.DurationTimeUnit {
					value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig].(time.Duration)
					expireTime = time.Now().UTC().Add(value)
				} else {
					value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig].(int)
					//if timeUnit == string(app.HourTimeUnit) {
					//	expireTime = time.Now().UTC().Add(
					//		time.Hour*time.Duration(value) +
					//			time.Minute*time.Duration(0) +
					//			time.Second*time.Duration(0))
					//} else {
					expireTime = time.Now().UTC().Add(
						time.Hour*time.Duration(0) +
							time.Minute*time.Duration(value) +
							time.Second*time.Duration(0))
					//}
				}

				order.UpdatedAt = time.Now().UTC()
				for i := 0; i < len(order.Packages); i++ {
					order.Packages[i].UpdatedAt = time.Now().UTC()
					for j := 0; j < len(order.Packages[i].Subpackages); j++ {
						schedulers := []*entities.SchedulerData{
							{
								order.OrderId,
								order.Packages[i].PId,
								order.Packages[i].Subpackages[j].SId,
								state.Name(),
								state.Index(),
								states.SchedulerJobName,
								states.SchedulerGroupName,
								scheduler_action.PaymentFail.ActionName(),
								0,
								app.Globals.FlowManagerConfig[app.FlowManagerSchedulerRetryPaymentPendingStateConfig].(int32),
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
						state.UpdateSubPackageWithScheduler(ctx, order.Packages[i].Subpackages[j], schedulers, nil)
					}
				}

				//state.UpdateOrderAllSubPkg(ctx, order, nil)
				_, err := app.Globals.OrderRepository.Save(ctx, *order)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("OrderRepository.Save failed",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"error", err)

					paymentAction := &entities.Action{
						Name:      system_action.PaymentFail.ActionName(),
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

					state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetData(order.OrderPayment[0].PaymentResponse.CallBackUrl + strconv.Itoa(int(order.OrderId))).
						//SetError(future.InternalError, "Unknown Error", err).
						Send()

					failAction := state.GetAction(system_action.PaymentFail.ActionName())
					state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
					return
				}

				app.Globals.Logger.FromContext(ctx).Debug("Order state of all subpackages update",
					"fn", "Process",
					"state", state.Name(),
					"oid", order.OrderId)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetData(paymentResponse.CallbackUrl).
					Send()
				return
			}
		}
	} else if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) &&
		iFrame.Header().KeyExists(string(frame.HeaderPaymentResult)) {
		order, err := app.Globals.OrderRepository.FindById(ctx, iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64))
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("OrderRepository.FindById failed",
				"fn", "Process",
				"state", state.Name(),
				"oid", iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64),
				"paymentResult", iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult),
				"error", err)

			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetCapacity(1).SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).
				Send()

			paymentAction := &entities.Action{
				Name:      system_action.PaymentFail.ActionName(),
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

			state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
			failAction := state.GetAction(system_action.PaymentFail.ActionName())
			state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
			return
		}

		ctx = context.WithValue(ctx, string(utils.CtxUserID), order.BuyerInfo.BuyerId)
		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetCapacity(1).Send()

		order.OrderPayment[0].PaymentResult = iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult)
		app.Globals.Logger.FromContext(ctx).Info("Order Received",
			"fn", "Process",
			"state", state.Name(),
			"oid", order.OrderId)
		if order.OrderPayment[0].PaymentResult.Result == false {
			app.Globals.Logger.FromContext(ctx).Info("PaymentResult failed",
				"orderId", order.OrderId)
			paymentAction := &entities.Action{
				Name:      system_action.PaymentFail.ActionName(),
				Type:      "",
				UId:       order.BuyerInfo.BuyerId,
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

			state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
			//updatedOrder, err := app.Globals.OrderRepository.Save(ctx, *order)
			//if err != nil {
			//	logger.Err("Process() => Singletons.OrderRepository.Save failed, orderId: %d, error: %v", order.OrderId, err)
			//	return
			//}
			failAction := state.GetAction(system_action.PaymentFail.ActionName())
			state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
			return
		} else {
			var voucherAction *entities.Action
			var voucherAmount decimal.Decimal
			if order.Invoice.Voucher != nil && order.Invoice.Voucher.Price != nil {
				var e error
				voucherAmount, e = decimal.NewFromString(order.Invoice.Voucher.Price.Amount)
				if e != nil {
					app.Globals.Logger.FromContext(ctx).Error("order.Invoice.Voucher.Price.Amount invalid",
						"fn", "Process",
						"state", state.Name(),
						"price", order.Invoice.Voucher.Price.Amount,
						"oid", order.OrderId,
						"error", err)
				}
			}

			if order.Invoice.Voucher != nil && (order.Invoice.Voucher.Percent > 0 || !voucherAmount.IsZero()) {
				iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
				futureData := iFuture.Get()
				if futureData.Error() != nil {
					app.Globals.Logger.FromContext(ctx).Error("VoucherService.VoucherSettlement failed",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"voucher Code", order.Invoice.Voucher.Code,
						"error", futureData.Error().Reason())
					timestamp := time.Now().UTC()
					order.Invoice.Voucher.Settlement = string(states.ActionFail)
					order.Invoice.Voucher.SettlementAt = &timestamp
					voucherAction = &entities.Action{
						Name:      system_action.VoucherSettlement.ActionName(),
						Type:      "",
						UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
						UTP:       actions.System.ActionName(),
						Perm:      "",
						Priv:      "",
						Policy:    "",
						Result:    string(states.ActionFail),
						Note:      "",
						Reasons:   nil,
						Data:      nil,
						CreatedAt: timestamp,
						Extended:  nil,
					}
				} else {
					if order.Invoice.Voucher.Percent > 0 {
						app.Globals.Logger.FromContext(ctx).Info("voucher applied in invoice of order success",
							"fn", "Process",
							"state", state.Name(),
							"oid", order.OrderId,
							"voucher Percent", order.Invoice.Voucher.Percent,
							"voucher Code", order.Invoice.Voucher.Code)
					} else {
						app.Globals.Logger.FromContext(ctx).Info("voucher applied in invoice of order success",
							"fn", "Process",
							"state", state.Name(),
							"oid", order.OrderId,
							"voucher Amount", voucherAmount,
							"voucher Code", order.Invoice.Voucher.Code)
					}

					timestamp := time.Now().UTC()
					order.Invoice.Voucher.Settlement = string(states.ActionSuccess)
					order.Invoice.Voucher.SettlementAt = &timestamp
					voucherAction = &entities.Action{
						Name:      system_action.VoucherSettlement.ActionName(),
						Type:      "",
						UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
						UTP:       actions.System.ActionName(),
						Perm:      "",
						Priv:      "",
						Policy:    "",
						Result:    string(states.ActionSuccess),
						Note:      "",
						Reasons:   nil,
						Data:      nil,
						CreatedAt: timestamp,
						Extended:  nil,
					}
				}

				if order.Invoice.Voucher.Percent > 0 {
					app.Globals.Logger.FromContext(ctx).Info("VoucherSettlement success",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"voucher Percent", order.Invoice.Voucher.Percent,
						"voucherCode", order.Invoice.Voucher.Code)
				} else {
					app.Globals.Logger.FromContext(ctx).Info("VoucherSettlement success",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"voucher Amount", voucherAmount,
						"voucherCode", order.Invoice.Voucher.Code)
				}
			} else {
				app.Globals.Logger.FromContext(ctx).Info("Order Invoice hasn't voucher",
					"state", state.Name(),
					"oid", order.OrderId)
			}

			paymentAction := &entities.Action{
				Name:      system_action.PaymentSuccess.ActionName(),
				Type:      "",
				UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
				UTP:       actions.System.ActionName(),
				Perm:      "",
				Priv:      "",
				Policy:    "",
				Result:    string(states.ActionSuccess),
				Note:      "",
				Reasons:   nil,
				Data:      nil,
				CreatedAt: time.Now().UTC(),
				Extended:  nil,
			}

			app.Globals.Logger.FromContext(ctx).Info("PaymentResult success",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId)
			state.UpdateOrderAllSubPkg(ctx, order, paymentAction, voucherAction)
			//_, err = app.Globals.OrderRepository.Save(ctx, *order)
			//if err != nil {
			//	logger.Err("Process() => OrderRepository.Save in %s state failed, orderId: %d, error: %v", state.Name(), order.OrderId, err)
			//	return
			//}
			successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
			state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
		}

	} else if iFrame.Header().KeyExists(string(frame.HeaderEvent)) {
		event, ok := iFrame.Header().Value(string(frame.HeaderEvent)).(events.IEvent)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("received frame doesn't have a event",
				"fn", "Process",
				"state", state.Name(),
				"frame", iFrame)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", nil).Send()
			return
		}

		if event.EventType() == events.Action {
			order, ok := iFrame.Body().Content().(*entities.Order)
			if !ok {
				app.Globals.Logger.FromContext(ctx).Error("content of frame body isn't order",
					"fn", "Process",
					"state", state.Name(),
					"event", event,
					"frame", iFrame)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", errors.New("frame body invalid")).Send()
				return
			}

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
					"event", event)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.NotAccepted, "Action Not Accepted", errors.New("Action Not Accepted")).Send()
				return
			}

			var expireTime time.Time
			timeUnit := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig].(string)
			if timeUnit == app.DurationTimeUnit {
				value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig].(time.Duration)
				expireTime = time.Now().UTC().Add(value)
			} else {
				value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig].(int)
				//if timeUnit == string(app.HourTimeUnit) {
				//	expireTime = time.Now().UTC().Add(
				//		time.Hour*time.Duration(value) +
				//			time.Minute*time.Duration(0) +
				//			time.Second*time.Duration(0))
				//} else {
				expireTime = time.Now().UTC().Add(
					time.Hour*time.Duration(0) +
						time.Minute*time.Duration(value) +
						time.Second*time.Duration(0))
				//}
			}

			var findFlag = false
			for i := 0; i < len(order.Packages); i++ {
				for j := 0; j < len(order.Packages[i].Subpackages); j++ {
					for _, schedulerData := range order.Packages[i].Subpackages[j].Tracking.State.Schedulers {
						if schedulerData.Type == string(states.SchedulerSubpackageStateExpire) {
							if schedulerData.Retry > 0 {
								findFlag = true
								schedulerData.Retry -= 1
								schedulerData.Data = expireTime
							}
						}
					}
				}
			}

			if findFlag {
				app.Globals.Logger.FromContext(ctx).Info("try get result from PaymentService.GetPaymentResult",
					"fn", "Process",
					"state", state.Name(),
					"oid", order.OrderId)
				iFuture := app.Globals.PaymentService.GetPaymentResult(ctx, order.OrderId)
				futureData := iFuture.Get()
				if futureData.Error() != nil {
					app.Globals.Logger.FromContext(ctx).Error("PaymentService.GetPaymentResult failed",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"event", event)
					_, err := app.Globals.OrderRepository.Save(ctx, *order)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("OrderRepository.Save of update scheduler data failed",
							"fn", "Process",
							"state", state.Name(),
							"oid", order.OrderId,
							"event", event)
					}
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.InternalError, "Unknown Err", futureData.Error().Reason()).Send()
					return
				}

				paymentResult := futureData.Data().(payment_service.PaymentQueryResult)
				if paymentResult.Status != payment_service.PaymentRequestPending {
					order.OrderPayment[0].PaymentResult = &entities.PaymentResult{
						Result:    paymentResult.Status == payment_service.PaymentRequestSuccess,
						Reason:    "",
						PaymentId: paymentResult.PaymentId,
						InvoiceId: paymentResult.InvoiceId,
						Price: &entities.Money{
							Amount:   strconv.Itoa(int(paymentResult.Amount)),
							Currency: "IRR",
						},
						CardNumMask: paymentResult.CardMask,
						CreatedAt:   time.Now().UTC(),
						Extended:    nil,
					}

					if !order.OrderPayment[0].PaymentResult.Result {
						app.Globals.Logger.FromContext(ctx).Error("retry get PaymentQueryResult failed",
							"fn", "Process",
							"state", state.Name(),
							"oid", order.OrderId,
							"result", paymentResult)
						paymentAction := &entities.Action{
							Name:      system_action.PaymentFail.ActionName(),
							Type:      "",
							UId:       order.BuyerInfo.BuyerId,
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

						state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
						response := events.ActionResponse{
							OrderId: order.OrderId,
							SIds:    nil,
						}
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()
						failAction := state.GetAction(system_action.PaymentFail.ActionName())
						state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetOrderId(order.OrderId).SetBody(order).Build())
					} else {
						var voucherAction *entities.Action
						var voucherAmount = 0
						if order.Invoice.Voucher != nil && order.Invoice.Voucher.Price != nil {
							voucherAmount, _ = strconv.Atoi(order.Invoice.Voucher.Price.Amount)
						}

						if order.Invoice.Voucher != nil && (order.Invoice.Voucher.Percent > 0 || voucherAmount > 0) {
							iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
							futureData := iFuture.Get()
							if futureData.Error() != nil {
								app.Globals.Logger.FromContext(ctx).Error("VoucherService.VoucherSettlement failed",
									"fn", "Process",
									"state", state.Name(),
									"oid", order.OrderId,
									"voucher Code", order.Invoice.Voucher.Code,
									"error", futureData.Error().Reason())
								timestamp := time.Now().UTC()
								order.Invoice.Voucher.Settlement = string(states.ActionFail)
								order.Invoice.Voucher.SettlementAt = &timestamp
								voucherAction = &entities.Action{
									Name:      system_action.VoucherSettlement.ActionName(),
									Type:      "",
									UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
									UTP:       actions.System.ActionName(),
									Perm:      "",
									Priv:      "",
									Policy:    "",
									Result:    string(states.ActionFail),
									Note:      "",
									Reasons:   nil,
									Data:      nil,
									CreatedAt: timestamp,
									Extended:  nil,
								}
							} else {
								if order.Invoice.Voucher.Percent > 0 {
									app.Globals.Logger.FromContext(ctx).Info("Invoice paid by voucher Percent order success",
										"fn", "Process",
										"state", state.Name(),
										"oid", order.OrderId,
										"voucher Percent", order.Invoice.Voucher.Percent,
										"voucher Code", order.Invoice.Voucher.Code)
								} else {
									app.Globals.Logger.FromContext(ctx).Info("Invoice paid by voucher Amount order success",
										"fn", "Process",
										"state", state.Name(),
										"oid", order.OrderId,
										"voucher Amount", voucherAmount,
										"voucher Code", order.Invoice.Voucher.Code)
								}

								timestamp := time.Now().UTC()
								order.Invoice.Voucher.Settlement = string(states.ActionSuccess)
								order.Invoice.Voucher.SettlementAt = &timestamp
								voucherAction = &entities.Action{
									Name:      system_action.VoucherSettlement.ActionName(),
									Type:      "",
									UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
									UTP:       actions.System.ActionName(),
									Perm:      "",
									Priv:      "",
									Policy:    "",
									Result:    string(states.ActionSuccess),
									Note:      "",
									Reasons:   nil,
									Data:      nil,
									CreatedAt: timestamp,
									Extended:  nil,
								}
							}

							if order.Invoice.Voucher.Percent > 0 {
								app.Globals.Logger.FromContext(ctx).Info("VoucherSettlement success",
									"fn", "Process",
									"state", state.Name(),
									"oid", order.OrderId,
									"voucher Percent", order.Invoice.Voucher.Percent,
									"voucherCode", order.Invoice.Voucher.Code)
							} else {
								app.Globals.Logger.FromContext(ctx).Info("VoucherSettlement success",
									"fn", "Process",
									"state", state.Name(),
									"oid", order.OrderId,
									"voucher Amount", voucherAmount,
									"voucher Code", order.Invoice.Voucher.Code)
							}
						} else {
							app.Globals.Logger.FromContext(ctx).Info("Order Invoice hasn't voucher",
								"fn", "Process",
								"state", state.Name(),
								"oid", order.OrderId)
						}

						paymentAction := &entities.Action{
							Name:      system_action.PaymentSuccess.ActionName(),
							Type:      "",
							UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
							UTP:       actions.System.ActionName(),
							Perm:      "",
							Priv:      "",
							Policy:    "",
							Result:    string(states.ActionSuccess),
							Note:      "",
							Reasons:   nil,
							Data:      nil,
							CreatedAt: time.Now().UTC(),
							Extended:  nil,
						}

						app.Globals.Logger.FromContext(ctx).Info("PaymentQueryResult success",
							"fn", "Process",
							"state", state.Name(),
							"oid", order.OrderId,
							"result", paymentResult)
						state.UpdateOrderAllSubPkg(ctx, order, paymentAction, voucherAction)
						response := events.ActionResponse{
							OrderId: order.OrderId,
							SIds:    nil,
						}
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()
						successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
						state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetOrderId(order.OrderId).SetBody(order).Build())
					}
				} else {
					app.Globals.Logger.FromContext(ctx).Info("PaymentQueryResult is Pending status",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"result", paymentResult)
					_, err := app.Globals.OrderRepository.Save(ctx, *order)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("OrderRepository.Save of update scheduler data failed",
							"fn", "Process",
							"state", state.Name(),
							"oid", order.OrderId,
							"event", event)
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetError(future.InternalError, "Unknown Err", futureData.Error().Reason()).Send()
						return
					}

					response := events.ActionResponse{
						OrderId: order.OrderId,
						SIds:    nil,
					}

					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()
					app.Globals.Logger.FromContext(ctx).Info("update scheduler data of order success",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId)
				}
			} else {
				order.OrderPayment[0].PaymentResult = &entities.PaymentResult{
					Result:      false,
					Reason:      "PaymentService Down",
					PaymentId:   "",
					InvoiceId:   0,
					Price:       nil,
					CardNumMask: "",
					CreatedAt:   time.Now().UTC(),
					Extended:    nil,
				}

				app.Globals.Logger.FromContext(ctx).Info("Process() => Get PaymentQueryResult from PaymentService failed",
					"oid", order.OrderId)
				paymentAction := &entities.Action{
					Name:      system_action.PaymentFail.ActionName(),
					Type:      "",
					UId:       order.BuyerInfo.BuyerId,
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

				state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
				response := events.ActionResponse{
					OrderId: order.OrderId,
					SIds:    nil,
				}

				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()
				failAction := state.GetAction(system_action.PaymentFail.ActionName())
				state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetOrderId(order.OrderId).SetBody(order).Build())
			}
		} else {
			app.Globals.Logger.FromContext(ctx).Error("Process() => event type not supported",
				"fn", "Process",
				"state", state.Name(),
				"event", event,
				"frame", iFrame)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", errors.New("event type invalid")).Send()
			return
		}
	} else {
		app.Globals.Logger.FromContext(ctx).Error("Frame Header/Body Invalid",
			"fn", "Process",
			"state", state.Name(),
			"iframe", iFrame)
	}
}
