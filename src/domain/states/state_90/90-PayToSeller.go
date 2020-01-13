package state_90

import (
	"bytes"
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/events"
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
	stepName  string = "Pay_To_Seller"
	stepIndex int    = 90
)

type payToSellerState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &payToSellerState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &payToSellerState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &payToSellerState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state payToSellerState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderSIds)) {
		//subpackages, ok := iFrame.Header().Value(string(frame.HeaderSubpackages)).([]*entities.Subpackage)
		//if !ok {
		//	logger.Err("iFrame.Header() not a subpackages, frame: %v, %s state ", iFrame, state.Name())
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

		smsTemplate, err := template.New("SMS").Parse(app.Globals.SMSTemplate.OrderNotifyBuyerReturnRejectedToPayToSellerState)
		if err != nil {
			logger.Err("Process() => smsTemplate.Parse failed, state: %s, orderId: %d, message: %s, err: %s",
				state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifyBuyerReturnRejectedToPayToSellerState, err)
		} else {
			var buf bytes.Buffer
			err = smsTemplate.Execute(&buf, templateData)
			newBuf := bytes.NewBuffer(bytes.Replace(buf.Bytes(), []byte("\\n"), []byte{10}, -1))
			if err != nil {
				logger.Err("Process() => smsTemplate.Execute failed, state: %s, orderId: %d, message: %s, err: %s",
					state.Name(), pkgItem.OrderId, app.Globals.SMSTemplate.OrderNotifyBuyerReturnRejectedToPayToSellerState, err)
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

		var settlementStockAction *entities.Action
		if err := state.settlementStock(ctx, sids, pkgItem); err != nil {
			settlementStockAction = &entities.Action{
				Name:      system_action.StockSettlement.ActionName(),
				Type:      "",
				UId:       0,
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
			settlementStockAction = &entities.Action{
				Name:      system_action.StockSettlement.ActionName(),
				Type:      "",
				UId:       0,
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

		for i := 0; i < len(sids); i++ {
			for j := 0; j < len(pkgItem.Subpackages); j++ {
				if pkgItem.Subpackages[j].SId == sids[i] {
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], buyerNotificationAction)
					state.UpdateSubPackage(ctx, pkgItem.Subpackages[j], settlementStockAction)
				}
			}
		}

		order, err := app.Globals.OrderRepository.FindById(ctx, pkgItem.OrderId)
		if err != nil {
			logger.Err("OrderRepository.FindById in %s state failed, orderId: %d, pid: %d, sids: %v, error: %v",
				state.Name(), pkgItem.OrderId, pkgItem.PId, sids, err)
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
				logger.Audit("set pkgItem status to closed, state: %s, orderId: %d, pid: %d",
					state.Name(), order.Packages[i].OrderId, order.Packages[i].PId)
			}
		}

		// TODO optimize write performance with journal and w options
		updatePkgItem, err := app.Globals.PkgItemRepository.Update(ctx, *pkgItem)
		if err != nil {
			logger.Err("Process() => PkgItemRepository.Update failed, state: %s, orderId: %d, pid: %d, sids: %v, error: %v",
				state.Name(), pkgItem.OrderId, pkgItem.PId, sids, err)
			return
		}

		response := events.ActionResponse{
			OrderId: pkgItem.OrderId,
			SIds:    sids,
		}

		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).SetData(response).Send()
		logger.Audit("Process() => Set State of subpackages success, state: %s, orderId: %d, pid: %d, sids: %v",
			state.Name(), updatePkgItem.OrderId, updatePkgItem.PId, sids)

		findFlag = true
		for i := 0; i < len(order.Packages); i++ {
			if order.Packages[i].Status != string(states.PackageClosedStatus) {
				findFlag = false
				break
			}
		}

		if findFlag {
			state.SetOrderStatus(ctx, order, states.OrderClosedStatus)
			err = app.Globals.OrderRepository.UpdateStatus(ctx, order)
			if err != nil {
				logger.Err("update order status to closed failed, state: %s, orderId: %d, error: %v",
					state.Name(), order.OrderId, err)
			} else {
				logger.Audit("update order status to closed success, state: %s, orderId: %d", state.Name(), order.OrderId)
			}
		}
	} else {
		logger.Err("iFrame.Header() not a subpackage or pid not found, state: %s iframe: %v", state.Name(), iFrame)
	}
}

func (state payToSellerState) settlementStock(ctx context.Context, sids []uint64, pkgItem *entities.PackageItem) error {

	var inventories = make(map[string]int, 32)
	for i := 0; i < len(pkgItem.Subpackages); i++ {
		for z := 0; z < len(pkgItem.Subpackages[i].Items); z++ {
			item := pkgItem.Subpackages[i].Items[z]
			inventories[item.InventoryId] = int(item.Quantity)
		}
	}

	iFuture := app.Globals.StockService.BatchStockActions(ctx, inventories,
		system_action.New(system_action.StockSettlement))
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		logger.Err("Settlement stock from stockService failed, state: %s, orderId: %d, pid: %d, sids: %v, error: %s",
			state.Name(), pkgItem.OrderId, pkgItem.PId, sids, futureData.Error())
		return futureData.Error().Reason()
	}

	logger.Audit("Settlement stock success, state: %s, orderId: %d, pid: %d, sids: %v",
		state.Name(), pkgItem.OrderId, pkgItem.PId, sids)
	return nil
}
