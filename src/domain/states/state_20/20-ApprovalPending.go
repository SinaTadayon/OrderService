package state_20

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
	stepName  string = "Approval_Pending"
	stepIndex int    = 20
	//Approved               = "Approved"
	//ApprovalPending        = "ApprovalPending"
	//StockReleased          = "StockReleased"
	//AutoReject             = "AutoReject"
)

type approvalPendingState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &approvalPendingState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &approvalPendingState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &approvalPendingState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state approvalPendingState) Process(ctx context.Context, iFrame frame.IFrame) {

	if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) {
		if iFrame.Body().Content() == nil {
			logger.Err("Process() => iFrame.Body().Content() is nil, orderId: %d, %s state ",
				iFrame.Header().Value(string(frame.HeaderOrderId)), state.Name())
			return
		}

		order, ok := iFrame.Body().Content().(*entities.Order)
		if !ok {
			logger.Err("Process() => iFrame.Body().Content() not a order, orderId: %d, %s state ",
				iFrame.Header().Value(string(frame.HeaderOrderId)), state.Name())
			return
		}

		// TODO must be read from reids config
		expiredTime := time.Now().Add(time.Hour*
			time.Duration(72) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(0))

		state.UpdateOrderAllSubPkg(ctx, order)
		for i := 0; i < len(order.Packages); i++ {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if order.Packages[i].Subpackages[j].Tracking.State != nil {
					order.Packages[i].Subpackages[j].Tracking.State.Data = map[string]interface{}{
						"expiredTime": expiredTime,
					}
				}
			}
		}
		_, err := global.Singletons.OrderRepository.Save(ctx, *order)
		if err != nil {
			logger.Err("Process() => OrderRepository.Save in %s state failed, orderId: %d, error: %s", state.Name(), order.OrderId, err.Error())
		} else {
			logger.Audit("Process() => Status of all subpackage update to ApprovalPending, orderId: %d", order.OrderId)
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
			pkgItem, ok := iFrame.Body().Content().(entities.PackageItem)
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
			var approvalPendingAction *entities.Action

			// iterate subpackages
			for _, eventSubPkg := range actionData.SubPackages {
				for i := 0; i < len(pkgItem.Subpackages); i++ {
					if eventSubPkg.ItemId == pkgItem.Subpackages[i].ItemId && pkgItem.Subpackages[i].Status == state.Name() {
						var findAction = false
						for action, nextState := range state.StatesMap() {
							if action.ActionType().ActionName() == event.Action().ActionType().ActionName() &&
								action.ActionEnum().ActionName() == event.Action().ActionEnum().ActionName() {
								findAction = true

								//var newSubPkg *entities.Subpackage
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

													approvalPendingAction = &entities.Action{
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
												pkgItem.Subpackages[i].Items[j].Invoice.Total = pkgItem.Subpackages[i].Items[j].Invoice.Unit * uint64(pkgItem.Subpackages[i].Items[j].Quantity)
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
												if newPkgItems != nil {
													newPkgItem := pkgItem.Subpackages[i].Items[j].DeepCopy()
													newPkgItems = append(newPkgItems, *newPkgItem)
												}
											}
										}
									}
								}

								// create diff packages
								if newPkgItems != nil {
									pkgItem.Subpackages[i].Items = newPkgItems
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

			if newSubPackage != nil {
				state.UpdateSubPackage(ctx, newSubPackage, approvalPendingAction)
				err := global.Singletons.SubPkgRepository.Save(ctx, newSubPackage)
				if err != nil {
					logger.Err("Process() => SubPkgRepository.Save in %s state failed, orderId: %d, sellerId: %d, action: %s, error: %s", state.Name(),
						newSubPackage.OrderId, newSubPackage.SellerId, approvalPendingAction.Name, err.Error())
					// TODO must distinct system error from update version error
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.InternalError, "Unknown Err", err).Send()
					return
				} else {
					logger.Audit("Process() => Status of new subpackage update to %s action, orderId: %d, sellerId: %d, itemId: %d",
						approvalPendingAction.Name, newSubPackage.OrderId, newSubPackage.SellerId, newSubPackage.ItemId)
				}

				//for z := 0; z < len(newSubPackages); z++ {
				//	state.UpdateSubPackage(ctx, &newSubPackages[z], approvalPendingAction)
				//	err := global.Singletons.SubPkgRepository.Save(ctx, &newSubPackages[z])
				//	if err != nil {
				//		logger.Err("Process() => SubPkgRepository.Save in %s state failed, orderId: %d, sellerId: %d, action: %s, error: %s", state.Name(),
				//			newSubPackages[z].OrderId, newSubPackages[z].SellerId, approvalPendingAction.Name, err.Error())
				//		// TODO must distinct system error from update version error
				//		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				//			SetError(future.InternalError, "Unknown Err", err).Send()
				//		return
				//	} else {
				//		logger.Audit("Process() => Status of new subpackage update to %s action, orderId: %d, sellerId: %d, itemId: %d",
				//			approvalPendingAction.Name, newSubPackages[z].OrderId, newSubPackages[z].SellerId, newSubPackages[z].ItemId)
				//	}
				//}
				if nextActionState != nil {
					if event.Action().ActionEnum() != seller_action.Approve {
						var rejectedSubtotal uint64 = 0
						var rejectedDiscount uint64 = 0

						for i := 0; i < len(newSubPackage.Items); i++ {
							rejectedSubtotal += newSubPackage.Items[i].Invoice.Total
							rejectedDiscount += newSubPackage.Items[i].Invoice.Discount
						}
						pkgItem.Invoice.Subtotal -= rejectedSubtotal
						pkgItem.Invoice.Discount -= rejectedDiscount
					}
					newPkgItem, err := global.Singletons.PkgItemRepository.Update(ctx, pkgItem)
					if err != nil {
						logger.Err("Process() => PkgItemRepository.Update in %s state failed, orderId: %d, sellerId: %d, event: %v, error: %s", state.Name(),
							pkgItem.OrderId, pkgItem.SellerId, event, err.Error())
						newPkgItem = nil
					} else {
						nextActionState.Process(ctx, frame.FactoryOf(iFrame).SetSubpackage(newSubPackage).SetBody(newPkgItem).Build())
						return
					}

					nextActionState.Process(ctx, frame.FactoryOf(iFrame).SetSubpackage(newSubPackage).Build())
				}
			} else {
				for i := 0; i < len(pkgItem.Subpackages); i++ {
					state.UpdateSubPackage(ctx, &pkgItem.Subpackages[i], approvalPendingAction)
					err := global.Singletons.SubPkgRepository.Save(ctx, &pkgItem.Subpackages[i])
					if err != nil {
						logger.Err("Process() => SubPkgRepository.Save in %s state failed, orderId: %d, sellerId: %d, action: %s, error: %s", state.Name(),
							&pkgItem.Subpackages[i].OrderId, &pkgItem.Subpackages[i].SellerId, approvalPendingAction.Name, err.Error())
						// TODO must distinct system error from update version error
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetError(future.InternalError, "Unknown Err", err).Send()
						return
					} else {
						logger.Audit("Process() => Status of new subpackage update to %s action, orderId: %d, sellerId: %d, itemId: %d",
							approvalPendingAction.Name, &pkgItem.Subpackages[i].OrderId, &pkgItem.Subpackages[i].SellerId, pkgItem.Subpackages[i].ItemId)
					}
				}
				if nextActionState != nil {
					nextActionState.Process(ctx, iFrame)
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

//func (sellerApprovalPending approvalPendingState) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {
//
//	if param == nil {
//		logger.Audit("Order Received in %s step, orderId: %d, Actions: %s", sellerApprovalPending.Name(), order.OrderId, ApprovalPending)
//		sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, false)
//		sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, ApprovalPending, true, "", true, states.OrderInProgressStatus)
//		if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//		}
//		returnChannel := make(chan future.IDataFuture, 1)
//		defer close(returnChannel)
//		returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
//		return future.NewFuture(returnChannel, 1, 1)
//	} else {
//		logger.Audit("Order Received in %s step, orderId: %d, Actions: %s", sellerApprovalPending.Name(), order.OrderId, Approved)
//		req, ok := param.(*message.RequestSellerOrderAction)
//		if ok != true {
//			if param == "actionExpired" {
//				iPromise := global.Singletons.StockService.BatchStockActions(ctx, order, itemsId, StockReleased)
//				futureData := iPromise.Get()
//				if futureData == nil {
//					if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//					}
//					logger.Err("StockService future channel has been closed, order: %d", order.OrderId)
//				} else if futureData.Ex != nil {
//					if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//					}
//					logger.Err("released stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
//					returnChannel := make(chan future.IDataFuture, 1)
//					defer close(returnChannel)
//					returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//					return future.NewFuture(returnChannel, 1, 1)
//				}
//
//				if len(order.Items) == len(itemsId) {
//					sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderClosedStatus, true)
//				} else {
//					sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, true)
//				}
//
//				sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, AutoReject, false, "Actions Expired", false, states.OrderClosedStatus)
//				if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//					returnChannel := make(chan future.IDataFuture, 1)
//					defer close(returnChannel)
//					returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//					return future.NewFuture(returnChannel, 1, 1)
//				}
//				return sellerApprovalPending.Childes()[1].ProcessOrder(ctx, order, itemsId, nil)
//
//			} else {
//				logger.Err("param not a message.RequestSellerOrderAction type , order: %v", order)
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//		}
//
//		if !sellerApprovalPending.validateAction(ctx, &order, itemsId) {
//			logger.Err("%s step received invalid action, order: %v, action: %s", sellerApprovalPending.Name(), order, req.Action)
//			returnChannel := make(chan future.IDataFuture, 1)
//			defer close(returnChannel)
//			returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.NotAccepted, Reason: "Actions Expired"}}
//			return future.NewFuture(returnChannel, 1, 1)
//		}
//
//		if req.Action == "success" {
//			sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, true)
//			sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, Approved, true, "", false, states.OrderInProgressStatus)
//			if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//			return sellerApprovalPending.Childes()[0].ProcessOrder(ctx, order, itemsId, nil)
//		} else if req.Action == "failed" {
//			if req.Data == nil {
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.BadRequest, Reason: "Reason Get Required"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			actionData := req.Data.(*message.RequestSellerOrderAction_Failed)
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
//				if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//				}
//				logger.Err("StockService future channel has been closed, order: %d", order.OrderId)
//			} else if futureData.Ex != nil {
//				if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//				}
//				logger.Err("released stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			if len(order.Items) == len(itemsId) {
//				sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderClosedStatus, true)
//			} else {
//				sellerApprovalPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, true)
//			}
//			sellerApprovalPending.updateOrderItemsProgress(ctx, &order, itemsId, Approved, false, actionData.Failed.Reason, false, states.OrderClosedStatus)
//			if err := sellerApprovalPending.persistOrder(ctx, &order); err != nil {
//				returnChannel := make(chan future.IDataFuture, 1)
//				defer close(returnChannel)
//				returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//				return future.NewFuture(returnChannel, 1, 1)
//			}
//
//			return sellerApprovalPending.Childes()[1].ProcessOrder(ctx, order, itemsId, nil)
//		}
//
//		logger.Err("%s step received invalid action, order: %v, action: %s", sellerApprovalPending.Name(), order, req.Action)
//		returnChannel := make(chan future.IDataFuture, 1)
//		defer close(returnChannel)
//		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		return future.NewFuture(returnChannel, 1, 1)
//	}
//}
//
//func (sellerApprovalPending approvalPendingState) validateAction(ctx context.Context, order *entities.Order, itemsId []uint64) bool {
//	if itemsId != nil && len(itemsId) > 0 {
//		for _, id := range itemsId {
//			for i := 0; i < len(order.Items); i++ {
//				length := len(order.Items[i].Progress.StepsHistory) - 1
//				if order.Items[i].ItemId == id && order.Items[i].Progress.StepsHistory[length].Name != sellerApprovalPending.Name() {
//					return false
//				}
//			}
//		}
//	} else {
//		for i := 0; i < len(order.Items); i++ {
//			length := len(order.Items[i].Progress.StepsHistory) - 1
//			if order.Items[i].Progress.StepsHistory[length].Name != sellerApprovalPending.Name() {
//				return false
//			}
//		}
//	}
//
//	return true
//}
//
//func (sellerApprovalPending approvalPendingState) persistOrder(ctx context.Context, order *entities.Order) error {
//	_, err := global.Singletons.OrderRepository.Save(*order)
//	if err != nil {
//		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", sellerApprovalPending.Name(), order, err.Error())
//	}
//
//	return err
//}
//
//func (sellerApprovalPending approvalPendingState) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, reason string, isSetExpireTime bool, itemStatus string) {
//
//	findFlag := false
//	if itemsId != nil && len(itemsId) > 0 {
//		for _, id := range itemsId {
//			findFlag = false
//			for i := 0; i < len(order.Items); i++ {
//				if order.Items[i].ItemId == id {
//					sellerApprovalPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
//					findFlag = true
//					break
//				}
//			}
//			if findFlag == false {
//				logger.Err("%s received itemId %d not exist in order, orderId: %d", sellerApprovalPending.Name(), id, order.OrderId)
//			}
//		}
//	} else {
//		for i := 0; i < len(order.Items); i++ {
//			sellerApprovalPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, reason, isSetExpireTime, itemStatus)
//		}
//	}
//}
//
//// TODO set time from redis config
//func (sellerApprovalPending approvalPendingState) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
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
//			time.Duration(24) +
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
////import "gitlab.faza.io/order-project/order-service"
////
////func ApprovalPendingApproved(ppr PaymentPendingRequest) error {
////	err := main.MoveOrderToNewState("seller", "", main.ShipmentPending, "shipment-pending", ppr)
////	if err != nil {
////		return err
////	}
////	return nil
////}
////
////// TODO: Improvement ApprovalPendingRejected
////func ApprovalPendingRejected(ppr PaymentPendingRequest, reason string) error {
////	err := main.MoveOrderToNewState("seller", reason, main.ShipmentRejectedBySeller, "shipment-rejected-by-seller", ppr)
////	if err != nil {
////		return err
////	}
////	newPpr, err := main.GetOrder(ppr.OrderNumber)
////	if err != nil {
////		return err
////	}
////	err = main.MoveOrderToNewState("system", reason, main.PayToBuyer, "pay-to-buyer", newPpr)
////	if err != nil {
////		return err
////	}
////	return nil
////}
