package state_30

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"time"
)

const (
	stepName  string = "Shipment_Pending"
	stepIndex int    = 30
)

type shipmentPendingState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &shipmentPendingState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &shipmentPendingState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &shipmentPendingState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state shipmentPendingState) Process(ctx context.Context, iFrame frame.IFrame) {

	if iFrame.Header().KeyExists(string(frame.HeaderSubpackage)) {
		subpkg, ok := iFrame.Header().Value(string(frame.HeaderSubpackage)).(*entities.Subpackage)
		if !ok {
			logger.Err("iFrame.Header() not a subpackage, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		if iFrame.Body().Content() == nil {
			logger.Err("Process() => iFrame.Body().Content() is nil, orderId: %d, sellerId: %d, itemId: %d, %s state ",
				subpkg.OrderId, subpkg.SellerId, subpkg.ItemId, state.Name())
			return
		}

		pkgItem, ok := iFrame.Body().Content().(*entities.PackageItem)
		if !ok {
			logger.Err("Process() => iFrame.Body().Content() is nil, orderId: %d, sellerId: %d, itemId: %d, %s state ",
				subpkg.OrderId, subpkg.SellerId, subpkg.ItemId, state.Name())
			return
		}

		// TODO must be read from reids config
		expiredTime := time.Now().UTC().Add(time.Hour*
			time.Duration(pkgItem.ShipmentSpec.ReactionTime) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(0))

		state.UpdateSubPackage(ctx, subpkg, nil)
		if subpkg.Tracking.State != nil {
			subpkg.Tracking.State.Data = map[string]interface{}{
				"expiredTime": expiredTime,
			}
			logger.Audit("Process() => set expiredTime: %s , orderId: %d, sellerId: %d, itemId: %d, %s state ",
				expiredTime, subpkg.OrderId, subpkg.SellerId, subpkg.ItemId, state.Name())
		}

		_, err := global.Singletons.SubPkgRepository.Update(ctx, *subpkg)
		if err != nil {
			logger.Err("Process() => SubPkgRepository.Update in %s state failed, orderId: %d, sellerId: %d, itemId: %d, error: %s",
				state.Name(), subpkg.OrderId, subpkg.SellerId, subpkg.ItemId, err.Error())
		} else {
			logger.Audit("Process() => Status of subpackage update to %s state, orderId: %d, sellerId: %d, itemId: %d",
				state.Name(), subpkg.OrderId, subpkg.SellerId, subpkg.ItemId)
		}

	} else if iFrame.Header().KeyExists(string(frame.HeaderPackage)) {
		pkgItem, ok := iFrame.Header().Value(string(frame.HeaderPackage)).(*entities.PackageItem)
		if !ok {
			logger.Err("iFrame.Header() not a sellerId, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		// TODO must be read from reids config
		expiredTime := time.Now().Add(time.Hour*
			time.Duration(72) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(0))

		for j := 0; j < len(pkgItem.Subpackages); j++ {
			state.UpdateSubPackage(ctx, &pkgItem.Subpackages[j], nil)
			if pkgItem.Subpackages[j].Tracking.State != nil {
				pkgItem.Subpackages[j].Tracking.State.Data = map[string]interface{}{
					"expiredTime": expiredTime,
				}
				logger.Audit("Process() => set expiredTime: %s , orderId: %d, sellerId: %d, itemId: %d, %s state ",
					expiredTime, pkgItem.Subpackages[j].OrderId, pkgItem.Subpackages[j].SellerId, pkgItem.Subpackages[j].ItemId, state.Name())
			}
		}

		_, err := global.Singletons.PkgItemRepository.Update(ctx, *pkgItem)
		if err != nil {
			logger.Err("PkgItemRepository.Update in %s state failed, orderId: %d, sellerId: %d, error: %s",
				state.Name(), pkgItem.OrderId, pkgItem.SellerId, err.Error())
		} else {
			logger.Audit("Process() => Status of subpackage update to %s state, orderId: %d, sellerId: %d",
				state.Name(), pkgItem.OrderId, pkgItem.SellerId)
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
			pkgItem, ok := iFrame.Body().Content().(*entities.PackageItem)
			if !ok {
				logger.Err("Process() => received frame body not a PackageItem, state: %s, event: %v, frame: %v", state.String(), event, iFrame)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", errors.New("frame body invalid")).Send()
				return
			}

			actionData, ok := event.Data().(events.ActionData)
			if !ok {
				logger.Err("Process() => received action event data invalid, state: %s, event: %v", state.String(), event)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", errors.New("Action Data event invalid")).Send()
				return
			}

			// TODO cleaning subpackage after merging subpackages
			var newSubPackage *entities.Subpackage
			var nextActionState states.IState
			var shipmentPendingAction *entities.Action

			// iterate subpackages
			for _, eventSubPkg := range actionData.SubPackages {
				for i := 0; i < len(pkgItem.Subpackages); i++ {
					if eventSubPkg.ItemId == pkgItem.Subpackages[i].ItemId && pkgItem.Subpackages[i].Status == state.Name() {
						var findAction = false
						for action, nextState := range state.StatesMap() {
							if action.ActionType().ActionName() == event.Action().ActionType().ActionName() &&
								action.ActionEnum().ActionName() == event.Action().ActionEnum().ActionName() {
								findAction = true
								var newPkgItems []entities.Item

								// iterate items
								for _, actionItem := range eventSubPkg.Items {
									for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
										if actionItem.InventoryId == pkgItem.Subpackages[i].Items[j].InventoryId {
											nextActionState = nextState

											if actionItem.Quantity != pkgItem.Subpackages[i].Items[j].Quantity {
												if newSubPackage == nil {
													newSubPackage = pkgItem.Subpackages[i].DeepCopy()
													newSubPackage.ItemId = 0
													newSubPackage.Items = make([]entities.Item, 0, len(eventSubPkg.Items))

													shipmentPendingAction = &entities.Action{
														Name:      action.ActionEnum().ActionName(),
														Type:      action.ActionType().ActionName(),
														Result:    string(states.ActionSuccess),
														Reasons:   actionItem.Reasons,
														CreatedAt: time.Now().UTC(),
													}
												}

												if newPkgItems == nil {
													newPkgItems = make([]entities.Item, 0, len(pkgItem.Subpackages[i].Items))
												}

												pkgItem.Subpackages[i].Items[j].Quantity -= actionItem.Quantity
												pkgItem.Subpackages[i].Items[j].Invoice.Total = pkgItem.Subpackages[i].Items[j].Invoice.Unit *
													uint64(pkgItem.Subpackages[i].Items[j].Quantity)
												newPkgItem := pkgItem.Subpackages[i].Items[j].DeepCopy()
												newPkgItems = append(newPkgItems, *newPkgItem)

												newItem := pkgItem.Subpackages[i].Items[j].DeepCopy()
												newItem.Quantity = actionItem.Quantity
												newItem.Reasons = actionItem.Reasons
												newItem.Invoice.Total = newItem.Invoice.Unit * uint64(newItem.Quantity)
												if newSubPackage != nil {
													newSubPackage.Items = append(newSubPackage.Items, *newItem)
												}
											} else {
												newItem := pkgItem.Subpackages[i].Items[j].DeepCopy()
												newItem.Quantity = actionItem.Quantity
												newItem.Reasons = actionItem.Reasons
												newItem.Invoice.Total = newItem.Invoice.Unit * uint64(newItem.Quantity)
												if newSubPackage == nil {
													newSubPackage = pkgItem.Subpackages[i].DeepCopy()
													newSubPackage.ItemId = 0
													newSubPackage.Items = make([]entities.Item, 0, len(eventSubPkg.Items))

													shipmentPendingAction = &entities.Action{
														Name:      action.ActionEnum().ActionName(),
														Type:      action.ActionType().ActionName(),
														Result:    string(states.ActionSuccess),
														Reasons:   actionItem.Reasons,
														CreatedAt: time.Now().UTC(),
													}
												}
												newSubPackage.Items = append(newSubPackage.Items, *newItem)
											}
										}
									}
								}

								// create diff packages
								if newPkgItems != nil {
									pkgItem.Subpackages[i].Items = newPkgItems
								} else {
									if newSubPackage != nil &&
										len(newSubPackage.Items) == len(pkgItem.Subpackages[i].Items) {
										pkgItem.Subpackages[i].Items = nil
									}
								}
							}
						}

						if !findAction {
							logger.Err("Process() => received action not acceptable, state: %s, event: %v", state.String(), event)
							future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
								SetError(future.NotAccepted, "Action Not Accepted", errors.New("Action Not Accepted")).Send()
							return
						}
					}
				}
			}

			// remove subpackage with zero of items
			var subpackages = make([]entities.Subpackage, 0, len(pkgItem.Subpackages))
			for i := 0; i < len(pkgItem.Subpackages); i++ {
				if len(pkgItem.Subpackages[i].Items) > 0 {
					subpackages = append(subpackages, pkgItem.Subpackages[i])
				}
			}
			pkgItem.Subpackages = subpackages

			// update and save newSubpackage
			if newSubPackage != nil {
				state.UpdateSubPackage(ctx, newSubPackage, shipmentPendingAction)
				err := global.Singletons.SubPkgRepository.Save(ctx, newSubPackage)
				if err != nil {
					logger.Err("Process() => SubPkgRepository.Save in %s state failed, orderId: %d, sellerId: %d, event: %v, error: %s", state.Name(),
						newSubPackage.OrderId, newSubPackage.SellerId, event, err.Error())
					// TODO must distinct system error from update version error
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.InternalError, "Unknown Err", err).Send()
					return
				} else {
					logger.Audit("Process() => Status of new subpackage update to %v event, orderId: %d, sellerId: %d, itemId: %d",
						event, newSubPackage.OrderId, newSubPackage.SellerId, newSubPackage.ItemId)
				}

				if nextActionState != nil {
					if event.Action().ActionEnum() == seller_action.Cancel {
						var rejectedSubtotal uint64 = 0
						var rejectedDiscount uint64 = 0

						for i := 0; i < len(newSubPackage.Items); i++ {
							rejectedSubtotal += newSubPackage.Items[i].Invoice.Total
							rejectedDiscount += newSubPackage.Items[i].Invoice.Discount
						}
						pkgItem.Invoice.Subtotal -= rejectedSubtotal
						pkgItem.Invoice.Discount -= rejectedDiscount
					}
					_, err := global.Singletons.PkgItemRepository.Update(ctx, *pkgItem)
					if err != nil {
						logger.Err("Process() => PkgItemRepository.Update in %s state failed, orderId: %d, sellerId: %d, event: %v, error: %s", state.Name(),
							pkgItem.OrderId, pkgItem.SellerId, event, err.Error())
					}

					response := events.ActionResponse{
						OrderId: newSubPackage.OrderId,
						ItemsId: newSubPackage.ItemId,
					}

					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetData(response).Send()
					nextActionState.Process(ctx, frame.Factory().SetSubpackage(newSubPackage).SetBody(pkgItem).Build())
				}
			} else {
				for i := 0; i < len(pkgItem.Subpackages); i++ {
					state.UpdateSubPackage(ctx, &pkgItem.Subpackages[i], shipmentPendingAction)
					err := global.Singletons.SubPkgRepository.Save(ctx, &pkgItem.Subpackages[i])
					if err != nil {
						logger.Err("Process() => SubPkgRepository.Save in %s state failed, orderId: %d, sellerId: %d, action: %s, error: %s", state.Name(),
							&pkgItem.Subpackages[i].OrderId, &pkgItem.Subpackages[i].SellerId, shipmentPendingAction.Name, err.Error())
						// TODO must distinct system error from update version error
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetError(future.InternalError, "Unknown Err", err).Send()
						return
					} else {
						logger.Audit("Process() => Status of new subpackage update to %s action, orderId: %d, sellerId: %d, itemId: %d",
							shipmentPendingAction.Name, &pkgItem.Subpackages[i].OrderId, &pkgItem.Subpackages[i].SellerId, pkgItem.Subpackages[i].ItemId)
					}
				}
				if nextActionState != nil {
					response := events.ActionResponse{
						OrderId: pkgItem.OrderId,
						ItemsId: pkgItem.Subpackages[0].ItemId,
					}

					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetData(response).Send()
					nextActionState.Process(ctx, frame.Factory().SetSellerId(pkgItem.SellerId).SetPackage(pkgItem).Build())
				}
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

//func (shipmentPending shipmentPendingState) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
//
//	if param == nil {
//		shipmentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//		shipmentPending.updateOrderItemsProgress(ctx, &order, itemsId, SellerShipmentPending, true, "", nil, true, states.OrderInProgressStatus)
//		if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//			returnChannel := make(chan future.IDataFuture, 1)
//			defer close(returnChannel)
//			returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//			return future.NewFuture(returnChannel, 1, 1)
//		}
//		returnChannel := make(chan future.IDataFuture, 1)
//		defer close(returnChannel)
//		returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
//		return future.NewFuture(returnChannel, 1, 1)
//	} else {
//		req, ok := param.(*message.RequestSellerOrderAction)
//		if ok != true {
//			if param == "actionExpired" {
//				iPromise := global.Singletons.StockService.BatchStockActions(ctx, nil, StockReleased)
//				futureData := iPromise.Get()
//				if futureData == nil {
//					if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//					}
//					logger.Err("StockService future channel has been closed, order: %d", order.OrderId)
//				} else if futureData.Ex != nil {
//					if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//					}
//					logger.Err("released stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
//					returnChannel := make(chan future.IDataFuture, 1)
//					defer close(returnChannel)
//					returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//					return future.NewFuture(returnChannel, 1, 1)
//				}
//
//				if len(order.Items) == len(itemsId) {
//					shipmentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderClosedStatus, false)
//				} else {
//					shipmentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//				}
//
//				shipmentPending.updateOrderItemsProgress(ctx, &order, itemsId, AutoReject, false, "Actions Expired", nil, false, states.OrderClosedStatus)
//				if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//					returnChannel := make(chan future.IDataFuture, 1)
//					defer close(returnChannel)
//					returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//					return future.NewFuture(returnChannel, 1, 1)
//				}
//
//				return shipmentPending.Childes()[1].ProcessOrder(ctx, order, itemsId, nil)
//			} else {
//				logger.Err("param not a message.RequestSellerOrderAction type , order: %v", order)
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//		}
//
//		if !shipmentPending.validateAction(ctx, &order, itemsId) {
//			logger.Err("%s step received invalid action, order: %v, action: %s", shipmentPending.Name(), order, req.Action)
//			returnChannel := make(chan future.IDataFuture, 1)
//			defer close(returnChannel)
//			returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.NotAccepted, Reason: "Actions Expired"}}
//			return future.NewFuture(returnChannel, 1, 1)
//		}
//
//		if req.Data == nil {
//			returnChannel := make(chan future.IDataFuture, 1)
//			defer close(returnChannel)
//			returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.BadRequest, Reason: "Reason Get Required"}}
//			return future.NewFuture(returnChannel, 1, 1)
//		}
//
//		if req.Action == "success" {
//			actionData, ok := req.Data.(*message.RequestSellerOrderAction_Success)
//			if ok != true {
//				logger.Err("request data not a message.RequestSellerOrderAction_Success type , order: %v", order)
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			shipmentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//			shipmentPending.updateOrderItemsProgress(ctx, &order, itemsId, Shipped, true, "", actionData, false, states.OrderInProgressStatus)
//			if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			return shipmentPending.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
//		} else if req.Action == "failed" {
//			actionData, ok := req.Data.(*message.RequestSellerOrderAction_Failed)
//			if ok != true {
//				logger.Err("request data not a message.RequestSellerOrderAction_Failed type , order: %v", order)
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			iPromise := global.Singletons.StockService.BatchStockActions(ctx, nil, StockReleased)
//			futureData := iPromise.Get()
//			if futureData == nil {
//				if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//				}
//				logger.Err("StockService future channel has been closed, orderId: %d", order.OrderId)
//			} else if futureData.Ex != nil {
//				if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//				}
//				logger.Err("released stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			if len(order.Items) == len(itemsId) {
//				shipmentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderClosedStatus, false)
//			} else {
//				shipmentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//			}
//
//			shipmentPending.updateOrderItemsProgress(ctx, &order, itemsId, Shipped, false, actionData.Failed.Reason, nil, false, states.OrderClosedStatus)
//			if err := shipmentPending.persistOrder(ctx, &order); err != nil {
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			return shipmentPending.Childes()[1].ProcessOrder(ctx, order, itemsId, nil)
//		}
//
//		logger.Err("%s step received invalid action, order: %v, action: %s", shipmentPending.Name(), order, req.Action)
//		returnChannel := make(chan future.IDataFuture, 1)
//		defer close(returnChannel)
//		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		return future.NewFuture(returnChannel, 1, 1)
//	}
//}
//
//func (shipmentPending shipmentPendingState) persistOrder(ctx context.Context, order *entities.Order) error {
//	_, err := global.Singletons.OrderRepository.Save(*order)
//	if err != nil {
//		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", shipmentPending.Name(), order, err.Error())
//	}
//
//	return err
//}
//
//func (shipmentPending shipmentPendingState) validateAction(ctx context.Context, order *entities.Order, itemsId []uint64) bool {
//	if itemsId != nil && len(itemsId) > 0 {
//		for _, id := range itemsId {
//			for i := 0; i < len(order.Items); i++ {
//				length := len(order.Items[i].Progress.StepsHistory) - 1
//				if order.Items[i].ItemId == id && order.Items[i].Progress.StepsHistory[length].Name != shipmentPending.Name() {
//					return false
//				}
//			}
//		}
//	} else {
//		for i := 0; i < len(order.Items); i++ {
//			length := len(order.Items[i].Progress.StepsHistory) - 1
//			if order.Items[i].Progress.StepsHistory[length].Name != shipmentPending.Name() {
//				return false
//			}
//		}
//	}
//
//	return true
//}
//
//func (shipmentPending shipmentPendingState) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, reason string, req *message.RequestSellerOrderAction_Success, isSetExpireTime bool, itemStatus string) {
//
//	findFlag := false
//	if itemsId != nil && len(itemsId) > 0 {
//		for _, id := range itemsId {
//			findFlag = false
//			for i := 0; i < len(order.Items); i++ {
//				if order.Items[i].ItemId == id {
//					findFlag = true
//					if req != nil {
//						order.Items[i].ShipmentDetails.SellerShipmentDetail = entities.ShippingDetail{
//							TrackingNumber: req.Success.TrackingId,
//							ShippingMethod: req.Success.ShipmentMethod,
//						}
//						break
//					} else {
//						shipmentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
//					}
//				}
//			}
//			if !findFlag {
//				logger.Err("%s received itemId %d not exist in order, orderId: %d", shipmentPending.Name(), id, order.OrderId)
//			}
//		}
//	} else {
//		for i := 0; i < len(order.Items); i++ {
//			shipmentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
//		}
//	}
//}
//
//func (shipmentPending shipmentPendingState) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
//	actionName string, result bool, reason string, isSetExpireTime bool, itemStatus string) {
//
//	order.Items[index].Status = itemStatus
//	order.Items[index].UpdatedAt = time.Now().UTC()
//
//	length := len(order.Items[index].Progress.StepsHistory) - 1
//
//	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
//		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
//	}
//
//	var action entities.Action
//	if isSetExpireTime {
//		expiredTime := order.Items[index].UpdatedAt.Add(time.Hour*
//			time.Duration(order.Items[index].ShipmentSpec.ReactionTime) +
//			time.Minute*time.Duration(0) +
//			time.Second*time.Duration(0))
//
//		action = entities.Action{
//			Name:   actionName,
//			Result: result,
//			Reason: reason,
//			Data: map[string]interface{}{
//				"expiredTime": expiredTime,
//			},
//			CreatedAt: order.Items[index].UpdatedAt,
//		}
//	} else {
//		action = entities.Action{
//			Name:      actionName,
//			Result:    result,
//			Reason:    reason,
//			CreatedAt: order.Items[index].UpdatedAt,
//		}
//	}
//
//	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
//}
//
////
////import (
////	"gitlab.faza.io/order-project/order-service"
////	OrderService "gitlab.faza.io/protos/order"
////)
////
////func ShipmentPendingEnteredDetail(ppr PaymentPendingRequest, req *OrderService.ShipmentDetailRequest) error {
////	ppr.ShippingDetail.ShippingDetail.ShipmentProvider = req.ShipmentProvider
////	ppr.ShippingDetail.ShippingDetail.ShipmentTrackingNumber = req.ShipmentTrackingNumber
////	ppr.ShippingDetail.ShippingDetail.Description = req.GetDescription()
////	err := main.MoveOrderToNewState("seller", "", main.Shipped, "shipped", ppr)
////	if err != nil {
////		return err
////	}
////	return nil
////}
