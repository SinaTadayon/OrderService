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
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
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
				SetError(future.InternalError, "Unknown Err", errors.New("Unknown Err")).
				Send()
			return
		}
		state.paymentHandler(ctx, iFrame, order)

	} else if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) && iFrame.Header().KeyExists(string(frame.HeaderPaymentResult)) {
		state.paymentResultHandler(ctx, iFrame)

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
		state.eventHandler(ctx, iFrame, event)

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

func (state paymentPendingState) paymentHandler(ctx context.Context, iFrame frame.IFrame, order *entities.Order) {
	grandTotal, err := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("order.Invoice.GrandTotal.Amount invalid",
			"fn", "paymentHandler",
			"state", state.Name(),
			"amount", order.Invoice.GrandTotal.Amount,
			"oid", order.OrderId,
			"error", err)
		state.actionFailOfPaymentHandler(ctx, iFrame, order)
		return
	}

	if grandTotal.IsZero() && order.Invoice.Voucher != nil {
		app.Globals.Logger.FromContext(ctx).Info("invoice order with zero grand total and full voucher applied",
			"fn", "paymentHandler",
			"state", state.Name(),
			"oid", order.OrderId)
		state.voucherWithZeroGrandTotalHandler(ctx, iFrame, order)
	} else {
		if order.Invoice.Voucher != nil {
			app.Globals.Logger.FromContext(ctx).Info("invoice order with voucher applied",
				"fn", "paymentHandler",
				"state", state.Name(),
				"oid", order.OrderId)
		} else {
			app.Globals.Logger.FromContext(ctx).Info("invoice order without voucher applied",
				"fn", "paymentHandler",
				"state", state.Name(),
				"oid", order.OrderId)
		}

		var paymentRequest payment_service.PaymentRequest
		if order.Invoice.PaymentMethod == "IPG" {
			paymentRequest = payment_service.PaymentRequest{
				Method:   payment_service.IPG,
				Amount:   grandTotal.IntPart(),
				Currency: order.Invoice.GrandTotal.Currency,
				Gateway:  order.Invoice.PaymentGateway,
				OrderId:  order.OrderId,
				Mobile:   order.BuyerInfo.Mobile,
			}
		} else if order.Invoice.PaymentMethod == "MPG" {
			paymentRequest = payment_service.PaymentRequest{
				Method:   payment_service.MPG,
				Amount:   grandTotal.IntPart(),
				Currency: order.Invoice.GrandTotal.Currency,
				Gateway:  "",
				OrderId:  order.OrderId,
				Mobile:   order.BuyerInfo.Mobile,
			}
		}

		order.OrderPayment = []entities.PaymentService{
			{
				PaymentRequest: &entities.PaymentRequest{
					Price: &entities.Money{
						Amount:   order.Invoice.GrandTotal.Amount,
						Currency: order.Invoice.GrandTotal.Currency,
					},
					Gateway:   order.Invoice.PaymentGateway,
					CreatedAt: time.Now().UTC(),
					Mobile:    order.BuyerInfo.Mobile,
					Data:      nil,
					Extended:  nil,
				},
			},
		}

		iFuture := app.Globals.PaymentService.OrderPayment(ctx, paymentRequest)
		futureData := iFuture.Get()
		if futureData.Error() != nil {
			order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
				Result:      false,
				Reason:      strconv.Itoa(int(futureData.Error().Code())),
				Description: "",
				Response:    nil,
				CreatedAt:   time.Now().UTC(),
				Extended:    nil,
			}

			state.actionFailOfPaymentHandler(ctx, iFrame, order)
			return
		} else {
			if paymentResponse, ok := futureData.Data().(payment_service.IPGPaymentResponse); ok {
				ipgResponse := entities.PaymentIPGResponse{
					CallBackUrl: paymentResponse.CallbackUrl,
					InvoiceId:   paymentResponse.InvoiceId,
					PaymentId:   paymentResponse.PaymentId,
					Extended:    nil,
				}
				order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
					Result:      true,
					Reason:      "",
					Description: "",
					Response:    ipgResponse,
					CreatedAt:   time.Now().UTC(),
					Extended:    nil,
				}
			} else if paymentResponse, ok := futureData.Data().(payment_service.MPGPaymentResponse); ok {
				ipgResponse := entities.PaymentMPGResponse{
					HostRequest:     paymentResponse.HostRequest,
					HostRequestSign: paymentResponse.HostRequestSign,
					PaymentId:       paymentResponse.PaymentId,
					Extended:        nil,
				}
				order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
					Result:      true,
					Reason:      "",
					Description: "",
					Response:    ipgResponse,
					CreatedAt:   time.Now().UTC(),
					Extended:    nil,
				}
			}

			var expireTime time.Time
			timeUnit := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerStateTimeUintConfig].(string)
			if timeUnit == app.DurationTimeUnit {
				value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig].(time.Duration)
				expireTime = time.Now().UTC().Add(value)
			} else {
				value := app.Globals.FlowManagerConfig[app.FlowManagerSchedulerPaymentPendingStateConfig].(int)
				expireTime = time.Now().UTC().Add(
					time.Hour*time.Duration(0) +
						time.Minute*time.Duration(value) +
						time.Second*time.Duration(0))
			}

			app.Globals.Logger.FromContext(ctx).Debug("scheduler expireTime",
				"fn", "paymentHandler",
				"state", state.Name(),
				"oid", order.OrderId,
				"timeUnit", timeUnit,
				"expireTime", expireTime.UTC().String())

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

			state.UpdateOrderUpdateAt(ctx, order)
			_, err := app.Globals.CQRSRepository.CmdR().OrderCR().Update(ctx, *order)
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("OrderRepository.Save failed",
					"fn", "paymentHandler",
					"state", state.Name(),
					"oid", order.OrderId,
					"error", err)

				state.actionFailOfPaymentHandler(ctx, iFrame, order)
				return
			}

			app.Globals.Logger.FromContext(ctx).Debug("Order state of all subpackages update",
				"fn", "paymentHandler",
				"state", state.Name(),
				"oid", order.OrderId)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetData(order).
				Send()
			return
		}
	}
}

