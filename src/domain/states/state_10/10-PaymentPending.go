package state_10

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/go-framework/logger"
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
			logger.Err("Process() => iFrame.Body().Content() not a order, orderId: %d, state: %s ", iFrame.Header().Value(string(frame.HeaderOrderId)), state.Name())
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetData(app.Globals.Config.App.OrderPaymentCallbackUrlFail).
				Send()
			return
		}

		grandTotal, err := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
		if err != nil {
			logger.Err("Process() => order.Invoice.GrandTotal.Amount invalid, amount: %s, orderId: %d, error: %s", order.Invoice.GrandTotal.Amount, order.OrderId, err)
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
				logger.Err("Process() => order.Invoice.Voucher.Price.Amount invalid, price: %s, orderId: %d, error: %s", order.Invoice.Voucher.Price.Amount, order.OrderId, err)
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
				logger.Err("Process() => VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Error().Reason())
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
					Data:      nil,
					CreatedAt: time.Now().UTC(),
					Extended:  nil,
				}
			} else {
				if order.Invoice.Voucher.Percent > 0 {
					logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
				} else {
					logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
				}
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
					Data:      nil,
					CreatedAt: time.Now().UTC(),
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
			logger.Audit("Process() => Order state of all subpackages update to %s state, orderId: %d", state.Name(), order.OrderId)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetData(order.OrderPayment[0].PaymentResponse.CallBackUrl).
				Send()
			successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
			state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
			//}
		} else {
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

				logger.Err("Process() => PaymentService.OrderPayment in %s state failed, orderId: %d, error: %v",
					state.Name(), order.OrderId, futureData.Error().Reason())
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

				order.UpdatedAt = time.Now().UTC()
				for i := 0; i < len(order.Packages); i++ {
					order.Packages[i].UpdatedAt = time.Now().UTC()
					for j := 0; j < len(order.Packages[i].Subpackages); j++ {
						order.Packages[i].Subpackages[j].Tracking.State.Schedulers = []*entities.SchedulerData{
							{
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
								nil,
							},
						}
					}
				}

				state.UpdateOrderAllSubPkg(ctx, order, nil)
				_, err := app.Globals.OrderRepository.Save(ctx, *order)
				if err != nil {
					logger.Err("Process() => OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)

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

				logger.Audit("Process() => Order state of all subpackages update to %s state, orderId: %d", state.Name(), order.OrderId)
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
			logger.Err("Process() => OrderRepository.FindById failed, orderId: %d, paymentResult: %v, error: %v",
				iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64),
				iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult), err)

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
		logger.Audit("Process() => Order Received in %s state, orderId: %d", state.Name(), order.OrderId)
		if order.OrderPayment[0].PaymentResult.Result == false {
			logger.Audit("Process() => PaymentResult failed, orderId: %d", order.OrderId)
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
				Data:      nil,
				CreatedAt: time.Now().UTC(),
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
			var voucherAmount = 0
			if order.Invoice.Voucher != nil && order.Invoice.Voucher.Price != nil {
				voucherAmount, _ = strconv.Atoi(order.Invoice.Voucher.Price.Amount)
			}

			if order.Invoice.Voucher != nil && (order.Invoice.Voucher.Percent > 0 || voucherAmount > 0) {
				iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
				futureData := iFuture.Get()
				if futureData.Error() != nil {
					logger.Err("Process() => VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Error().Reason())
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
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					}
				} else {
					if order.Invoice.Voucher.Percent > 0 {
						logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
					} else {
						logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
					}

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
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					}
				}

				if order.Invoice.Voucher.Percent > 0 {
					logger.Audit("Process() => VoucherSettlement success, orderId: %d, voucher Percent: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
				} else {
					logger.Audit("Process() => VoucherSettlement success, orderId: %d, voucher Amount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
				}
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
				Reasons:   nil,
				Data:      nil,
				CreatedAt: time.Now().UTC(),
				Extended:  nil,
			}

			logger.Audit("Process() => PaymentResult success, orderId: %d", order.OrderId)
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
			logger.Err("Process() => received frame doesn't have a event, state: %s, frame: %v", state.String(), iFrame)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", nil).Send()
			return
		}

		if event.EventType() == events.Action {
			order, ok := iFrame.Body().Content().(*entities.Order)
			if !ok {
				logger.Err("Process() => received frame body not a PackageItem, state: %s, event: %v, frame: %v", state.String(), event, iFrame)
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
				logger.Err("Process() => received action not acceptable, state: %s, event: %v", state.String(), event)
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
				iFuture := app.Globals.PaymentService.GetPaymentResult(ctx, order.OrderId)
				futureData := iFuture.Get()
				if futureData.Error() != nil {
					logger.Err("Process() => PaymentService.GetPaymentResult failed, state: %s, orderId: %d, event: %v", state.String(), order.OrderId, event)
					_, err := app.Globals.OrderRepository.Save(ctx, *order)
					if err != nil {
						logger.Err("Process() => OrderRepository.Save of update scheduler data failed,  state: %s, orderId: %d, event: %v", state.String(), order.OrderId, event)
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
						logger.Audit("Process() => PaymentQueryResult failed, orderId: %d, result: %v", order.OrderId, paymentResult)
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
							Data:      nil,
							CreatedAt: time.Now().UTC(),
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
								logger.Err("Process() => VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Error().Reason())
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
									Data:      nil,
									CreatedAt: time.Now().UTC(),
									Extended:  nil,
								}
							} else {
								if order.Invoice.Voucher.Percent > 0 {
									logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
								} else {
									logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
								}

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
									Data:      nil,
									CreatedAt: time.Now().UTC(),
									Extended:  nil,
								}
							}

							if order.Invoice.Voucher.Percent > 0 {
								logger.Audit("Process() => VoucherSettlement success, orderId: %d, voucher Percent: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
							} else {
								logger.Audit("Process() => VoucherSettlement success, orderId: %d, voucher Amount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
							}
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
							Reasons:   nil,
							Data:      nil,
							CreatedAt: time.Now().UTC(),
							Extended:  nil,
						}

						logger.Audit("Process() => PaymentQueryResult success, orderId: %d, result: %v", order.OrderId, paymentResult)
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
					logger.Audit("Process() => PaymentQueryResult is Pending status, orderId: %d, result: %v", order.OrderId, paymentResult)
					_, err := app.Globals.OrderRepository.Save(ctx, *order)
					if err != nil {
						logger.Err("Process() => OrderRepository.Save of update scheduler data failed,  state: %s, orderId: %d, event: %v", state.String(), order.OrderId, event)
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetError(future.InternalError, "Unknown Err", futureData.Error().Reason()).Send()
						return
					}

					response := events.ActionResponse{
						OrderId: order.OrderId,
						SIds:    nil,
					}

					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()
					logger.Audit("Process() => update scheduler data of order success, state: %s, orderId: %d", state.Name(), order.OrderId)
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

				logger.Audit("Process() => Get PaymentQueryResult from PaymentService failed, orderId: %d", order.OrderId)
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
					Data:      nil,
					CreatedAt: time.Now().UTC(),
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
			logger.Err("Process() => event type not supported, state: %s, event: %v, frame: %v", state.String(), event, iFrame)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", errors.New("event type invalid")).Send()
			return
		}
	} else {
		logger.Err("HeaderOrderId or HeaderEvent of iFrame.Header not found, state: %s iframe: %v", state.Name(), iFrame)
	}
}
