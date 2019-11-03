package payment_pending_step

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	message "gitlab.faza.io/protos/order"
	"time"
)

const (
	stepName string 	= "Payment_Pending"
	stepIndex int		= 10
	
	PaymentCallbackUrlRequest = "PaymentCallbackUrlRequest"
	OrderPayment              = "OrderPayment"
	StockReleased             = "StockReleased"
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

func (paymentPending paymentPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) promise.IPromise {
	panic("implementation required")
}

func (paymentPending paymentPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []string, param interface{}) promise.IPromise {

	//orderPaymentState, ok := paymentPending.StatesMap()[0].(launcher_state.ILauncherState)
	//if ok != true || orderPaymentState.ActiveType() != actives.OrderPaymentAction {
	//	logger.Err("orderPayment state doesn't exist in index 0 of statesMap, order: %v", order)
	//	returnChannel := make(chan promise.FutureData, 1)
	//	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	//	defer close(returnChannel)
	//	return promise.NewPromise(returnChannel, 1, 1)
	//}


	paymentAction := param.(string)

	if paymentAction == PaymentCallbackUrlRequest {
		logger.Audit("Order Received in %s step, orderId: %s, Action: %s", paymentPending.Name(), order.OrderId, PaymentCallbackUrlRequest)
		paymentPending.UpdateOrderStep(ctx, &order, itemsId, "NEW", false)
		//return orderPaymentState.ActionLauncher(ctx, order, nil, nil)

		paymentRequest := payment_service.PaymentRequest{
			Amount:   int64(order.Amount.Total),
			Gateway:  order.Amount.PaymentOption,
			Currency: order.Amount.Currency,
			OrderId:  order.OrderId,
		}

		order.PaymentService = []entities.PaymentService{
			{
				PaymentRequest: &entities.PaymentRequest{
					Amount:      uint64(paymentRequest.Amount),
					Currency:    paymentRequest.Currency,
					Gateway:     paymentRequest.Gateway,
					CreatedAt:   time.Now().UTC(),
				},
			},
		}

		iPromise := global.Singletons.PaymentService.OrderPayment(ctx, paymentRequest)
		futureData := iPromise.Data()
		if futureData == nil {
			order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
				Result:      false,
				Reason:      "PaymentService.OrderPayment in orderPaymentState failed",
				CreatedAt:   time.Now().UTC(),
			}

			paymentPending.UpdateOrderStep(ctx, &order, nil, "CLOSED", true)
			paymentPending.updateOrderItemsProgress(ctx, &order, nil, PaymentCallbackUrlRequest, false)
			paymentPending.releasedStock(ctx, &order)
			if err := paymentPending.persistOrder(ctx, &order); err != nil {}

			logger.Err("PaymentService promise channel has been closed, order: %v", order)
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		if futureData.Ex != nil {
			order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
				Result:      false,
				Reason:      futureData.Ex.Error(),
				CreatedAt:   time.Now().UTC(),
			}

			paymentPending.UpdateOrderStep(ctx, &order, nil, "CLOSED", true)
			paymentPending.updateOrderItemsProgress(ctx, &order, nil, PaymentCallbackUrlRequest, false)
			paymentPending.releasedStock(ctx, &order)
			if err := paymentPending.persistOrder(ctx, &order); err != nil {}
			logger.Err("PaymentService.OrderPayment in orderPaymentState failed, order: %v, error", order, futureData.Ex.Error())
			returnChannel := make(chan promise.FutureData, 1)
			defer close(returnChannel)
			returnChannel <- promise.FutureData{Data:nil, Ex:futureData.Ex}
			return promise.NewPromise(returnChannel, 1, 1)
		}

		paymentResponse := futureData.Data.(payment_service.PaymentResponse)
		//activeEvent := active_event.NewActiveEvent(order, itemsId, actives.OrderPaymentAction, order_payment_action.NewOf(order_payment_action.OrderPaymentAction),
		//	paymentResponse, time.Now())

		order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
			Result:      true,
			CallBackUrl: paymentResponse.CallbackUrl,
			InvoiceId:   paymentResponse.InvoiceId,
			PaymentId:   paymentResponse.PaymentId,
			CreatedAt:   time.Now().UTC(),
		}

		paymentPending.updateOrderItemsProgress(ctx, &order, nil, PaymentCallbackUrlRequest, true)
		if err := paymentPending.persistOrder(ctx, &order); err != nil {}

		returnChannel := make(chan promise.FutureData, 1)
		defer close(returnChannel)
		returnChannel <- promise.FutureData{Data:paymentResponse, Ex:nil}
		return promise.NewPromise(returnChannel, 1, 1)

	} else if paymentAction == OrderPayment {
		logger.Audit("Order Received in %s step, orderId: %s, Action: %s", paymentPending.Name(), order.OrderId, OrderPayment)
		if order.PaymentService[0].PaymentResult.Result == false {
			logger.Audit("PaymentResult of order failed, order: %v", order)
			paymentPending.updateOrderItemsProgress(ctx, &order, nil, OrderPayment, false)
			paymentPending.UpdateOrderStep(ctx, &order, nil, "CLOSED", true)
			paymentPending.releasedStock(ctx, &order)
			if err := paymentPending.persistOrder(ctx, &order); err != nil {}
			return paymentPending.Childes()[0].ProcessOrder(ctx, order, nil, nil)
		}

		logger.Audit("PaymentResult of order success, order: %v", order)
		paymentPending.updateOrderItemsProgress(ctx, &order, nil, OrderPayment, true)
		return paymentPending.Childes()[1].ProcessOrder(ctx, order, nil, nil)
	}

	logger.Err("%s step received invalid action, order: %v, action: %s", paymentPending.Name(), order, paymentAction)
	returnChannel := make(chan promise.FutureData, 1)
	defer close(returnChannel)
	returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.InternalError, Reason:"Unknown Error"}}
	return promise.NewPromise(returnChannel, 1, 1)
	//orderPayment.persistOrderState(ctx, &order, itemsId, order_payment_action.OrderPaymentAction,
	//	true, "", &paymentResponse)
	//return paymentState.ActionListener(ctx, activeEvent, nil)
}

