package state_12

import (
	"bytes"
	"context"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"text/template"
	"time"
)

const (
	stepName  string = "Payment_Failed"
	stepIndex int    = 12
)

type paymentFailedState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentFailedState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentFailedState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &paymentFailedState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state paymentFailedState) Process(ctx context.Context, iFrame frame.IFrame) {

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
		smsTemplate, err := template.New("SMS").Parse(app.Globals.SMSTemplate.OrderNotifyBuyerPaymentFailedState)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("smsTemplate.Parse failed",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId,
				"message", app.Globals.SMSTemplate.OrderNotifyBuyerPaymentFailedState,
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
					"message", app.Globals.SMSTemplate.OrderNotifyBuyerPaymentFailedState,
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
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					}
				} else {
					app.Globals.Logger.FromContext(ctx).Debug("NotifyService.NotifyBySMS success",
						"fn", "Process",
						"state", state.Name(),
						"oid", order.OrderId,
						"request", buyerNotify)
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

		if buyerNotificationAction != nil {
			state.UpdateOrderAllSubPkg(ctx, order, buyerNotificationAction)
		}

		state.releasedStock(ctx, order)
		state.UpdateOrderAllStatus(ctx, order, states.OrderClosedStatus, states.PackageClosedStatus)
		_, err = app.Globals.OrderRepository.Save(ctx, *order)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("OrderRepository.Save failed",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId,
				"error", err)
		} else {
			app.Globals.Logger.FromContext(ctx).Debug("Order state of all subpackages update",
				"fn", "Process",
				"state", state.Name(),
				"oid", order.OrderId)
		}
	} else {
		app.Globals.Logger.FromContext(ctx).Error("Frame Header/Body Invalid",
			"fn", "Process",
			"state", state.Name(),
			"iframe", iFrame)
	}
}

func (state paymentFailedState) releasedStock(ctx context.Context, order *entities.Order) {

	for i := 0; i < len(order.Packages); i++ {
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			result := true
			stockActionDataList := make([]entities.StockActionData, 0, 32)
			for z := 0; z < len(order.Packages[i].Subpackages[j].Items); z++ {
				item := order.Packages[i].Subpackages[j].Items[z]
				requestStock := stock_service.RequestStock{
					InventoryId: item.InventoryId,
					Count:       int(item.Quantity),
				}

				iFuture := app.Globals.StockService.SingleStockAction(ctx, requestStock, order.OrderId,
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
						app.Globals.Logger.FromContext(ctx).Error("Released stock from stockService failed",
							"fn", "releasedStock",
							"state", state.Name(),
							"oid", order.OrderId,
							"response", response,
							"error", futureData.Error())
					} else {
						actionData := entities.StockActionData{
							InventoryId: requestStock.InventoryId,
							Quantity:    requestStock.Count,
							Result:      false,
						}
						stockActionDataList = append(stockActionDataList, actionData)
						app.Globals.Logger.FromContext(ctx).Error("Released stock from stockService failed",
							"fn", "releasedStock",
							"state", state.Name(),
							"oid", order.OrderId,
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
						"oid", order.OrderId)
				}
			}
			var stockAction *entities.Action
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

			state.UpdateSubPackage(ctx, order.Packages[i].Subpackages[j], stockAction)
		}
	}
}
