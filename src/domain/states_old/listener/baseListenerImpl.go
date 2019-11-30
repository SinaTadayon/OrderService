package listener_state

import (
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"strconv"
)

type BaseListenerImpl struct {
	name      string
	index     int
	childes   []states_old.IState
	parents   []states_old.IState
	action    actions.IAction
	actorType actors.ActorType
}

func NewBaseListener(name string, index int, childes, parents []states_old.IState,
	action actions.IAction, actorType actors.ActorType) *BaseListenerImpl {
	return &BaseListenerImpl{name, index, childes, parents,
		action, actorType}
}

func (listener *BaseListenerImpl) SetName(name string) {
	listener.name = name
}

func (listener *BaseListenerImpl) SetIndex(index int) {
	listener.index = index
}

func (listener *BaseListenerImpl) SetChildes(childes []states_old.IState) {
	listener.childes = childes
}

func (listener *BaseListenerImpl) SetParents(parents []states_old.IState) {
	listener.parents = parents
}

func (listener *BaseListenerImpl) SetAction(action actions.IAction) {
	listener.action = action
}

func (listener *BaseListenerImpl) Name() string {
	return listener.String()
}

func (listener BaseListenerImpl) Index() int {
	return listener.index
}

func (listener BaseListenerImpl) Childes() []states_old.IState {
	return listener.childes
}

func (listener BaseListenerImpl) Parents() []states_old.IState {
	return listener.parents
}

func (listener BaseListenerImpl) Actions() actions.IAction {
	return listener.action
}

func (listener BaseListenerImpl) ActorType() actors.ActorType {
	return listener.actorType
}

func (listener BaseListenerImpl) String() string {
	return strconv.Itoa(listener.index) + "." + listener.name
}
