package main

import (
	"gitlab.faza.io/go-framework/kafkaadapter"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"gitlab.faza.io/order-project/order-service/server/grpc"
	"os"

	_ "github.com/devfeel/mapper"
)

var App struct {
	config 			*configs.Cfg
	kafka  			*kafkaadapter.Kafka
	orderRepository repository.IOrderRepository
	flowManager		domain.IFlowManager
	grpcServer		grpc.Server
}
var brokers []string

const (
	PaymentUrl                        = "PaymentURL"
	//MongoDB                           = "orders"
	//Orders                            = "orders"
	OrderRollbackMongoError           = "can not rollback on kafka"
	StateMachineNextStateNotAvailable = "can not go to next state"
)

// TODO Add worker scheduler and start from main
func main() {

	if App.config.App.ServiceMode == "server" {
		App.grpcServer.Start()
	}

	//switch App.Cfg.Kafka.ConsumerTopic {
	//case "payment-pending":
	//	logger.Audit("starting grpc ...")
	//	server.startGrpc()
	//case "payment-success":
	//	logger.Audit("starting " + App.Cfg.Kafka.ConsumerTopic)
	//	startPaymentSuccess(App.Cfg.Kafka.Version, App.Cfg.Kafka.ConsumerTopic)
	//case "seller-approval-pending":
	//	logger.Audit("starting " + App.Cfg.Kafka.ConsumerTopic)
	//	server.startGrpc()
	//default:
	//	logger.Err("consumer topic env is wrong:" + App.Cfg.Kafka.ConsumerTopic)
	//}
}

func init() {
	var err error
	if os.Getenv("APP_ENV") == "dev" {
		App.config, err = configs.LoadConfig("./testdata/.env")
	} else {
		App.config, err = configs.LoadConfig("")
	}
	if err != nil {
		logger.Err("LoadConfig of main init failed, %s ", err.Error())
		panic("LoadConfig of main init failed, " + err.Error())
	}

	 App.orderRepository ,err = repository.NewOrderRepository(App.config)
	 if err != nil {
		 logger.Err("repository creation failed, %s ", err.Error())
		 panic("repository creation failed, " + err.Error())
	 }

	 // TODO create item repository
	 App.flowManager, err = domain.NewFlowManager(App.orderRepository, nil)
	if err != nil {
		logger.Err("flowManager creation failed, %s ", err.Error())
		panic("flowManager creation failed, " + err.Error())
	}

	 App.grpcServer = grpc.NewServer(App.config.GRPCServer.Address, uint16(App.config.GRPCServer.Port), App.flowManager, App.orderRepository)

	//brokers = strings.Split(App.config.Kafka.Brokers, ",")
	//if App.config.App.Port == "" {
	//	logger.Err("grpc PORT env not defined")
	//	//return errors.New("grpc PORT env not defined")
	//}




	//// store in mongo
	//mongoConf := &mongoadapter.MongoConfig{
	//	Host:     App.config.Mongo.Host,
	//	Port:     App.config.Mongo.Port,
	//	Username: App.config.Mongo.User,
	//	//Password:     App.Cfg.Mongo.Pass,
	//	ConnTimeout:  time.Duration(App.config.Mongo.ConnectionTimeout),
	//	ReadTimeout:  time.Duration(App.config.Mongo.ReadTimeout),
	//	WriteTimeout: time.Duration(App.config.Mongo.WriteTimeout),
	//}
	//
	//App.mongo, err = mongoadapter.NewMongo(mongoConf)
	//if err != nil {
	//	logger.Err("New Mongo: %v", err.Error())
	//}
	//_, err = App.mongo.AddUniqueIndex(MongoDB, Orders, "orderId")
	//if err != nil {
	//	logger.Err(err.Error())
	//}

	//err = initTopics()
	//if err != nil {
	//	logger.Err(err.Error())
	//	os.Exit(1)
	//}
}

