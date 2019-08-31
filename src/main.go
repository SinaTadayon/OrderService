package main

import (
	"errors"
	"os"
	"strings"

	"gitlab.faza.io/go-framework/logger"

	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

var App struct {
	config configuration
}
var brokers []string

func main() {
	switch App.config.Kafka.ConsumerTopic {
	case "payment-pending":
		startPaymentPending(App.config.Kafka.Version, App.config.Kafka.ConsumerTopic)
	case "payment-success":
		startPaymentSuccess(App.config.Kafka.Version, App.config.Kafka.ConsumerTopic)
	case "payment-failed":
		startPaymentFailed(App.config.Kafka.Version, App.config.Kafka.ConsumerTopic)
	case "payment-control":
		startPaymentControl(App.config.Kafka.Version, App.config.Kafka.ConsumerTopic)
	}

}

func init() {
	err := LoadConfig()
	if err != nil {
		logger.Err(err.Error())
	}

	//err = initTopics()
	//if err != nil {
	//	logger.Err(err.Error())
	//	os.Exit(1)
	//}
}

func LoadConfig() error {
	if os.Getenv("APP_ENV") == "dev" {
		err := godotenv.Load(".env")
		if err != nil {
			return errors.New("Error loading .env file")
		}
	}

	// Get environment variables for config
	_, err := env.UnmarshalFromEnviron(&App.config)
	if err != nil {
		return err
	}
	brokers = strings.Split(App.config.Kafka.Brokers, ",")
	return nil
}
