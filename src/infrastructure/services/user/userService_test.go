package user_service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	client "gitlab.faza.io/services/user-app-client"
	"google.golang.org/grpc"
	pb1 "gitlab.faza.io/protos/user"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

var config *configs.Cfg
var userService *iUserServiceImpl


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

	userService = &iUserServiceImpl {
		client:        nil,
		serverAddress: config.UserService.Address,
		serverPort:    config.UserService.Port,
	}
}

func TestGetSellerInfo(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5 * time.Second)
	err := userService.getUserService(ctx)
	assert.Nil(t, err)

	result, err := userService.client.RegisterUser(createCustomer(), "", ctx)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, int(result.Code))
	data, err := userService.client.UserList(1,1,nil,nil,"", ctx)
	assert.Nil(t, err)
	assert.Nil(t, data)
}

func CreateRandomMobileNumber(prefix string) string {
	var min = 1000000
	var max = 9999999
	rand.Seed(time.Now().UnixNano())
	return prefix + strconv.Itoa(rand.Intn(max-min)+min)
}

func createCustomer() *client.UserFields {
	random := CreateRandomMobileNumber("")
	user := &client.UserFields{}
	user.FirstName = "Client Sample FN"
	user.LastName = "Client Sample LN"
	user.Mobile = "0937"+random
	user.Email = "client@gmail.com"
	user.UserType = "customer"
	user.Password = "123456"
	user.NationalCode = "1234567891"
	user.CardNumber = "1234123412341234"
	user.Iban = "IR123456789123456789123456"
	user.Gender = "male"
	user.BirthDate = "1990-01-06"
	return user
}



