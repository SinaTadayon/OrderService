package domain

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/buyer"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/reports"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/domain/states/state_01"
	"gitlab.faza.io/order-project/order-service/domain/states/state_10"
	"gitlab.faza.io/order-project/order-service/domain/states/state_11"
	"gitlab.faza.io/order-project/order-service/domain/states/state_12"
	"gitlab.faza.io/order-project/order-service/domain/states/state_13"
	"gitlab.faza.io/order-project/order-service/domain/states/state_14"
	"gitlab.faza.io/order-project/order-service/domain/states/state_15"
	"gitlab.faza.io/order-project/order-service/domain/states/state_20"
	"gitlab.faza.io/order-project/order-service/domain/states/state_21"
	"gitlab.faza.io/order-project/order-service/domain/states/state_22"
	"gitlab.faza.io/order-project/order-service/domain/states/state_30"
	"gitlab.faza.io/order-project/order-service/domain/states/state_31"
	"gitlab.faza.io/order-project/order-service/domain/states/state_32"
	"gitlab.faza.io/order-project/order-service/domain/states/state_33"
	"gitlab.faza.io/order-project/order-service/domain/states/state_34"
	"gitlab.faza.io/order-project/order-service/domain/states/state_35"
	"gitlab.faza.io/order-project/order-service/domain/states/state_36"
	"gitlab.faza.io/order-project/order-service/domain/states/state_40"
	"gitlab.faza.io/order-project/order-service/domain/states/state_41"
	"gitlab.faza.io/order-project/order-service/domain/states/state_42"
	"gitlab.faza.io/order-project/order-service/domain/states/state_50"
	"gitlab.faza.io/order-project/order-service/domain/states/state_51"
	"gitlab.faza.io/order-project/order-service/domain/states/state_52"
	"gitlab.faza.io/order-project/order-service/domain/states/state_53"
	"gitlab.faza.io/order-project/order-service/domain/states/state_54"
	"gitlab.faza.io/order-project/order-service/domain/states/state_55"
	"gitlab.faza.io/order-project/order-service/domain/states/state_56"
	"gitlab.faza.io/order-project/order-service/domain/states/state_80"
	"gitlab.faza.io/order-project/order-service/domain/states/state_90"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"os"
	"strconv"
	"time"

	//"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	////"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
	//"github.com/golang/protobuf/ptypes"
	//"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
)

const (
	// ISO8601 standard time format
	ISO8601 = "2006-01-02T15:04:05-0700"
)

type iFlowManagerImpl struct {
	//nameStateMap  map[string]states.IState
	statesMap map[states.IEnumState]states.IState
}

func NewFlowManager() (IFlowManager, error) {
	//nameStepsMap := make(map[string]states.IState, 64)
	statesMap := make(map[states.IEnumState]states.IState, 64)

	iFlowManagerImpl := &iFlowManagerImpl{statesMap}
	if err := iFlowManagerImpl.setupFlowManager(); err != nil {
		return nil, err
	}

	return iFlowManagerImpl, nil
}

func (flowManager *iFlowManagerImpl) GetState(state states.IEnumState) states.IState {
	return flowManager.statesMap[state]
}

