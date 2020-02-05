module gitlab.faza.io/order-project/order-service

go 1.13

require (
	github.com/Netflix/go-env v0.0.0-20180529183433-1e80ef5003ef
	github.com/devfeel/mapper v0.7.2
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/joho/godotenv v1.3.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.3.0
	github.com/shopspring/decimal v0.0.0-20191130220710-360f2bc03045
	github.com/stretchr/testify v1.4.0
	gitlab.faza.io/go-framework/acl v0.0.3
	gitlab.faza.io/go-framework/logger v0.0.10
	gitlab.faza.io/go-framework/mongoadapter v0.0.9
	gitlab.faza.io/protos/cart v0.0.14
	gitlab.faza.io/protos/notification v0.0.3
	gitlab.faza.io/protos/order v0.0.59
	gitlab.faza.io/protos/payment-gateway v0.0.14
	gitlab.faza.io/protos/stock-proto.git v0.0.8
	gitlab.faza.io/protos/user v0.0.41
	gitlab.faza.io/services/user-app-client v0.0.20
	go.mongodb.org/mongo-driver v1.2.0
	go.uber.org/zap v1.13.0
	google.golang.org/grpc v1.26.0
)
