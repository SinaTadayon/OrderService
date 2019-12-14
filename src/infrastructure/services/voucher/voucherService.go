package voucher_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type IVoucherService interface {
	VoucherSettlement(ctx context.Context, voucherCode string,
		orderId uint64, buyerId uint64) future.IFuture
	Connect() error
	Disconnect() error
}
