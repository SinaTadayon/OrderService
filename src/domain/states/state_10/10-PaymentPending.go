package state_10

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	payment_action "gitlab.faza.io/order-project/order-service/domain/actions/payment"
	voucher_action "gitlab.faza.io/order-project/order-service/domain/actions/voucher"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"strconv"
	"time"
)

const (
	stepName  string = "Payment_Pending"
	stepIndex int    = 10
)

type paymentPendingState struct {
	*states.BaseStateImpl
}

func New(childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentPendingState{states.NewBaseStep(stepName, stepIndex, childes, parents, actionStateMap)}
}

func NewOf(name string, index int, childes, parents []states.IState, actionStateMap map[actions.IAction]states.IState) states.IState {
	return &paymentPendingState{states.NewBaseStep(name, index, childes, parents, actionStateMap)}
}

func NewFrom(base *states.BaseStateImpl) states.IState {
	return &paymentPendingState{base}
}

func NewValueOf(base *states.BaseStateImpl, params ...interface{}) states.IState {
	panic("implementation required")
}

func (state paymentPendingState) Process(ctx context.Context, iFrame frame.IFrame) {

	if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) && iFrame.Body().Content() != nil {
		order, ok := iFrame.Body().Content().(*entities.Order)
		if !ok {
			logger.Err("iFrame.Body().Content() not a order, orderId: %d, %s state ", iFrame.Header().Value(string(frame.HeaderOrderId)), state.Name())
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Error", errors.New("Frame body invalid")).
				Send()
			return
		}

		if order.Invoice.GrandTotal == 0 && order.Invoice.Voucher.Amount > 0 {
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

			paymentAction := &entities.Action{
				Name:      payment_action.BuyerPaymentPendingRequest.ActionName(),
				Type:      actions.Payment.ActionName(),
				Data:      nil,
				Result:    string(states.ActionSuccess),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}

			// TODO check it voucher amount and if voucherSettlement failed can be cancel order
			var voucherAction *entities.Action
			iFuture := global.Singletons.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				logger.Err("VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Error().Reason())
				voucherAction = &entities.Action{
					Name:      voucher_action.Settlement.ActionName(),
					Type:      actions.Voucher.ActionName(),
					Data:      nil,
					Result:    string(states.ActionFail),
					Reasons:   nil,
					CreatedAt: time.Now().UTC(),
				}
			} else {
				logger.Audit("Invoice paid by voucher order success, orderId: %d, voucherAmount: %d, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Amount, order.Invoice.Voucher.Code)
				voucherAction = &entities.Action{
					Name:      voucher_action.Settlement.ActionName(),
					Type:      actions.Voucher.ActionName(),
					Data:      nil,
					Result:    string(states.ActionSuccess),
					Reasons:   nil,
					CreatedAt: time.Now().UTC(),
				}
			}

			state.UpdateOrderAllSubPkg(ctx, order, paymentAction, voucherAction)
			orderUpdated, err := global.Singletons.OrderRepository.Save(ctx, *order)
			if err != nil {
				errStr := fmt.Sprintf("OrderRepository.Save in %s state failed, order: %v, error: %s", state.Name(), order, err.Error())
				logger.Err(errStr)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, errStr, err).
					Send()
			} else {
				successAction := state.GetAction(payment_action.Success.ActionName())
				state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(orderUpdated).Build())
			}
		} else {
			paymentRequest := payment_service.PaymentRequest{
				Amount:   int64(order.Invoice.GrandTotal),
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

			iFuture := global.Singletons.PaymentService.OrderPayment(ctx, paymentRequest)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
					Result:    false,
					Reason:    strconv.Itoa(int(futureData.Error().Code())),
					CreatedAt: time.Now().UTC(),
				}

				paymentAction := &entities.Action{
					Name:      payment_action.BuyerPaymentPendingRequest.ActionName(),
					Type:      actions.Payment.ActionName(),
					Data:      nil,
					Result:    string(states.ActionFail),
					Reasons:   nil,
					CreatedAt: time.Now().UTC(),
				}

				state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
				orderUpdated, err := global.Singletons.OrderRepository.Save(ctx, *order)
				if err != nil {
					logger.Err("Singletons.OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.InternalError, "Unknown Error", err).
						Send()
					return
				}

				logger.Err("PaymentService.OrderPayment in orderPaymentState failed, orderId: %d, error: %s",
					order.OrderId, futureData.Error().Reason())

				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetErrorOf(futureData.Error()).Send()

				failAction := state.GetAction(payment_action.Fail.ActionName())
				state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(orderUpdated).Build())
				return
			} else {
				paymentResponse := futureData.Data().(payment_service.PaymentResponse)
				order.PaymentService[0].PaymentResponse = &entities.PaymentResponse{
					Result:      true,
					CallBackUrl: paymentResponse.CallbackUrl,
					InvoiceId:   paymentResponse.InvoiceId,
					PaymentId:   paymentResponse.PaymentId,
					CreatedAt:   time.Now().UTC(),
				}

				paymentAction := &entities.Action{
					Name:      payment_action.BuyerPaymentPendingRequest.ActionName(),
					Type:      actions.Payment.ActionName(),
					Data:      nil,
					Result:    string(states.ActionSuccess),
					Reasons:   nil,
					CreatedAt: time.Now().UTC(),
				}

				state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
				_, err := global.Singletons.OrderRepository.Save(ctx, *order)
				if err != nil {
					logger.Err("Singletons.OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.InternalError, "Unknown Error", err).
						Send()
					return
				}

				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetData(paymentResponse.CallbackUrl).
					Send()
				return
			}
		}
	} else if iFrame.Header().KeyExists(string(frame.HeaderOrderId)) &&
		iFrame.Header().KeyExists(string(frame.HeaderPaymentResult)) {
		order, err := global.Singletons.OrderRepository.FindById(ctx, iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64))
		if err != nil {
			logger.Err("Singletons.OrderRepository.Save failed, orderId: %d, paymentResult: %v, error: %s",
				iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64),
				iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult), err)
			return
		}

		order.PaymentService[0].PaymentResult = iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult)
		logger.Audit("Order Received in %s state, orderId: %d", state.Name(), order.OrderId)
		if order.PaymentService[0].PaymentResult.Result == false {
			logger.Audit("PaymentResult failed, orderId: %d", order.OrderId)
			paymentAction := &entities.Action{
				Name:      payment_action.BuyerPaymentPendingResponse.ActionName(),
				Type:      actions.Payment.ActionName(),
				Data:      nil,
				Result:    string(states.ActionFail),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}

			state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
			if _, err := global.Singletons.OrderRepository.Save(ctx, *order); err != nil {
				logger.Err("Singletons.OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)
			}
			failAction := state.GetAction(payment_action.Fail.ActionName())
			state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
			return
		} else {
			var voucherAction *entities.Action
			if order.Invoice.Voucher.Amount > 0 {
				iFuture := global.Singletons.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
				futureData := iFuture.Get()
				if futureData.Error() != nil {
					logger.Err("VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Error().Reason())
					voucherAction = &entities.Action{
						Name:      voucher_action.Settlement.ActionName(),
						Type:      actions.Voucher.ActionName(),
						Data:      nil,
						Result:    string(states.ActionFail),
						Reasons:   nil,
						CreatedAt: time.Now().UTC(),
					}
				} else {
					logger.Audit("Invoice paid by voucher order success, orderId: %d, voucherAmount: %d, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Amount, order.Invoice.Voucher.Code)
					voucherAction = &entities.Action{
						Name:      voucher_action.Settlement.ActionName(),
						Type:      actions.Voucher.ActionName(),
						Data:      nil,
						Result:    string(states.ActionSuccess),
						Reasons:   nil,
						CreatedAt: time.Now().UTC(),
					}
				}

				logger.Audit("VoucherSettlement success, orderId: %d, voucherAmount: %d, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Amount, order.Invoice.Voucher.Code)
			}

			paymentAction := &entities.Action{
				Name:      payment_action.BuyerPaymentPendingResponse.ActionName(),
				Type:      actions.Payment.ActionName(),
				Data:      nil,
				Result:    string(states.ActionSuccess),
				Reasons:   nil,
				CreatedAt: time.Now().UTC(),
			}

			logger.Audit("PaymentResult success, orderId: %d", order.OrderId)
			state.UpdateOrderAllSubPkg(ctx, order, paymentAction, voucherAction)
			_, err = global.Singletons.OrderRepository.Save(ctx, *order)
			if err != nil {
				errStr := fmt.Sprintf("OrderRepository.Save in %s state failed, orderId: %d, error: %s", state.Name(), order.OrderId, err.Error())
				logger.Err(errStr)
			}
			successAction := state.GetAction(payment_action.Success.ActionName())
			state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
		}
	}
}
