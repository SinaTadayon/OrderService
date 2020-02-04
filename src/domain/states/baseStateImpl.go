package states

import (
	"context"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
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
	return base.name
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

func (base BaseStateImpl) GetAction(action string) actions.IAction {
	for key, _ := range base.actionStateMap {
		if key.ActionEnum().ActionName() == action {
			return key
		}
	}
	return nil
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
	status OrderStatus, pkgStatus PackageStatus, actions ...*entities.Action) {
	order.UpdatedAt = time.Now().UTC()
	order.Status = string(status)
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].UpdatedAt = time.Now().UTC()
		order.Packages[i].Status = string(pkgStatus)
		for z := 0; z < len(actions); z++ {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				base.UpdateSubPackage(ctx, order.Packages[i].Subpackages[j], actions[z])
			}
		}
	}
}

func (base BaseStateImpl) UpdateOrderAllSubPkg(ctx context.Context, order *entities.Order, actions ...*entities.Action) {
	order.UpdatedAt = time.Now().UTC()
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].UpdatedAt = time.Now().UTC()
		if actions != nil && len(actions) > 0 {
			for z := 0; z < len(actions); z++ {
				for j := 0; j < len(order.Packages[i].Subpackages); j++ {
					base.UpdateSubPackage(ctx, order.Packages[i].Subpackages[j], actions[z])
				}
			}
		} else {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				base.UpdateSubPackage(ctx, order.Packages[i].Subpackages[j], nil)
			}
		}
	}
}

func (base BaseStateImpl) UpdatePackageAllSubPkg(ctx context.Context, packageItem *entities.PackageItem, actions ...*entities.Action) {
	for i := 0; i < len(packageItem.Subpackages); i++ {
		packageItem.UpdatedAt = time.Now().UTC()
		if actions != nil && len(actions) > 0 {
			for j := 0; j < len(actions); j++ {
				base.UpdateSubPackage(ctx, packageItem.Subpackages[i], actions[j])
			}
		} else {
			for i := 0; i < len(packageItem.Subpackages); i++ {
				base.UpdateSubPackage(ctx, packageItem.Subpackages[i], nil)
			}
		}
	}
}

func (base BaseStateImpl) UpdateSubPackage(ctx context.Context, subpackage *entities.Subpackage, action *entities.Action) {
	subpackage.UpdatedAt = time.Now().UTC()
	subpackage.Status = base.Name()
	subpackage.Tracking.Action = action
	if subpackage.Tracking.State == nil {
		state := entities.State{
			Name:       base.Name(),
			Index:      base.Index(),
			Schedulers: nil,
			Data:       nil,
			Actions:    nil,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
			Extended:   nil,
		}
		if action != nil {
			state.Actions = make([]entities.Action, 0, 8)
			state.Actions = append(state.Actions, *action)
		}

		if subpackage.Tracking.History == nil {
			subpackage.Tracking.History = make([]entities.State, 0, 3)
		}
		subpackage.Tracking.State = &state
		subpackage.Tracking.History = append(subpackage.Tracking.History, state)
	} else {
		if subpackage.Tracking.State.Index != base.Index() {
			newState := entities.State{
				Name:       base.Name(),
				Index:      base.Index(),
				Schedulers: nil,
				Data:       nil,
				Actions:    nil,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
				Extended:   nil,
			}
			if action != nil {
				newState.Actions = make([]entities.Action, 0, 8)
				newState.Actions = append(newState.Actions, *action)
			}
			if subpackage.Tracking.History == nil {
				subpackage.Tracking.History = make([]entities.State, 0, 3)
			}
			subpackage.Tracking.State = &newState
			subpackage.Tracking.History = append(subpackage.Tracking.History, newState)
		} else {
			subpackage.Tracking.State.UpdatedAt = time.Now().UTC()
			if action != nil {
				subpackage.Tracking.State.Actions = append(subpackage.Tracking.State.Actions, *action)
				subpackage.Tracking.Action = action
			}
			subpackage.Tracking.History[len(subpackage.Tracking.History)-1] = *subpackage.Tracking.State
		}
	}
}

