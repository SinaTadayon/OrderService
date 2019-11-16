package active_event

import (
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"time"
)

type iActiveEventImpl struct {
	*events.BaseEventImpl
	order      entities.Order
	itemsId    []uint64
	activeType actives.ActiveType
	action     actives.IActiveAction
}

func NewActiveEvent(order entities.Order, itemsId []uint64, activeType actives.ActiveType,
	action actives.IActiveAction, data interface{}, timestamp time.Time) IActiveEvent {
	return &iActiveEventImpl{events.NewBaseEventImpl(events.ActiveEvent, data, timestamp),
		order, itemsId, activeType, action}
}

func (activeEvent iActiveEventImpl) Order() entities.Order {
	return activeEvent.order
}

func (activeEvent iActiveEventImpl) ItemsId() []uint64 {
	return activeEvent.itemsId
}

func (activeEvent iActiveEventImpl) ActiveType() actives.ActiveType {
	return activeEvent.activeType
}

func (activeEvent iActiveEventImpl) ActiveAction() actives.IActiveAction {
	return activeEvent.action
}
