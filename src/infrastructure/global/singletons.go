package global

import (
	"gitlab.faza.io/go-framework/kafkaadapter"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/item"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
)

type CtxKey int
const (
	CtxUserID 		CtxKey = iota
	CtxAuthToken
	CtxStepName
	CtxStepIndex
	CtxStepTimestamp
)

var Singletons struct {
	Kafka           *kafkaadapter.Kafka
	OrderRepository order_repository.IOrderRepository
	ItemRepository  item_repository.IItemRepository
	Converter       converter.IConverter
	StockService	stock.IStockService
	PaymentService 	payment.IPaymentService
}
