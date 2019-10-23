module gitlab.faza.io/order-project/order-service

go 1.13

require (
	github.com/Netflix/go-env v0.0.0-20180529183433-1e80ef5003ef
	github.com/Shopify/sarama v1.24.0
	github.com/cheekybits/is v0.0.0-20150225183255-68e9c0620927 // indirect
	github.com/devfeel/mapper v0.0.0-20190905045745-405b6c90b771
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/pretty v1.0.0 // indirect
	gitlab.faza.io/go-framework/kafkaadapter v0.0.1
	gitlab.faza.io/go-framework/logger v0.0.3
	gitlab.faza.io/go-framework/mongoadapter v0.0.6
	gitlab.faza.io/protos/order v0.0.0-20191014173539-2bcc7283a98d
	gitlab.faza.io/services/notification-client v0.0.3
	go.mongodb.org/mongo-driver v1.1.2
	google.golang.org/grpc v1.24.0
)
