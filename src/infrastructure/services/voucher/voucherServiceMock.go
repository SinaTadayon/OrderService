package voucher_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type iVoucherServiceMock struct {
}

func NewVoucherServiceMock() IVoucherService {
	return &iVoucherServiceMock{}
}

func (voucherService iVoucherServiceMock) VoucherSettlement(ctx context.Context, voucherCode string,
	orderId uint64, buyerId uint64) future.IFuture {
	returnChannel := make(chan future.IDataFuture, 1)
	defer close(returnChannel)
	returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
	return future.NewFuture(returnChannel, 1, 1)
}

func (voucherService iVoucherServiceMock) Connect() error {
	return nil
}

func (voucherService iVoucherServiceMock) Disconnect() error {
	return nil
}
