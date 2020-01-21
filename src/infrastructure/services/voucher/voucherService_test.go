package voucher_service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	user_service "gitlab.faza.io/order-project/order-service/infrastructure/services/user"
	voucherProto "gitlab.faza.io/protos/cart"
	"google.golang.org/grpc/metadata"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

var config *configs.Config
var voucherSrv iVoucherServiceImpl
var userService user_service.IUserService

func createAuthenticatedContext() (context.Context, error) {
	ctx, _ := context.WithCancel(context.Background())
	futureData := userService.UserLogin(ctx, "989200000000", "123456").Get()

	if futureData.Error() != nil {
		return nil, futureData.Error().Reason()
	}

	loginTokens, ok := futureData.Data().(user_service.LoginTokens)
	if ok != true {
		return nil, errors.New("data does not LoginTokens type")
	}

	var authorization = map[string]string{
		"authorization": fmt.Sprintf("Bearer %v", loginTokens.AccessToken),
		"userId":        "1000002",
	}
	md := metadata.New(authorization)
	ctxToken := metadata.NewOutgoingContext(ctx, md)

	return ctxToken, nil
}

func TestMain(m *testing.M) {
	var err error
	var path string
	if os.Getenv("APP_MODE") == "dev" {
		path = "../../../testdata/.env"
	} else {
		path = ""
	}

	config, _, err = configs.LoadConfigs(path, "")
	if err != nil {
		logger.Err(err.Error())
		os.Exit(1)
	}

	userService = user_service.NewUserService(config.UserService.Address, config.UserService.Port, config.UserService.Timeout)

	voucherSrv = iVoucherServiceImpl{
		voucherClient:  nil,
		grpcConnection: nil,
		serverAddress:  config.VoucherService.Address,
		serverPort:     config.VoucherService.Port,
		timeout:        config.VoucherService.Timeout,
	}

	// Running Tests
	code := m.Run()
	os.Exit(code)
}

func TestVoucherSettlement(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	testName1 := "test-" + strconv.Itoa(rand.Int())

	cT := &voucherProto.CouponTemplate{
		Title:           testName1,
		Prefix:          testName1,
		UseLimit:        1,
		Count:           1,
		Length:          5,
		StartDate:       time.Date(2019, 07, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		EndDate:         time.Date(2019, 07, 25, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Categories:      nil,
		Products:        nil,
		Sellers:         nil,
		IsFirstPurchase: true,
		CouponDiscount: &voucherProto.CouponDiscount{
			Type:             "fixed",
			Amount:           100000,
			MaxDiscountValue: 0,
			MinBasketValue:   1500000,
		},
	}

	err := voucherSrv.Connect()
	require.Nil(t, err)

	defer voucherSrv.Disconnect()

	ctx, err := createAuthenticatedContext()
	require.Nil(t, err)

	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	result, err := voucherSrv.voucherClient.CreateCouponTemplate(ctx, cT)
	require.Nil(t, err)
	require.NotNil(t, result)
	require.Equal(t, 200, int(result.Code))

	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	voucherRequest := &voucherProto.GetVoucherByTemplateNameRequest{
		Page:        1,
		Perpage:     1,
		VoucherName: testName1,
	}
	allVouchers, err := voucherSrv.voucherClient.GetVoucherByTemplateName(ctx, voucherRequest)
	require.Nil(t, err)
	require.NotNil(t, allVouchers.Vouchers[0])
	require.NotEmpty(t, allVouchers.Vouchers[0].Code)

	ctx, _ = context.WithCancel(context.Background())
	iFuture := voucherSrv.VoucherSettlement(ctx, allVouchers.Vouchers[0].Code, 123456789776, 1000001)
	futureData := iFuture.Get()
	require.Nil(t, futureData.Data())
	require.Nil(t, futureData.Error())
}
