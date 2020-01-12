package configs

import (
	"flag"
	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
	"gitlab.faza.io/go-framework/logger"
	"os"
)

type SmsTemplate struct {
	OrderNotifyBuyerPaymentSuccessState                         string `env:"ORDER_NOTIFY_BUYER_PAYMENT_SUCCESS_STATE"`
	OrderNotifyBuyerPaymentFailedState                          string `env:"ORDER_NOTIFY_BUYER_PAYMENT_FAILED_STATE"`
	OrderNotifySellerApprovalPendingState                       string `env:"ORDER_NOTIFY_SELLER_APPROVAL_PENDING_STATE"`
	OrderNotifyBuyerShipmentPendingState                        string `env:"ORDER_NOTIFY_BUYER_SHIPMENT_PENDING_STATE"`
	OrderNotifySellerShipmentDelayedState                       string `env:"ORDER_NOTIFY_SELLER_SHIPMENT_DELAYED_STATE"`
	OrderNotifyBuyerShipmentDelayedState                        string `env:"ORDER_NOTIFY_BUYER_SHIPMENT_DELAYED_STATE"`
	OrderNotifySellerCanceledByBuyerState                       string `env:"ORDER_NOTIFY_SELLER_CANCELED_BY_BUYER_STATE"`
	OrderNotifyBuyerCanceledByBuyerState                        string `env:"ORDER_NOTIFY_BUYER_CANCELED_BY_BUYER_STATE"`
	OrderNotifyBuyerCanceledBySellerState                       string `env:"ORDER_NOTIFY_BUYER_CANCELED_BY_SELLER_STATE"`
	OrderNotifyBuyerDeliveryPendingState                        string `env:"ORDER_NOTIFY_BUYER_DELIVERY_PENDING_STATE"`
	OrderNotifySellerReturnRequestPendingState                  string `env:"ORDER_NOTIFY_SELLER_RETURN_REQUEST_PENDING_STATE"`
	OrderNotifyBuyerReturnRequestPendingState                   string `env:"ORDER_NOTIFY_BUYER_RETURN_REQUEST_PENDING_STATE"`
	OrderNotifyBuyerReturnShipmentPendingState                  string `env:"ORDER_NOTIFY_BUYER_RETURN_SHIPMENT_PENDING_STATE"`
	OrderNotifySellerReturnRequestRejectedState                 string `env:"ORDER_NOTIFY_SELLER_RETURN_REQUEST_REJECTED_STATE"`
	OrderNotifyBuyerReturnCanceledState                         string `env:"ORDER_NOTIFY_BUYER_RETURN_CANCELED_STATE"`
	OrderNotifyBuyerReturnDeliveryPendingToReturnDeliveredState string `env:"ORDER_NOTIFY_BUYER_RETURN_DELIVERY_PENDING_TO_RETURN_DELIVERED_STATE"`
	OrderNotifyBuyerReturnDeliveryDelayedToReturnDeliveredState string `env:"ORDER_NOTIFY_BUYER_RETURN_DELIVERY_DELAYED_TO_RETURN_DELIVERED_STATE"`
	OrderNotifyBuyerReturnDeliveredToPayToBuyerState            string `env:"ORDER_NOTIFY_BUYER_RETURN_DELIVERED_TO_PAY_TO_BUYER_STATE"`
	OrderNotifyBuyerReturnRejectedToPayToBuyerState             string `env:"ORDER_NOTIFY_BUYER_RETURN_REJECTED_TO_PAY_TO_BUYER_STATE"`
	OrderNotifyBuyerReturnRejectedToPayToSellerState            string `env:"ORDER_NOTIFY_BUYER_RETURN_REJECTED_TO_PAY_TO_SELLER_STATE"`
}

