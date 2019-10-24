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
		Version       string `env:"ORDER_SERVICE_KAFKA_VERSION"`
		Brokers       string `env:"ORDER_SERVICE_KAFKA_BROKERS"`
		ConsumerTopic string `env:"ORDER_SERVICE_KAFKA_CONSUMER_TOPIC"`
		ConsumerGroup string `env:"ORDER_SERVICE_KAFKA_CONSUMER_GROUP"`
		Partition     string `env:"ORDER_SERVICE_KAFKA_PARTITION"`
		Replica       string `env:"ORDER_SERVICE_KAFKA_REPLICA"`
	}
	Mongo struct {
		User              string `env:"ORDER_SERVICE_MONGO_USER"`
		Pass              string `env:"ORDER_SERVICE_MONGO_PASS"`
		Host              string `env:"ORDER_SERVICE_MONGO_HOST"`
		Port              int    `env:"ORDER_SERVICE_MONGO_PORT"`
		ConnectionTimeout int    `env:"ORDER_SERVICE_MONGO_CONN_TIMEOUT"`
		ReadTimeout       int    `env:"ORDER_SERVICE_MONGO_READ_TIMEOUT"`
		WriteTimeout      int    `env:"ORDER_SERVICE_MONGO_WRITE_TIMEOUT"`
		MaxConnIdleTime	  int	 `env:"ORDER_SERVICE_MONGO_MAX_CONN_IDLE_TIME"`
		MaxPoolSize		  int	 `env:"ORDER_SERVICE_MONGO_MAX_POOL_SIZE"`
		MinPoolSize		  int	 `env:"ORDER_SERVICE_MONGO_MIN_POOL_SIZE"`
	}
}

func LoadConfig(path string) (*Cfg, error) {
	var config = &Cfg{}
	if os.Getenv("APP_ENV") == "dev" {
		if path != "" {
			err := godotenv.Load(path)
			if err != nil {
				logger.Err("Error loading testdata .env file, path: %s", path)
			}
		} else if flag.Lookup("test.v") != nil {
			// test mode
			err := godotenv.Load("../testdata/.env")
			//err := godotenv.Load(path)
			if err != nil {
				logger.Err("Error loading testdata .env file")
			}
		} else {
			//err := godotenv.Load(path)
			err := godotenv.Load("../.env")
			if err != nil {
				logger.Err("Error loading .env file")
			}
		}
	}

	//else if len(path) != 0 {
	//	err := godotenv.Load(path)
	//	if err != nil {
	//		logger.Err("Error loading .env file, path: %s", path)
	//	}
	//}

	// Get environment variables for Cfg
	_, err := env.UnmarshalFromEnviron(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

//func LoadConfigWithPath(path string) (*Cfg, error) {
//	var config = &Cfg{}
//
//	if os.Getenv("APP_ENV") == "dev" {
//		if flag.Lookup("test.v") != nil {
//			// test mode
//			err := godotenv.Load("../testdata/.env")
//			if err != nil {
//				logger.Err("Error loading testdata .env file")
//			}
//		} else {
//			err := godotenv.Load("../.env")
//			if err != nil {
//				logger.Err("Error loading .env file")
//			}
//		}
//	} else if len(path) != 0 {
//		err := godotenv.Load(path)
//		if err != nil {
//			logger.Err("Error loading .env file, path: %s", path)
//		}
//	}
//
//	// Get environment variables for Cfg
//	_, err1 := env.UnmarshalFromEnviron(config)
//	if err1 != nil {
//		return nil, err1
//	}
//
//	return config, nil
//}