func (state paymentPendingState) voucherWithZeroGrandTotalHandler(ctx context.Context, iFrame frame.IFrame, order *entities.Order) {
	order.OrderPayment = []entities.PaymentService{
		{
			PaymentRequest: &entities.PaymentRequest{
				Price:     nil,
				Gateway:   "",
				CreatedAt: time.Now().UTC(),
				Mobile:    order.BuyerInfo.Mobile,
				Data:      nil,
				Extended:  nil,
			},

			PaymentResult: &entities.PaymentResult{
				Result:      true,
				Reason:      "Invoice paid by voucher",
				PaymentId:   "",
				InvoiceId:   0,
				Price:       nil,
				CardNumMask: "",
				Data:        nil,
				CreatedAt:   time.Now().UTC(),
				Extended:    nil,
			},
		},
	}

	order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
		Result:      true,
		Reason:      "",
		Description: "",
		Response: entities.PaymentIPGResponse{
			CallBackUrl: app.Globals.Config.App.OrderPaymentCallbackUrlSuccess + strconv.Itoa(int(order.OrderId)),
			InvoiceId:   0,
			PaymentId:   "",
			Extended:    nil,
		},
		CreatedAt: time.Now().UTC(),
		Extended:  nil,
	}

	var voucherAction *entities.Action
	iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		app.Globals.Logger.FromContext(ctx).Error("VoucherService.VoucherSettlement failed",
			"fn", "voucherWithZeroGrandTotalHandler",
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
			Reasons:   nil,
			Note:      "",
			Data:      nil,
			CreatedAt: timestamp,
			Extended:  nil,
		}
	} else {
		app.Globals.Logger.FromContext(ctx).Info("voucher applied in invoice of order success",
			"fn", "voucherWithZeroGrandTotalHandler",
			"state", state.Name(),
			"oid", order.OrderId,
			"voucher Amount", order.Invoice.Voucher.RoundupAppliedPrice.Amount,
			"voucher Code", order.Invoice.Voucher.Code)
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
	app.Globals.Logger.FromContext(ctx).Debug("Order state of all subpackages update",
		"fn", "voucherWithZeroGrandTotalHandler",
		"state", state.Name(),
		"oid", order.OrderId)
	future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
		SetData(order).
		Send()
	successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
	state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
}

