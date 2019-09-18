package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"time"

	"gitlab.faza.io/go-framework/kafkaadapter"

	"gitlab.faza.io/go-framework/mongoadapter"

	"gitlab.faza.io/go-framework/logger"

	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

var App struct {
	config configuration
	mongo  *mongoadapter.Mongo
	kafka  *kafkaadapter.Kafka
}
var brokers []string

const (
	PaymentUrl                        = "PaymentURL"
	MongoDB                           = "orders"
	Orders                            = "orders"
	OrderRollbackMongoError           = "can not rollback on kafka"
	StateMachineNextStateNotAvailable = "can not go to next state"
)

func main() {
	switch App.config.Kafka.ConsumerTopic {
	case "payment-pending":
		logger.Audit("starting grpc ...")
		startGrpc()
	case "payment-success":
		logger.Audit("starting " + App.config.Kafka.ConsumerTopic)
		startPaymentSuccess(App.config.Kafka.Version, App.config.Kafka.ConsumerTopic)
	case "seller-approval-pending":
		logger.Audit("starting " + App.config.Kafka.ConsumerTopic)
		startGrpc()
	default:
		logger.Err("consumer topic env is wrong:" + App.config.Kafka.ConsumerTopic)
	}
}

func init() {
	err := LoadConfig()
	if err != nil {
		logger.Err(err.Error())
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     App.config.Mongo.Host,
		Port:     App.config.Mongo.Port,
		Username: App.config.Mongo.User,
		//Password:     App.config.Mongo.Pass,
		ConnTimeout:  time.Duration(App.config.Mongo.ConnectionTimeout),
		ReadTimeout:  time.Duration(App.config.Mongo.ReadTimeout),
		WriteTimeout: time.Duration(App.config.Mongo.WriteTimeout),
	}

	App.mongo, err = mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("New Mongo: %v", err.Error())
	}
	_, err = App.mongo.AddUniqueIndex(MongoDB, Orders, "ordernumber")
	if err != nil {
		logger.Err(err.Error())
	}

	err = initTopics()
	if err != nil {
		logger.Err(err.Error())
		os.Exit(1)
	}
}

func LoadConfig() error {
	if os.Getenv("APP_ENV") == "dev" {
		if flag.Lookup("test.v") != nil {
			// test mode
			err := godotenv.Load("testdata/.env")
			if err != nil {
				logger.Err("Error loading testdata .env file")
			}
		} else {
			err := godotenv.Load(".env")
			if err != nil {
				logger.Err("Error loading .env file")
			}
		}
	}

	// Get environment variables for config
	_, err := env.UnmarshalFromEnviron(&App.config)
	if err != nil {
		return err
	}
	brokers = strings.Split(App.config.Kafka.Brokers, ",")
	if App.config.App.Port == "" {
		logger.Err("grpc PORT env not defined")
		return errors.New("grpc PORT env not defined")
	}
	return nil
}
