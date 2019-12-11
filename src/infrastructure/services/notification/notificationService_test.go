package notify_service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"os"
	"testing"
)

var config *configs.Config
var notify iNotificationServiceImpl

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

	notify = iNotificationServiceImpl{nil, nil,
		config.NotifyService.Address, config.NotifyService.Port}
}

func TestNotifySMS(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	request := SMSRequest{
		Phone: "09373969041",
		Body:  "سلام، این اس ام اس تستی هست",
	}
	iFuture := notify.NotifyBySMS(ctx, request)
	futureData := iFuture.Get()

	assert.Nil(t, futureData.Data())
	assert.Nil(t, futureData.Error())
}

func TestNotifyEmail(t *testing.T) {

}
