package state_35

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
	"time"
)

const (
	stepName  string = "Delivery_Delayed"
	stepIndex int    = 35
)

type DeliveryDelayedState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &DeliveryDelayedState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &DeliveryDelayedState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &DeliveryDelayedState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state DeliveryDelayedState) Process(ctx context.Context, iFrame frame.IFrame) {
	if iFrame.Header().KeyExists(string(frame.HeaderSubpackages)) {
		subpackages, ok := iFrame.Header().Value(string(frame.HeaderSubpackages)).([]*entities.Subpackage)
		if !ok {
			logger.Err("iFrame.Header() not a subpackages, frame: %v, %s state ", iFrame, state.Name())
			return
		}

		if iFrame.Body().Content() == nil {
			logger.Err("Process() => iFrame.Body().Content() is nil, orderId: %d, pid: %d, sid: %d, %s state ",
				subpackages[0].OrderId, subpackages[0].PId, subpackages[0].SId, state.Name())
			return
		}

		//pkgItem, ok := iFrame.Body().Content().(*entities.PackageItem)
		//if !ok {
		//	logger.Err("Process() => iFrame.Body().Content() is nil, orderId: %d, pid: %d, sid: %d, %s state ",
		//		subpackages[0].OrderId, subpackages[0].PId, subpackages[0].SId, state.Name())
		//	return
		//}

		for _, subpackage := range subpackages {
			state.UpdateSubPackage(ctx, subpackage, nil)
			_, err := app.Globals.SubPkgRepository.Update(ctx, *subpackage)
			if err != nil {
				logger.Err("Process() => SubPkgRepository.Update in %s state failed, orderId: %d, pid: %d, sid: %d, error: %s",
					state.Name(), subpackage.OrderId, subpackage.PId, subpackage.SId, err.Error())
			} else {
				logger.Audit("Process() => Status of subpackages update to %s state, orderId: %d, pid: %d, sid: %d",
					state.Name(), subpackage.OrderId, subpackage.PId, subpackage.SId)
			}
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

			var newSubPackages []*entities.Subpackage
			var requestAction *entities.Action
			var newSubPkg *entities.Subpackage
			var fullItems []entities.Item
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

			// iterate subpackages
			for _, eventSubPkg := range actionData.SubPackages {
				for i := 0; i < len(pkgItem.Subpackages); i++ {
					if eventSubPkg.SId == pkgItem.Subpackages[i].SId && pkgItem.Subpackages[i].Status == state.Name() {
						newSubPkg = nil
						fullItems = nil
						var findItem = false

						// iterate items
						for _, actionItem := range eventSubPkg.Items {
							findItem = false
							for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
								if actionItem.InventoryId == pkgItem.Subpackages[i].Items[j].InventoryId {
									findItem = true

									// create new subpackages which contains new items along
									// with new quantity and recalculated related invoice
									if actionItem.Quantity < pkgItem.Subpackages[i].Items[j].Quantity {
										if newSubPkg == nil {
											newSubPkg = pkgItem.Subpackages[i].DeepCopy()
											newSubPkg.SId = 0
											newSubPkg.Items = make([]entities.Item, 0, len(eventSubPkg.Items))

											requestAction = &entities.Action{
												Name:       actionState.ActionEnum().ActionName(),
												Type:       "",
												UId:        ctx.Value(string(utils.CtxUserID)).(uint64),
												UTP:        actionState.ActionType().ActionName(),
												Permission: "",
												Privilege:  "",
												Policy:     "",
												Result:     string(states.ActionSuccess),
												Reasons:    actionItem.Reasons,
												Data:       nil,
												CreatedAt:  time.Now().UTC(),
												Extended:   nil,
											}
										}

										pkgItem.Subpackages[i].Items[j].Quantity -= actionItem.Quantity
										pkgItem.Subpackages[i].Items[j].Invoice.Total = pkgItem.Subpackages[i].Items[j].Invoice.Unit *
											uint64(pkgItem.Subpackages[i].Items[j].Quantity)

										// create new item from requested action item
										newItem := pkgItem.Subpackages[i].Items[j].DeepCopy()
										newItem.Quantity = actionItem.Quantity
										newItem.Reasons = actionItem.Reasons
										newItem.Invoice.Total = newItem.Invoice.Unit * uint64(newItem.Quantity)
										newSubPkg.Items = append(newSubPkg.Items, *newItem)

									} else if actionItem.Quantity > pkgItem.Subpackages[i].Items[j].Quantity {
										logger.Err("Process() => received action not acceptable, Requested quantity greater than item quantity, state: %s, event: %v", state.String(), event)
										future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
											SetError(future.NotAccepted, "Requested quantity greater than item quantity", errors.New("Action Not Accepted")).Send()
										return

									} else {
										if fullItems == nil {
											fullItems = make([]entities.Item, 0, len(pkgItem.Subpackages[i].Items))
											requestAction = &entities.Action{
												Name:       actionState.ActionEnum().ActionName(),
												Type:       "",
												UId:        ctx.Value(string(utils.CtxUserID)).(uint64),
												UTP:        actionState.ActionType().ActionName(),
												Permission: "",
												Privilege:  "",
												Policy:     "",
												Result:     string(states.ActionSuccess),
												Reasons:    actionItem.Reasons,
												Data:       nil,
												CreatedAt:  time.Now().UTC(),
												Extended:   nil,
											}
										}
										fullItems = append(fullItems, pkgItem.Subpackages[i].Items[j])
										pkgItem.Subpackages[i].Items[len(pkgItem.Subpackages[i].Items)-1], pkgItem.Subpackages[i].Items[j] =
											pkgItem.Subpackages[i].Items[j], pkgItem.Subpackages[i].Items[len(pkgItem.Subpackages[i].Items)-1]
										pkgItem.Subpackages[i].Items = pkgItem.Subpackages[i].Items[:len(pkgItem.Subpackages[i].Items)-1]
									}
								}
							}
							if !findItem {
								logger.Err("Process() => received action item inventory not found, Requested action item inventory not found in requested subpackage, inventoryId: %s, state: %s, event: %v", actionItem.InventoryId, state.String(), event)
								future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
									SetError(future.NotFound, "Request action item not found", errors.New("Action Item Not Found")).Send()
								return
							}
						}

						newSubPackages = make([]*entities.Subpackage, 0, len(actionData.SubPackages))
						if newSubPkg != nil {
							if fullItems != nil {
								for z := 0; z < len(fullItems); z++ {
									newSubPkg.Items = append(newSubPkg.Items, fullItems[z])
								}
							}
							newSubPackages = append(newSubPackages, newSubPkg)
						} else {
							for z := 0; z < len(fullItems); z++ {
								pkgItem.Subpackages[i].Items = append(pkgItem.Subpackages[i].Items, fullItems[z])
							}
							newSubPackages = append(newSubPackages, &pkgItem.Subpackages[i])
						}
					}
				}
			}

			if newSubPackages != nil {
				var sids = make([]uint64, 0, 32)
				for i := 0; i < len(newSubPackages); i++ {
					if newSubPackages[i].SId == 0 {
						// TODO must be optimized performance
						state.UpdateSubPackage(ctx, newSubPackages[i], requestAction)
						err := app.Globals.SubPkgRepository.Save(ctx, newSubPackages[i])
						if err != nil {
							logger.Err("Process() => SubPkgRepository.Save in %s state failed, orderId: %d, sellerId: %d, event: %v, error: %s", state.Name(),
								newSubPackages[i].OrderId, newSubPackages[i].PId, event, err.Error())
							// TODO must distinct system error from update version error
							future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
								SetError(future.InternalError, "Unknown Err", err).Send()
							return
						}

						pkgItem.Subpackages = append(pkgItem.Subpackages, *newSubPackages[i])
						logger.Audit("Process() => Status of new subpackage update to %v event, orderId: %d, sellerId: %d, sid: %d",
							event, newSubPackages[i].OrderId, newSubPackages[i].PId, newSubPackages[i].SId)
					} else {
						state.UpdateSubPackage(ctx, newSubPackages[i], requestAction)
					}
					sids = append(sids, newSubPackages[i].SId)
				}

				if event.Action().ActionEnum() == operator_action.DeliveryFail {
					var rejectedSubtotal uint64 = 0
					var rejectedDiscount uint64 = 0

					for _, subpackage := range newSubPackages {
						for j := 0; j < len(subpackage.Items); j++ {
							rejectedSubtotal += subpackage.Items[j].Invoice.Total
							rejectedDiscount += subpackage.Items[j].Invoice.Discount
						}
					}

					if rejectedSubtotal < pkgItem.Invoice.Subtotal && rejectedDiscount < pkgItem.Invoice.Discount {
						pkgItem.Invoice.Subtotal -= rejectedSubtotal
						pkgItem.Invoice.Discount -= rejectedDiscount
						logger.Audit("Process() => calculate package invoice success, orderId: %d, pid:%d, action: %s, subtotal: %d, discount: %d",
							pkgItem.OrderId, pkgItem.PId, event.Action().ActionEnum().ActionName(), pkgItem.Invoice.Subtotal, pkgItem.Invoice.Discount)

					} else if rejectedSubtotal > pkgItem.Invoice.Subtotal || rejectedDiscount > pkgItem.Invoice.Discount {
						logger.Err("Process() => calculate package invoice failed, orderId: %d, pid:%d, action: %s, subtotal: %d, discount: %d",
							pkgItem.OrderId, pkgItem.PId, event.Action().ActionEnum().ActionName(), pkgItem.Invoice.Subtotal, pkgItem.Invoice.Discount)
						future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
							SetError(future.InternalError, "Unknown Error", errors.New("Package Invoice Invalid")).Send()
						return
					}
				}

				pkgItemUpdated, err := app.Globals.PkgItemRepository.Update(ctx, *pkgItem)
				if err != nil {
					logger.Err("Process() => PkgItemRepository.Update in %s state failed, orderId: %d, sellerId: %d, event: %v, error: %s", state.Name(),
						pkgItem.OrderId, pkgItem.PId, event, err.Error())
					// TODO must distinct system error from update version error
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.InternalError, "Unknown Err", err).Send()
					return
				}
				pkgItem = pkgItemUpdated

				response := events.ActionResponse{
					OrderId: pkgItem.OrderId,
					SIds:    sids,
				}

				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetData(response).Send()
				nextActionState.Process(ctx, frame.Factory().SetSIds(sids).SetSubpackages(newSubPackages).SetBody(pkgItem).Build())
			} else {
				logger.Err("Process() => event action data invalid, state: %s, event: %v, frame: %v", state.String(), event, iFrame)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.BadRequest, "Event Action Data Invalid", errors.New("event action data invalid")).Send()
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
