package user_service

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"google.golang.org/grpc/metadata"

	"os"
	"testing"
	"time"
)

var config *configs.Config
var userService *iUserServiceImpl

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

	userService = &iUserServiceImpl{
		client:        nil,
		serverAddress: config.UserService.Address,
		serverPort:    config.UserService.Port,
		timeout:       config.UserService.Timeout,
	}

	// Running Tests
	code := m.Run()
	os.Exit(code)

}

func TestGetSellerInfo(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := userService.getUserService(ctx)
	require.Nil(t, err)

	ctx, _ = context.WithCancel(context.Background())

	// user service create dummy user with id 1000001
	iFuture := userService.GetSellerProfile(ctx, "1000001")
	futureData := iFuture.Get()
	require.Nil(t, futureData.Error())
	require.Equal(t, futureData.Data().(*entities.SellerProfile).SellerId, int64(1000001))
}

func TestAuthenticationToken(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := userService.getUserService(ctx)
	require.Nil(t, err)

	result, err := userService.client.Login("989100000002", "123456", ctx)
	require.Nil(t, err)
	require.NotNil(t, result)
	require.Equal(t, 200, int(result.Code))

	var authorization = map[string]string{"authorization": fmt.Sprintf("Bearer %v", result.Data.AccessToken)}
	md := metadata.New(authorization)
	ctxToken := metadata.NewIncomingContext(context.Background(), md)

	acl, err := userService.AuthenticateContextToken(ctxToken)
	require.Nil(t, err)
	require.Equal(t, acl.User().UserID, int64(1000001))
}

//func CreateRandomMobileNumber(prefix string) string {
//	var min = 1000000
//	var max = 9999999
//	rand.Seed(time.Now().UnixNano())
//	return prefix + strconv.Itoa(rand.Intn(max-min)+min)
//}
//
//func createCustomer() *client.UserFields {
//	random := CreateRandomMobileNumber("")
//	user := &client.UserFields{}
//	user.FirstName = "Client Sample FN"
//	user.LastName = "Client Sample LN"
//	user.Mobile = "0937"+random
//	user.Email = "client@gmail.com"
//	user.UserType = "customer"
//	user.Password = "123456"
//	user.NationalCode = "1234567891"
//	user.CardNumber = "1234123412341234"
//	user.Iban = "IR123456789123456789123456"
//	user.Gender = "male"
//	user.BirthDate = "1990-01-06"
//	return user
//}
