package payment_pending_step

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	message "gitlab.faza.io/protos/order/general"
)

const (
	stepName string 	= "Payment_Pending"
	stepIndex int		= 10
)

type paymentPendingStep struct {
	*steps.BaseStepImpl
}

func New(childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentPendingStep{steps.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []steps.IStep, states ...states.IState) steps.IStep {
	return &paymentPendingStep{steps.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *steps.BaseStepImpl) steps.IStep {
	return &paymentPendingStep{base}
}

func NewValueOf(base *steps.BaseStepImpl, params ...interface{}) steps.IStep {
	panic("implementation required")
}

func (paymentPending paymentPendingStep) ProcessMessage(ctx context.Context, request *message.Request) promise.IPromise {
	panic("implementation required")
}

func (paymentPending paymentPendingStep) ProcessOrder(ctx context.Context, order entities.Order) promise.IPromise {
	panic("implementation required")
}


//func (ppr *PaymentPendingRequest) validate() error {
//	var errValidation []string
//	// Validate order number
//	errPaymentRequest := validation.ValidateStruct(ppr,
//		validation.Field(&ppr.OrderNumber, validation.Required, validation.Length(5, 250)),
//	)
//	if errPaymentRequest != nil {
//		errValidation = append(errValidation, errPaymentRequest.Error())
//	}
//
//	// Validate Buyer
//	errPaymentRequestBuyer := validation.ValidateStruct(&ppr.Buyer,
//		validation.Field(&ppr.Buyer.FirstName, validation.Required),
//		validation.Field(&ppr.Buyer.LastName, validation.Required),
//		validation.Field(&ppr.Buyer.Email, validation.Required, is.Email),
//		validation.Field(&ppr.Buyer.NationalId, validation.Required, validation.Length(10, 10)),
//		validation.Field(&ppr.Buyer.Mobile, validation.Required),
//	)
//	if errPaymentRequestBuyer != nil {
//		errValidation = append(errValidation, errPaymentRequestBuyer.Error())
//	}
//
//	// Validate Buyer finance
//	errPaymentRequestBuyerFinance := validation.ValidateStruct(&ppr.Buyer.Finance,
//		validation.Field(&ppr.Buyer.Finance.Iban, validation.Required, validation.Length(26, 26)),
//	)
//	if errPaymentRequestBuyerFinance != nil {
//		errValidation = append(errValidation, errPaymentRequestBuyerFinance.Error())
//	}
//
//	// Validate Buyer address
//	errPaymentRequestBuyerAddress := validation.ValidateStruct(&ppr.Buyer.Address,
//		validation.Field(&ppr.Buyer.Address.Address, validation.Required),
//		validation.Field(&ppr.Buyer.Address.State, validation.Required),
//		validation.Field(&ppr.Buyer.Address.City, validation.Required),
//		validation.Field(&ppr.Buyer.Address.Country, validation.Required),
//		validation.Field(&ppr.Buyer.Address.ZipCode, validation.Required),
//		validation.Field(&ppr.Buyer.Address.Phone, validation.Required),
//	)
//	if errPaymentRequestBuyerAddress != nil {
//		errValidation = append(errValidation, errPaymentRequestBuyerAddress.Error())
//	}
//
//	// Validate amount
//	errPaymentRequestAmount := validation.ValidateStruct(&ppr.Amount,
//		validation.Field(&ppr.Amount.Total, validation.Required),
//		validation.Field(&ppr.Amount.Discount, validation.Required),
//		validation.Field(&ppr.Amount.Payable, validation.Required),
//	)
//	if errPaymentRequestAmount != nil {
//		errValidation = append(errValidation, errPaymentRequestAmount.Error())
//	}
//
//	if len(ppr.Items) != 0 {
//		for i := range ppr.Items {
//			// Validate amount
//			errPaymentRequestItems := validation.ValidateStruct(&ppr.Items[i],
//				validation.Field(&ppr.Items[i].Sku, validation.Required),
//				validation.Field(&ppr.Items[i].Quantity, validation.Required),
//				validation.Field(&ppr.Items[i].Title, validation.Required),
//				validation.Field(&ppr.Items[i].Categories, validation.Required),
//				validation.Field(&ppr.Items[i].Brand, validation.Required),
//			)
//			if errPaymentRequestItems != nil {
//				errValidation = append(errValidation, errPaymentRequestItems.Error())
//			}
//
//			errPaymentRequestItemsSeller := validation.ValidateStruct(&ppr.Items[i].Seller,
//				validation.Field(&ppr.Items[i].Seller.Title, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.FirstName, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.LastName, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Mobile, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Email, validation.Required),
//			)
//			if errPaymentRequestItemsSeller != nil {
//				errValidation = append(errValidation, errPaymentRequestItemsSeller.Error())
//			}
//
//			errPaymentRequestItemsSellerFinance := validation.ValidateStruct(&ppr.Items[i].Seller.Finance,
//				validation.Field(&ppr.Items[i].Seller.Finance.Iban, validation.Required),
//			)
//			if errPaymentRequestItemsSellerFinance != nil {
//				errValidation = append(errValidation, errPaymentRequestItemsSellerFinance.Error())
//			}
//
//			errPaymentRequestItemsSellerAddress := validation.ValidateStruct(&ppr.Items[i].Seller.Address,
//				validation.Field(&ppr.Items[i].Seller.Address.Title, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.Address, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.Phone, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.Country, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.State, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.City, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.ZipCode, validation.Required),
//			)
//			if errPaymentRequestItemsSellerAddress != nil {
//				errValidation = append(errValidation, errPaymentRequestItemsSellerAddress.Error())
//			}
//
//			errPaymentRequestItemsPrice := validation.ValidateStruct(&ppr.Items[i].Price,
//				validation.Field(&ppr.Items[i].Price.Unit, validation.Required),
//				validation.Field(&ppr.Items[i].Price.Total, validation.Required),
//				validation.Field(&ppr.Items[i].Price.Payable, validation.Required),
//				validation.Field(&ppr.Items[i].Price.Discount, validation.Required),
//				validation.Field(&ppr.Items[i].Price.SellerCommission, validation.Required),
//			)
//			if errPaymentRequestItemsPrice != nil {
//				errValidation = append(errValidation, errPaymentRequestItemsPrice.Error())
//			}
//
//			errPaymentRequestItemsShipment := validation.ValidateStruct(&ppr.Items[i].Shipment,
//				validation.Field(&ppr.Items[i].Shipment.ProviderName, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ReactionTime, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ShippingTime, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ReturnTime, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ShipmentDetail, validation.Required),
//			)
//			if errPaymentRequestItemsShipment != nil {
//				errValidation = append(errValidation, errPaymentRequestItemsShipment.Error())
//			}
//		}
//	}
//
//	res := strings.Join(errValidation, " ")
//	// return nil
//	if res == "" {
//		return nil
//	}
//	return errors.New(res)
//}
