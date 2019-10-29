package global

import (
	"gitlab.faza.io/go-framework/kafkaadapter"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/item"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/order"
)

var Singletons struct {
	Kafka           *kafkaadapter.Kafka
	OrderRepository order.IOrderRepository
	ItemRepository  item.IItemRepository
	Converter       converter.IConverter
}
