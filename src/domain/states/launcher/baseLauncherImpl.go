package launcher_state

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"strconv"
)

type BaseLauncherImpl struct {
	name       string
	index      int
	childes    []states.IState
	parents    []states.IState
	actions    actions.IAction
	activeType actives.ActiveType
}

func NewBaseLauncher(name string, index int, childes, parents []states.IState,
	actions actions.IAction, launcherType actives.ActiveType) *BaseLauncherImpl {
	return &BaseLauncherImpl{name, index, childes, parents,
		actions, launcherType}
}

func (launcher *BaseLauncherImpl) SetName(name string) {
	launcher.name = name
}

func (launcher *BaseLauncherImpl) SetIndex(index int) {
	launcher.index = index
}

func (launcher *BaseLauncherImpl) SetChildes(states []states.IState) {
	launcher.childes = states
}

func (launcher *BaseLauncherImpl) SetParents(states []states.IState ) {
	launcher.parents = states
}

func (launcher *BaseLauncherImpl) SetActions(action actions.IAction) {
	launcher.actions = action
}

func (launcher *BaseLauncherImpl) SetLauncherType(activeType actives.ActiveType) {
	launcher.activeType = activeType
}

func (launcher BaseLauncherImpl) Name() string {
	return launcher.String()
}

func (launcher BaseLauncherImpl) Index() int {
	return launcher.index
}

func (launcher BaseLauncherImpl) Childes()	[]states.IState {
	return launcher.childes
}

func (launcher BaseLauncherImpl) Parents()	[]states.IState {
	return launcher.parents
}

func (launcher BaseLauncherImpl) Actions() actions.IAction {
	return launcher.actions
}

func (launcher BaseLauncherImpl) ActiveType() actives.ActiveType {
	return launcher.activeType
}

func (launcher BaseLauncherImpl) String() string {
	return strconv.Itoa(launcher.index) + "." + launcher.name
}
