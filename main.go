package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/go-framework/redisadapter"
)

var App struct {
	config configuration
	// mode:
	// server = grpc server
	// consumer = kafka consumer
	mode  string
	mongo *mongoadapter.Mongo
}
var mode string

func main() {
	logger.Audit("App started...")

	if App.mode == "server" {
		logger.Audit("started as grpc server")
		//startGrpcServer()
	} else {
		logger.Audit("started as consumer")
		//startConsumer(App.config.Kafka.Version, App.config.Kafka.ConsumerTopic)
	}
}

// Check and verify configs
func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Err("Error loading .env file")
	}

	// Get environment variables for config
	_, err = env.UnmarshalFromEnviron(&App.config)
	if err != nil {
		log.Fatal(err)
	}

	// Validate configs
	if App.config.Redis.Host == "" ||
		App.config.Redis.Port == "" {
		logger.Err("Cant get configs")
	}

	redisPort, err := strconv.Atoi(App.config.Redis.Port)
	if err != nil {
		logger.Err("Cant convert redis port %v", err.Error())
		os.Exit(1)
	}

	mongoPort, err := strconv.Atoi(App.config.Mongo.Port)
	if err != nil {
		logger.Err("Cant convert mongo port %v", err.Error())
		os.Exit(1)
	}
	mongoConnTimeout, err := strconv.Atoi(App.config.Mongo.ConnTimeout)
	if err != nil {
		logger.Err("Cant convert mongo connTimeout %v", err.Error())
		os.Exit(1)
	}
	mongoReadTimeout, err := strconv.Atoi(App.config.Mongo.ReadTimeout)
	if err != nil {
		logger.Err("Cant convert mongo ReadTimeout %v", err.Error())
		os.Exit(1)
	}
	mongoWriteTimeout, err := strconv.Atoi(App.config.Mongo.WriteTimeout)
	if err != nil {
		logger.Err("Cant convert mongo WriteTimeout %v", err.Error())
		os.Exit(1)
	}

	logger.Audit("App start in " + App.config.App.Mode + " mode")

	if App.config.App.Mode == "" {
		mode = "server"
	} else {
		mode = App.config.App.Mode
	}

	if mode == "server" {
		App.mode = "server"
	} else if mode == "consumer" {
		App.mode = "consumer"
	} else {
		App.mode = "server"
	}

	App.redis, err = redisadapter.NewRedis(&redisadapter.RedisConfig{Host: App.config.Redis.Host, Port: redisPort}, nil)
	if err != nil {
		logger.Err("Cant get redis, %v", err.Error())
		os.Exit(1)
	}
	err = App.redis.Connect()
	if err != nil {
		logger.Err("Cant connect to Redis, %v", err.Error())
		os.Exit(1)
	}

	mongoConfig := mongoadapter.MongoConfig{
		Host:         App.config.Mongo.Host,
		Port:         mongoPort,
		Username:     App.config.Mongo.Username,
		Password:     App.config.Mongo.Password,
		ConnTimeout:  time.Duration(mongoConnTimeout) * time.Second,
		ReadTimeout:  time.Duration(mongoReadTimeout) * time.Second,
		WriteTimeout: time.Duration(mongoWriteTimeout) * time.Second,
	}
	App.mongo, err = mongoadapter.NewMongo(&mongoConfig)
	if err != nil {
		logger.Err("Cant get mongo connection")
	}

	MongoMigrations()
}
