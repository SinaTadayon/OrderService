package payment_service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"os"
	"testing"
)

var config *configs.Config
var payment iPaymentServiceImpl

func init() {
	var err error
	var path string
	if os.Getenv("APP_ENV") == "dev" {
		path = "../../../testdata/.env"
	} else {
		path = ""
	}

	config, err = configs.LoadConfig(path)
	if err != nil {
		logger.Err(err.Error())
		panic("configs.LoadConfig failed")
	}

	payment = iPaymentServiceImpl{nil, nil,
		config.PaymentGatewayService.Address, config.PaymentGatewayService.Port}
}

func TestOrderPayment_Success(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	if err := payment.ConnectToPaymentService(); err != nil {
		logger.Err(err.Error())
		panic("stockService.ConnectToPaymentService() failed")
	}

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
	assert.NotNil(t, futureData)
	assert.Nil(t, futureData.Error())

	paymentResponse := futureData.Data().(PaymentResponse)
	assert.NotEmpty(t, paymentResponse.CallbackUrl)
}
