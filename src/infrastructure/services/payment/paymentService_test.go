package payment_service

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"os"
	"testing"
	"time"
)

var config *configs.Config
var payment iPaymentServiceImpl

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

	payment = iPaymentServiceImpl{nil, nil,
		config.PaymentGatewayService.Address,
		config.PaymentGatewayService.Port, config.PaymentGatewayService.CallbackTimeout, config.PaymentGatewayService.PaymentResultTimeout}

	// Running Tests
	code := m.Run()
	os.Exit(code)
}

func TestGetQueryOrder(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	err := payment.ConnectToPaymentService()
	require.Nil(t, err)

	defer func() {
		if err := payment.grpcConnection.Close(); err != nil {
		}
	}()

	iFuture := payment.GetPaymentResult(ctx, 123456789)
	futureData := iFuture.Get()
	require.NotNil(t, futureData)
	require.Error(t, futureData.Error().Reason())
}

func TestOrderPayment_Success(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	err := payment.ConnectToPaymentService()
	require.Nil(t, err)

	defer func() {
		if err := payment.grpcConnection.Close(); err != nil {
		}
	}()

	request := PaymentRequest{
		Gateway:  "asanpardakht",
		Amount:   1200000,
		Currency: "IRR",
		OrderId:  123456789,
	}

	iFuture := payment.OrderPayment(ctx, request)
	futureData := iFuture.Get()
	require.NotNil(t, futureData)
	//require.Nil(t, futureData.Error())
	//
	//paymentResponse := futureData.Data().(PaymentResponse)
	//require.NotEmpty(t, paymentResponse.CallbackUrl)
}