func (paymentPending paymentPendingStep) releasedStock(ctx context.Context, order *entities.Order) {
	itemStocks := make(map[string]int, len(order.Items))
	for i:= 0; i < len(order.Items); i++ {
		if value, ok := itemStocks[order.Items[i].InventoryId]; ok {
			itemStocks[order.Items[i].InventoryId] = value + 1
		} else {
			itemStocks[order.Items[i].InventoryId] = 1
		}
	}

	iPromise := global.Singletons.StockService.BatchStockActions(ctx, itemStocks, StockReleased)
	futureData := iPromise.Data()
	if futureData == nil {
		paymentPending.updateOrderItemsProgress(ctx, order, nil, StockReleased, false)
		logger.Err("StockService promise channel has been closed, step: %s, order: %v",  paymentPending.Name(), order)
		return
	}

	if futureData.Ex != nil {
		paymentPending.updateOrderItemsProgress(ctx, order, nil, StockReleased, false)
		logger.Err("Reserved stock from stockService failed, step: %s, order: %v, error: %s", paymentPending.Name(), order, futureData.Ex.Error())
		return
	}

	paymentPending.updateOrderItemsProgress(ctx, order, nil, StockReleased, true)
	logger.Audit("Reserved stock from stockService success, step: %s, order: %v", paymentPending.Name(), order)
}

func (paymentPending paymentPendingStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_ , err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, order: %v, error: %s", paymentPending.Name(), order, err.Error())
	}

	return err
}

func (paymentPending paymentPendingStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []string, action string, result bool) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					paymentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result)
					findFlag = true
					break
				}
			}

			if !findFlag {
				logger.Err("%s received itemId %s not exist in order, orderId: %v", paymentPending.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			paymentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result)
		}
	}
}

