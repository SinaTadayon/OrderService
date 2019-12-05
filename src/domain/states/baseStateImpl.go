package states

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"strconv"
	"time"
)

type BaseStateImpl struct {
	name           string
	index          int
	childes        []IState
	parents        []IState
	actions        []actions.IAction
	actionStateMap map[actions.IAction]IState
	configs        map[string]interface{}
}

func NewBaseStep(name string, index int, childes, parents []IState, actionStateMap map[actions.IAction]IState) *BaseStateImpl {
	actionList := make([]actions.IAction, 0, len(actionStateMap))
	for key, _ := range actionStateMap {
		actionList = append(actionList, key)
	}

	return &BaseStateImpl{name, index,
		childes, parents, actionList, actionStateMap, nil}
}

func NewBaseStepWithConfig(name string, index int, childes, parents []IState,
	actionStateMap map[actions.IAction]IState, configs map[string]interface{}) *BaseStateImpl {
	actionList := make([]actions.IAction, 0, len(actionStateMap))
	for key, _ := range actionStateMap {
		actionList = append(actionList, key)
	}

	return &BaseStateImpl{name, index, childes, parents,
		actionList, actionStateMap, configs}
}

func (base *BaseStateImpl) SetName(name string) {
	base.name = name
}

func (base *BaseStateImpl) SetIndex(index int) {
	base.index = index
}

func (base *BaseStateImpl) SetChildes(iSteps []IState) {
	base.childes = iSteps
}

func (base *BaseStateImpl) SetParents(iSteps []IState) {
	base.parents = iSteps
}

func (base BaseStateImpl) Name() string {
	return base.String()
}

func (base BaseStateImpl) Index() int {
	return base.index
}

func (base BaseStateImpl) Childes() []IState {
	return base.childes
}

func (base BaseStateImpl) Parents() []IState {
	return base.parents
}

func (base BaseStateImpl) Actions() []actions.IAction {
	return base.actions
}

func (base BaseStateImpl) IsActionValid(action actions.IAction) bool {
	for key, _ := range base.actionStateMap {
		if key.ActionType() == action.ActionType() &&
			key.ActionEnum().ActionName() == action.ActionEnum().ActionName() {
			return true
		}
	}
	return false
}

func (base BaseStateImpl) StatesMap() map[actions.IAction]IState {
	return base.actionStateMap
}

func (base *BaseStateImpl) BaseState() *BaseStateImpl {
	return base
}

func (base *BaseStateImpl) GetConfigs() map[string]interface{} {
	return base.configs
}

func (base BaseStateImpl) String() string {
	return strconv.Itoa(base.index) + "." + base.name
}

func (base BaseStateImpl) SetOrderPkgStatus(ctx context.Context, order *entities.Order, status OrderStatus, pkgStatus PackageStatus) {
	order.UpdatedAt = time.Now().UTC()
	order.Status = string(status)
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].UpdatedAt = time.Now().UTC()
		order.Packages[i].Status = string(pkgStatus)
	}
}

func (base BaseStateImpl) SetOrderStatus(ctx context.Context, order *entities.Order, status OrderStatus) {
	order.UpdatedAt = time.Now().UTC()
	order.Status = string(status)
}

func (base BaseStateImpl) SetPkgStatus(ctx context.Context, packageItem *entities.PackageItem, status PackageStatus) {
	packageItem.UpdatedAt = time.Now().UTC()
	packageItem.Status = string(status)
}

func (base BaseStateImpl) UpdateOrderAllStatus(ctx context.Context, order *entities.Order,
	action *entities.Action, status OrderStatus, pkgStatus PackageStatus) {
	order.UpdatedAt = time.Now().UTC()
	order.Status = string(status)
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].UpdatedAt = time.Now().UTC()
		order.Packages[i].Status = string(pkgStatus)
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			base.UpdateSubPackage(ctx, &order.Packages[i].Subpackages[j], action)
		}
	}
}

