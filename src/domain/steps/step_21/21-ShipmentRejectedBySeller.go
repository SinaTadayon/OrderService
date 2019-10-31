package shipment_rejected_by_seller_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName string 	= "Shipment_Rejected_By_Seller"
	stepIndex int		= 21
)

type paymentRejectedBySellerStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentRejectedBySellerStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentRejectedBySellerStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &paymentRejectedBySellerStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentRejectedBySeller paymentRejectedBySellerStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (paymentRejectedBySeller paymentRejectedBySellerStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string) promise.IPromise {
	panic("implementation required")
}


//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//// TODO Must be implement
//func ShipmentRejectedBySellerMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.ShipmentRejectedBySeller)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func ShipmentRejectedBySellerAction(message *sarama.ConsumerMessage) error {
//
//	err := ShipmentRejectedBySellerProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func ShipmentRejectedBySellerProduce(topic string, payload []byte) error {
//	return nil
//}
