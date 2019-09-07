module gitlab.faza.io/payment-project/payment-app

go 1.12

require (
	github.com/Netflix/go-env v0.0.0-20180529183433-1e80ef5003ef
	github.com/Shopify/sarama v1.23.0
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/joho/godotenv v1.3.0
	github.com/rs/xid v1.2.1
	github.com/stretchr/testify v1.3.0
	gitlab.faza.io/go-framework/kafkaadapter v0.0.1
	gitlab.faza.io/go-framework/logger v0.0.3
	gitlab.faza.io/go-framework/mongoadapter v0.0.3
	gitlab.faza.io/protos/payment v0.0.0-20190907080532-1fc2628fc20a
	google.golang.org/grpc v1.23.0
)
