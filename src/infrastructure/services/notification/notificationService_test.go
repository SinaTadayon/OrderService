package notify_service

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"os"
	"sync"
	"testing"
)

var config *configs.Config
var notify iNotificationServiceImpl

func TestMain(m *testing.M) {
	var err error
	var path string
	if os.Getenv("APP_MODE") == "dev" {
		path = "../../../testdata/.env"
	} else {
		path = ""
	}

	applog.GLog.ZapLogger = applog.InitZap()
	applog.GLog.Logger = logger.NewZapLogger(applog.GLog.ZapLogger)

	config, _, err = configs.LoadConfigs(path, "")
	if err != nil {
		applog.GLog.Logger.Error("configs.LoadConfig failed",
			"error", err)
		os.Exit(1)
	}

	notify = iNotificationServiceImpl{nil, nil,
		config.NotifyService.Address, config.NotifyService.Port,
		config.NotifyService.NotifySeller, config.NotifyService.NotifyBuyer,
		config.NotifyService.Timeout, sync.Mutex{}}

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
		User:  SellerUser,
	}
	iFuture := notify.NotifyBySMS(ctx, request)
	futureData := iFuture.Get()

	require.Nil(t, futureData.Data())
	require.Nil(t, futureData.Error())
}

func TestNotifyEmail(t *testing.T) {

}
