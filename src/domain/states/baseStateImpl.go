package states

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"strconv"
	"time"
)

type BaseStateImpl struct {
	name    string
	index   int
	childes []IState
	parents []IState
	configs map[string]interface{}
}

func NewBaseStep(name string, index int, childes, parents []IState) *BaseStateImpl {
	return &BaseStateImpl{name, index,
		childes, parents, make(map[string]interface{}, 4)}
}

func NewBaseStepWithConfig(name string, index int, childes, parents []IState, stateList []states_old.IState, configs map[string]interface{}) *BaseStateImpl {
	statesMap := make(map[int]states_old.IState, len(stateList))
	for i, v := range stateList {
		statesMap[int(i)] = v
	}
	return &BaseStateImpl{name, index, childes, parents,
		stateList, statesMap, configs}
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

func (base *BaseStateImpl) SetStates(statesList ...states_old.IState) {
	base.states = statesList
	statesMap := make(map[int]states_old.IState, len(statesList))
	for i, v := range statesList {
		statesMap[int(i)] = v
	}

	base.statesMap = statesMap
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

func (base BaseStateImpl) States() []states_old.IState {
	return base.states
}

func (base BaseStateImpl) StatesMap() map[int]states_old.IState {
	return base.statesMap
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

func (base BaseStateImpl) UpdateAllOrderStatus(ctx context.Context, order *entities.Order, itemsId []uint64, orderStatus string, isUpdateOnlyOrderStatus bool) {

	if isUpdateOnlyOrderStatus == true {
		order.UpdatedAt = time.Now().UTC()
		order.Status = orderStatus
	} else {
		order.UpdatedAt = time.Now().UTC()
		order.Status = orderStatus
		findFlag := true
		if itemsId != nil && len(itemsId) > 0 {
			for _, id := range itemsId {
				findFlag = false
				for i := 0; i < len(order.Items); i++ {
					if order.Items[i].ItemId == id {
						base.doUpdateOrderStep(ctx, order, i)
						findFlag = true
						break
					}
				}
				if !findFlag {
					logger.Err("%s received itemId %d not exist in order, orderId: %d", base.Name(), id, order.OrderId)
				}
			}
		} else {
			for i := 0; i < len(order.Items); i++ {
				base.doUpdateOrderStep(ctx, order, i)
			}
		}
	}
}

func (base BaseStateImpl) doUpdateOrderStep(ctx context.Context, order *entities.Order, index int) {
	order.Items[index].Progress.CreatedAt = time.Now().UTC()
	order.Items[index].Progress.CurrentStepName = base.Name()
	order.Items[index].Progress.CurrentStepIndex = base.Index()

	stepHistory := entities.StateHistory{
		Name:      base.Name(),
		Index:     base.Index(),
		CreatedAt: order.Items[index].Progress.CreatedAt,
		//ActionHistory: make([]entities.Action, 0, 1),
	}

	if order.Items[index].Progress.StepsHistory == nil || len(order.Items[index].Progress.StepsHistory) == 0 {
		order.Items[index].Progress.StepsHistory = make([]entities.StateHistory, 0, 5)
	}

	order.Items[index].Progress.StepsHistory = append(order.Items[index].Progress.StepsHistory, stepHistory)
}
