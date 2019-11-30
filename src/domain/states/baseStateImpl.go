package states

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"strconv"
	"time"
)

//422 - Validation Errors, an array of objects, each object containing the field and the value (message) of the error
//400 - Bad Request - Any request not properly formatted for the server to understand and parse it
//403 - Forbidden - This can be used for any authentication errors, a user not being logged in etc.
//404 - Any requested entity which is not being found on the server
//406 - Not Accepted - The example usage for this code, is an attempt on an expired or timed-out action. Such as trying to cancel an order which cannot be cancelled any more
//409 - Conflict - Anything which causes conflicts on the server, the most famous one, a not unique email error, a duplicate entity...

const (
	BadRequest      = 400
	ForBidden       = 403
	NotFound        = 404
	NotAccepted     = 406
	Conflict        = 409
	ValidationError = 422
)

type BaseStepImpl struct {
	name      string
	index     int
	childes   []IStep
	parents   []IStep
	states    []states_old.IState
	statesMap map[int]states_old.IState
	configs   map[string]interface{}
}

func NewBaseStep(name string, index int, childes, parents []IStep, stateList []states_old.IState) *BaseStepImpl {
	statesMap := make(map[int]states_old.IState, len(stateList))
	for i, v := range stateList {
		statesMap[int(i)] = v
	}
	return &BaseStepImpl{name, index,
		childes, parents, stateList,
		statesMap, make(map[string]interface{}, 4)}
}

func NewBaseStepWithConfig(name string, index int, childes, parents []IStep, stateList []states_old.IState, configs map[string]interface{}) *BaseStepImpl {
	statesMap := make(map[int]states_old.IState, len(stateList))
	for i, v := range stateList {
		statesMap[int(i)] = v
	}
	return &BaseStepImpl{name, index, childes, parents,
		stateList, statesMap, configs}
}

func (base *BaseStepImpl) SetName(name string) {
	base.name = name
}

func (base *BaseStepImpl) SetIndex(index int) {
	base.index = index
}

func (base *BaseStepImpl) SetChildes(iSteps []IStep) {
	base.childes = iSteps
}

func (base *BaseStepImpl) SetParents(iSteps []IStep) {
	base.parents = iSteps
}

func (base *BaseStepImpl) SetStates(statesList ...states_old.IState) {
	base.states = statesList
	statesMap := make(map[int]states_old.IState, len(statesList))
	for i, v := range statesList {
		statesMap[int(i)] = v
	}

	base.statesMap = statesMap
}

func (base BaseStepImpl) Name() string {
	return base.String()
}

func (base BaseStepImpl) Index() int {
	return base.index
}

func (base BaseStepImpl) Childes() []IStep {
	return base.childes
}

func (base BaseStepImpl) Parents() []IStep {
	return base.parents
}

func (base BaseStepImpl) States() []states_old.IState {
	return base.states
}

func (base BaseStepImpl) StatesMap() map[int]states_old.IState {
	return base.statesMap
}

func (base *BaseStepImpl) BaseStep() *BaseStepImpl {
	return base
}

func (base *BaseStepImpl) GetConfigs() map[string]interface{} {
	return base.configs
}

func (base BaseStepImpl) String() string {
	return strconv.Itoa(base.index) + "." + base.name
}

func (base BaseStepImpl) UpdateAllOrderStatus(ctx context.Context, order *entities.Order, itemsId []uint64, orderStatus string, isUpdateOnlyOrderStatus bool) {

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

func (base BaseStepImpl) doUpdateOrderStep(ctx context.Context, order *entities.Order, index int) {
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