func (flowManager *iFlowManagerImpl) setupFlowManager() error {
	var emptyState []states.IState
	var emptyActionState map[actions.IAction]states.IState

	//////////////////////////////////////////////////////////////////
	// Pay To SellerInfo
	// create empty step90 which is required for step92
	state := state_90.New(emptyState, emptyState, emptyActionState)

	// add to flowManager maps
	flowManager.statesMap[states.PayToSeller] = state

	//////////////////////////////////////////////////////////////////
	// Pay To Buyer
	// create empty step80 which is required for step82
	state = state_80.New(emptyState, emptyState, emptyActionState)

	// add to flowManager maps
	flowManager.statesMap[states.PayToBuyer] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap := map[actions.IAction]states.IState{
		system_action.New(system_action.Close): flowManager.statesMap[states.PayToSeller],
	}
	childStates := []states.IState{flowManager.statesMap[states.PayToSeller]}
	state = state_56.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnDeliveryFailed] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		operator_action.New(operator_action.Accept): flowManager.statesMap[states.PayToBuyer],
		operator_action.New(operator_action.Reject): flowManager.statesMap[states.PayToSeller],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToBuyer],
		flowManager.statesMap[states.PayToSeller],
	}
	state = state_55.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnRejected] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		seller_action.New(seller_action.Reject):       flowManager.statesMap[states.ReturnRejected],
		seller_action.New(seller_action.Accept):       flowManager.statesMap[states.PayToBuyer],
		scheduler_action.New(scheduler_action.Accept): flowManager.statesMap[states.PayToBuyer],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToBuyer],
		flowManager.statesMap[states.ReturnRejected],
	}
	state = state_52.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnDelivered] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		operator_action.New(operator_action.Deliver):      flowManager.statesMap[states.ReturnDelivered],
		operator_action.New(operator_action.DeliveryFail): flowManager.statesMap[states.ReturnDeliveryFailed],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ReturnDeliveryFailed],
		flowManager.statesMap[states.ReturnDelivered],
	}
	state = state_54.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnDeliveryDelayed] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		seller_action.New(seller_action.Deliver):      flowManager.statesMap[states.ReturnDelivered],
		seller_action.New(seller_action.DeliveryFail): flowManager.statesMap[states.ReturnDeliveryDelayed],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ReturnDelivered],
		flowManager.statesMap[states.ReturnDeliveryDelayed],
	}
	state = state_53.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnDeliveryPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		scheduler_action.New(scheduler_action.DeliveryPending): flowManager.statesMap[states.ReturnDeliveryPending],
		seller_action.New(seller_action.Deliver):               flowManager.statesMap[states.ReturnDelivered],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ReturnDeliveryPending],
		flowManager.statesMap[states.ReturnDelivered],
	}
	state = state_51.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnShipped] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		buyer_action.New(buyer_action.Cancel):              flowManager.statesMap[states.PayToSeller],
		scheduler_action.New(scheduler_action.Cancel):      flowManager.statesMap[states.PayToSeller],
		buyer_action.New(buyer_action.EnterShipmentDetail): flowManager.statesMap[states.ReturnShipped],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToSeller],
		flowManager.statesMap[states.ReturnShipped],
	}
	state = state_50.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnShipmentPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.Close): flowManager.statesMap[states.PayToSeller],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToSeller],
	}
	state = state_42.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnCanceled] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		operator_action.New(operator_action.Accept): flowManager.statesMap[states.ReturnShipmentPending],
		operator_action.New(operator_action.Reject): flowManager.statesMap[states.PayToSeller],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToSeller],
		flowManager.statesMap[states.ReturnShipmentPending],
	}
	state = state_41.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnRequestRejected] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		buyer_action.New(buyer_action.Cancel):         flowManager.statesMap[states.ReturnCanceled],
		seller_action.New(seller_action.Reject):       flowManager.statesMap[states.ReturnRequestRejected],
		seller_action.New(seller_action.Accept):       flowManager.statesMap[states.ReturnShipmentPending],
		scheduler_action.New(scheduler_action.Accept): flowManager.statesMap[states.ReturnShipmentPending],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ReturnCanceled],
		flowManager.statesMap[states.PayToSeller],
		flowManager.statesMap[states.ReturnShipmentPending],
	}
	state = state_40.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnRequestPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		buyer_action.New(buyer_action.SubmitReturnRequest): flowManager.statesMap[states.ReturnRequestPending],
		scheduler_action.New(scheduler_action.Close):       flowManager.statesMap[states.PayToSeller],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ReturnRequestPending],
		flowManager.statesMap[states.PayToSeller],
	}
	state = state_32.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.Delivered] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.Close): flowManager.statesMap[states.PayToBuyer],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToBuyer],
	}
	state = state_36.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.DeliveryFailed] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		operator_action.New(operator_action.Deliver):      flowManager.statesMap[states.Delivered],
		operator_action.New(operator_action.DeliveryFail): flowManager.statesMap[states.DeliveryFailed],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.Delivered],
		flowManager.statesMap[states.DeliveryFailed],
	}
	state = state_35.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.DeliveryDelayed] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		scheduler_action.New(scheduler_action.Deliver):      flowManager.statesMap[states.Delivered],
		scheduler_action.New(scheduler_action.Notification): nil,
		operator_action.New(operator_action.DeliveryDelay):  flowManager.statesMap[states.DeliveryDelayed],
		buyer_action.New(buyer_action.DeliveryDelay):        flowManager.statesMap[states.DeliveryDelayed],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.Delivered],
		flowManager.statesMap[states.DeliveryDelayed],
	}
	state = state_34.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.DeliveryPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		scheduler_action.New(scheduler_action.DeliveryPending): flowManager.statesMap[states.DeliveryPending],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.DeliveryPending],
	}
	state = state_31.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.Shipped] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.Close): flowManager.statesMap[states.PayToBuyer],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToBuyer],
	}
	state = state_21.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.CanceledBySeller] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.Close): flowManager.statesMap[states.PayToBuyer],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToBuyer],
	}
	state = state_22.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.CanceledByBuyer] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		buyer_action.New(buyer_action.Cancel):                flowManager.statesMap[states.CanceledByBuyer],
		seller_action.New(seller_action.Cancel):              flowManager.statesMap[states.CanceledBySeller],
		seller_action.New(seller_action.EnterShipmentDetail): flowManager.statesMap[states.Shipped],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.CanceledByBuyer],
		flowManager.statesMap[states.CanceledBySeller],
		flowManager.statesMap[states.Shipped],
	}
	state = state_33.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ShipmentDelayed] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		scheduler_action.New(scheduler_action.Cancel):        flowManager.statesMap[states.ShipmentDelayed],
		seller_action.New(seller_action.Cancel):              flowManager.statesMap[states.CanceledBySeller],
		seller_action.New(seller_action.EnterShipmentDetail): flowManager.statesMap[states.Shipped],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ShipmentDelayed],
		flowManager.statesMap[states.CanceledBySeller],
		flowManager.statesMap[states.Shipped],
	}
	state = state_30.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ShipmentPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		scheduler_action.New(scheduler_action.Cancel): flowManager.statesMap[states.CanceledBySeller],
		seller_action.New(seller_action.Approve):      flowManager.statesMap[states.ShipmentPending],
		seller_action.New(seller_action.Reject):       flowManager.statesMap[states.CanceledBySeller],
		buyer_action.New(buyer_action.Cancel):         flowManager.statesMap[states.CanceledByBuyer],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.CanceledByBuyer],
		flowManager.statesMap[states.CanceledBySeller],
		flowManager.statesMap[states.ShipmentPending],
	}
	state = state_20.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ApprovalPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.NextToState): flowManager.statesMap[states.ApprovalPending],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ApprovalPending],
	}
	state = state_14.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.OrderVerificationSuccess] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.Close): flowManager.statesMap[states.PayToBuyer],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToBuyer],
	}
	state = state_15.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.OrderVerificationFailed] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.Success): flowManager.statesMap[states.OrderVerificationSuccess],
		system_action.New(system_action.Fail):    flowManager.statesMap[states.OrderVerificationFailed],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.OrderVerificationSuccess],
		flowManager.statesMap[states.OrderVerificationFailed],
	}
	state = state_13.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.OrderVerificationPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.NextToState): flowManager.statesMap[states.OrderVerificationPending],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.OrderVerificationPending],
	}
	state = state_11.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.PaymentSuccess] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{}
	childStates = []states.IState{}
	state = state_12.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.PaymentFailed] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.PaymentSuccess):    flowManager.statesMap[states.PaymentSuccess],
		system_action.New(system_action.PaymentFail):       flowManager.statesMap[states.PaymentFailed],
		scheduler_action.New(scheduler_action.PaymentFail): flowManager.statesMap[states.PaymentFailed],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PaymentSuccess],
		flowManager.statesMap[states.PaymentFailed],
	}
	state = state_10.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.PaymentPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.NextToState): flowManager.statesMap[states.PaymentPending],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PaymentPending],
	}
	state = state_01.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.NewOrder] = state

	return nil
}

