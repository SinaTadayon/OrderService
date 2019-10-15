package configs

import (
	"flag"
	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
	"gitlab.faza.io/go-framework/logger"
	"os"
)

type Cfg struct {
	App struct {
		Port                                 string `env:"PORT"`
		SmsTemplateDir                       string `env:"NOTIFICATION_SMS_TEMPLATES"`
		EmailTemplateNotifySellerForNewOrder string `env:"EMAIL_TMP_NOTIFY_SELLER_FOR_NEW_ORDER"`
	}
	Kafka struct {
		Version       string `env:"PAYMENT_KAFKA_VERSION"`
		Brokers       string `env:"PAYMENT_KAFKA_BROKERS"`
		ConsumerTopic string `env:"PAYMENT_KAFKA_CONSUMER_TOPIC"`
		ConsumerGroup string `env:"PAYMENT_KAFKA_CONSUMER_GROUP"`
		Partition     string `env:"PAYMENT_KAFKA_PARTITION"`
		Replica       string `env:"PAYMENT_KAFKA_REPLICA"`
	}
	Mongo struct {
		User              string `env:"PAYMENT_MONGO_USER"`
		Pass              string `env:"PAYMENT_MONGO_PASS"`
		Host              string `env:"PAYMENT_MONGO_HOST"`
		Port              int    `env:"PAYMENT_MONGO_PORT"`
		ConnectionTimeout int    `env:"PAYMENT_MONGO_CONN_TIMEOUT"`
		ReadTimeout       int    `env:"PAYMENT_MONGO_READ_TIMEOUT"`
		WriteTimeout      int    `env:"PAYMENT_MONGO_WRITE_TIMEOUT"`
	}
}

func LoadConfig() (*Cfg, error) {
	var config = &Cfg{}
	if os.Getenv("APP_ENV") == "dev" {
		if flag.Lookup("test.v") != nil {
			// test mode
			err := godotenv.Load("../testdata/.env")
			if err != nil {
				logger.Err("Error loading testdata .env file")
			}
		} else {
			err := godotenv.Load("../.env")
			if err != nil {
				logger.Err("Error loading .env file")
			}
		}
	}

	// Get environment variables for Cfg
	_, err := env.UnmarshalFromEnviron(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