func (state paymentPendingState) actionFailOfPaymentHandler(ctx context.Context, iFrame frame.IFrame, order *entities.Order) {
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

	response := entities.PaymentIPGResponse{
		CallBackUrl: app.Globals.Config.App.OrderPaymentCallbackUrlFail + strconv.Itoa(int(order.OrderId)),
		InvoiceId:   0,
		PaymentId:   "",
		Extended:    nil,
	}

	order.OrderPayment[0].PaymentResponse.Response = response

	future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
		SetData(order).
		Send()

	state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
	failAction := state.GetAction(system_action.PaymentFail.ActionName())
	state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
}

func (state paymentPendingState) paymentResultHandler(ctx context.Context, iFrame frame.IFrame) {
	order, err := app.Globals.CQRSRepository.QueryR().OrderQR().FindById(ctx, iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64))
	if err != nil {
		if err.Code() == repository.InternalErr {
			app.Globals.Logger.FromContext(ctx).Error("app.Globals.OrderRepository.FindById of payment result failed",
				"fn", "paymentResultHandler",
				"state", state.Name(),
				"oid", iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64),
				"paymentResult", iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult),
				"error", err)
		} else {
			app.Globals.Logger.FromContext(ctx).Error("orderId of payment result received not found",
				"fn", "paymentResultHandler",
				"state", state.Name(),
				"oid", iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64),
				"paymentResult", iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult),
				"error", err)
		}

		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetCapacity(1).SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).
			Send()
		return
	}

	ctx = context.WithValue(ctx, string(utils.CtxUserID), order.BuyerInfo.BuyerId)
	future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
		SetCapacity(1).Send()

	// issue SELLER-22
	if states.OrderStatus(order.Status) != states.OrderNewStatus {
		app.Globals.Logger.FromContext(ctx).Info("order status is not a NEW status",
			"fn", "paymentResultHandler",
			"state", state.Name(),
			"oid", order.OrderId,
			"status", order.Status,
			"paymentResult", iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult))
		return
	}

	order.OrderPayment[0].PaymentResult = iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult)
	app.Globals.Logger.FromContext(ctx).Info("PaymentResult received",
		"fn", "paymentResultHandler",
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
		failAction := state.GetAction(system_action.PaymentFail.ActionName())
		state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
		return
	} else {
		var voucherAction *entities.Action
		if order.Invoice.Voucher != nil {
			iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				app.Globals.Logger.FromContext(ctx).Error("VoucherService.VoucherSettlement failed",
					"fn", "paymentResultHandler",
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

				app.Globals.Logger.FromContext(ctx).Info("VoucherSettlement success",
					"fn", "paymentResultHandler",
					"state", state.Name(),
					"oid", order.OrderId,
					"voucher Amount", order.Invoice.Voucher.RoundupAppliedPrice.Amount,
					"voucher Code", order.Invoice.Voucher.Code)

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
		} else {
			app.Globals.Logger.FromContext(ctx).Info("Order Invoice hasn't voucher",
				"fn", "paymentResultHandler",
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
			"fn", "paymentResultHandler",
			"state", state.Name(),
			"oid", order.OrderId)
		state.UpdateOrderAllSubPkg(ctx, order, paymentAction, voucherAction)
		successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
		state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
	}
}

func (state paymentPendingState) eventHandler(ctx context.Context, iFrame frame.IFrame, event events.IEvent) {
	if event.EventType() == events.Action {
		order, ok := iFrame.Body().Content().(*entities.Order)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("content of frame body isn't order",
				"fn", "eventHandler",
				"state", state.Name(),
				"event", event,
				"frame", iFrame)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", errors.New("frame body invalid")).Send()
			return
		}

		// issue SELLER-22
		if states.OrderStatus(order.Status) != states.OrderNewStatus {
			app.Globals.Logger.FromContext(ctx).Info("order status is not a NEW status",
				"fn", "eventHandler",
				"state", state.Name(),
				"oid", order.OrderId,
				"status", order.Status,
				"event", event)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.NotAccepted, "Action Not Accepted", errors.New("Action Not Accepted")).Send()
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
				"fn", "eventHandler",
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

		app.Globals.Logger.FromContext(ctx).Debug("scheduler expireTime",
			"fn", "eventHandler",
			"state", state.Name(),
			"oid", order.OrderId,
			"timeUnit", timeUnit,
			"expireTime", expireTime.UTC().String())

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
				"fn", "eventHandler",
				"state", state.Name(),
				"oid", order.OrderId)
			iFuture := app.Globals.PaymentService.GetPaymentResult(ctx, order.OrderId)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				app.Globals.Logger.FromContext(ctx).Error("PaymentService.GetPaymentResult failed",
					"fn", "eventHandler",
					"state", state.Name(),
					"oid", order.OrderId,
					"event", event)

				state.UpdateOrderUpdateAt(ctx, order)
				_, err := app.Globals.CQRSRepository.CmdR().OrderCR().Update(ctx, *order)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("OrderRepository.Save of update scheduler data failed",
						"fn", "eventHandler",
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
						"fn", "eventHandler",
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
					if order.Invoice.Voucher != nil {
						iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
						futureData := iFuture.Get()
						if futureData.Error() != nil {
							app.Globals.Logger.FromContext(ctx).Error("VoucherService.VoucherSettlement failed",
								"fn", "eventHandler",
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
							app.Globals.Logger.FromContext(ctx).Info("Invoice paid by voucher Amount order success",
								"fn", "eventHandler",
								"state", state.Name(),
								"oid", order.OrderId,
								"voucher Amount", order.Invoice.Voucher.RoundupAppliedPrice.Amount,
								"voucher Code", order.Invoice.Voucher.Code)

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

						app.Globals.Logger.FromContext(ctx).Info("VoucherSettlement success",
							"fn", "eventHandler",
							"state", state.Name(),
							"oid", order.OrderId,
							"voucher Amount", order.Invoice.Voucher.RoundupAppliedPrice.Amount,
							"voucher Code", order.Invoice.Voucher.Code)

					} else {
						app.Globals.Logger.FromContext(ctx).Info("Order Invoice hasn't voucher",
							"fn", "eventHandler",
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
						"fn", "eventHandler",
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
					"fn", "eventHandler",
					"state", state.Name(),
					"oid", order.OrderId,
					"result", paymentResult)

				state.UpdateOrderUpdateAt(ctx, order)
				_, err := app.Globals.CQRSRepository.CmdR().OrderCR().Update(ctx, *order)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("OrderRepository.Save of update scheduler data failed",
						"fn", "eventHandler",
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
					"fn", "eventHandler",
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

			app.Globals.Logger.FromContext(ctx).Info("Get PaymentQueryResult from PaymentService failed",
				"fn", "eventHandler",
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
		app.Globals.Logger.FromContext(ctx).Error("event type not supported",
			"fn", "eventHandler",
			"state", state.Name(),
			"event", event,
			"frame", iFrame)
		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetError(future.InternalError, "Unknown Err", errors.New("event type invalid")).Send()
		return
	}
}
