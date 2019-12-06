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
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
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
	"strconv"
	"time"

	//"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	//pb "gitlab.faza.io/protos/order"
	////"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
	pg "gitlab.faza.io/protos/payment-gateway"
	//"github.com/golang/protobuf/ptypes"
	//"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	//message "gitlab.faza.io/protos/order"
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
		scheduler_action.New(scheduler_action.Reject):       flowManager.statesMap[states.PayToSeller],
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

//func (flowManager *iFlowManagerImpl) createStep94() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	// Create Finalize Status
//	finalizeState := finalize_state.New(1, emptyState, emptyState,
//		finalize_action.NewOf(finalize_action.MarketFinalizeAction))
//
//	// Create Notification Status
//	notificationState := notification_state.New(0, []states_old.IState{finalizeState}, emptyState,
//		notification_action.NewOf(notification_action.MarketNotificationAction))
//	step94 := pay_to_market_success_step.New(emptyStep, emptyStep, notificationState, finalizeState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step94.Index()] = step94
//	flowManager.nameStateMap[step94.Name()] = step94
//}
//
//func (flowManager *iFlowManagerImpl) createStep95() {
//	var emptyState []states_old.IState
//
//	step93 := flowManager.statesMap[93]
//	step94 := flowManager.statesMap[94]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		retry_action.RetryAction:                          step93,
//		manual_payment_action.ManualPaymentToMarketAction: step94,
//	}
//
//	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
//		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	manualPaymentState := manual_payment_state.New(2, []states_old.IState{nextToStep}, emptyState,
//		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToMarketAction))
//	operatorNotificationState := notification_state.New(1, []states_old.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
//	retryState := retry_state.New(0, []states_old.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))
//
//	step95 := pay_to_market_failed_step.New([]states.IStep{step94}, []states.IStep{step93},
//		retryState, operatorNotificationState, manualPaymentState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step95.Index()] = step95
//	flowManager.nameStateMap[step95.Name()] = step95
//}
//
//func (flowManager *iFlowManagerImpl) createStep93(baseStep93 states.IBaseStep) {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step94 := flowManager.statesMap[94]
//	step95 := flowManager.statesMap[95]
//
//	payToMarketActions := pay_to_market_action.NewOf(pay_to_market_action.SuccessAction, pay_to_market_action.FailedAction)
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		pay_to_market_action.SuccessAction: step94,
//		pay_to_market_action.FailedAction:  step95,
//	}
//
//	nextToStepState := next_to_step_state.New(1, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	payToMarketState := pay_to_market_state.New(0, []states_old.IState{nextToStepState}, emptyState, payToMarketActions)
//
//	// TODO check change baseStep3 cause of change Step93 settings
//	baseStep93.BaseStep().SetChildes([]states.IStep{step94, step95})
//	baseStep93.BaseStep().SetParents(emptyStep)
//	baseStep93.BaseStep().SetStates(payToMarketState, nextToStepState)
//
//	// add to flowManager maps
//	//flowManager.statesMap[step93.Index()] = step93
//	//flowManager.nameStateMap[step93.ActionName()] = step93
//}
//
//func (flowManager *iFlowManagerImpl) createStep91() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step93 := flowManager.statesMap[93]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		next_to_step_action.NextToStepAction: step93,
//	}
//
//	nextToStep93 := next_to_step_state.New(1, emptyState, emptyState,
//		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//
//	// Create Notification Status
//	notificationState := notification_state.New(0, []states_old.IState{nextToStep93}, emptyState,
//		notification_action.NewOf(notification_action.SellerNotificationAction))
//
//	step91 := pay_to_seller_success_step.New([]states.IStep{step93}, emptyStep, notificationState, nextToStep93)
//
//	// add to flowManager maps
//	flowManager.statesMap[step91.Index()] = step91
//	flowManager.nameStateMap[step91.Name()] = step91
//}
//
//// TODO: checking flow and next to step states_old sequence
//func (flowManager *iFlowManagerImpl) createStep92() {
//	var emptyState []states_old.IState
//
//	step90 := flowManager.statesMap[90]
//	step91 := flowManager.statesMap[91]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		retry_action.RetryAction:                          step90,
//		manual_payment_action.ManualPaymentToSellerAction: step91,
//	}
//
//	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
//		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//
//	manualPaymentState := manual_payment_state.New(2, []states_old.IState{nextToStep}, emptyState,
//		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToSellerAction))
//
//	operatorNotificationState := notification_state.New(1, []states_old.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
//	retryState := retry_state.New(0, []states_old.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))
//
//	step92 := pay_to_seller_failed_step.New([]states.IStep{step91}, []states.IStep{step90},
//		retryState, operatorNotificationState, manualPaymentState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step92.Index()] = step92
//	flowManager.nameStateMap[step92.Name()] = step92
//}
//
//// TODO settlement stock must be call once as a result it must be save in db
//func (flowManager *iFlowManagerImpl) createStep90(baseStep90 states.IBaseStep) {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step91 := flowManager.statesMap[91]
//	step92 := flowManager.statesMap[92]
//
//	payToSellerActions := pay_to_seller_action.NewOf(pay_to_seller_action.SuccessAction, pay_to_seller_action.FailedAction)
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		pay_to_seller_action.SuccessAction: step91,
//		pay_to_seller_action.FailedAction:  step92,
//	}
//
//	nextToStepState := next_to_step_state.New(1, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	payToSellerState := pay_to_seller_state.New(0, []states_old.IState{nextToStepState}, emptyState, payToSellerActions)
//
//	// TODO check change baseStep90 cause of change Step90 settings
//	baseStep90.BaseStep().SetChildes([]states.IStep{step91, step92})
//	baseStep90.BaseStep().SetParents(emptyStep)
//	baseStep90.BaseStep().SetStates(payToSellerState, nextToStepState)
//
//	// add to flowManager maps
//	//flowManager.statesMap[step90.Index()] = step90
//	//flowManager.nameStateMap[step90.ActionName()] = step90
//}
//
//func (flowManager *iFlowManagerImpl) createStep81() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	// Create Finalize Status
//	finalizeState := finalize_state.New(1, emptyState, emptyState,
//		finalize_action.NewOf(finalize_action.BuyerFinalizeAction))
//
//	// Create Notification Status
//	notificationState := notification_state.New(0, []states_old.IState{finalizeState}, emptyState,
//		notification_action.NewOf(notification_action.BuyerNotificationAction))
//	step81 := pay_to_buyer_success_step.New(emptyStep, emptyStep, notificationState, finalizeState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step81.Index()] = step81
//	flowManager.nameStateMap[step81.Name()] = step81
//}
//
//// TODO: checking flow and next to step states_old sequence
//func (flowManager *iFlowManagerImpl) createStep82() {
//	var emptyState []states_old.IState
//
//	step80 := flowManager.statesMap[80]
//	step81 := flowManager.statesMap[81]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		retry_action.RetryAction:                         step80,
//		manual_payment_action.ManualPaymentToBuyerAction: step81,
//	}
//
//	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
//		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//
//	manualPaymentState := manual_payment_state.New(2, []states_old.IState{nextToStep}, emptyState,
//		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToBuyerAction))
//	operatorNotificationState := notification_state.New(1, []states_old.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
//	retryState := retry_state.New(0, []states_old.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))
//
//	step82 := pay_to_buyer_failed_step.New([]states.IStep{step81}, []states.IStep{step80},
//		retryState, operatorNotificationState, manualPaymentState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step82.Index()] = step82
//	flowManager.nameStateMap[step82.Name()] = step82
//}
//
//// TODO release stock must be call once as a result it must be save in db
////func (flowManager *iFlowManagerImpl) createStep80(baseStep80 states.IBaseStep) {
////	var emptyState []states_old.IState
////	var emptyStep []states.IStep
////
////	step81 := flowManager.statesMap[81]
////	step82 := flowManager.statesMap[82]
////
////	payToBuyerActions := pay_to_buyer_action.NewOf(pay_to_buyer_action.SuccessAction, pay_to_buyer_action.FailedAction)
////	actionStepMap := map[actions.IEnumAction]states.IStep{
////		pay_to_buyer_action.SuccessAction: step81,
////		pay_to_buyer_action.FailedAction:  step82,
////	}
////
////	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
////	payToBuyerState := pay_to_buyer_state.New(1, []states_old.IState{nextToStepState}, emptyState, payToBuyerActions)
////	stockReleaseActionState := stock_action_state.New(0, []states_old.IState{payToBuyerState}, emptyState, stock_action.NewOf(stock_action.ReleasedAction))
////
////	// TODO check change baseStep80 cause of change Step82 settings
////	baseStep80.BaseStep().SetChildes([]states.IStep{step81, step82})
////	baseStep80.BaseStep().SetParents(emptyStep)
////	baseStep80.BaseStep().SetStates(stockReleaseActionState, payToBuyerState, nextToStepState)
////
////	// add to flowManager maps
////	//flowManager.statesMap[step93.Index()] = step93
////	//flowManager.nameStateMap[step93.ActionName()] = step93
////}
//
//func (flowManager *iFlowManagerImpl) createStep55() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step80 := flowManager.statesMap[80]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		next_to_step_action.NextToStepAction: step80,
//	}
//
//	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//
//	step55 := return_shipment_success_step.New([]states.IStep{step80}, emptyStep, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step55.Index()] = step55
//	flowManager.nameStateMap[step55.Name()] = step55
//}
//
//func (flowManager *iFlowManagerImpl) createStep54() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step90 := flowManager.statesMap[90]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		next_to_step_action.NextToStepAction: step90,
//	}
//
//	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	step54 := return_shipment_canceled_step.New([]states.IStep{step90}, emptyStep, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step54.Index()] = step54
//	flowManager.nameStateMap[step54.Name()] = step54
//}
//
//func (flowManager *iFlowManagerImpl) createStep53() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step55 := flowManager.statesMap[55]
//	step54 := flowManager.statesMap[54]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		operator_action.ReturnCanceledAction: step54,
//		operator_action.ReturnedAction:       step55,
//	}
//
//	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnCanceledAction, operator_action.ReturnedAction))
//	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
//
//	step53 := return_shipment_delivery_problem_step.New([]states.IStep{step54, step55}, emptyStep,
//		notificationState, operatorActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step53.Index()] = step53
//	flowManager.nameStateMap[step53.Name()] = step53
//}
//
//func (flowManager *iFlowManagerImpl) createStep50() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step53 := flowManager.statesMap[53]
//	step55 := flowManager.statesMap[55]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		seller_action.NeedSupportAction:     step53,
//		seller_action.ApprovedAction:        step55,
//		scheduler_action.AutoApprovedAction: step55,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.AutoApprovedAction))
//	sellerApprovedActionState := seller_action_state.New("Return_Seller_Approval_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.ApprovedAction, seller_action.NeedSupportAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerApprovedActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))
//
//	step50 := return_shipment_delivered_step.New([]states.IStep{step53, step55}, emptyStep,
//		notificationActionState, composeActorsActionState, sellerApprovedActionState, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step50.Index()] = step50
//	flowManager.nameStateMap[step50.Name()] = step50
//}
//
//func (flowManager *iFlowManagerImpl) createStep52() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step50 := flowManager.statesMap[50]
//	step54 := flowManager.statesMap[54]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		operator_action.ReturnDeliveredAction: step50,
//		operator_action.ReturnCanceledAction:  step54,
//	}
//
//	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnCanceledAction, operator_action.ReturnDeliveredAction))
//	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
//
//	step52 := return_shipment_delivery_delayed_step.New([]states.IStep{step54, step50}, emptyStep,
//		notificationState, operatorActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step52.Index()] = step52
//	flowManager.nameStateMap[step52.Name()] = step52
//}
//
//// TODO schedulers need a config for timeout or auto approved
//func (flowManager *iFlowManagerImpl) createStep51() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step50 := flowManager.statesMap[50]
//	step52 := flowManager.statesMap[52]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		buyer_action.NeedSupportAction:                 step52,
//		scheduler_action.NoActionForXDaysTimeoutAction: step50,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
//	buyerNeedSupportActionState := buyer_action_state.New("Buyer_Return_Shipment_Delivery_Pending_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.NeedSupportAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerNeedSupportActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))
//
//	step51 := return_shipment_delivery_pending_step.New([]states.IStep{step50, step52}, emptyStep,
//		notificationActionState, composeActorsActionState, buyerNeedSupportActionState, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step51.Index()] = step51
//	flowManager.nameStateMap[step51.Name()] = step51
//}
//
//func (flowManager *iFlowManagerImpl) createStep42() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step51 := flowManager.statesMap[51]
//	step50 := flowManager.statesMap[50]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		buyer_action.DeliveredAction:                      step50,
//		scheduler_action.WaitForShippingDaysTimeoutAction: step51,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.WaitForShippingDaysTimeoutAction))
//	buyerReturnShipmentDeliveryActionState := buyer_action_state.New("Buyer_Return_Shipment_Delivery_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.DeliveredAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerReturnShipmentDeliveryActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))
//
//	step42 := return_shipped_step.New([]states.IStep{step50, step51}, emptyStep,
//		notificationActionState, composeActorsActionState, buyerReturnShipmentDeliveryActionState, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step42.Index()] = step42
//	flowManager.nameStateMap[step42.Name()] = step42
//}
//
//func (flowManager *iFlowManagerImpl) createStep40() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step90 := flowManager.statesMap[90]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		next_to_step_action.NextToStepAction: step90,
//	}
//
//	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	step40 := shipment_success_step.New([]states.IStep{step90}, emptyStep, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step40.Index()] = step40
//	flowManager.nameStateMap[step40.Name()] = step40
//}
//
//func (flowManager *iFlowManagerImpl) createStep44() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step40 := flowManager.statesMap[40]
//	step42 := flowManager.statesMap[42]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		buyer_action.EnterReturnShipmentDetailAction: step42,
//		scheduler_action.WaitXDaysTimeoutAction:      step40,
//	}
//
//	nextToStepState := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStepState}, emptyState, scheduler_action.NewOf(scheduler_action.WaitXDaysTimeoutAction))
//	buyerEnterReturnShipmentDetailDelayedActionState := buyer_action_state.New("Buyer_Enter_Return_Shipment_Detail_Delayed_Action_State", 2, []states_old.IState{nextToStepState}, emptyState, buyer_action.NewOf(buyer_action.EnterReturnShipmentDetailAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerEnterReturnShipmentDetailDelayedActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))
//
//	step44 := return_shipment_detail_delayed_step.New([]states.IStep{step40, step42}, emptyStep,
//		notificationState, composeActorsActionState, buyerEnterReturnShipmentDetailDelayedActionState,
//		schedulerActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step44.Index()] = step44
//	flowManager.nameStateMap[step44.Name()] = step44
//}
//
//func (flowManager *iFlowManagerImpl) createStep41() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step42 := flowManager.statesMap[42]
//	step44 := flowManager.statesMap[44]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		buyer_action.EnterReturnShipmentDetailAction:   step42,
//		scheduler_action.NoActionForXDaysTimeoutAction: step44,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
//	BuyerEnterReturnShipmentDetailAction := buyer_action_state.New("Buyer_Enter_Return_Shipment_Detail_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.EnterReturnShipmentDetailAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{BuyerEnterReturnShipmentDetailAction, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))
//
//	step41 := return_shipment_pending_step.New([]states.IStep{step42, step44}, emptyStep,
//		notificationActionState, composeActorsActionState, BuyerEnterReturnShipmentDetailAction, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step41.Index()] = step41
//	flowManager.nameStateMap[step41.Name()] = step41
//}
//
//func (flowManager *iFlowManagerImpl) createStep43() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step41 := flowManager.statesMap[41]
//	step40 := flowManager.statesMap[40]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		operator_action.ReturnedAction:       step41,
//		operator_action.ReturnCanceledAction: step40,
//	}
//
//	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnedAction, operator_action.ReturnCanceledAction))
//	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
//
//	step43 := shipment_delivery_problem_step.New([]states.IStep{step40, step41}, emptyStep,
//		notificationState, operatorActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step43.Index()] = step43
//	flowManager.nameStateMap[step43.Name()] = step43
//}
//
//func (flowManager *iFlowManagerImpl) createStep36() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step80 := flowManager.statesMap[80]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		next_to_step_action.NextToStepAction: step80,
//	}
//
//	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	step36 := shipment_canceled_step.New([]states.IStep{step80}, emptyStep, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step36.Index()] = step36
//	flowManager.nameStateMap[step36.Name()] = step36
//}
//
//func (flowManager *iFlowManagerImpl) createStep32() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step43 := flowManager.statesMap[43]
//	step41 := flowManager.statesMap[41]
//	step40 := flowManager.statesMap[40]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		scheduler_action.AutoApprovedAction: step40,
//		buyer_action.ApprovedAction:         step40,
//		buyer_action.NeedSupportAction:      step43,
//		buyer_action.ReturnIfPossibleAction: step41,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.AutoApprovedAction))
//	buyerApprovalActionState := buyer_action_state.New("Buyer_Approval_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.ApprovedAction, buyer_action.NeedSupportAction, buyer_action.ReturnIfPossibleAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerApprovalActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))
//
//	step32 := shipment_delivered_step.New([]states.IStep{step43, step41, step40}, emptyStep,
//		notificationActionState, composeActorsActionState, buyerApprovalActionState, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step32.Index()] = step32
//	flowManager.nameStateMap[step32.Name()] = step32
//}
//
//func (flowManager *iFlowManagerImpl) createStep35() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step36 := flowManager.statesMap[36]
//	step32 := flowManager.statesMap[32]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		operator_action.CanceledAction:  step36,
//		operator_action.DeliveredAction: step32,
//	}
//
//	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.DeliveredAction, operator_action.CanceledAction))
//	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
//
//	step35 := shipment_delivery_delayed_step.New([]states.IStep{step32, step36}, emptyStep,
//		notificationState, operatorActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step35.Index()] = step35
//	flowManager.nameStateMap[step35.Name()] = step35
//}
//
//func (flowManager *iFlowManagerImpl) createStep34() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step35 := flowManager.statesMap[35]
//	step32 := flowManager.statesMap[32]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		seller_action.NeedSupportAction:                step35,
//		scheduler_action.NoActionForXDaysTimeoutAction: step32,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
//	sellerShipmentDeliveryPendingActionState := seller_action_state.New("Seller_Shipment_Delivery_Pending_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.NeedSupportAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerShipmentDeliveryPendingActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))
//
//	step34 := shipment_delivery_pending_step.New([]states.IStep{step35, step32}, emptyStep,
//		notificationActionState, composeActorsActionState, sellerShipmentDeliveryPendingActionState, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step34.Index()] = step34
//	flowManager.nameStateMap[step34.Name()] = step34
//}
//
//func (flowManager *iFlowManagerImpl) createStep31() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step32 := flowManager.statesMap[32]
//	step34 := flowManager.statesMap[34]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		seller_action.DeliveredAction:                     step32,
//		scheduler_action.WaitForShippingDaysTimeoutAction: step34,
//	}
//
//	nextToStep := next_to_step_state.New(5, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(4, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.WaitForShippingDaysTimeoutAction))
//	sellerShipmentDeliveryActionState := buyer_action_state.New("Seller_Shipment_Delivery_Action_State", 3, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.DeliveredAction))
//	composeActorsActionState := system_action_state.New(2, []states_old.IState{sellerShipmentDeliveryActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(1, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))
//	stockSettlementActionState := stock_action_state.New(0, []states_old.IState{notificationActionState}, emptyState, stock_action.NewOf(stock_action.SettlementAction))
//
//	step31 := shipped_step.New([]states.IStep{step32, step34}, emptyStep,
//		stockSettlementActionState, notificationActionState, composeActorsActionState,
//		sellerShipmentDeliveryActionState, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step31.Index()] = step31
//	flowManager.nameStateMap[step31.Name()] = step31
//}
//
//func (flowManager *iFlowManagerImpl) createStep33() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step31 := flowManager.statesMap[31]
//	step36 := flowManager.statesMap[36]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		seller_action.EnterShipmentDetailAction: step31,
//		buyer_action.CanceledAction:             step36,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	buyerWaitForSellerEnterShipmentDetailDelayedActionState := buyer_action_state.New("Buyer_Wait_For_Seller_Enter_Shipment_Detail_Delayed_Action_State", 3, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.CanceledAction))
//	sellerEnterShipmentDetailDelayedActionState := seller_action_state.New("Seller_Enter_Shipment_Detail_Delayed_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.EnterShipmentDetailAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerWaitForSellerEnterShipmentDetailDelayedActionState, sellerEnterShipmentDetailDelayedActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction, notification_action.BuyerNotificationAction))
//
//	step33 := shipment_detail_delayed_step.New([]states.IStep{step31, step36}, emptyStep,
//		notificationActionState, composeActorsActionState, sellerEnterShipmentDetailDelayedActionState, buyerWaitForSellerEnterShipmentDetailDelayedActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step33.Index()] = step33
//	flowManager.nameStateMap[step33.Name()] = step33
//}
//
//func (flowManager *iFlowManagerImpl) createStep30() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step31 := flowManager.statesMap[31]
//	step33 := flowManager.statesMap[33]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		seller_action.EnterShipmentDetailAction:        step31,
//		scheduler_action.NoActionForXDaysTimeoutAction: step33,
//	}
//
//	nextToStepState := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStepState}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
//	sellerEnterShipmentDetailActionState := seller_action_state.New("Seller_Enter_Shipment_Detail_Action_State", 2, []states_old.IState{nextToStepState}, emptyState, seller_action.NewOf(seller_action.EnterShipmentDetailAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerEnterShipmentDetailActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))
//
//	step30 := shipment_pending_step.New([]states.IStep{step31, step33}, emptyStep,
//		notificationState, composeActorsActionState, sellerEnterShipmentDetailActionState,
//		schedulerActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step30.Index()] = step30
//	flowManager.nameStateMap[step30.Name()] = step30
//}
//
//func (flowManager *iFlowManagerImpl) createStep21() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step80 := flowManager.statesMap[80]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		next_to_step_action.NextToStepAction: step80,
//	}
//
//	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	step21 := shipment_rejected_by_seller_step.New([]states.IStep{step80}, emptyStep, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step21.Index()] = step21
//	flowManager.nameStateMap[step21.Name()] = step21
//}
//
//func (flowManager *iFlowManagerImpl) createStep20() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step30 := flowManager.statesMap[30]
//	step21 := flowManager.statesMap[21]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		seller_action.ApprovedAction:                   step30,
//		seller_action.RejectAction:                     step21,
//		scheduler_action.NoActionForXDaysTimeoutAction: step21,
//	}
//
//	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
//	sellerApprovalActionState := seller_action_state.New("Seller_Approval_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.ApprovedAction, seller_action.RejectAction))
//	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerApprovalActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
//	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))
//
//	step20 := seller_approval_pending_step.New([]states.IStep{step30, step21}, emptyStep,
//		notificationActionState, composeActorsActionState, sellerApprovalActionState, schedulerActionState, nextToStep)
//
//	// add to flowManager maps
//	flowManager.statesMap[step20.Index()] = step20
//	flowManager.nameStateMap[step20.Name()] = step20
//}
//
//func (flowManager *iFlowManagerImpl) createStep14() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step80 := flowManager.statesMap[80]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		next_to_step_action.NextToStepAction: step80,
//	}
//
//	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	step14 := payment_rejected_step.New([]states.IStep{step80}, emptyStep, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step14.Index()] = step14
//	flowManager.nameStateMap[step14.Name()] = step14
//}
//
//func (flowManager *iFlowManagerImpl) createStep11() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step14 := flowManager.statesMap[14]
//	step20 := flowManager.statesMap[20]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		buyer_action.RejectAction:   step14,
//		buyer_action.ApprovedAction: step20,
//	}
//
//	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	buyerPaymentApprovalActionState := buyer_action_state.New("Buyer_Payment_Approval_Action_State", 1, []states_old.IState{nextToStepState}, emptyState, buyer_action.NewOf(buyer_action.ApprovedAction, buyer_action.RejectAction))
//	notificationState := notification_state.New(0, []states_old.IState{buyerPaymentApprovalActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))
//
//	step11 := payment_success_step.New([]states.IStep{step14, step20}, emptyStep,
//		notificationState, buyerPaymentApprovalActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step11.Index()] = step11
//	flowManager.nameStateMap[step11.Name()] = step11
//}
//
//func (flowManager *iFlowManagerImpl) createStep12() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	// Create Finalize Status
//	finalizeState := finalize_state.New(2, emptyState, emptyState,
//		finalize_action.NewOf(finalize_action.PaymentFailedFinalizeAction))
//
//	// Create Notification Status
//	notificationState := notification_state.New(1, []states_old.IState{finalizeState}, emptyState,
//		notification_action.NewOf(notification_action.BuyerNotificationAction))
//
//	stockReleaseActionState := stock_action_state.New(0, []states_old.IState{notificationState, finalizeState}, emptyState, stock_action.NewOf(stock_action.ReleasedAction, stock_action.FailedAction))
//	step12 := payment_failed_step.New(emptyStep, emptyStep,
//		stockReleaseActionState, notificationState, finalizeState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step12.Index()] = step12
//	flowManager.nameStateMap[step12.Name()] = step12
//}
//
//func (flowManager *iFlowManagerImpl) createStep10() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step12 := flowManager.statesMap[12]
//	step11 := flowManager.statesMap[11]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		order_payment_action.OrderPaymentFailedAction: step12,
//		payment_action.FailedAction:                   step12,
//		payment_action.SuccessAction:                  step11,
//	}
//
//	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	paymentActionState := payment_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, payment_action.NewOf(payment_action.SuccessAction, payment_action.FailedAction))
//	orderPaymentActionState := order_payment_action_state.New(0, []states_old.IState{paymentActionState, nextToStepState}, emptyState, order_payment_action.NewOf(order_payment_action.OrderPaymentAction, order_payment_action.OrderPaymentFailedAction))
//
//	step10 := payment_pending_step.New([]states.IStep{step12, step11}, emptyStep,
//		orderPaymentActionState, paymentActionState, nextToStepState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step10.Index()] = step10
//	flowManager.nameStateMap[step10.Name()] = step10
//}
//
//func (flowManager *iFlowManagerImpl) createStep1() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	// Create Finalize Status
//	finalizeState := finalize_state.New(1, emptyState, emptyState,
//		finalize_action.NewOf(finalize_action.OrderFailedFinalizeAction))
//
//	step1 := new_order_failed_step.New(emptyStep, emptyStep, finalizeState)
//
//	// add to flowManager maps
//	flowManager.statesMap[step1.Index()] = step1
//	flowManager.nameStateMap[step1.Name()] = step1
//}
//
//func (flowManager *iFlowManagerImpl) createStep0() {
//	var emptyState []states_old.IState
//	var emptyStep []states.IStep
//
//	step1 := flowManager.statesMap[1]
//	step10 := flowManager.statesMap[10]
//
//	actionStepMap := map[actions.IEnumAction]states.IStep{
//		stock_action.FailedAction:   step1,
//		stock_action.ReservedAction: step10,
//	}
//
//	nextToStepState := next_to_step_state.New(3, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
//	stockReservedActionState := stock_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, stock_action.NewOf(stock_action.ReservedAction, stock_action.FailedAction))
//	checkoutStateAction := checkout_action_state.New(0, []states_old.IState{stockReservedActionState, nextToStepState}, emptyState, checkout_action.NewOf(checkout_action.NewOrderAction))
//
//	step0 := new_order_step.New([]states.IStep{step1, step10}, emptyStep,
//		checkoutStateAction, stockReservedActionState, nextToStepState)
//	// add to flowManager maps
//	flowManager.statesMap[step0.Index()] = step0
//	flowManager.nameStateMap[step0.Name()] = step0
//}

