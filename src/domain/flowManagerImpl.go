package domain

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/buyer"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	payment_action "gitlab.faza.io/order-project/order-service/domain/actions/payment"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
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
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"strconv"
	"time"

	//"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	////"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
	pg "gitlab.faza.io/protos/payment-gateway"
	//"github.com/golang/protobuf/ptypes"
	//"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
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
		flowManager.statesMap[states.PayToBuyer],
		flowManager.statesMap[states.PayToSeller],
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
		buyer_action.New(buyer_action.Cancel):               flowManager.statesMap[states.PayToSeller],
		scheduler_action.New(scheduler_action.Close):        flowManager.statesMap[states.PayToSeller],
		buyer_action.New(buyer_action.EnterShipmentDetails): flowManager.statesMap[states.ReturnShipped],
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
	//actionStateMap = map[actions.IAction]states.IState{
	//	system_action.New(system_action.NextToState): flowManager.statesMap[states.PayToSeller],
	//}
	//childStates = []states.IState{
	//	flowManager.statesMap[states.PayToSeller],
	//}
	//state = state_42.New(childStates, emptyState, actionStateMap)
	//// add to flowManager maps
	//flowManager.statesMap[states.ReturnCanceled] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		buyer_action.New(buyer_action.Cancel):         flowManager.statesMap[states.PayToSeller],
		seller_action.New(seller_action.Reject):       flowManager.statesMap[states.ReturnRequestRejected],
		seller_action.New(seller_action.Accept):       flowManager.statesMap[states.ReturnShipmentPending],
		scheduler_action.New(scheduler_action.Accept): flowManager.statesMap[states.ReturnShipmentPending],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToSeller],
		flowManager.statesMap[states.ReturnShipmentPending],
	}
	state = state_40.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnRequestPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		buyer_action.New(buyer_action.SubmitReturnRequest): flowManager.statesMap[states.ReturnRequestPending],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.ReturnRequestPending],
	}
	state = state_32.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnRequestPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		system_action.New(system_action.NextToState): flowManager.statesMap[states.PayToSeller],
	}
	childStates = []states.IState{
		flowManager.statesMap[states.PayToSeller],
	}
	state = state_36.New(childStates, emptyState, actionStateMap)
	// add to flowManager maps
	flowManager.statesMap[states.ReturnRequestPending] = state

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
	flowManager.statesMap[states.ReturnRequestPending] = state

	////////////////////////////////////////////////////////////////////
	actionStateMap = map[actions.IAction]states.IState{
		scheduler_action.New(scheduler_action.Deliver):     flowManager.statesMap[states.Delivered],
		operator_action.New(operator_action.DeliveryDelay): flowManager.statesMap[states.DeliveryDelayed],
		buyer_action.New(buyer_action.DeliveryDelay):       flowManager.statesMap[states.DeliveryDelayed],
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
	flowManager.statesMap[states.Shipped] = state

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
	flowManager.statesMap[states.ShipmentPending] = state

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
		payment_action.New(payment_action.Success): flowManager.statesMap[states.PaymentSuccess],
		payment_action.New(payment_action.Fail):    flowManager.statesMap[states.PaymentFailed],
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
		logger.Err("PaymentGatewayResult() => request orderId invalid, OrderRepository.FindById failed, order: %s, error: %s",
			req.OrderID, err)

		return future.Factory().
			SetError(future.BadRequest, "OrderId Invalid", errors.Wrap(err, "strconv.Atoi() Failed")).
			BuildAndSend()
	}

	paymentResult := &entities.PaymentResult{
		Result:      req.Result,
		Reason:      "",
		PaymentId:   req.PaymentId,
		InvoiceId:   req.InvoiceId,
		Amount:      uint64(req.Amount),
		CardNumMask: req.CardMask,
		CreatedAt:   time.Now().UTC(),
	}

	iframe := frame.Factory().SetDefaultHeader(frame.HeaderPaymentResult, paymentResult).SetOrderId(uint64(orderId)).Build()

	//order, err := global.Singletons.OrderRepository.FindById(ctx, uint64(orderId))
	//if err != nil {
	//	logger.Err("PaymentGatewayResult() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
	//		req.OrderID, err)
	//	returnChannel := make(chan future.FutureData, 1)
	//	returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "OrderId Not Found"}}
	//	defer close(returnChannel)
	//	return future.NewPromise(returnChannel, 1, 1)
	//}
	//

	//order.PaymentService[0].PaymentResult = &entities.PaymentResult{
	//	Result:      req.Result,
	//	Reason:      "",
	//	PaymentId:   req.PaymentId,
	//	InvoiceId:   req.InvoiceId,
	//	Amount:      uint64(req.Amount),
	//	CardNumMask: req.CardMask,
	//	CreatedAt:   time.Now().UTC(),
	//}

	flowManager.statesMap[states.PaymentPending].Process(ctx, iframe)
	return future.Factory().BuildAndSend()
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
	//if err := ptypes.UnmarshalAny(request.Data, &requestNewOrder); err != nil {
	//	logger.Err("Could not unmarshal requestNewOrder from anything field, error: %s, request: %v", err, request)
	//	returnChannel := make(chan future.IDataFuture, 1)
	//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.BadRequest, Reason: "Invalid requestNewOrder"}}
	//	close(returnChannel)
	//	return future.NewFuture(returnChannel, 1, 1)
	//}
	//
	//timestamp, err := ptypes.Timestamp(request.Time)
	//if err != nil {
	//	logger.Err("timestamp of requestNewOrder invalid, error: %s, requestNewOrder: %v", err, requestNewOrder)
	//	returnChannel := make(chan future.IDataFuture, 1)
	//	returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.BadRequest, Reason:"Invalid Request Timestamp"}}
	//	defer close(returnChannel)
	//	return future.NewFuture(returnChannel, 1, 1)
	//}

	value, err := global.Singletons.Converter.Map(requestNewOrder, entities.Order{})
	if err != nil {
		logger.Err("Converter.Map requestNewOrder to order object failed, error: %s, requestNewOrder: %v", err, requestNewOrder)
		future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
			SetError(future.BadRequest, "Received requestNewOrder invalid", err).
			Send()
	}

	newOrder := value.(*entities.Order)

	var inventories = make(map[string]int, 32)
	for i := 0; i < len(newOrder.Packages); i++ {
		for j := 0; j < len(newOrder.Packages[i].Subpackages); j++ {
			for z := 0; z < len(newOrder.Packages[i].Subpackages[j].Items); z++ {
				item := newOrder.Packages[i].Subpackages[j].Items[z]
				inventories[item.InventoryId] = int(item.Quantity)
			}
		}
	}

	iFuture := global.Singletons.StockService.BatchStockActions(ctx, inventories,
		stock_action.New(stock_action.Release))
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		logger.Err("Reserved stock from stockService failed, newOrder: %v, error: %s",
			newOrder, futureData.Error())
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
		pkgItem, err := global.Singletons.PkgItemRepository.FindById(ctx, event.OrderId(), event.PackageId())
		if err != nil {
			logger.Err("EventHandler => SubPkgRepository.FindByOrderAndSellerId failed, event: %v, error: %s ", event, err)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", err).Send()
		}

		state := states.FromIndex(event.StateIndex())
		if state == nil {
			logger.Err("EventHandler => stateIndex invalid, event: %v, error: %s ", event, err)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", err).Send()
		}

		if state, ok := flowManager.statesMap[state]; ok {
			state.Process(ctx, frame.FactoryOf(iFrame).SetBody(pkgItem).Build())
		} else {
			logger.Err("EventHandler => state in flowManager.statesMap no found, state: %s, event: %v, error: %s ", state, event, err)
			future.FactoryOf(iFrame.Header().Value(string(frame.HeaderFuture)).(future.IFuture)).
				SetError(future.InternalError, "Unknown Err", err).Send()
		}
	}
}

