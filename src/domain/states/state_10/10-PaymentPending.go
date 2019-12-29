package state_10

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	payment_service "gitlab.faza.io/order-project/order-service/infrastructure/services/payment"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"
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
			logger.Err("Process() => iFrame.Body().Content() not a order, orderId: %d, state: %s ", iFrame.Header().Value(string(frame.HeaderOrderId)), state.Name())
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Error", errors.New("Frame body invalid")).
				Send()
			return
		}

		grandTotal, err := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
		if err != nil {
			logger.Err("Process() => order.Invoice.GrandTotal.Amount invalid, amount: %s, orderId: %d, error: %s", order.Invoice.GrandTotal.Amount, order.OrderId, err)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Error", errors.New("Frame body invalid")).
				Send()
			return
		}

		var voucherAmount decimal.Decimal
		if order.Invoice.Voucher.Price != nil {
			voucherAmount, err = decimal.NewFromString(order.Invoice.Voucher.Price.Amount)
			if err != nil {
				logger.Err("Process() => order.Invoice.Voucher.Price.Amount invalid, price: %s, orderId: %d, error: %s", order.Invoice.Voucher.Price.Amount, order.OrderId, err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Error", errors.New("Voucher.Price invalid")).
					Send()
				return
			}
		}

		if grandTotal.IsZero() && order.Invoice.Voucher != nil &&
			(order.Invoice.Voucher.Percent > 0 || !voucherAmount.IsZero()) {
			order.OrderPayment = []entities.PaymentService{
				{
					PaymentRequest: &entities.PaymentRequest{
						Price:     nil,
						Gateway:   "",
						CreatedAt: time.Now().UTC(),
					},

					PaymentResult: &entities.PaymentResult{
						Result:      true,
						Reason:      "Invoice paid by voucher",
						PaymentId:   "",
						InvoiceId:   0,
						Price:       nil,
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

			// TODO check it voucher amount and if voucherSettlement failed can be cancel order
			var voucherAction *entities.Action
			iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				logger.Err("Process() => VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Error().Reason())
				voucherAction = &entities.Action{
					Name:      system_action.VoucherSettlement.ActionName(),
					Type:      "",
					UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
					UTP:       actions.System.ActionName(),
					Perm:      "",
					Priv:      "",
					Policy:    "",
					Result:    string(states.ActionFail),
					Reasons:   nil,
					Data:      nil,
					CreatedAt: time.Now().UTC(),
					Extended:  nil,
				}
			} else {
				if order.Invoice.Voucher.Percent > 0 {
					logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
				} else {
					logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
				}
				voucherAction = &entities.Action{
					Name:      system_action.VoucherSettlement.ActionName(),
					Type:      "",
					UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
					UTP:       actions.System.ActionName(),
					Perm:      "",
					Priv:      "",
					Policy:    "",
					Result:    string(states.ActionSuccess),
					Reasons:   nil,
					Data:      nil,
					CreatedAt: time.Now().UTC(),
					Extended:  nil,
				}
			}

			state.UpdateOrderAllSubPkg(ctx, order, voucherAction)
			orderUpdated, err := app.Globals.OrderRepository.Save(ctx, *order)
			if err != nil {
				errStr := fmt.Sprintf("Process() => OrderRepository.Save in %s state failed, order: %v, error: %s", state.Name(), order, err.Error())
				logger.Err(errStr)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, errStr, err).
					Send()
			} else {
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetData(order.OrderPayment[0].PaymentResponse.CallBackUrl).
					Send()
				successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
				state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(orderUpdated).Build())
			}
		} else {
			// TODO check it for IntPart()
			paymentRequest := payment_service.PaymentRequest{
				Amount:  int64(grandTotal.IntPart()),
				Gateway: order.Invoice.PaymentGateway,
				OrderId: order.OrderId,
			}

			order.OrderPayment = []entities.PaymentService{
				{
					PaymentRequest: &entities.PaymentRequest{
						Price: &entities.Money{
							Amount:   strconv.Itoa(int(paymentRequest.Amount)),
							Currency: "IRR",
						},
						Gateway:   paymentRequest.Gateway,
						CreatedAt: time.Now().UTC(),
					},
				},
			}

			iFuture := app.Globals.PaymentService.OrderPayment(ctx, paymentRequest)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
					Result:    false,
					Reason:    strconv.Itoa(int(futureData.Error().Code())),
					CreatedAt: time.Now().UTC(),
				}

				paymentAction := &entities.Action{
					Name:      system_action.PaymentFail.ActionName(),
					Type:      "",
					UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
					UTP:       actions.System.ActionName(),
					Perm:      "",
					Priv:      "",
					Policy:    "",
					Result:    string(states.ActionFail),
					Reasons:   nil,
					Data:      nil,
					CreatedAt: time.Now().UTC(),
					Extended:  nil,
				}

				state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
				orderUpdated, err := app.Globals.OrderRepository.Save(ctx, *order)
				if err != nil {
					logger.Err("Process() => Singletons.OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)
					future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
						SetError(future.InternalError, "Unknown Error", err).
						Send()
					return
				}

				logger.Err("Process() => OrderPayment.OrderPayment in orderPaymentState failed, orderId: %d, error: %s",
					order.OrderId, futureData.Error().Reason())

				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetErrorOf(futureData.Error()).Send()

				failAction := state.GetAction(system_action.PaymentFail.ActionName())
				state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(orderUpdated).Build())
				return
			} else {
				paymentResponse := futureData.Data().(payment_service.PaymentResponse)
				order.OrderPayment[0].PaymentResponse = &entities.PaymentResponse{
					Result:      true,
					CallBackUrl: paymentResponse.CallbackUrl,
					InvoiceId:   paymentResponse.InvoiceId,
					PaymentId:   paymentResponse.PaymentId,
					CreatedAt:   time.Now().UTC(),
				}

				_, err := app.Globals.OrderRepository.Save(ctx, *order)
				if err != nil {
					logger.Err("Process() => Singletons.OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)
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
		order, err := app.Globals.OrderRepository.FindById(ctx, iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64))
		if err != nil {
			logger.Err("Process() => Singletons.OrderRepository.Save failed, orderId: %d, paymentResult: %v, error: %s",
				iFrame.Header().Value(string(frame.HeaderOrderId)).(uint64),
				iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult), err)

			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetCapacity(1).SetError(future.NotFound, "OrderId Not Found", err).
				Send()
			return
		}

		ctx = context.WithValue(ctx, string(utils.CtxUserID), order.BuyerInfo.BuyerId)
		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetCapacity(1).Send()

		order.OrderPayment[0].PaymentResult = iFrame.Header().Value(string(frame.HeaderPaymentResult)).(*entities.PaymentResult)
		logger.Audit("Process() => Order Received in %s state, orderId: %d", state.Name(), order.OrderId)
		if order.OrderPayment[0].PaymentResult.Result == false {
			logger.Audit("Process() => PaymentResult failed, orderId: %d", order.OrderId)
			paymentAction := &entities.Action{
				Name:      system_action.PaymentFail.ActionName(),
				Type:      "",
				UId:       order.BuyerInfo.BuyerId,
				UTP:       actions.System.ActionName(),
				Perm:      "",
				Priv:      "",
				Policy:    "",
				Result:    string(states.ActionFail),
				Reasons:   nil,
				Data:      nil,
				CreatedAt: time.Now().UTC(),
			}

			state.UpdateOrderAllSubPkg(ctx, order, paymentAction)
			updatedOrder, err := app.Globals.OrderRepository.Save(ctx, *order)
			if err != nil {
				logger.Err("Process() => Singletons.OrderRepository.Save failed, orderId: %d, error: %s", order.OrderId, err)
			}
			failAction := state.GetAction(system_action.PaymentFail.ActionName())
			state.StatesMap()[failAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(updatedOrder).Build())
			return
		} else {
			var voucherAction *entities.Action
			var voucherAmount = 0
			if order.Invoice.Voucher.Price != nil {
				voucherAmount, err = strconv.Atoi(order.Invoice.Voucher.Price.Amount)
			}

			if order.Invoice.Voucher != nil && (order.Invoice.Voucher.Percent > 0 || voucherAmount > 0) {
				iFuture := app.Globals.VoucherService.VoucherSettlement(ctx, order.Invoice.Voucher.Code, order.OrderId, order.BuyerInfo.BuyerId)
				futureData := iFuture.Get()
				if futureData.Error() != nil {
					logger.Err("Process() => VoucherService.VoucherSettlement failed, orderId: %d, voucherCode: %s, error: %s", order.OrderId, order.Invoice.Voucher.Code, futureData.Error().Reason())
					voucherAction = &entities.Action{
						Name:      system_action.VoucherSettlement.ActionName(),
						Type:      "",
						UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
						UTP:       actions.System.ActionName(),
						Perm:      "",
						Priv:      "",
						Policy:    "",
						Result:    string(states.ActionFail),
						Reasons:   nil,
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					}
				} else {
					if order.Invoice.Voucher.Percent > 0 {
						logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
					} else {
						logger.Audit("Process() => Invoice paid by voucher order success, orderId: %d, voucherAmount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
					}

					voucherAction = &entities.Action{
						Name:      system_action.VoucherSettlement.ActionName(),
						Type:      "",
						UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
						UTP:       actions.System.ActionName(),
						Perm:      "",
						Priv:      "",
						Policy:    "",
						Result:    string(states.ActionSuccess),
						Reasons:   nil,
						Data:      nil,
						CreatedAt: time.Now().UTC(),
						Extended:  nil,
					}
				}

				if order.Invoice.Voucher.Percent > 0 {
					logger.Audit("Process() => VoucherSettlement success, orderId: %d, voucher Percent: %v, voucherCode: %s", order.OrderId, order.Invoice.Voucher.Percent, order.Invoice.Voucher.Code)
				} else {
					logger.Audit("Process() => VoucherSettlement success, orderId: %d, voucher Amount: %v, voucherCode: %s", order.OrderId, voucherAmount, order.Invoice.Voucher.Code)
				}
			}

			paymentAction := &entities.Action{
				Name:      system_action.PaymentSuccess.ActionName(),
				Type:      "",
				UId:       ctx.Value(string(utils.CtxUserID)).(uint64),
				UTP:       actions.System.ActionName(),
				Perm:      "",
				Priv:      "",
				Policy:    "",
				Result:    string(states.ActionSuccess),
				Reasons:   nil,
				Data:      nil,
				CreatedAt: time.Now().UTC(),
				Extended:  nil,
			}

			logger.Audit("Process() => PaymentResult success, orderId: %d", order.OrderId)
			state.UpdateOrderAllSubPkg(ctx, order, paymentAction, voucherAction)
			_, err = app.Globals.OrderRepository.Save(ctx, *order)
			if err != nil {
				errStr := fmt.Sprintf("Process() => OrderRepository.Save in %s state failed, orderId: %d, error: %s", state.Name(), order.OrderId, err.Error())
				logger.Err(errStr)
			}
			successAction := state.GetAction(system_action.PaymentSuccess.ActionName())
			state.StatesMap()[successAction].Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
		}
	}
}