func (flowManager iFlowManagerImpl) PaymentGatewayResult(ctx context.Context, req *pg.PaygateHookRequest) future.IFuture {
	orderId, err := strconv.Atoi(req.OrderID)
	if err != nil {
		app.Globals.Logger.Error("request orderId invalid, OrderRepository.FindById failed",
			"fn", "PaymentGatewayResult",
			"orderId", req.OrderID,
			"error", err)

		return future.Factory().
			SetError(future.BadRequest, "OrderId Invalid", errors.Wrap(err, "strconv.Atoi() Failed")).
			BuildAndSend()
	}

	paymentResult := &entities.PaymentResult{
		Result:    req.Result,
		Reason:    "",
		PaymentId: req.PaymentId,
		InvoiceId: req.InvoiceId,
		Price: &entities.Money{
			Amount:   strconv.Itoa(int(req.Amount)),
			Currency: "IRR",
		},
		CardNumMask: req.CardMask,
		CreatedAt:   time.Now().UTC(),
	}

	iFuture := future.Factory().SetCapacity(1).Build()
	iframe := frame.Factory().SetOrderId(uint64(orderId)).
		SetDefaultHeader(frame.HeaderPaymentResult, paymentResult).
		SetFuture(iFuture).
		SetOrderId(uint64(orderId)).Build()

	flowManager.statesMap[states.PaymentPending].Process(ctx, iframe)
	return iFuture
}

