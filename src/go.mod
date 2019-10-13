module order-service

go 1.13

require (
	github.com/Netflix/go-env v0.0.0-20180529183433-1e80ef5003ef
	github.com/Shopify/sarama v1.24.0
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/joho/godotenv v1.3.0
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/rs/xid v1.2.1
	github.com/stretchr/testify v1.4.0
	gitlab.faza.io/go-framework/kafkaadapter v0.0.1
	gitlab.faza.io/go-framework/logger v0.0.3
	gitlab.faza.io/go-framework/mongoadapter v0.0.3
	gitlab.faza.io/protos/order v0.0.0-20191013164541-b913dc0b7c67
	gitlab.faza.io/services/notification-client v0.0.3
	go.mongodb.org/mongo-driver v1.1.2
	google.golang.org/grpc v1.24.0
)
