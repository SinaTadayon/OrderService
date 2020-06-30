package state_11

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"text/template"
	"time"
)

const (
	stepName  string = "Payment_Success"
	stepIndex int    = 11
)

type paymentSuccessState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentSuccessState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentSuccessState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &paymentSuccessState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state paymentSuccessState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) && iFrame.Body().Content() != nil {
		order, ok := iFrame.Body().Content().(*entities.Order)
		if !ok {
			app.Globals.Logger.FromContext(ctx).Error("Content of frame body isn't an order",
				"fn", "Process",
				"state", state.Name(),
				"oid", iFrame.Header().Value(string(frame.HeaderOrderId)),
				"content", iFrame.Body().Content())
			return
		}

		var buyerNotificationAction *entities.Action = nil
		smsTemplate, err := template.New("SMS").Parse(app.Globals.SMSTemplate.OrderNotifyBuyerPaymentSuccessState)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("smsTemplate.Parse failed",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId,
				"message", app.Globals.SMSTemplate.OrderNotifyBuyerPaymentSuccessState,
				"error", err)
		} else {
			var buf bytes.Buffer
			err = smsTemplate.Execute(&buf, order.OrderId)
			newBuf := bytes.NewBuffer(bytes.Replace(buf.Bytes(), []byte("\\n"), []byte{10}, -1))
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("smsTemplate.Execute failed",
					"fn", "Process",
					"state", state.Name(),
					"oid", order.OrderId,
					"message", app.Globals.SMSTemplate.OrderNotifyBuyerPaymentSuccessState,
					"error", err)
			} else {
				buyerNotify := notify_service.SMSRequest{
					Phone: order.BuyerInfo.ShippingAddress.Mobile,
					Body:  newBuf.String(),
					User:  notify_service.BuyerUser,
				}

				buyerFutureData := app.Globals.NotifyService.NotifyBySMS(ctx, buyerNotify).Get()
				if buyerFutureData.Error() != nil {
					app.Globals.Logger.FromContext(ctx).Error("NotifyService.NotifyBySMS failed",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
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
						"request", buyerNotify,
						"oid", order.OrderId)
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

		paymentAction := &entities.Action{
			Name:      system_action.NextToState.ActionName(),
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

		if buyerNotificationAction != nil {
			state.UpdateOrderAllSubPkg(ctx, order, buyerNotificationAction)
		}
		state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
		app.Globals.Logger.FromContext(ctx).Debug("Order state of all subpackages update",
			"fn", "Process",
			"state", state.Name(),
			"oid", order.OrderId)
		state.StatesMap()[state.Actions()[0]].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
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