type Config struct {
	App struct {
		ServiceMode                          string `env:"ORDER_SERVICE_MODE"`
		SmsTemplates                         string `env:"NOTIFICATION_SMS_TEMPLATES"`
		EmailTemplateNotifySellerForNewOrder string `env:"EMAIL_TMP_NOTIFY_SELLER_FOR_NEW_ORDER"`
		PrometheusPort                       int    `env:"PROMETHEUS_PORT"`

		OrderPaymentCallbackUrlSuccess             string `env:"ORDER_PAYMENT_CALLBACK_URL_SUCCESS"`
		OrderPaymentCallbackUrlFail                string `env:"ORDER_PAYMENT_CALLBACK_URL_FAIL"`
		OrderPaymentCallbackUrlAsanpardakhtSuccess string `env:"ORDER_PAYMENT_CALLBACK_URL_ASANPARDAKHT_SUCCESS"`
		OrderPaymentCallbackUrlAsanpardakhtFail    string `env:"ORDER_PAYMENT_CALLBACK_URL_ASANPARDAKHT_FAIL"`

		SchedulerTimeUint            string `env:"ORDER_SCHEDULER_TIME_UNIT"`
		SchedulerStates              string `env:"ORDER_SCHEDULER_STATES"`
		SchedulerInterval            string `env:"ORDER_SCHEDULER_INTERVAL"`
		SchedulerParentWorkerTimeout string `env:"ORDER_SCHEDULER_PARENT_WORKER_TIMEOUT"`
		SchedulerWorkerTimeout       string `env:"ORDER_SCHEDULER_WORKER_TIMEOUT"`

		SchedulerStateTimeUint              string `env:"ORDER_SCHEDULER_STATE_TIME_UINT"`
		SchedulerSellerReactionTime         string `env:"ORDER_SCHEDULER_SELLER_REACTION_TIME"`
		SchedulerPaymentPendingState        string `env:"ORDER_SCHEDULER_PAYMENT_PENDING_STATE"`
		SchedulerApprovalPendingState       string `env:"ORDER_SCHEDULER_APPROVAL_PENDING_STATE"`
		SchedulerShipmentPendingState       string `env:"ORDER_SCHEDULER_SHIPMENT_PENDING_STATE"`
		SchedulerShippedState               string `env:"ORDER_SCHEDULER_SHIPPED_STATE"`
		SchedulerDeliveryPendingState       string `env:"ORDER_SCHEDULER_DELIVERY_PENDING_STATE"`
		SchedulerNotifyDeliveryPendingState string `env:"ORDER_SCHEDULER_NOTIFY_DELIVERY_PENDING_STATE"`
		SchedulerDeliveredState             string `env:"ORDER_SCHEDULER_DELIVERED_STATE"`
		SchedulerReturnShippedState         string `env:"ORDER_SCHEDULER_RETURN_SHIPPED_STATE"`
		SchedulerReturnRequestPendingState  string `env:"ORDER_SCHEDULER_RETURN_REQUEST_PENDING_STATE"`
		SchedulerReturnShipmentPendingState string `env:"ORDER_SCHEDULER_RETURN_SHIPMENT_PENDING_STATE"`
		SchedulerReturnDeliveredState       string `env:"ORDER_SCHEDULER_RETURN_DELIVERED_STATE"`
	}

	GRPCServer struct {
		Address string `env:"ORDER_SERVER_ADDRESS"`
		Port    int    `env:"ORDER_SERVER_PORT"`
	}

	UserService struct {
		Address string `env:"USER_SERVICE_ADDRESS"`
		Port    int    `env:"USER_SERVICE_PORT"`
	}

	NotifyService struct {
		Address string `env:"NOTIFY_SERVICE_ADDRESS"`
		Port    int    `env:"NOTIFY_SERVICE_PORT"`
	}

	VoucherService struct {
		Address     string `env:"VOUCHER_SERVICE_ADDRESS"`
		Port        int    `env:"VOUCHER_SERVICE_PORT"`
		MockEnabled bool   `env:"VOUCHER_SERVICE_MOCK_ENABLED"`
	}

	PaymentGatewayService struct {
		Address     string `env:"PAYMENT_GATEWAY_ADDRESS"`
		Port        int    `env:"PAYMENT_GATEWAY_PORT"`
		MockEnabled bool   `env:"PAYMENT_SERVICE_MOCK_ENABLED"`
	}

	StockService struct {
		Address     string `env:"STOCK_SERVICE_ADDRESS"`
		Port        int    `env:"STOCK_SERVICE_PORT"`
		MockEnabled bool   `env:"STOCK_SERVICE_MOCK_ENABLED"`
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
		MaxConnIdleTime   int    `env:"ORDER_SERVICE_MONGO_MAX_CONN_IDLE_TIME"`
		MaxPoolSize       int    `env:"ORDER_SERVICE_MONGO_MAX_POOL_SIZE"`
		MinPoolSize       int    `env:"ORDER_SERVICE_MONGO_MIN_POOL_SIZE"`
	}
}

