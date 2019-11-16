package active_event

import (
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

type IActiveEvent interface {
	events.IEvent
	Order() entities.Order
	ItemsId() []uint64
	ActiveType() actives.ActiveType
	ActiveAction() actives.IActiveAction
}
