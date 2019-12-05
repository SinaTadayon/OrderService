package voucher_service

import (
	"context"
	"fmt"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	voucherProto "gitlab.faza.io/protos/cart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"strconv"
	"time"
)

type iVoucherServiceImpl struct {
	voucherClient  voucherProto.CouponServiceClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
}

func NewVoucherService(serverAddress string, serverPort int) IVoucherService {
	return &iVoucherServiceImpl{
		serverAddress: serverAddress,
		serverPort:    serverPort,
	}
}

func (voucherService *iVoucherServiceImpl) Connect() error {
	if voucherService.grpcConnection == nil || voucherService.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		voucherService.grpcConnection, err = grpc.DialContext(ctx, voucherService.serverAddress+":"+fmt.Sprint(voucherService.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			logger.Err("GRPC connect dial to voucher service failed, address: %s, port: %d, err: %s",
				voucherService.serverAddress, voucherService.serverPort, err.Error())
			return err
		}
		voucherService.voucherClient = voucherProto.NewCouponServiceClient(voucherService.grpcConnection)
	}
	return nil
}

func (voucherService *iVoucherServiceImpl) Disconnect() error {
	if voucherService.grpcConnection != nil {
		err := voucherService.grpcConnection.Close()
		if err != nil {
			logger.Err("voucherService Disconnect() failed, error: %s ", err)
			return nil
		}
	}

	voucherService.grpcConnection = nil
	voucherService.voucherClient = nil
	return nil
}

func (voucherService iVoucherServiceImpl) VoucherSettlement(ctx context.Context,
	voucherCode string, orderId uint64, buyerId uint64) future.IFuture {
	if err := voucherService.Connect(); err != nil {
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{
			Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}

	couponReq := &voucherProto.CouponUseRequest{
		Code:  voucherCode,
		User:  strconv.Itoa(int(buyerId)),
		Order: strconv.Itoa(int(orderId)),
	}

	result, err := voucherService.voucherClient.CouponUsed(ctx, couponReq)
	if err != nil {
		logger.Err("VoucherSettlement() => voucherClient.CouponUsed internal error, "+
			"voucherCode: %s, orderId: %d, buyerId: %d, error: %s", voucherCode, orderId, buyerId, result)
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{
			Code: future.InternalError, Reason: "Unknown Error"}}
		return future.NewFuture(returnChannel, 1, 1)
	}

	if result.Code != 200 {
		logger.Err("VoucherSettlement() => voucherClient.CouponUsed failed, "+
			"voucherCode: %s, orderId: %d, buyerId: %d, error: %s", voucherCode, orderId, buyerId, result)
		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{
			Code: result.Code, Reason: result.Message}}
		return future.NewFuture(returnChannel, 1, 1)
	}

	logger.Audit("VoucherSettlement() => voucherClient.CouponUsed success, "+
		"voucherCode: %s, orderId: %d, buyerId: %d, error: %s", voucherCode, orderId, buyerId, result)
	returnChannel := make(chan future.IDataFuture, 1)
	defer close(returnChannel)
	returnChannel <- future.IDataFuture{Data: nil, Ex: nil}
	return future.NewFuture(returnChannel, 1, 1)
}