func (base BaseStateImpl) UpdateOrderAllSubPkg(ctx context.Context, order *entities.Order, action *entities.Action) {
	order.UpdatedAt = time.Now().UTC()
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].UpdatedAt = time.Now().UTC()
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			base.UpdateSubPackage(ctx, &order.Packages[i].Subpackages[j], action)
		}
	}
}

func (base BaseStateImpl) UpdateSubPackage(ctx context.Context, subpackage *entities.Subpackage, action *entities.Action) {
	subpackage.UpdatedAt = time.Now().UTC()
	subpackage.Status = base.Name()
	subpackage.Tracking.StateName = base.Name()
	subpackage.Tracking.StateIndex = base.Index()
	subpackage.Tracking.Action = action

	if subpackage.Tracking.States == nil {
		subpackage.Tracking.States = make([]entities.State, 0, 3)
		state := entities.State{
			Name:      base.Name(),
			Index:     base.Index(),
			CreatedAt: time.Now().UTC(),
		}
		state.Actions = make([]entities.Action, 0, 8)
		state.Actions = append(state.Actions, *action)
		subpackage.Tracking.States = append(subpackage.Tracking.States, state)
	} else {
		state := subpackage.Tracking.States[len(subpackage.Tracking.States)-1]
		if state.Index != base.Index() {
			state := entities.State{
				Name:      base.Name(),
				Index:     base.Index(),
				CreatedAt: time.Now().UTC(),
			}
			state.Actions = make([]entities.Action, 0, 8)
			state.Actions = append(state.Actions, *action)
			subpackage.Tracking.States = append(subpackage.Tracking.States, state)
		} else {
			state.Actions = append(state.Actions, *action)
		}
	}
}

func (base BaseStateImpl) SaveOrUpdateOrder(ctx context.Context, order *entities.Order) error {
	var err error
	order, err = global.Singletons.OrderRepository.Save(ctx, *order)
	return errors.Wrap(err, "OrderRepository.Save failed")
}

//func (base BaseStateImpl) UpdatePackageStatus(ctx context.Context, packageItem *entities.PackageItem, status PackageStatus) {
//
//	if isUpdateOnlyOrderStatus == true {
//		order.UpdatedAt = time.Now().UTC()
//		order.Status = orderStatus
//	} else {
//		order.UpdatedAt = time.Now().UTC()
//		order.Status = orderStatus
//		findFlag := true
//		if itemsId != nil && len(itemsId) > 0 {
//			for _, id := range itemsId {
//				findFlag = false
//				for i := 0; i < len(order.Items); i++ {
//					if order.Items[i].ItemId == id {
//						base.doUpdateOrderStep(ctx, order, i)
//						findFlag = true
//						break
//					}
//				}
//				if !findFlag {
//					logger.Err("%s received itemId %d not exist in order, orderId: %d", base.Name(), id, order.OrderId)
//				}
//			}
//		} else {
//			for i := 0; i < len(order.Items); i++ {
//				base.doUpdateOrderStep(ctx, order, i)
//			}
//		}
//	}
//}
//
//func (base BaseStateImpl) doUpdateOrderStep(ctx context.Context, order *entities.Order, index int) {
//	order.Items[index].Progress.CreatedAt = time.Now().UTC()
//	order.Items[index].Progress.CurrentStepName = base.Name()
//	order.Items[index].Progress.CurrentStepIndex = base.Index()
//
//	stepHistory := entities.StateHistory{
//		Name:      base.Name(),
//		Index:     base.Index(),
//		CreatedAt: order.Items[index].Progress.CreatedAt,
//		//ActionHistory: make([]entities.Actions, 0, 1),
//	}
//
//	if order.Items[index].Progress.StepsHistory == nil || len(order.Items[index].Progress.StepsHistory) == 0 {
//		order.Items[index].Progress.StepsHistory = make([]entities.StateHistory, 0, 5)
//	}
//
//	order.Items[index].Progress.StepsHistory = append(order.Items[index].Progress.StepsHistory, stepHistory)
//}
