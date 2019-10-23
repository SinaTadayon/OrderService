package pay_to_buyer_success_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Pay_To_Buyer_Success"
	stepIndex int		= 81
)

type payToBuyerSuccessStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToBuyerSuccessStep{steps.NewBaseStep(stepName, stepIndex, childes,
		parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &payToBuyerSuccessStep{steps.NewBaseStep(name, index, childes,
		parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &payToBuyerSuccessStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (payToBuyerSuccess payToBuyerSuccessStep) ProcessMessage(ctx context.Context, request message.Request) (message.Response, error) {
	panic("implementation required")
}

func (payToBuyerSuccess payToBuyerSuccessStep) ProcessOrder(ctx context.Context, order entities.Order) error {
	panic("implementation required")
}


//
//import (
//	"gitlab.faza.io/order-project/order-service"
//	pb "gitlab.faza.io/protos/order"
//)
//
//func PayToBuyerSuccessAction(ppr PaymentPendingRequest, req *pb.PayToBuyerSuccessRequest) error {
//	err := main.MoveOrderToNewState("operator", req.GetDescription(), main.PayToBuyerSuccess, "pay-to-buyer-success", ppr)
//	if err != nil {
//		return err
//	}
//	return nil
//}