func (base BaseStateImpl) UpdateSubPackageWithScheduler(ctx context.Context, subpackage *entities.Subpackage, schedulers []*entities.SchedulerData, action *entities.Action) {
	subpackage.UpdatedAt = time.Now().UTC()
	subpackage.Status = base.Name()
	subpackage.Tracking.Action = action
	if subpackage.Tracking.State == nil {
		state := entities.State{
			Name:       base.Name(),
			Index:      base.Index(),
			Schedulers: schedulers,
			Data:       nil,
			Actions:    nil,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
			Extended:   nil,
		}
		if action != nil {
			state.Actions = make([]entities.Action, 0, 8)
			state.Actions = append(state.Actions, *action)
		}

		if subpackage.Tracking.History == nil {
			subpackage.Tracking.History = make([]entities.State, 0, 3)
		}
		subpackage.Tracking.State = &state
		subpackage.Tracking.History = append(subpackage.Tracking.History, state)
	} else {
		if subpackage.Tracking.State.Index != base.Index() {
			newState := entities.State{
				Name:       base.Name(),
				Index:      base.Index(),
				Schedulers: schedulers,
				Data:       nil,
				Actions:    nil,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
				Extended:   nil,
			}
			if action != nil {
				newState.Actions = make([]entities.Action, 0, 8)
				newState.Actions = append(newState.Actions, *action)
			}
			if subpackage.Tracking.History == nil {
				subpackage.Tracking.History = make([]entities.State, 0, 3)
			}
			subpackage.Tracking.State = &newState
			subpackage.Tracking.History = append(subpackage.Tracking.History, newState)
		} else {
			subpackage.Tracking.State.UpdatedAt = time.Now().UTC()
			if action != nil {
				subpackage.Tracking.State.Actions = append(subpackage.Tracking.State.Actions, *action)
				subpackage.Tracking.Action = action
			}
			subpackage.Tracking.History[len(subpackage.Tracking.History)-1] = *subpackage.Tracking.State
		}
	}
}

func (base BaseStateImpl) UpdateSubPackageWithData(ctx context.Context, subpackage *entities.Subpackage, data map[string]interface{}, action *entities.Action) {
	subpackage.UpdatedAt = time.Now().UTC()
	subpackage.Status = base.Name()
	subpackage.Tracking.Action = action
	if subpackage.Tracking.State == nil {
		state := entities.State{
			Name:       base.Name(),
			Index:      base.Index(),
			Schedulers: nil,
			Data:       data,
			Actions:    nil,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
			Extended:   nil,
		}
		if action != nil {
			state.Actions = make([]entities.Action, 0, 8)
			state.Actions = append(state.Actions, *action)
		}

		if subpackage.Tracking.History == nil {
			subpackage.Tracking.History = make([]entities.State, 0, 3)
		}
		subpackage.Tracking.State = &state
		subpackage.Tracking.History = append(subpackage.Tracking.History, state)
	} else {
		if subpackage.Tracking.State.Index != base.Index() {
			newState := entities.State{
				Name:       base.Name(),
				Index:      base.Index(),
				Schedulers: nil,
				Data:       data,
				Actions:    nil,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
				Extended:   nil,
			}
			if action != nil {
				newState.Actions = make([]entities.Action, 0, 8)
				newState.Actions = append(newState.Actions, *action)
			}
			if subpackage.Tracking.History == nil {
				subpackage.Tracking.History = make([]entities.State, 0, 3)
			}
			subpackage.Tracking.State = &newState
			subpackage.Tracking.History = append(subpackage.Tracking.History, newState)
		} else {
			subpackage.Tracking.State.UpdatedAt = time.Now().UTC()
			if action != nil {
				subpackage.Tracking.State.Actions = append(subpackage.Tracking.State.Actions, *action)
				subpackage.Tracking.Action = action
			}
			subpackage.Tracking.History[len(subpackage.Tracking.History)-1] = *subpackage.Tracking.State
		}
	}
}

func (base BaseStateImpl) SaveOrUpdateOrder(ctx context.Context, order *entities.Order) error {
	var err error
	order, err = app.Globals.OrderRepository.Save(ctx, *order)
	return err
}
