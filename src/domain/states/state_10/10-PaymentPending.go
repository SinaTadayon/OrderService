package state_10

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	message "gitlab.faza.io/protos/order"
	"strconv"
	"time"
)

const (
	stepName  string = "Payment_Pending"
	stepIndex int    = 10

	PaymentCallbackUrlRequest = "PaymentCallbackUrlRequest"
	OrderPayment              = "OrderPayment"
	StockReleased             = "StockReleased"
)

type paymentPendingStep struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &paymentPendingStep{states.NewBaseStep(stepName, stepIndex, childes, parents, states)}
}

func NewOf(name string, index int, childes, parents []states.IState, states ...states_old.IState) states.IState {
	return &paymentPendingStep{states.NewBaseStep(name, index, childes, parents, states)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &paymentPendingStep{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (paymentPending paymentPendingStep) ProcessMessage(ctx context.Context, request *message.MessageRequest) future.IFuture {
	panic("implementation required")
}

func (paymentPending paymentPendingStep) ProcessOrder(ctx context.Context, order entities.Order, itemsId []uint64, param interface{}) future.IFuture {

	//orderPaymentState, ok := paymentPending.StatesMap()[0].(launcher_state.ILauncherState)
	//if ok != true || orderPaymentState.ActiveType() != actives.OrderPaymentAction {
	//	logger.Err("orderPayment state doesn't exist in index 0 of statesMap, order: %v", order)
	//	returnChannel := make(chan future.IDataFuture, 1)
	//	returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.InternalError, Reason:"Unknown Error"}}
	//	defer close(returnChannel)
	//	return future.NewFuture(returnChannel, 1, 1)
	//}

	paymentAction := param.(string)

	if paymentAction == PaymentCallbackUrlRequest {
		logger.Audit("Order Received in %s step, orderId: %d, Actions: %s", paymentPending.Name(), order.OrderId, PaymentCallbackUrlRequest)

		// handle amount == 0 because of full voucher
		if order.Invoice.Total == 0 && order.Invoice.Voucher.Amount > 0 {
			order.PaymentService = []entities.PaymentService{
				{
					PaymentRequest: &entities.PaymentRequest{
						Amount:    0,
						Currency:  "IRR",
						Gateway:   "Assanpardakht",
						CreatedAt: time.Now().UTC(),
					},

					PaymentResult: &entities.PaymentResult{
						Result:      true,
						Reason:      "Invoice paid by voucher",
						PaymentId:   "",
						InvoiceId:   0,
						Amount:      0,
						ReqBody:     "",
						ResBody:     "",
						CardNumMask: "",
						CreatedAt:   time.Now().UTC(),
					},

					PaymentResponse: &entities.PaymentResponse{
						Result:      true,
						CallBackUrl: "http://staging.faza.io/callback-success?orderid=" + strconv.Itoa(int(order.OrderId)),
						InvoiceId:   0,
						PaymentId:   "",
						CreatedAt:   time.Now().UTC(),
					},
				},
			}

			iPromise := global.Singletons.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
			futureData := iPromise.Get()
			if futureData.Ex != nil {
				logger.Err("VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Ex.Error())
				returnChannel := make(chan future.IDataFuture, 1)
				defer close(returnChannel)
				returnChannel <- future.IDataFuture{Data: nil, Ex: futureData.Ex}
				return future.NewFuture(returnChannel, 1, 1)
			}

			logger.Audit("Invoice paid by voucher order success, orderId: %d, voucherAmount: %d, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Amount, order.Invoice.Voucher.Code)
			paymentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, true)
			paymentPending.updateOrderItemsProgress(ctx, &order, nil, OrderPayment, true, states.OrderInProgressStatus)
			if err := paymentPending.persistOrder(ctx, &order); err != nil {
			}

			go func() {
				paymentPending.Childes()[1].ProcessOrder(ctx, order, nil, nil)
			}()

			returnChannel := make(chan future.IDataFuture, 1)
			defer close(returnChannel)
			returnChannel <- future.IDataFuture{Data: order.PaymentService[0].PaymentResponse.CallBackUrl, Ex: nil}
			return future.NewFuture(returnChannel, 1, 1)
		}

		paymentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderNewStatus, false)
		//return orderPaymentState.ActionLauncher(ctx, order, nil, nil)

		paymentRequest := payment_service.PaymentRequest{
			Amount:   int64(order.Invoice.Total),
			Gateway:  order.Invoice.PaymentGateway,
			Currency: order.Invoice.Currency,
			OrderId:  order.OrderId,
		}

		order.PaymentService = []entities.PaymentService{
			{
				PaymentRequest: &entities.PaymentRequest{
					Amount:    uint64(paymentRequest.Amount),
					Currency:  paymentRequest.Currency,
					Gateway:   paymentRequest.Gateway,
					CreatedAt: time.Now().UTC(),
				},
			},
		}

		iPromise := global.Singletons.PaymentService.OrderPayment(ctx, paymentRequest)
		futureData := iPromise.Get()
		if futureData == nil {
			order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
				Result:    false,
				Reason:    "PaymentService.OrderPayment in orderPaymentState failed",
				CreatedAt: time.Now().UTC(),
			}

			paymentPending.UpdateAllOrderStatus(ctx, &order, nil, states.OrderClosedStatus, true)
			paymentPending.updateOrderItemsProgress(ctx, &order, nil, PaymentCallbackUrlRequest, false, states.OrderClosedStatus)
			paymentPending.releasedStock(ctx, &order)
			if err := paymentPending.persistOrder(ctx, &order); err != nil {
			}

			logger.Err("PaymentService future channel has been closed, orderId: %d", order.OrderId)
			returnChannel := make(chan future.IDataFuture, 1)
			defer close(returnChannel)
			returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
			return future.NewFuture(returnChannel, 1, 1)
		}

		if futureData.Ex != nil {
			order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
				Result:    false,
				Reason:    futureData.Ex.Error(),
				CreatedAt: time.Now().UTC(),
			}

			paymentPending.UpdateAllOrderStatus(ctx, &order, nil, states.OrderClosedStatus, true)
			paymentPending.updateOrderItemsProgress(ctx, &order, nil, PaymentCallbackUrlRequest, false, states.OrderClosedStatus)
			paymentPending.releasedStock(ctx, &order)
			if err := paymentPending.persistOrder(ctx, &order); err != nil {
			}
			logger.Err("PaymentService.OrderPayment in orderPaymentState failed, orderId: %d, error: %s", order.OrderId, futureData.Ex.Error())
			returnChannel := make(chan future.IDataFuture, 1)
			defer close(returnChannel)
			returnChannel <- future.IDataFuture{Data: nil, Ex: futureData.Ex}
			return future.NewFuture(returnChannel, 1, 1)
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

		paymentPending.updateOrderItemsProgress(ctx, &order, nil, PaymentCallbackUrlRequest, true, states.OrderNewStatus)
		if err := paymentPending.persistOrder(ctx, &order); err != nil {
		}

		returnChannel := make(chan future.IDataFuture, 1)
		defer close(returnChannel)
		returnChannel <- future.IDataFuture{Data: paymentResponse.CallbackUrl, Ex: nil}
		return future.NewFuture(returnChannel, 1, 1)

	} else if paymentAction == OrderPayment {
		logger.Audit("Order Received in %s step, orderId: %d, Actions: %s", paymentPending.Name(), order.OrderId, OrderPayment)
		if order.PaymentService[0].PaymentResult.Result == false {
			logger.Audit("PaymentResult of order failed, orderId: %d", order.OrderId)
			paymentPending.updateOrderItemsProgress(ctx, &order, nil, OrderPayment, false, states.OrderClosedStatus)
			paymentPending.UpdateAllOrderStatus(ctx, &order, nil, states.OrderClosedStatus, true)
			paymentPending.releasedStock(ctx, &order)
			if err := paymentPending.persistOrder(ctx, &order); err != nil {
			}
			return paymentPending.Childes()[0].ProcessOrder(ctx, order, nil, nil)
		}

		if order.Invoice.Voucher.Amount > 0 {
			iPromise := global.Singletons.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
			futureData := iPromise.Get()
			if futureData.Ex != nil {
				logger.Err("VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Ex.Error())
				returnChannel := make(chan future.IDataFuture, 1)
				defer close(returnChannel)
				returnChannel <- future.IDataFuture{Data: nil, Ex: futureData.Ex}
				return future.NewFuture(returnChannel, 1, 1)
			}
			logger.Audit("VoucherSettlement success, orderId: %d, voucherAmount: %d, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Amount, order.Invoice.Voucher.Code)
		}

		logger.Audit("PaymentResult of order success, orderId: %d", order.OrderId)
		paymentPending.UpdateAllOrderStatus(ctx, &order, itemsId, states.OrderInProgressStatus, true)
		paymentPending.updateOrderItemsProgress(ctx, &order, nil, OrderPayment, true, states.OrderInProgressStatus)
		if err := paymentPending.persistOrder(ctx, &order); err != nil {
		}
		return paymentPending.Childes()[1].ProcessOrder(ctx, order, nil, nil)
	}

	logger.Err("%s step received invalid action, orderId: %d, action: %s", paymentPending.Name(), order.OrderId, paymentAction)
	returnChannel := make(chan future.IDataFuture, 1)
	defer close(returnChannel)
	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
	return future.NewFuture(returnChannel, 1, 1)
	//orderPayment.persistOrderState(ctx, &order, itemsId, order_payment_action.OrderPaymentAction,
	//	true, "", &paymentResponse)
	//return paymentState.ActionListener(ctx, activeEvent, nil)
}

func (paymentPending paymentPendingStep) releasedStock(ctx context.Context, order *entities.Order) {
	iPromise := global.Singletons.StockService.BatchStockActions(ctx, nil, StockReleased)
	futureData := iPromise.Get()
	if futureData == nil {
		paymentPending.updateOrderItemsProgress(ctx, order, nil, StockReleased, false, states.OrderClosedStatus)
		logger.Err("StockService future channel has been closed, step: %s, orderId: %d", paymentPending.Name(), order.OrderId)
		return
	}

	if futureData.Ex != nil {
		paymentPending.updateOrderItemsProgress(ctx, order, nil, StockReleased, false, states.OrderClosedStatus)
		logger.Err("Reserved stock from stockService failed, step: %s, orderId: %d, error: %s", paymentPending.Name(), order.OrderId, futureData.Ex.Error())
		return
	}

	paymentPending.updateOrderItemsProgress(ctx, order, nil, StockReleased, true, states.OrderClosedStatus)
	logger.Audit("Reserved stock from stockService success, step: %s, orderId: %d", paymentPending.Name(), order.OrderId)
}

func (paymentPending paymentPendingStep) persistOrder(ctx context.Context, order *entities.Order) error {
	_, err := global.Singletons.OrderRepository.Save(*order)
	if err != nil {
		logger.Err("OrderRepository.Save in %s step failed, orderId: %d, error: %s", paymentPending.Name(), order.OrderId, err.Error())
	}

	return err
}

func (paymentPending paymentPendingStep) updateOrderItemsProgress(ctx context.Context, order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					paymentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
					findFlag = true
					break
				}
			}

			if !findFlag {
				logger.Err("%s received itemId %d not exist in order, orderId: %d", paymentPending.Name(), id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			paymentPending.doUpdateOrderItemsProgress(ctx, order, i, action, result, itemStatus)
		}
	}
}

func (paymentPending paymentPendingStep) doUpdateOrderItemsProgress(ctx context.Context, order *entities.Order, index int,
	actionName string, result bool, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	action := entities.Action{
		Name:      actionName,
		Result:    result,
		CreatedAt: order.Items[index].UpdatedAt,
	}

	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
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
//		validation.Field(&ppr.Buyer.Address.Status, validation.Required),
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
//	errPaymentCallbackUrlRequestAmount := validation.ValidateStruct(&ppr.Invoice,
//		validation.Field(&ppr.Invoice.total, validation.Required),
//		validation.Field(&ppr.Invoice.Discount, validation.Required),
//		validation.Field(&ppr.Invoice.Subtotal, validation.Required),
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
//				validation.Field(&ppr.Items[i].Category, validation.Required),
//				validation.Field(&ppr.Items[i].Brand, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItems != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItems.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsSeller := validation.ValidateStruct(&ppr.Items[i].SellerInfo,
//				validation.Field(&ppr.Items[i].SellerInfo.Title, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.FirstName, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.LastName, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Mobile, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Email, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsSeller != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsSeller.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsSellerFinance := validation.ValidateStruct(&ppr.Items[i].SellerInfo.Finance,
//				validation.Field(&ppr.Items[i].SellerInfo.Finance.Iban, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsSellerFinance != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsSellerFinance.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsSellerAddress := validation.ValidateStruct(&ppr.Items[i].SellerInfo.Address,
//				validation.Field(&ppr.Items[i].SellerInfo.Address.Title, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Address.Address, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Address.Phone, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Address.Country, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Address.Status, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Address.City, validation.Required),
//				validation.Field(&ppr.Items[i].SellerInfo.Address.ZipCode, validation.Required),
//			)
//			if errPaymentCallbackUrlRequestItemsSellerAddress != nil {
//				errValidation = append(errValidation, errPaymentCallbackUrlRequestItemsSellerAddress.Error())
//			}
//
//			errPaymentCallbackUrlRequestItemsPrice := validation.ValidateStruct(&ppr.Items[i].ItemInvoice,
//				validation.Field(&ppr.Items[i].ItemInvoice.Unit, validation.Required),
//				validation.Field(&ppr.Items[i].ItemInvoice.total, validation.Required),
//				validation.Field(&ppr.Items[i].ItemInvoice.Subtotal, validation.Required),
//				validation.Field(&ppr.Items[i].ItemInvoice.Discount, validation.Required),
//				validation.Field(&ppr.Items[i].ItemInvoice.SellerCommission, validation.Required),
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
//				validation.Field(&ppr.Items[i].Shipment.ShippingDetail, validation.Required),
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
