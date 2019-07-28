package main

type configuration struct {
	App struct {
		Mode string `env:"PAYMENT_APP_MODE"`
		Port string `env:"PAYMENT_APP_PORT"`
	}
	Kafka struct {
		Version       string `env:"PAYMENT_KAFKA_VERSION"`
		Brokers       string `env:"PAYMENT_KAFKA_BROKERS"`
		ConsumerTopic string `env:"PAYMENT_KAFKA_CONSUMER_TOPIC"`
		ConsumerGroup string `env:"PAYMENT_KAFKA_CONSUMER_GROUP"`
	}
	Redis struct {
		Host string `env:"PAYMENT_REDIS_HOST"`
		Port string `env:"PAYMENT_REDIS_PORT"`
	}
	Mongo struct {
		Host         string `env:"PAYMENT_MONGO_HOST"`
		Port         string `env:"PAYMENT_MONGO_PORT"`
		Username     string `env:"PAYMENT_MONGO_USER"`
		Password     string `env:"PAYMENT_MONGO_PASS"`
		ConnTimeout  string `env:"PAYMENT_MONGO_CONN_TIMEOUT"`
		ReadTimeout  string `env:"PAYMENT_MONGO_READ_TIMEOUT"`
		WriteTimeout string `env:"PAYMENT_MONGO_WRITE_TIMEOUT"`
	}
}