func LoadConfig(path string) (*Config, *SmsTemplate, error) {
	var config = &Config{}
	var smsTemplate = &SmsTemplate{}
	currntPath, err := os.Getwd()
	if err != nil {
		logger.Err("get current working directory failed, error %s", err)
	}

	if os.Getenv("APP_MODE") == "dev" {
		if path != "" {
			err := godotenv.Load(path)
			if err != nil {
				logger.Err("Error loading testdata .env file, Working Directory: %s  path: %s, error: %s", currntPath, path, err)
			}
		} else if flag.Lookup("test.v") != nil {
			// test mode
			err := godotenv.Load("../testdata/.env")
			if err != nil {
				logger.Err("Error loading testdata .env file, error: %s", err)
			}
		}
	} else if os.Getenv("APP_MODE") == "docker" {
		err := godotenv.Load(path)
		if err != nil {
			logger.Err("Error loading .docker-env file, " + path)
		}
	}

	// Get environment variables for Config
	_, err = env.UnmarshalFromEnviron(config)
	if err != nil {
		logger.Err("env.UnmarshalFromEnviron config failed")
		return nil, nil, err
	}

	if config.App.SmsTemplates != "" {
		err := godotenv.Load(config.App.SmsTemplates)
		if err != nil {
			logger.Err("Error loading " + config.App.SmsTemplates + " file")
			return nil, nil, err
		}
	}

	_, err = env.UnmarshalFromEnviron(smsTemplate)
	if err != nil {
		logger.Err("env.UnmarshalFromEnviron smsTemplate failed")
		return nil, nil, err
	}

	return config, smsTemplate, nil
}

func LoadConfigs(configPath string, smsTemplatePath string) (*Config, *SmsTemplate, error) {
	var config = &Config{}
	var smsTemplate = &SmsTemplate{}
	currntPath, err := os.Getwd()
	if err != nil {
		logger.Err("get current working directory failed, error %s", err)
	}

	if os.Getenv("APP_MODE") == "dev" {
		if configPath != "" {
			err := godotenv.Load(configPath)
			if err != nil {
				logger.Err("Error loading testdata .env file, Working Directory: %s  path: %s, error: %s", currntPath, configPath, err)
			}
		} else if flag.Lookup("test.v") != nil {
			// test mode
			err := godotenv.Load("../testdata/.env")
			if err != nil {
				logger.Err("Error loading testdata .env file, error: %s", err)
			}
		}
	}

	// Get environment variables for Config
	_, err = env.UnmarshalFromEnviron(config)
	if err != nil {
		logger.Err("env.UnmarshalFromEnviron config failed")
		return nil, nil, err
	}

	if smsTemplatePath != "" {
		err = godotenv.Load(smsTemplatePath)
		if err != nil {
			logger.Err("Error loading " + smsTemplatePath + " file")
			return nil, nil, err
		}

		_, err = env.UnmarshalFromEnviron(smsTemplate)
		if err != nil {
			logger.Err("env.UnmarshalFromEnviron smsTemplate failed")
			return nil, nil, err
		}
		return config, smsTemplate, nil
	}

	return config, nil, nil
}