func (paymentPending paymentPendingStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool) {

	order.Items[index].Status = paymentPending.Name()
	order.Items[index].UpdatedAt = time.Now().UTC()

	if order.Items[index].Progress.ActionHistory == nil || len(order.Items[index].Progress.ActionHistory) == 0 {
		order.Items[index].Progress.ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.ActionHistory = append(order.Items[index].Progress.ActionHistory, action)
}



//func (ppr *PaymentPendingRequest) validate() error {
//	var errValidation []string
//	// Validate order number
//	errPaymentCallbackUrlRequest := validation.ValidateStruct(ppr,
//		validation.Field(&ppr.OrderNumber, validation.Required, validation.Length(5, 250)),
//	)
//	if errPaymentCallbackUrlRequest != nil {
//		errValidation = append(errValidation, errPaymentCallbackUrlRequest.Error())
//	}
//
//	// Validate Buyer
//	errPaymentCallbackUrlRequestBuyer := validation.ValidateStruct(&ppr.Buyer,
//		validation.Field(&ppr.Buyer.FirstName, validation.Required),
//		validation.Field(&ppr.Buyer.LastName, validation.Required),
//		validation.Field(&ppr.Buyer.Email, validation.Required, is.Email),
//		validation.Field(&ppr.Buyer.NationalId, validation.Required, validation.Length(10, 10)),
//		validation.Field(&ppr.Buyer.Mobile, validation.Required),
//	)
//	if errPaymentCallbackUrlRequestBuyer != nil {
//		errValidation = append(errValidation, errPaymentCallbackUrlRequestBuyer.Error())
//	}
//
//	// Validate Buyer finance
//	errPaymentCallbackUrlRequestBuyerFinance := validation.ValidateStruct(&ppr.Buyer.Finance,
//		validation.Field(&ppr.Buyer.Finance.Iban, validation.Required, validation.Length(26, 26)),
//	)
//	if errPaymentCallbackUrlRequestBuyerFinance != nil {
//		errValidation = append(errValidation, errPaymentCallbackUrlRequestBuyerFinance.Error())
//	}
//
//	// Validate Buyer address
//	errPaymentCallbackUrlRequestBuyerAddress := validation.ValidateStruct(&ppr.Buyer.Address,
//		validation.Field(&ppr.Buyer.Address.Address, validation.Required),
//		validation.Field(&ppr.Buyer.Address.State, validation.Required),
//		validation.Field(&ppr.Buyer.Address.City, validation.Required),
//		validation.Field(&ppr.Buyer.Address.Country, validation.Required),
//		validation.Field(&ppr.Buyer.Address.ZipCode, validation.Required),
//		validation.Field(&ppr.Buyer.Address.Phone, validation.Required),
//	)
//	if errPaymentCallbackUrlRequestBuyerAddress != nil {
//		errValidation = append(errValidation, errPaymentCallbackUrlRequestBuyerAddress.Error())
//	}
//
//	// Validate amount
//	errPaymentCallbackUrlRequestAmount := validation.ValidateStruct(&ppr.Amount,
//		validation.Field(&ppr.Amount.Total, validation.Required),
//		validation.Field(&ppr.Amount.Discount, validation.Required),
//		validation.Field(&ppr.Amount.Payable, validation.Required),
//	)
//	if errPaymentCallbackUrlRequestAmount != nil {
//		errValidation = append(errValidation, errPaymentCallbackUrlRequestAmount.Error())
//	}
//
//	if len(ppr.Items) != 0 {
//		for i := range ppr.Items {
//			// Validate amount
//			errPaymentCallbackUrlRequestItems := validation.ValidateStruct(&ppr.Items[i],
//				validation.Field(&ppr.Items[i].Sku, validation.Required),
//				validation.Field(&ppr.Items[i].Quantity, validation.Required),
//				validation.Field(&ppr.Items[i].Title, validation.Required),
//				validation.Field(&ppr.Items[i].Categories, validation.Required),
//				validation.Field(&ppr.Items[i].Brand, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItems != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItems.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsSeller := validation.ValidateStruct(&ppr.Items[i].Seller,
//				validation.Field(&ppr.Items[i].Seller.Title, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.FirstName, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.LastName, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Mobile, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Email, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsSeller != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsSeller.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsSellerFinance := validation.ValidateStruct(&ppr.Items[i].Seller.Finance,
//				validation.Field(&ppr.Items[i].Seller.Finance.Iban, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsSellerFinance != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsSellerFinance.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsSellerAddress := validation.ValidateStruct(&ppr.Items[i].Seller.Address,
//				validation.Field(&ppr.Items[i].Seller.Address.Title, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.Address, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.Phone, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.Country, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.State, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.City, validation.Required),
//				validation.Field(&ppr.Items[i].Seller.Address.ZipCode, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsSellerAddress != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsSellerAddress.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsPrice := validation.ValidateStruct(&ppr.Items[i].Price,
//				validation.Field(&ppr.Items[i].Price.Unit, validation.Required),
//				validation.Field(&ppr.Items[i].Price.Total, validation.Required),
//				validation.Field(&ppr.Items[i].Price.Payable, validation.Required),
//				validation.Field(&ppr.Items[i].Price.Discount, validation.Required),
//				validation.Field(&ppr.Items[i].Price.SellerCommission, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsPrice != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsPrice.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsShipment := validation.ValidateStruct(&ppr.Items[i].Shipment,
//				validation.Field(&ppr.Items[i].Shipment.ProviderName, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ReactionTime, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ShippingTime, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ReturnTime, validation.Required),
//				validation.Field(&ppr.Items[i].Shipment.ShipmentDetail, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsShipment != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsShipment.Error())
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
