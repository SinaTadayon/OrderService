package pay_to_seller_success_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Pay_To_Seller_Success"
	stepIndex int		= 91
)

type payToSellerSuccessStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToSellerSuccessStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToSellerSuccessStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &payToSellerSuccessStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (payToSellerSuccess payToSellerSuccessStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (payToSellerSuccess payToSellerSuccessStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}



//
//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func PayToSellerSuccessAction(ppr PaymentPendingRequest, req *pb.PayToSellerSuccessRequest) error {
//	err := main.MoveOrderToNewState("operator", req.GetDescription(), main.PayToSellerSuccess, "pay-to-seller-success", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}