// TODO Must be refactored
//func (flowManager iFlowManagerImpl) MessageHandler(ctx context.Context, req *message.MessageRequest) future.IPromise {
//	// received New Order Request
//	//if len(req.OrderId) == 0 {
//
//	step0 := flowManager.statesMap[0]
//	return step0.ProcessMessage(ctx, req)
//
//	//var requestNewOrder pb.RequestNewOrder
//	//if err := ptypes.UnmarshalAny(request.Data, &requestNewOrder); err != nil {
//	//	logger.Err("Could not unmarshal requestNewOrder from anything field, error: %s, request: %v", err, request)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.BadRequest, Reason: "Invalid requestNewOrder"}}
//	//	close(returnChannel)
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//
//	//timestamp, err := ptypes.Timestamp(request.Time)
//	//if err != nil {
//	//	logger.Err("timestamp of requestNewOrder invalid, error: %s, requestNewOrder: %v", err, requestNewOrder)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.BadRequest, Reason:"Invalid Request Timestamp"}}
//	//	defer close(returnChannel)
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//
//	//value, err := global.Singletons.Converter.Map(requestNewOrder, entities.Order{})
//	//if err != nil {
//	//	logger.Err("Converter.Map requestNewOrder to order object failed, error: %s, requestNewOrder: %v", err, requestNewOrder)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.BadRequest, Reason: "Received requestNewOrder invalid"}}
//	//	defer close(returnChannel)
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//	//
//	//state := value.(*entities.Order)
//	//newOrderEvent := actor_event.NewActorEvent(actors.CheckoutActor, checkout_action.NewOf(checkout_action.NewOrderAction),
//	//	state, nil, nil, timestamp)
//	//
//	//checkoutState, ok := state.StatesMap()[0].(listener_state.IListenerState)
//	//if ok != true || checkoutState.ActorType() != actors.CheckoutActor {
//	//	logger.Err("checkout state doesn't exist in index 0 of statesMap, requestNewOrder: %v", requestNewOrder)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	returnChannel <- future.IDataFuture{Get:nil, Ex:future.FutureError{Code: future.InternalError, Reason:"Unknown Error"}}
//	//	defer close(returnChannel)
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//
//	//state.UpdateAllOrderStatus(ctx, state, nil, states.OrderNewStatus, false)
//	//order, err := global.Singletons.OrderRepository.Save(*state)
//	//if err != nil {
//	//	logger.Err("Save NewOrder Step Failed, error: %s, order: %v", err, state)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//	//	defer close(returnChannel)
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//
//	//iPromise := global.Singletons.StockService.BatchStockActions(ctx, *order, nil, StockReserved)
//	//futureData := iPromise.Get()
//	//if futureData == nil {
//	//	state.UpdateAllOrderStatus(ctx, order, nil, states.OrderClosedStatus, true)
//	//	state.updateOrderItemsProgress(ctx, order, nil, StockReserved, false, states.OrderClosedStatus)
//	//	if err := state.persistOrder(ctx, order); err != nil {
//	//	}
//	//	logger.Err("StockService future channel has been closed, order: %v", order)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	defer close(returnChannel)
//	//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//	//	go func() {
//	//		state.Childes()[0].ProcessOrder(ctx, *order, nil, nil)
//	//	}()
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//	//
//	//if futureData.Ex != nil {
//	//	state.UpdateAllOrderStatus(ctx, order, nil, states.OrderClosedStatus, true)
//	//	state.updateOrderItemsProgress(ctx, order, nil, StockReserved, false, states.OrderClosedStatus)
//	//	if err := state.persistOrder(ctx, order); err != nil {
//	//	}
//	//	logger.Err("Reserved stock from stockService failed, error: %s, orderId: %d", futureData.Ex.Error(), order.OrderId)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	defer close(returnChannel)
//	//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//	//	go func() {
//	//		state.Childes()[0].ProcessOrder(ctx, *order, nil, nil)
//	//	}()
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//	//
//	//state.updateOrderItemsProgress(ctx, order, nil, StockReserved, true, states.OrderNewStatus)
//	//if err := state.persistOrder(ctx, order); err != nil {
//	//	state.releasedStock(ctx, order)
//	//	returnChannel := make(chan future.IDataFuture, 1)
//	//	defer close(returnChannel)
//	//	returnChannel <- future.IDataFuture{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//	//
//	//	go func() {
//	//		state.Childes()[0].ProcessOrder(ctx, *order, nil, nil)
//	//	}()
//	//
//	//	return future.NewFuture(returnChannel, 1, 1)
//	//}
//	//
//	//return state.Childes()[1].ProcessOrder(ctx, *order, nil, "PaymentCallbackUrlRequest")
//	////return checkoutState.ActionListener(ctx, newOrderEvent, nil)
//
//	// TODO must be implement
//	//}
//	//} else if len(req.ItemId) != 0 {
//
//	//} else {
//	//	order, err := global.Singletons.OrderRepository.FindById(req.OrderId)
//	//	if err != nil {
//	//		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
//	//			req.OrderId, err)
//	//		returnChannel := make(chan future.FutureData, 1)
//	//		returnChannel <- future.FutureData{Get:nil, Ex:future.FutureError{Code: future.NotFound, Reason:"OrderId Not Found"}}
//	//		defer close(returnChannel)
//	//		return future.NewPromise(returnChannel, 1, 1)
//	//	}
//	//}
//}

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
//		if order.Items[i].SellerInfo.SellerId == req.SellerId {
//			itemsId = append(itemsId, order.Items[i].ItemId)
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
//	//	if order.Items[i].SellerInfo.SellerId == req.ItemId[i] {
//	//		itemsId = append(itemsId,order.Items[i].ItemId)
//	//	}
//	//}
//
//	if req.ActionType == "Approved" {
//		return flowManager.statesMap[32].ProcessOrder(ctx, *order, req.ItemsId, req)
//	}
//
//	returnChannel := make(chan future.FutureData, 1)
//	returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "ActionType Not Found"}}
//	defer close(returnChannel)
//	return future.NewPromise(returnChannel, 1, 1)
//}

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