// TODO Must be refactored
func (flowManager iFlowManagerImpl) MessageHandler(ctx context.Context, iFrame frame.IFrame) {
	// received New Order Request
	//if len(req.OrderId) == 0 {

	if iFrame.Header().KeyExists(string(frame.HeaderEvent)) {
		flowManager.EventHandler(ctx, iFrame)
	} else if iFrame.Header().KeyExists(string(frame.HeaderNewOrder)) {
		flowManager.newOrderHandler(ctx, iFrame)
	}
}

func (flowManager iFlowManagerImpl) newOrderHandler(ctx context.Context, iFrame frame.IFrame) {

	requestNewOrder := iFrame.Header().Value(string(frame.HeaderNewOrder))
	value, err := app.Globals.Converter.Map(requestNewOrder, entities.Order{})
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("Converter.Map requestNewOrder to order object failed",
			"fn", "newOrderHandler",
			"error", err, "requestNewOrder", requestNewOrder)
		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetError(future.BadRequest, "Received requestNewOrder invalid", err).
			Send()
		return
	}

	newOrder := value.(*entities.Order)

	var requestStockList = make([]stock_service.RequestStock, 0, 32)
	for i := 0; i < len(newOrder.Packages); i++ {
		for j := 0; j < len(newOrder.Packages[i].Subpackages); j++ {
			for z := 0; z < len(newOrder.Packages[i].Subpackages[j].Items); z++ {
				item := newOrder.Packages[i].Subpackages[j].Items[z]
				requestStock := stock_service.RequestStock{
					InventoryId: item.InventoryId,
					Count:       int(item.Quantity),
				}
				requestStockList = append(requestStockList, requestStock)
			}
		}
	}

	iFuture := app.Globals.StockService.BatchStockActions(ctx, requestStockList, 0,
		system_action.New(system_action.StockReserve))
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		app.Globals.Logger.FromContext(ctx).Error("Reserved stock from stockService failed",
			"fn", "newOrderHandler",
			"newOrder", newOrder,
			"error", futureData.Error())

		if responseStockList, ok := futureData.Data().([]stock_service.ResponseStock); ok {
			requestStockList = make([]stock_service.RequestStock, 0, 32)
			for _, response := range responseStockList {
				if response.Result {
					requestStock := stock_service.RequestStock{
						InventoryId: response.InventoryId,
						Count:       response.Count,
					}
					requestStockList = append(requestStockList, requestStock)
				}
			}
			iFuture := app.Globals.StockService.BatchStockActions(ctx, requestStockList, 0,
				system_action.New(system_action.StockRelease))
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				responseList, ok := futureData.Data().([]stock_service.ResponseStock)
				if ok {
					app.Globals.Logger.FromContext(ctx).Error("Rollback reserved stock from stockService failed",
						"fn", "newOrderHandler",
						"newOrder", newOrder,
						"response", responseList,
						"error", futureData.Error())
				} else {
					app.Globals.Logger.FromContext(ctx).Error("Rollback reserved stock from stockService failed",
						"fn", "newOrderHandler",
						"newOrder", newOrder,
						"error", futureData.Error())
				}
			} else {
				responseList := futureData.Data().([]stock_service.ResponseStock)
				app.Globals.Logger.FromContext(ctx).Debug("Rollback reserved stock from stockService success",
					"fn", "newOrderHandler",
					"newOrder", newOrder,
					"response", responseList)
			}
		}

		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetError(future.NotAccepted, "Received requestNewOrder invalid", err).
			Send()
		return
	}

	flowManager.statesMap[states.NewOrder].Process(ctx, frame.FactoryOf(iFrame).SetOrder(newOrder).Build())
}