//func (flowManager iFlowManagerImpl) SellerApprovalPending(ctx context.Context, req *message.RequestSellerOrderAction) future.IPromise {
//	order, err := global.Singletons.OrderRepository.FindById(req.OrderId)
//	if err != nil {
//		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, orderId: %d, error: %s",
//			req.OrderId, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "OrderId Not Found"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	itemsId := make([]uint64, 0, len(order.Items))
//	for i := 0; i < len(order.Items); i++ {
//		if order.Items[i].SellerInfo.PId == req.PId {
//			itemsId = append(itemsId, order.Items[i].SId)
//		}
//	}
//
//	if req.ActionType == "approved" {
//		return flowManager.statesMap[20].ProcessOrder(ctx, *order, itemsId, req)
//	} else if req.ActionType == "shipped" {
//		return flowManager.statesMap[30].ProcessOrder(ctx, *order, itemsId, req)
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "ActionType Not Found"}}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}
//
//func (flowManager iFlowManagerImpl) BuyerApprovalPending(ctx context.Context, req *message.RequestBuyerOrderAction) future.IPromise {
//	order, err := global.Singletons.OrderRepository.FindById(req.OrderId)
//	if err != nil {
//		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, orderId: %d, error: %s",
//			req.OrderId, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "OrderId Not Found"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	//itemsId := make([]string, 0, len(order.Items))
//	//for i:= 0; i < len(order.Items); i++ {
//	//	if order.Items[i].SellerInfo.PId == req.SId[i] {
//	//		itemsId = append(itemsId,order.Items[i].SId)
//	//	}
//	//}
//
//	if req.ActionType == "Approved" {
//		return flowManager.statesMap[32].ProcessOrder(ctx, *order, req.SIds, req)
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "ActionType Not Found"}}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}
//
//func (flowManager iFlowManagerImpl) OperatorActionPending(ctx context.Context, req *message.RequestBackOfficeOrderAction) future.IPromise {
//	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
//		return bson.D{{"items.sid", req.SId}}
//	})
//
//	if err != nil {
//		logger.Err("MessageHandler() => request sid not found, OrderRepository.FindById failed, sid: %d, error: %s",
//			req.SId, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if len(orders) == 0 {
//		logger.Err("MessageHandler() => request sid not found, sid: %d", req.SId)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "SId Not Found"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if len(orders) > 1 {
//		logger.Err("MessageHandler() => request sid found in multiple order, sid: %d, error: %s",
//			req.SId, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	itemsId := make([]uint64, 0, 1)
//	for i := 0; i < len(orders[0].Items); i++ {
//		if orders[0].Items[i].SId == req.SId {
//			itemsId = append(itemsId, orders[0].Items[i].SId)
//		}
//	}
//
//	if req.ActionType == "shipmentDelivered" {
//		return flowManager.statesMap[43].ProcessOrder(ctx, *orders[0], itemsId, req)
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "ActionType Not Found"}}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}
//
//func (flowManager iFlowManagerImpl) BackOfficeOrdersListView(ctx context.Context, req *message.RequestBackOfficeOrdersList) future.IPromise {
//	orders, total, err := global.Singletons.OrderRepository.FindAllWithPageAndSort(int64(req.Page), int64(req.PerPage), req.Sort, int(req.Direction))
//
//	if err != nil {
//		logger.Err("BackOfficeOrdersListView() => FindAllWithPageAndSort failed, error: %s", err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if len(orders) == 0 {
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "Orders Not Found"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	response := message.ResponseBackOfficeOrdersList{
//		Total:  total,
//		Orders: make([]*message.BackOfficeOrdersList, 0, len(orders)),
//	}
//
//	for _, order := range orders {
//		backOfficeOrder := &message.BackOfficeOrdersList{
//			OrderId:     order.OrderId,
//			PurchasedOn: order.CreatedAt.Unix(),
//			BasketSize:  0,
//			BillTo:      order.BuyerInfo.FirstName + order.BuyerInfo.LastName,
//			ShipTo:      order.BuyerInfo.ShippingAddress.FirstName + order.BuyerInfo.ShippingAddress.LastName,
//			TotalAmount: int64(order.Invoice.Total),
//			Status:      order.Status,
//			LastUpdated: order.UpdatedAt.Unix(),
//			Actions:     []string{"success", "cancel"},
//		}
//
//		itemsInventory := make(map[string]int, len(order.Items))
//		for i := 0; i < len(order.Items); i++ {
//			if _, ok := itemsInventory[order.Items[i].InventoryId]; !ok {
//				backOfficeOrder.BasketSize += order.Items[i].Quantity
//			}
//		}
//
//		if order.Invoice.Voucher != nil {
//			backOfficeOrder.PaidAmount = int64(order.Invoice.Total - order.Invoice.Voucher.Amount)
//			backOfficeOrder.Voucher = true
//		} else {
//			backOfficeOrder.Voucher = false
//			backOfficeOrder.PaidAmount = 0
//		}
//
//		response.Orders = append(response.Orders, backOfficeOrder)
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: &response, Ex: nil}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}
//
//// TODO check payment length
//func (flowManager iFlowManagerImpl) BackOfficeOrderDetailView(ctx context.Context, req *message.RequestIdentifier) future.IPromise {
//
//	orderId, err := strconv.Atoi(req.Id)
//	if err != nil {
//		logger.Err("BackOfficeOrderDetailView() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
//			req.Id, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.BadRequest, Reason: "OrderId Invalid"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	order, err := global.Singletons.OrderRepository.FindById(uint64(orderId))
//	if err != nil {
//		logger.Err("BackOfficeOrderDetailView() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
//			req.Id, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "OrderId Not Found"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	response := &message.ResponseOrderDetailView{
//		OrderId:   order.OrderId,
//		CreatedAt: order.CreatedAt.Unix(),
//		Ip:        order.BuyerInfo.IP,
//		Status:    order.Status,
//		Payment: &message.PaymentInfo{
//			PaymentMethod: order.Invoice.PaymentMethod,
//			PaymentOption: order.Invoice.PaymentGateway,
//		},
//		Billing: &message.BillingInfo{
//			BuyerId:    order.BuyerInfo.BuyerId,
//			FullName:   order.BuyerInfo.FirstName + order.BuyerInfo.LastName,
//			Phone:      order.BuyerInfo.Phone,
//			Mobile:     order.BuyerInfo.Mobile,
//			NationalId: order.BuyerInfo.NationalId,
//		},
//		ShippingInfo: &message.ShippingInfo{
//			FullName:     order.BuyerInfo.ShippingAddress.FirstName + order.BuyerInfo.ShippingAddress.LastName,
//			Country:      order.BuyerInfo.ShippingAddress.Country,
//			City:         order.BuyerInfo.ShippingAddress.City,
//			Province:     order.BuyerInfo.ShippingAddress.Province,
//			Neighborhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
//			Address:      order.BuyerInfo.ShippingAddress.Address,
//			ZipCode:      order.BuyerInfo.ShippingAddress.ZipCode,
//		},
//		Items: make([]*message.ItemInfo, 0, len(order.Items)),
//	}
//
//	for _, item := range order.Items {
//		itemInfo := &message.ItemInfo{
//			SId:      item.SId,
//			PId:    item.SellerInfo.PId,
//			InventoryId: item.InventoryId,
//			Quantity:    item.Quantity,
//			ItemStatus:  item.Status,
//			Price: &message.PriceInfo{
//				Unit:             item.Invoice.Unit,
//				Total:            item.Invoice.Total,
//				Original:         item.Invoice.Original,
//				Special:          item.Invoice.Special,
//				Discount:         item.Invoice.Discount,
//				SellerCommission: item.Invoice.SellerCommission,
//				Currency:         item.Invoice.Currency,
//			},
//			UpdatedAt: item.UpdatedAt.Unix(),
//			Actions:   []string{"success", "cancel"},
//		}
//
//		lastStep := item.Progress.StepsHistory[len(item.Progress.StepsHistory)-1]
//
//		if lastStep.ActionHistory != nil {
//			lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
//			itemInfo.StepStatus = lastAction.Name
//		} else {
//			itemInfo.StepStatus = "none"
//			logger.Audit("BackOfficeOrderDetailView() => Actions History is nil, orderId: %d, sid: %d", order.OrderId, item.SId)
//		}
//
//		//lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
//		//itemInfo.StepStatus = lastAction.ActionName
//
//		response.Items = append(response.Items, itemInfo)
//	}
//
//	if order.PaymentService != nil && len(order.PaymentService) == 1 {
//		if order.PaymentService[0].PaymentResult != nil {
//			response.Payment.Result = order.PaymentService[0].PaymentResponse.Result
//		} else {
//			response.Payment.Result = false
//		}
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: response, Ex: nil}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}
//
//func (flowManager iFlowManagerImpl) SellerReportOrders(req *message.RequestSellerReportOrders, srv message.OrderService_SellerReportOrdersServer) future.IPromise {
//	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
//		return bson.D{{"createdAt",
//			bson.D{{"$gte", time.Unix(int64(req.StartDateTime), 0).UTC()}}},
//			{"items.status", req.Status}, {"items.sellerInfo.sellerId", req.PId}}
//	})
//
//	if err != nil {
//		logger.Err("SellerReportOrders() => OrderRepository.FindByFilter failed, startDateTime: %v, status: %s, error: %s",
//			req.StartDateTime, req.Status, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if orders == nil || len(orders) == 0 {
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: nil}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	reports := make([]*reports2.SellerExportOrders, 0, len(orders))
//
//	for _, order := range orders {
//		for _, item := range order.Items {
//			if item.Status == req.Status {
//				itemReport := &reports2.SellerExportOrders{
//					OrderId:     order.OrderId,
//					SId:      item.SId,
//					ProductId:   item.InventoryId[0:8],
//					InventoryId: item.InventoryId,
//					PaidPrice:   item.Invoice.Total,
//					Commission:  item.Invoice.SellerCommission,
//					Category:    item.Category,
//					Status:      item.Status,
//				}
//
//				//localTime := item.CreatedAt.Local()
//				tempTime := time.Date(item.CreatedAt.Year(),
//					item.CreatedAt.Month(),
//					item.CreatedAt.Day(),
//					item.CreatedAt.Hour(),
//					item.CreatedAt.Minute(),
//					item.CreatedAt.Second(),
//					item.CreatedAt.Nanosecond(),
//					ptime.Iran())
//
//				pt := ptime.New(tempTime)
//				itemReport.CreatedAt = pt.String()
//
//				tempTime = time.Date(item.UpdatedAt.Year(),
//					item.UpdatedAt.Month(),
//					item.UpdatedAt.Day(),
//					item.UpdatedAt.Hour(),
//					item.UpdatedAt.Minute(),
//					item.UpdatedAt.Second(),
//					item.UpdatedAt.Nanosecond(),
//					ptime.Iran())
//
//				pt = ptime.New(tempTime)
//				itemReport.UpdatedAt = pt.String()
//				reports = append(reports, itemReport)
//			}
//		}
//	}
//
//	csvReports := make([][]string, 0, len(reports))
//	csvHeadLines := []string{
//		"OrderId", "SId", "ProductId", "InventoryId",
//		"PaidPrice", "Commission", "Category", "Status", "CreatedAt", "UpdatedAt",
//	}
//
//	csvReports = append(csvReports, csvHeadLines)
//	for _, itemReport := range reports {
//		csvRecord := []string{
//			strconv.Itoa(int(itemReport.OrderId)),
//			strconv.Itoa(int(itemReport.SId)),
//			itemReport.ProductId,
//			itemReport.InventoryId,
//			fmt.Sprint(itemReport.PaidPrice),
//			fmt.Sprint(itemReport.Commission),
//			itemReport.Category,
//			itemReport.Status,
//			itemReport.CreatedAt,
//			itemReport.UpdatedAt,
//		}
//		csvReports = append(csvReports, csvRecord)
//	}
//
//	reportTime := time.Unix(int64(req.StartDateTime), 0)
//	fileName := fmt.Sprintf("SellerReportOrders-%s.csv", fmt.Sprintf("%d", reportTime.UnixNano()))
//	f, err := os.Create("/tmp/" + fileName)
//	if err != nil {
//		logger.Err("SellerReportOrders() => create file %s failed, startDateTime: %v, status: %s, error: %s",
//			fileName, req.StartDateTime, req.Status, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	w := csv.NewWriter(f)
//	// calls Flush internally
//	if err := w.WriteAll(csvReports); err != nil {
//		logger.Err("SellerReportOrders() => write csv to file failed, startDateTime: %v, : status: %s, error: %s",
//			req.StartDateTime, req.Status, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if err := f.Close(); err != nil {
//		logger.Err("SellerReportOrders() => file close failed, filename: %s, error: %s", fileName, err)
//	}
//
//	file, err := os.Open("/tmp/" + fileName)
//	if err != nil {
//		logger.Err("SellerReportOrders() => write csv to file failed, startDateTime: %v, : status: %s, error: %s",
//			req.StartDateTime, req.Status, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	var fileErr, grpcErr error
//	var b [4096 * 1000]byte
//	for {
//		n, err := file.Read(b[:])
//		if err != nil {
//			if err != io.EOF {
//				fileErr = err
//			}
//			break
//		}
//		err = srv.Send(&message.ResponseDownloadFile{
//			Data: b[:n],
//		})
//		if err != nil {
//			grpcErr = err
//		}
//	}
//
//	if err := file.Close(); err != nil {
//		logger.Err("SellerReportOrders() => file close failed, filename: %s, error: %s", fileName, err)
//	}
//
//	if err := os.Remove("/tmp/" + fileName); err != nil {
//		logger.Err("SellerReportOrders() => remove file failed, filename: %s, error: %s", fileName, err)
//	}
//
//	if fileErr != nil {
//		logger.Err("SellerReportOrders() => read csv from file failed, filename: %s, error: %s", fileName, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if grpcErr != nil {
//		logger.Err("SellerReportOrders() => send cvs file failed, filename: %s, error: %s", fileName, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: nil, Ex: nil}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}
//
//func (flowManager iFlowManagerImpl) BackOfficeReportOrderItems(req *message.RequestBackOfficeReportOrderItems, srv message.OrderService_BackOfficeReportOrderItemsServer) future.IPromise {
//	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
//		return bson.D{{"createdAt",
//			bson.D{{"$gte", time.Unix(int64(req.StartDateTime), 0).UTC()},
//				{"$lte", time.Unix(int64(req.EndDataTime), 0).UTC()}}}}
//	})
//
//	if err != nil {
//		logger.Err("BackOfficeReportOrderItems() => request sid not found, OrderRepository.FindById failed, startDateTime: %v, endDateTime: %v, error: %s",
//			req.StartDateTime, req.EndDataTime, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if orders == nil || len(orders) == 0 {
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: nil}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	reports := make([]*reports2.BackOfficeExportItems, 0, len(orders))
//	sellerProfileMap := make(map[uint64]entities.SellerProfile)
//
//	for _, order := range orders {
//		for _, item := range order.Items {
//			itemReport := &reports2.BackOfficeExportItems{
//				SId:      item.SId,
//				InventoryId: item.InventoryId,
//				ProductId:   item.InventoryId[0:8],
//				BuyerId:     order.BuyerInfo.BuyerId,
//				BuyerPhone:  order.BuyerInfo.Phone,
//				PId:    item.SellerInfo.PId,
//				SellerName:  "",
//				Price:       item.Invoice.Total,
//				Status:      item.Status,
//			}
//
//			tempTime := time.Date(item.CreatedAt.Year(),
//				item.CreatedAt.Month(),
//				item.CreatedAt.Day(),
//				item.CreatedAt.Hour(),
//				item.CreatedAt.Minute(),
//				item.CreatedAt.Second(),
//				item.CreatedAt.Nanosecond(),
//				ptime.Iran())
//
//			pt := ptime.New(tempTime)
//			itemReport.CreatedAt = pt.String()
//
//			tempTime = time.Date(item.UpdatedAt.Year(),
//				item.UpdatedAt.Month(),
//				item.UpdatedAt.Day(),
//				item.UpdatedAt.Hour(),
//				item.UpdatedAt.Minute(),
//				item.UpdatedAt.Second(),
//				item.UpdatedAt.Nanosecond(),
//				ptime.Iran())
//
//			pt = ptime.New(tempTime)
//			itemReport.UpdatedAt = pt.String()
//			reports = append(reports, itemReport)
//
//			if sellerProfile, ok := sellerProfileMap[item.SellerInfo.PId]; !ok {
//				userCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
//				ipromise := global.Singletons.UserService.GetSellerProfile(userCtx, strconv.Itoa(int(item.SellerInfo.PId)))
//				futureData := ipromise.Get()
//				if futureData.Ex != nil {
//					logger.Err("BackOfficeReportOrderItems() => get sellerProfile failed, orderId: %d, sid: %d, sellerId: %d",
//						order.OrderId, item.SId, item.SellerInfo.PId)
//					continue
//				}
//
//				sellerInfo, ok := futureData.Data.(entities.SellerProfile)
//				if ok != true {
//					logger.Err("BackOfficeReportOrderItems() => get sellerProfile invalid, orderId: %d, sid: %d, sellerId: %d",
//						order.OrderId, item.SId, item.SellerInfo.PId)
//					continue
//				}
//
//				sellerProfileMap[item.SellerInfo.PId] = sellerProfile
//				itemReport.SellerName = sellerInfo.GeneralInfo.ShopDisplayName
//			} else {
//				itemReport.SellerName = sellerProfile.GeneralInfo.ShopDisplayName
//			}
//		}
//	}
//
//	csvReports := make([][]string, 0, len(reports))
//	csvHeadLines := []string{
//		"SId", "InventoryId", "ProductId", "BuyerId", "BuyerPhone", "PId",
//		"SellerName", "ItemInvoice", "Status", "CreatedAt", "UpdatedAt",
//	}
//
//	csvReports = append(csvReports, csvHeadLines)
//	for _, itemReport := range reports {
//		csvRecord := []string{
//			strconv.Itoa(int(itemReport.SId)),
//			itemReport.InventoryId,
//			itemReport.ProductId,
//			strconv.Itoa(int(itemReport.BuyerId)),
//			itemReport.BuyerPhone,
//			strconv.Itoa(int(itemReport.PId)),
//			itemReport.SellerName,
//			fmt.Sprint(itemReport.Price),
//			itemReport.Status,
//			itemReport.CreatedAt,
//			itemReport.UpdatedAt,
//		}
//		csvReports = append(csvReports, csvRecord)
//	}
//
//	reportTime := time.Unix(int64(req.StartDateTime), 0)
//	fileName := fmt.Sprintf("BackOfficeReport-%s.csv", fmt.Sprintf("%d", reportTime.UnixNano()))
//	f, err := os.Create("/tmp/" + fileName)
//	if err != nil {
//		logger.Err("BackOfficeReportOrderItems() => create file %s failed, startDateTime: %v, endDateTime: %v, error: %s",
//			fileName, req.StartDateTime, req.EndDataTime, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	w := csv.NewWriter(f)
//	// calls Flush internally
//	if err := w.WriteAll(csvReports); err != nil {
//		logger.Err("BackOfficeReportOrderItems() => write csv to file failed, startDateTime: %v, endDateTime: %v, error: %s",
//			req.StartDateTime, req.EndDataTime, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if err := f.Close(); err != nil {
//		logger.Err("BackOfficeReportOrderItems() => file close failed, filename: %s, error: %s", fileName, err)
//	}
//
//	file, err := os.Open("/tmp/" + fileName)
//	if err != nil {
//		logger.Err("BackOfficeReportOrderItems() => read csv from file failed, startDateTime: %v, endDateTime: %v, error: %s",
//			req.StartDateTime, req.EndDataTime, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	var fileErr, grpcErr error
//	var b [4096 * 1000]byte
//	for {
//		n, err := file.Read(b[:])
//		if err != nil {
//			if err != io.EOF {
//				fileErr = err
//			}
//			break
//		}
//		err = srv.Send(&message.ResponseDownloadFile{
//			Data: b[:n],
//		})
//		if err != nil {
//			grpcErr = err
//		}
//	}
//
//	if err := file.Close(); err != nil {
//		logger.Err("BackOfficeReportOrderItems() => file close failed, filename: %s, error: %s", file.Name(), err)
//	}
//
//	if err := os.Remove("/tmp/" + fileName); err != nil {
//		logger.Err("BackOfficeReportOrderItems() => remove file failed, filename: %s, error: %s", fileName, err)
//	}
//
//	if fileErr != nil {
//		logger.Err("BackOfficeReportOrderItems() => read csv from file failed, filename: %s, error: %s", fileName, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if grpcErr != nil {
//		logger.Err("BackOfficeReportOrderItems() => send cvs file failed, filename: %s, error: %s", fileName, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: nil, Ex: nil}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}
//
//func (flowManager iFlowManagerImpl) SchedulerEvents(event events.ISchedulerEvent) {
//	order, err := global.Singletons.OrderRepository.FindById(event.OrderId)
//	if err != nil {
//		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, schedulerEvent: %v, error: %s",
//			event, err)
//		return
//	}
//
//	//itemsId := make([]string, 0, len(order.Items))
//	//for i:= 0; i < len(event.SIds); i++ {
//	//	if order.Items[i].SId == event.SIds[i] && order.Items[i].SellerInfo.PId == event.PId {
//	//		itemsId = append(itemsId,order.Items[i].SId)
//	//	}
//	//}
//
//	ctx, _ := context.WithCancel(context.Background())
//
//	if event.Action == "ApprovalPending" {
//		flowManager.statesMap[20].ProcessOrder(ctx, *order, event.SIds, "actionExpired")
//
//	} else if event.Action == "SellerShipmentPending" {
//		flowManager.statesMap[30].ProcessOrder(ctx, *order, event.SIds, "actionExpired")
//
//	} else if event.Action == "ShipmentDeliveredPending" {
//		flowManager.statesMap[32].ProcessOrder(ctx, *order, event.SIds, "actionApproved")
//	}
//
//}
