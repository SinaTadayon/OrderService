package notify_service

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"os"
	"testing"
)

var config *configs.Config
var notify iNotificationServiceImpl

func TestMain(m *testing.M) {
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
		os.Exit(1)
	}

	notify = iNotificationServiceImpl{nil, nil,
		config.NotifyService.Address, config.NotifyService.Port}

	// Running Tests
	code := m.Run()
	os.Exit(code)
}

func TestNotifySMS(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	request := SMSRequest{
		//Phone: "09373969041",
		Phone: "09128085965",
		Body:  "سلام، این اس ام اس تستی هست",
	}
	iFuture := notify.NotifyBySMS(ctx, request)
	futureData := iFuture.Get()

	require.Nil(t, futureData.Data())
	require.Nil(t, futureData.Error())
}

func TestNotifyEmail(t *testing.T) {

}
