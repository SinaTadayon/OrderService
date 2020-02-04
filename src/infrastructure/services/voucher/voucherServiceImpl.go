package voucher_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	voucherProto "gitlab.faza.io/protos/cart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

type iVoucherServiceImpl struct {
	voucherClient  voucherProto.CouponServiceClient
	grpcConnection *grpc.ClientConn
	serverAddress  string
	serverPort     int
	timeout        int
}

func NewVoucherService(serverAddress string, serverPort int, timeout int) IVoucherService {
	return &iVoucherServiceImpl{
		serverAddress: serverAddress,
		serverPort:    serverPort,
		timeout:       timeout,
	}
}

func (voucherService *iVoucherServiceImpl) Connect() error {
	if voucherService.grpcConnection == nil || voucherService.grpcConnection.GetState() != connectivity.Ready {
		var err error
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		voucherService.grpcConnection, err = grpc.DialContext(ctx, voucherService.serverAddress+":"+fmt.Sprint(voucherService.serverPort),
			grpc.WithBlock(), grpc.WithInsecure())
		if err != nil {
			applog.GLog.Logger.Error("GRPC connect dial to voucher service failed",
				"fn", "connect",
				"address", voucherService.serverAddress,
				"port", voucherService.serverPort,
				"err", err.Error())
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
			applog.GLog.Logger.Error("voucherService Disconnect failed",
				"fn", "Disconnect",
				"error", err)
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
		applog.GLog.Logger.FromContext(ctx).Error("voucherClient.CouponUsed internal error",
			"fn", "VoucherSettlement",
			"voucherCode", voucherCode,
			"oid", orderId,
			"buyerId", buyerId,
			"error", err)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "voucherService.Connect() Failed")).
			BuildAndSend()
	}

	var outCtx context.Context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		outCtx = metadata.NewOutgoingContext(ctx, md)
	} else {
		outCtx = metadata.NewOutgoingContext(ctx, metadata.New(nil))
	}

	timeoutTimer := time.NewTimer(time.Duration(voucherService.timeout) * time.Second)

	couponReq := &voucherProto.CouponUseRequest{
		Code:  voucherCode,
		User:  strconv.Itoa(int(buyerId)),
		Order: strconv.Itoa(int(orderId)),
	}

	voucherFn := func() <-chan interface{} {
		voucherChan := make(chan interface{}, 0)
		go func() {
			result, err := voucherService.voucherClient.CouponUsed(outCtx, couponReq)
			if err != nil {
				voucherChan <- err
			} else {
				voucherChan <- result
			}
		}()
		return voucherChan
	}

	var obj interface{} = nil
	select {
	case obj = <-voucherFn():
		timeoutTimer.Stop()
		break
	case <-timeoutTimer.C:
		applog.GLog.Logger.FromContext(ctx).Error("voucherService.voucherClient.CouponUsed grpc timeout",
			"fn", "VoucherSettlement",
			"request", couponReq)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "Unknown Error", errors.New("CouponUsed Timeout")).
			BuildAndSend()
	}

	// TODO decode err code
	if err, ok := obj.(error); ok {
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("voucherClient.CouponUsed internal error",
				"fn", "VoucherSettlement",
				"voucherCode", voucherCode,
				"orderId", orderId,
				"buyerId", buyerId,
				"error", err)
			return future.Factory().SetCapacity(1).
				SetError(future.InternalError, "Unknown Error", errors.Wrap(err, "CouponUsed Failed")).
				BuildAndSend()
		}
	} else if result, ok := obj.(*voucherProto.Result); ok {
		if result.Code != 200 {
			applog.GLog.Logger.FromContext(ctx).Error("voucherClient.CouponUsed failed",
				"fn", "VoucherSettlement",
				"voucherCode", voucherCode,
				"oid", orderId,
				"buyerId", buyerId,
				"error", result)
			return future.Factory().SetCapacity(1).
				SetError(future.ErrorCode(result.Code), result.Message, errors.New("CouponUsed Failed")).
				BuildAndSend()
		}
	}

	applog.GLog.Logger.FromContext(ctx).Info("voucherClient.CouponUsed success",
		"fn", "VoucherSettlement",
		"voucherCode", voucherCode,
		"oid", orderId,
		"buyerId", buyerId)
	return future.Factory().SetCapacity(1).BuildAndSend()
}
