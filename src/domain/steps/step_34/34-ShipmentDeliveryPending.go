package shipment_delivery_pending_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order"
)

const (
	stepName  string = "Shipment_Delivery_Pending"
	stepIndex int    = 34
)

type shipmentDeliveryPendingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDeliveryPendingStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &shipmentDeliveryPendingStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &shipmentDeliveryPendingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (shipmentDeliveryPending shipmentDeliveryPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (shipmentDeliveryPending shipmentDeliveryPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) promise.IPromise {
	panic("implementation required")
}

//
//import (
//	"github.com/Shopify/sarama"
//	"gitlab.faza.io/order-project/order-service"
//)
//
//// TODO: Must be implement ShipmentDeliveryPendingAction
//func ShipmentDeliveryPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
//	mess, err := main.CheckOrderKafkaAndMongoStatus(message, main.ShipmentDeliveryPending)
//	if err != nil {
//		return mess, err
//	}
//	return message, nil
//}
//
//func ShipmentDeliveryPendingAction(message *sarama.ConsumerMessage) error {
//
//	err := ShipmentDeliveryPendingProduce("", []byte{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func ShipmentDeliveryPendingProduce(topic string, payload []byte) error {
//	return nil
//}