//func (flowManager iFlowManagerImpl) OperatorActionPending(ctx context.Context, req *message.RequestBackOfficeOrderAction) future.IPromise {
//	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
//		return bson.D{{"items.itemId", req.ItemId}}
//	})
//
//	if err != nil {
//		logger.Err("MessageHandler() => request itemId not found, OrderRepository.FindById failed, itemId: %d, error: %s",
//			req.ItemId, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if len(orders) == 0 {
//		logger.Err("MessageHandler() => request itemId not found, itemId: %d", req.ItemId)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.NotFound, Reason: "ItemId Not Found"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	if len(orders) > 1 {
//		logger.Err("MessageHandler() => request itemId found in multiple order, itemId: %d, error: %s",
//			req.ItemId, err)
//		returnChannel := make(chan future.FutureData, 1)
//		returnChannel <- future.FutureData{Data: nil, Ex: future.FutureError{Code: future.InternalError, Reason: "Unknown Error"}}
//		defer close(returnChannel)
//		return future.NewPromise(returnChannel, 1, 1)
//	}
//
//	itemsId := make([]uint64, 0, 1)
//	for i := 0; i < len(orders[0].Items); i++ {
//		if orders[0].Items[i].ItemId == req.ItemId {
//			itemsId = append(itemsId, orders[0].Items[i].ItemId)
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
//			ItemId:      item.ItemId,
//			SellerId:    item.SellerInfo.SellerId,
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
//			logger.Audit("BackOfficeOrderDetailView() => Actions History is nil, orderId: %d, itemId: %d", order.OrderId, item.ItemId)
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