func (flowManager iFlowManagerImpl) EventHandler(ctx context.Context, iFrame frame.IFrame) {
	event := iFrame.Header().Value(string(frame.HeaderEvent)).(events.IEvent)
	if event.EventType() == events.Action {
		if event.Action().ActionEnum() == scheduler_action.PaymentFail {
			order, err := app.Globals.OrderRepository.FindById(ctx, event.OrderId())
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("OrderRepository.FindById failed",
					"fn", "EventHandler",
					"event", event,
					"error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).Send()
				return
			}

			state := states.FromIndex(event.StateIndex())
			if state == nil {
				app.Globals.Logger.FromContext(ctx).Error("sIdx invalid",
					"fn", "EventHandler",
					"event", event, "error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", err).Send()
				return
			}

			if state, ok := flowManager.statesMap[state]; ok {
				state.Process(ctx, frame.FactoryOf(iFrame).SetBody(order).Build())
			} else {
				app.Globals.Logger.FromContext(ctx).Error("state in flowManager.statesMap no found",
					"fn", "EventHandler",
					"state", state.Name(),
					"event", event,
					"error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", err).Send()
			}
		} else {
			pkgItem, err := app.Globals.PkgItemRepository.FindById(ctx, event.OrderId(), event.PackageId())
			if err != nil {
				app.Globals.Logger.Error("PkgItemRepository.FindById failed",
					"fn", "EventHandler",
					"event", event, "error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.ErrorCode(err.Code()), err.Message(), err.Reason()).Send()
				return
			}

			state := states.FromIndex(event.StateIndex())
			if state == nil {
				app.Globals.Logger.FromContext(ctx).Error("sIdx invalid",
					"fn", "EventHandler",
					"event", event, "error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", err).Send()
				return
			}

			if state, ok := flowManager.statesMap[state]; ok {
				state.Process(ctx, frame.FactoryOf(iFrame).SetBody(pkgItem).Build())
			} else {
				app.Globals.Logger.FromContext(ctx).Error("state in flowManager.statesMap no found",
					"fn", "EventHandler",
					"state", state.Name(),
					"event", event,
					"error", err)
				future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
					SetError(future.InternalError, "Unknown Err", err).Send()
			}
		}
	}
}

func (flowManager iFlowManagerImpl) ReportOrderItems(ctx context.Context, req *pb.RequestReportOrderItems, srv pb.OrderService_ReportOrderItemsServer) future.IFuture {

	startTime, err := time.Parse(ISO8601, req.StartDateTime)
	if err != nil {
		return future.Factory().SetCapacity(1).SetError(future.BadRequest, "StartDateTime Invalid", err).Build()
	}

	endTime, err := time.Parse(ISO8601, req.EndDataTime)
	if err != nil {
		return future.Factory().SetCapacity(1).SetError(future.BadRequest, "EndDateTime Invalid", err).Build()
	}

	orders, _, e := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, func() (interface{}, string, int) {
		return bson.D{{"createdAt", bson.D{{"$gte", startTime}, {"$lte", endTime}}}},
			"createdAt", -1
	}, int64(1), int64(2000))

	if e != nil {
		app.Globals.Logger.Error("OrderRepository.FindByFilter failed",
			"fn", "ReportOrderItems",
			"startDateTime", startTime,
			"endDateTime", endTime,
			"error", e)
		return future.Factory().SetCapacity(1).SetError(future.ErrorCode(e.Code()), e.Message(), e.Reason()).Build()
	}

	if orders == nil || len(orders) == 0 {
		return future.Factory().SetCapacity(1).SetError(future.NotFound, "Order Not Found", errors.New("Order Not Found")).Build()
	}

	orderReports := make([]*reports.ExportOrderItems, 0, len(orders))
	sellerProfileMap := make(map[uint64]entities.SellerProfile, 16)

	for i := 0; i < len(orders); i++ {
		for j := 0; j < len(orders[i].Packages); j++ {
			for k := 0; k < len(orders[i].Packages[j].Subpackages); k++ {
				for z := 0; z < len(orders[i].Packages[j].Subpackages[k].Items); z++ {
					itemReport := &reports.ExportOrderItems{
						SId:               orders[i].Packages[j].Subpackages[k].SId,
						InventoryId:       orders[i].Packages[j].Subpackages[k].Items[z].InventoryId,
						ProductId:         orders[i].Packages[j].Subpackages[k].Items[z].SKU,
						BuyerId:           orders[i].BuyerInfo.BuyerId,
						BuyerPhone:        orders[i].BuyerInfo.Phone,
						SellerId:          orders[i].Packages[j].PId,
						SellerDisplayName: orders[i].Packages[j].ShopName,
						Price:             orders[i].Invoice.Subtotal.Amount,
						VoucherAmount:     nil,
						ShippingCost:      orders[i].Packages[j].Invoice.ShipmentAmount.Amount,
						Status:            orders[i].Packages[j].Subpackages[k].Status,
						CreatedAt:         "",
						UpdatedAt:         "",
					}

					if orders[i].Invoice.Voucher != nil {
						itemReport.VoucherAmount = orders[i].Invoice.Voucher.AppliedPrice.Amount
					}

					tempTime := time.Date(orders[i].CreatedAt.Year(),
						orders[i].CreatedAt.Month(),
						orders[i].CreatedAt.Day(),
						orders[i].CreatedAt.Hour(),
						orders[i].CreatedAt.Minute(),
						orders[i].CreatedAt.Second(),
						orders[i].CreatedAt.Nanosecond(),
						ptime.Iran())

					pt := ptime.New(tempTime)
					itemReport.CreatedAt = pt.String()

					tempTime = time.Date(orders[i].Packages[j].Subpackages[k].UpdatedAt.Year(),
						orders[i].Packages[j].Subpackages[k].UpdatedAt.Month(),
						orders[i].Packages[j].Subpackages[k].UpdatedAt.Day(),
						orders[i].Packages[j].Subpackages[k].UpdatedAt.Hour(),
						orders[i].Packages[j].Subpackages[k].UpdatedAt.Minute(),
						orders[i].Packages[j].Subpackages[k].UpdatedAt.Second(),
						orders[i].Packages[j].Subpackages[k].UpdatedAt.Nanosecond(),
						ptime.Iran())

					pt = ptime.New(tempTime)
					itemReport.UpdatedAt = pt.String()
					reports = append(reports, itemReport)

					if sellerProfile, ok := sellerProfileMap[item.SellerInfo.PId]; !ok {
						userCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
						ipromise := app.Globals.UserService.GetSellerProfile(userCtx, strconv.Itoa(int(item.SellerInfo.PId)))
						futureData := ipromise.Get()
						if futureData.Ex != nil {
							logger.Err("BackOfficeReportOrderItems() => get sellerProfile failed, orderId: %d, sid: %d, sellerId: %d",
								order.OrderId, item.SId, item.SellerInfo.PId)
							continue
						}

						sellerInfo, ok := futureData.Data.(entities.SellerProfile)
						if ok != true {
							logger.Err("BackOfficeReportOrderItems() => get sellerProfile invalid, orderId: %d, sid: %d, sellerId: %d",
								order.OrderId, item.SId, item.SellerInfo.PId)
							continue
						}

						sellerProfileMap[item.SellerInfo.PId] = sellerProfile
						itemReport.SellerDisplayName = sellerInfo.GeneralInfo.ShopDisplayName
					} else {
						itemReport.SellerDisplayName = sellerProfile.GeneralInfo.ShopDisplayName
					}
				}
			}
		}
	}

	for _, order := range orders {
		for _, item := range order.Items {

		}
	}

	csvReports := make([][]string, 0, len(reports))
	csvHeadLines := []string{
		"SId", "InventoryId", "ProductId", "BuyerId", "BuyerPhone", "PId",
		"SellerDisplayName", "ItemInvoice", "Status", "CreatedAt", "UpdatedAt",
	}

	csvReports = append(csvReports, csvHeadLines)
	for _, itemReport := range reports {
		csvRecord := []string{
			strconv.Itoa(int(itemReport.SId)),
			itemReport.InventoryId,
			itemReport.ProductId,
			strconv.Itoa(int(itemReport.BuyerId)),
			itemReport.BuyerPhone,
			strconv.Itoa(int(itemReport.PId)),
			itemReport.SellerName,
			fmt.Sprint(itemReport.Price),
			itemReport.Status,
			itemReport.CreatedAt,
			itemReport.UpdatedAt,
		}
		csvReports = append(csvReports, csvRecord)
	}

	reportTime := time.Unix(int64(req.StartDateTime), 0)
	fileName := fmt.Sprintf("BackOfficeReport-%s.csv", fmt.Sprintf("%d", reportTime.UnixNano()))
	f, err := os.Create("/tmp/" + fileName)
	if err != nil {
		logger.Err("BackOfficeReportOrderItems() => create file %s failed, startDateTime: %v, endDateTime: %v, error: %s",
			fileName, req.StartDateTime, req.EndDataTime, err)
		returnChannel := make(chan future.FutureData, 1)
		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewPromise(returnChannel, 1, 1)
	}

	w := csv.NewWriter(f)
	// calls Flush internally
	if err := w.WriteAll(csvReports); err != nil {
		logger.Err("BackOfficeReportOrderItems() => write csv to file failed, startDateTime: %v, endDateTime: %v, error: %s",
			req.StartDateTime, req.EndDataTime, err)
		returnChannel := make(chan future.FutureData, 1)
		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewPromise(returnChannel, 1, 1)
	}

	if err := f.Close(); err != nil {
		logger.Err("BackOfficeReportOrderItems() => file close failed, filename: %s, error: %s", fileName, err)
	}

	file, err := os.Open("/tmp/" + fileName)
	if err != nil {
		logger.Err("BackOfficeReportOrderItems() => read csv from file failed, startDateTime: %v, endDateTime: %v, error: %s",
			req.StartDateTime, req.EndDataTime, err)
		returnChannel := make(chan future.FutureData, 1)
		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewPromise(returnChannel, 1, 1)
	}

	var fileErr, grpcErr error
	var b [4096 * 1000]byte
	for {
		n, err := file.Read(b[:])
		if err != nil {
			if err != io.EOF {
				fileErr = err
			}
			break
		}
		err = srv.Send(&message.ResponseDownloadFile{
			Data: b[:n],
		})
		if err != nil {
			grpcErr = err
		}
	}

	if err := file.Close(); err != nil {
		logger.Err("BackOfficeReportOrderItems() => file close failed, filename: %s, error: %s", file.Name(), err)
	}

	if err := os.Remove("/tmp/" + fileName); err != nil {
		logger.Err("BackOfficeReportOrderItems() => remove file failed, filename: %s, error: %s", fileName, err)
	}

	if fileErr != nil {
		logger.Err("BackOfficeReportOrderItems() => read csv from file failed, filename: %s, error: %s", fileName, err)
		returnChannel := make(chan future.FutureData, 1)
		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewPromise(returnChannel, 1, 1)
	}

	if grpcErr != nil {
		logger.Err("BackOfficeReportOrderItems() => send cvs file failed, filename: %s, error: %s", fileName, err)
		returnChannel := make(chan future.FutureData, 1)
		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return future.NewPromise(returnChannel, 1, 1)
	}

	returnChannel := make(chan future.FutureData, 1)
	returnChannel <- future.FutureData{Data: nil, Ex: nil}
	defer close(returnChannel)
	return future.NewPromise(returnChannel, 1, 1)
}
