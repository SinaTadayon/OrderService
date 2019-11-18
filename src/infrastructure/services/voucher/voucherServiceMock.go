package voucher_service

import (
	"context"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
)

type iVoucherServiceMock struct {
}

func NewVoucherServiceMock() IVoucherService {
	return &iVoucherServiceMock{}
}

func (voucherService iVoucherServiceMock) VoucherSettlement(ctx context.Context, voucherCode string,
	orderId uint64, buyerId uint64) promise.IPromise {
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data: nil, Ex: nil}
	return promise.NewPromise(returnChannel, 1, 1)
}

func (voucherService iVoucherServiceMock) Connect() error {
	return nil
}

func (voucherService iVoucherServiceMock) Disconnect() error {
	return nil
}