//func (flowManager iFlowManagerImpl) SellerReportOrders(req *message.RequestSellerReportOrders, srv message.OrderService_SellerReportOrdersServer) future.IPromise {
//	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
//		return bson.D{{"createdAt",
//			bson.D{{"$gte", time.Unix(int64(req.StartDateTime), 0).UTC()}}},
//			{"items.status", req.Status}, {"items.sellerInfo.sellerId", req.SellerId}}
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
//					ItemId:      item.ItemId,
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
//		"OrderId", "ItemId", "ProductId", "InventoryId",
//		"PaidPrice", "Commission", "Category", "Status", "CreatedAt", "UpdatedAt",
//	}
//
//	csvReports = append(csvReports, csvHeadLines)
//	for _, itemReport := range reports {
//		csvRecord := []string{
//			strconv.Itoa(int(itemReport.OrderId)),
//			strconv.Itoa(int(itemReport.ItemId)),
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
//		logger.Err("BackOfficeReportOrderItems() => request itemId not found, OrderRepository.FindById failed, startDateTime: %v, endDateTime: %v, error: %s",
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
//				ItemId:      item.ItemId,
//				InventoryId: item.InventoryId,
//				ProductId:   item.InventoryId[0:8],
//				BuyerId:     order.BuyerInfo.BuyerId,
//				BuyerPhone:  order.BuyerInfo.Phone,
//				SellerId:    item.SellerInfo.SellerId,
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
//			if sellerProfile, ok := sellerProfileMap[item.SellerInfo.SellerId]; !ok {
//				userCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
//				ipromise := global.Singletons.UserService.GetSellerProfile(userCtx, strconv.Itoa(int(item.SellerInfo.SellerId)))
//				futureData := ipromise.Get()
//				if futureData.Ex != nil {
//					logger.Err("BackOfficeReportOrderItems() => get sellerProfile failed, orderId: %d, itemId: %d, sellerId: %d",
//						order.OrderId, item.ItemId, item.SellerInfo.SellerId)
//					continue
//				}
//
//				sellerInfo, ok := futureData.Data.(entities.SellerProfile)
//				if ok != true {
//					logger.Err("BackOfficeReportOrderItems() => get sellerProfile invalid, orderId: %d, itemId: %d, sellerId: %d",
//						order.OrderId, item.ItemId, item.SellerInfo.SellerId)
//					continue
//				}
//
//				sellerProfileMap[item.SellerInfo.SellerId] = sellerProfile
//				itemReport.SellerName = sellerInfo.GeneralInfo.ShopDisplayName
//			} else {
//				itemReport.SellerName = sellerProfile.GeneralInfo.ShopDisplayName
//			}
//		}
//	}
//
//	csvReports := make([][]string, 0, len(reports))
//	csvHeadLines := []string{
//		"ItemId", "InventoryId", "ProductId", "BuyerId", "BuyerPhone", "SellerId",
//		"SellerName", "ItemInvoice", "Status", "CreatedAt", "UpdatedAt",
//	}
//
//	csvReports = append(csvReports, csvHeadLines)
//	for _, itemReport := range reports {
//		csvRecord := []string{
//			strconv.Itoa(int(itemReport.ItemId)),
//			itemReport.InventoryId,
//			itemReport.ProductId,
//			strconv.Itoa(int(itemReport.BuyerId)),
//			itemReport.BuyerPhone,
//			strconv.Itoa(int(itemReport.SellerId)),
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
//	//for i:= 0; i < len(event.ItemsId); i++ {
//	//	if order.Items[i].ItemId == event.ItemsId[i] && order.Items[i].SellerInfo.SellerId == event.SellerId {
//	//		itemsId = append(itemsId,order.Items[i].ItemId)
//	//	}
//	//}
//
//	ctx, _ := context.WithCancel(context.Background())
//
//	if event.Action == "ApprovalPending" {
//		flowManager.statesMap[20].ProcessOrder(ctx, *order, event.ItemsId, "actionExpired")
//
//	} else if event.Action == "SellerShipmentPending" {
//		flowManager.statesMap[30].ProcessOrder(ctx, *order, event.ItemsId, "actionExpired")
//
//	} else if event.Action == "ShipmentDeliveredPending" {
//		flowManager.statesMap[32].ProcessOrder(ctx, *order, event.ItemsId, "actionApproved")
//	}
//
//}
