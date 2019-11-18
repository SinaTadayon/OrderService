package global

import (
	"gitlab.faza.io/go-framework/kafkaadapter"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/item"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	notify_service "gitlab.faza.io/order-project/order-service/infrastructure/services/notification"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucher_service "gitlab.faza.io/order-project/order-service/infrastructure/services/voucher"
)

type CtxKey int

const (
	CtxUserID CtxKey = iota
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
	StockService    stock_service.IStockService
	PaymentService  payment_service.IPaymentService
	NotifyService   notify_service.INotificationService
	UserService     user_service.IUserService
	VoucherService  voucher_service.IVoucherService
	//SchedulerService	scheduler_service.ISchedulerService
	//FlowManager		domain.IFlowManager
	//GRPCServer      grpc.Server
}
