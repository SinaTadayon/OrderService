package steps

import (
	"gitlab.faza.io/order-project/order-service/domain/states"
	"strconv"
)

type BaseStepImpl struct {
	name           	string
	index          	int
	childes        	[]IStep
	parents        	[]IStep
	states 			[]states.IState
	statesMap 		map[int]states.IState
	configs			map[string]interface{}
}

func NewBaseStep(name string, index int, childes, parents []IStep, stateList []states.IState) *BaseStepImpl {
	statesMap := make(map[int]states.IState, len(stateList))
	for i, v := range stateList {
		statesMap[int(i)] = v
	}
	return &BaseStepImpl{name, index, childes, parents,
		stateList, statesMap, make(map[string]interface{}, 4)}
}

func NewBaseStepWithConfig(name string, index int, childes, parents []IStep, stateList []states.IState, configs map[string]interface{}) *BaseStepImpl {
	statesMap := make(map[int]states.IState, len(stateList))
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

func (base *BaseStepImpl) SetStates(statesList ...states.IState) {
	base.states = statesList
	statesMap := make(map[int]states.IState, len(statesList))
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

func (base BaseStepImpl) Childes()	[]IStep {
	return base.childes
}

func (base BaseStepImpl) Parents()	[]IStep {
	return base.parents
}

func (base BaseStepImpl) States() []states.IState {
	return base.states
}

func (base BaseStepImpl) StatesMap() map[int]states.IState {
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