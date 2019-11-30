package domain

import (
	"context"
	"encoding/csv"
	"fmt"
	ptime "github.com/yaa110/go-persian-calendar"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	reports2 "gitlab.faza.io/order-project/order-service/domain/models/reports"
	order_payment_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/orderpayment"
	"gitlab.faza.io/order-project/order-service/infrastructure/global"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"os"
	"strconv"
	"time"

	//"github.com/pkg/errors"
	checkout_action "gitlab.faza.io/order-project/order-service/domain/actions/checkout"
	checkout_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener/checkout"
	"gitlab.faza.io/order-project/order-service/infrastructure/promise"
	//pb "gitlab.faza.io/protos/order"
	////"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
	pg "gitlab.faza.io/protos/payment-gateway"
	//"github.com/golang/protobuf/ptypes"
	//"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/buyer"
	notification_action "gitlab.faza.io/order-project/order-service/domain/actions/notification"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/operator"
	payment_action "gitlab.faza.io/order-project/order-service/domain/actions/payment"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/seller"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/stock"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/system"
	"gitlab.faza.io/order-project/order-service/domain/states"
	shipment_delivery_problem_step "gitlab.faza.io/order-project/order-service/domain/states/obsolete_state_43"
	new_order_step "gitlab.faza.io/order-project/order-service/domain/states/state_0"
	new_order_failed_step "gitlab.faza.io/order-project/order-service/domain/states/state_01"
	payment_pending_step "gitlab.faza.io/order-project/order-service/domain/states/state_10"
	payment_success_step "gitlab.faza.io/order-project/order-service/domain/states/state_11"
	payment_failed_step "gitlab.faza.io/order-project/order-service/domain/states/state_12"
	payment_rejected_step "gitlab.faza.io/order-project/order-service/domain/states/state_14"
	seller_approval_pending_step "gitlab.faza.io/order-project/order-service/domain/states/state_20"
	shipment_rejected_by_seller_step "gitlab.faza.io/order-project/order-service/domain/states/state_21"
	shipment_pending_step "gitlab.faza.io/order-project/order-service/domain/states/state_30"
	shipped_step "gitlab.faza.io/order-project/order-service/domain/states/state_31"
	shipment_delivered_step "gitlab.faza.io/order-project/order-service/domain/states/state_32"
	shipment_detail_delayed_step "gitlab.faza.io/order-project/order-service/domain/states/state_33"
	shipment_delivery_pending_step "gitlab.faza.io/order-project/order-service/domain/states/state_34"
	shipment_delivery_delayed_step "gitlab.faza.io/order-project/order-service/domain/states/state_35"
	shipment_canceled_step "gitlab.faza.io/order-project/order-service/domain/states/state_36"
	shipment_success_step "gitlab.faza.io/order-project/order-service/domain/states/state_40"
	return_shipment_pending_step "gitlab.faza.io/order-project/order-service/domain/states/state_41"
	return_shipped_step "gitlab.faza.io/order-project/order-service/domain/states/state_42"
	return_shipment_detail_delayed_step "gitlab.faza.io/order-project/order-service/domain/states/state_44"
	return_shipment_delivered_step "gitlab.faza.io/order-project/order-service/domain/states/state_50"
	return_shipment_delivery_pending_step "gitlab.faza.io/order-project/order-service/domain/states/state_51"
	return_shipment_delivery_delayed_step "gitlab.faza.io/order-project/order-service/domain/states/state_52"
	return_shipment_delivery_problem_step "gitlab.faza.io/order-project/order-service/domain/states/state_53"
	return_shipment_canceled_step "gitlab.faza.io/order-project/order-service/domain/states/state_54"
	return_shipment_success_step "gitlab.faza.io/order-project/order-service/domain/states/state_55"
	pay_to_buyer_step "gitlab.faza.io/order-project/order-service/domain/states/state_80"
	pay_to_seller_step "gitlab.faza.io/order-project/order-service/domain/states/state_90"
	"gitlab.faza.io/order-project/order-service/domain/states_old"
	finalize_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/finalize"
	manual_payment_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/manualpayment"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher/nextstep"
	notification_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/notification"
	pay_to_buyer_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/paytobuyer"
	pay_to_market_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/paytomarket"
	pay_to_seller_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/paytoseller"
	retry_state "gitlab.faza.io/order-project/order-service/domain/states_old/launcher/retry"
	"gitlab.faza.io/order-project/order-service/domain/states_old/launcher/stock"
	buyer_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener/buyer"
	operator_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener/operator"
	payment_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener/payment"
	scheduler_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener/scheduler"
	seller_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener/seller"
	system_action_state "gitlab.faza.io/order-project/order-service/domain/states_old/listener/system"
	message "gitlab.faza.io/protos/order"
)

type iFlowManagerImpl struct {
	nameStepsMap  map[string]states.IStep
	indexStepsMap map[int]states.IStep
}

func NewFlowManager() (IFlowManager, error) {
	nameStepsMap := make(map[string]states.IStep, 64)
	indexStepsMap := make(map[int]states.IStep, 64)

	iFlowManagerImpl := &iFlowManagerImpl{nameStepsMap, indexStepsMap}
	if err := iFlowManagerImpl.setupFlowManager(); err != nil {
		return nil, err
	}

	return iFlowManagerImpl, nil
}

func (flowManager *iFlowManagerImpl) GetNameStepsMap() map[string]states.IStep {
	return flowManager.nameStepsMap
}

func (flowManager *iFlowManagerImpl) GetIndexStepsMap() map[int]states.IStep {
	return flowManager.indexStepsMap
}

func (flowManager *iFlowManagerImpl) setupFlowManager() error {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	//////////////////////////////////////////////////////////////////
	// Pay To Market
	// create empty step93 that required for step95
	step93 := pay_to_market_step.New(emptyStep, emptyStep, emptyState...)
	baseStep93 := step93.(states.IBaseStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step93.Index()] = step93
	flowManager.nameStepsMap[step93.Name()] = step93

	flowManager.createStep94()
	flowManager.createStep95()
	flowManager.createStep93(baseStep93)

	//////////////////////////////////////////////////////////////////
	// Pay To SellerInfo
	// create empty step90 which is required for step92
	step90 := pay_to_seller_step.New(emptyStep, emptyStep, emptyState...)
	baseStep90 := step90.(states.IBaseStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step90.Index()] = step90
	flowManager.nameStepsMap[step90.Name()] = step90

	flowManager.createStep91()
	flowManager.createStep92()
	flowManager.createStep90(baseStep90)

	//////////////////////////////////////////////////////////////////
	// Pay To Buyer
	// create empty step80 which is required for step82
	step80 := pay_to_buyer_step.New(emptyStep, emptyStep, emptyState...)
	baseStep80 := step80.(states.IBaseStep)
	// add to flowManager maps
	flowManager.indexStepsMap[step80.Index()] = step80
	flowManager.nameStepsMap[step80.Name()] = step80

	flowManager.createStep81()
	flowManager.createStep82()
	flowManager.createStep80(baseStep80)

	//////////////////////////////////////////////////////////////////
	flowManager.createStep55()
	flowManager.createStep54()
	flowManager.createStep53()
	flowManager.createStep50()
	flowManager.createStep52()
	flowManager.createStep51()

	//////////////////////////////////////////////////////////////////
	flowManager.createStep42()
	flowManager.createStep40()
	flowManager.createStep44()
	flowManager.createStep41()
	flowManager.createStep43()

	//////////////////////////////////////////////////////////////////
	flowManager.createStep36()
	flowManager.createStep32()
	flowManager.createStep35()
	flowManager.createStep34()
	flowManager.createStep31()
	flowManager.createStep33()
	flowManager.createStep30()

	//////////////////////////////////////////////////////////////////
	flowManager.createStep21()
	flowManager.createStep20()

	//////////////////////////////////////////////////////////////////
	flowManager.createStep14()
	// TODO will be implement
	//flowManager.createStep13()
	flowManager.createStep11()
	flowManager.createStep12()
	flowManager.createStep10()

	//////////////////////////////////////////////////////////////////
	flowManager.createStep1()
	flowManager.createStep0()

	// Change route to support seller shipped reject route to shipment to
	flowManager.indexStepsMap[30].Childes()[1] = flowManager.indexStepsMap[21]
	flowManager.indexStepsMap[43].Childes()[1] = flowManager.indexStepsMap[80]
	return nil
}

func (flowManager *iFlowManagerImpl) createStep94() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(1, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.MarketFinalizeAction))

	// Create Notification State
	notificationState := notification_state.New(0, []states_old.IState{finalizeState}, emptyState,
		notification_action.NewOf(notification_action.MarketNotificationAction))
	step94 := pay_to_market_success_step.New(emptyStep, emptyStep, notificationState, finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step94.Index()] = step94
	flowManager.nameStepsMap[step94.Name()] = step94
}

func (flowManager *iFlowManagerImpl) createStep95() {
	var emptyState []states_old.IState

	step93 := flowManager.indexStepsMap[93]
	step94 := flowManager.indexStepsMap[94]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		retry_action.RetryAction:                          step93,
		manual_payment_action.ManualPaymentToMarketAction: step94,
	}

	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	manualPaymentState := manual_payment_state.New(2, []states_old.IState{nextToStep}, emptyState,
		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToMarketAction))
	operatorNotificationState := notification_state.New(1, []states_old.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
	retryState := retry_state.New(0, []states_old.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))

	step95 := pay_to_market_failed_step.New([]states.IStep{step94}, []states.IStep{step93},
		retryState, operatorNotificationState, manualPaymentState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step95.Index()] = step95
	flowManager.nameStepsMap[step95.Name()] = step95
}

func (flowManager *iFlowManagerImpl) createStep93(baseStep93 states.IBaseStep) {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step94 := flowManager.indexStepsMap[94]
	step95 := flowManager.indexStepsMap[95]

	payToMarketActions := pay_to_market_action.NewOf(pay_to_market_action.SuccessAction, pay_to_market_action.FailedAction)
	actionStepMap := map[actions.IEnumAction]states.IStep{
		pay_to_market_action.SuccessAction: step94,
		pay_to_market_action.FailedAction:  step95,
	}

	nextToStepState := next_to_step_state.New(1, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	payToMarketState := pay_to_market_state.New(0, []states_old.IState{nextToStepState}, emptyState, payToMarketActions)

	// TODO check change baseStep3 cause of change Step93 settings
	baseStep93.BaseStep().SetChildes([]states.IStep{step94, step95})
	baseStep93.BaseStep().SetParents(emptyStep)
	baseStep93.BaseStep().SetStates(payToMarketState, nextToStepState)

	// add to flowManager maps
	//flowManager.indexStepsMap[step93.Index()] = step93
	//flowManager.nameStepsMap[step93.Name()] = step93
}

func (flowManager *iFlowManagerImpl) createStep91() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step93 := flowManager.indexStepsMap[93]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		next_to_step_action.NextToStepAction: step93,
	}

	nextToStep93 := next_to_step_state.New(1, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)

	// Create Notification State
	notificationState := notification_state.New(0, []states_old.IState{nextToStep93}, emptyState,
		notification_action.NewOf(notification_action.SellerNotificationAction))

	step91 := pay_to_seller_success_step.New([]states.IStep{step93}, emptyStep, notificationState, nextToStep93)

	// add to flowManager maps
	flowManager.indexStepsMap[step91.Index()] = step91
	flowManager.nameStepsMap[step91.Name()] = step91
}

// TODO: checking flow and next to step states_old sequence
func (flowManager *iFlowManagerImpl) createStep92() {
	var emptyState []states_old.IState

	step90 := flowManager.indexStepsMap[90]
	step91 := flowManager.indexStepsMap[91]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		retry_action.RetryAction:                          step90,
		manual_payment_action.ManualPaymentToSellerAction: step91,
	}

	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)

	manualPaymentState := manual_payment_state.New(2, []states_old.IState{nextToStep}, emptyState,
		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToSellerAction))

	operatorNotificationState := notification_state.New(1, []states_old.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
	retryState := retry_state.New(0, []states_old.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))

	step92 := pay_to_seller_failed_step.New([]states.IStep{step91}, []states.IStep{step90},
		retryState, operatorNotificationState, manualPaymentState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step92.Index()] = step92
	flowManager.nameStepsMap[step92.Name()] = step92
}

// TODO settlement stock must be call once as a result it must be save in db
func (flowManager *iFlowManagerImpl) createStep90(baseStep90 states.IBaseStep) {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step91 := flowManager.indexStepsMap[91]
	step92 := flowManager.indexStepsMap[92]

	payToSellerActions := pay_to_seller_action.NewOf(pay_to_seller_action.SuccessAction, pay_to_seller_action.FailedAction)
	actionStepMap := map[actions.IEnumAction]states.IStep{
		pay_to_seller_action.SuccessAction: step91,
		pay_to_seller_action.FailedAction:  step92,
	}

	nextToStepState := next_to_step_state.New(1, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	payToSellerState := pay_to_seller_state.New(0, []states_old.IState{nextToStepState}, emptyState, payToSellerActions)

	// TODO check change baseStep90 cause of change Step90 settings
	baseStep90.BaseStep().SetChildes([]states.IStep{step91, step92})
	baseStep90.BaseStep().SetParents(emptyStep)
	baseStep90.BaseStep().SetStates(payToSellerState, nextToStepState)

	// add to flowManager maps
	//flowManager.indexStepsMap[step90.Index()] = step90
	//flowManager.nameStepsMap[step90.Name()] = step90
}

func (flowManager *iFlowManagerImpl) createStep81() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(1, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.BuyerFinalizeAction))

	// Create Notification State
	notificationState := notification_state.New(0, []states_old.IState{finalizeState}, emptyState,
		notification_action.NewOf(notification_action.BuyerNotificationAction))
	step81 := pay_to_buyer_success_step.New(emptyStep, emptyStep, notificationState, finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step81.Index()] = step81
	flowManager.nameStepsMap[step81.Name()] = step81
}

// TODO: checking flow and next to step states_old sequence
func (flowManager *iFlowManagerImpl) createStep82() {
	var emptyState []states_old.IState

	step80 := flowManager.indexStepsMap[80]
	step81 := flowManager.indexStepsMap[81]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		retry_action.RetryAction:                         step80,
		manual_payment_action.ManualPaymentToBuyerAction: step81,
	}

	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)

	manualPaymentState := manual_payment_state.New(2, []states_old.IState{nextToStep}, emptyState,
		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToBuyerAction))
	operatorNotificationState := notification_state.New(1, []states_old.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
	retryState := retry_state.New(0, []states_old.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))

	step82 := pay_to_buyer_failed_step.New([]states.IStep{step81}, []states.IStep{step80},
		retryState, operatorNotificationState, manualPaymentState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step82.Index()] = step82
	flowManager.nameStepsMap[step82.Name()] = step82
}

// TODO release stock must be call once as a result it must be save in db
func (flowManager *iFlowManagerImpl) createStep80(baseStep80 states.IBaseStep) {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step81 := flowManager.indexStepsMap[81]
	step82 := flowManager.indexStepsMap[82]

	payToBuyerActions := pay_to_buyer_action.NewOf(pay_to_buyer_action.SuccessAction, pay_to_buyer_action.FailedAction)
	actionStepMap := map[actions.IEnumAction]states.IStep{
		pay_to_buyer_action.SuccessAction: step81,
		pay_to_buyer_action.FailedAction:  step82,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	payToBuyerState := pay_to_buyer_state.New(1, []states_old.IState{nextToStepState}, emptyState, payToBuyerActions)
	stockReleaseActionState := stock_action_state.New(0, []states_old.IState{payToBuyerState}, emptyState, stock_action.NewOf(stock_action.ReleasedAction))

	// TODO check change baseStep80 cause of change Step82 settings
	baseStep80.BaseStep().SetChildes([]states.IStep{step81, step82})
	baseStep80.BaseStep().SetParents(emptyStep)
	baseStep80.BaseStep().SetStates(stockReleaseActionState, payToBuyerState, nextToStepState)

	// add to flowManager maps
	//flowManager.indexStepsMap[step93.Index()] = step93
	//flowManager.nameStepsMap[step93.Name()] = step93
}

func (flowManager *iFlowManagerImpl) createStep55() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)

	step55 := return_shipment_success_step.New([]states.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step55.Index()] = step55
	flowManager.nameStepsMap[step55.Name()] = step55
}

func (flowManager *iFlowManagerImpl) createStep54() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step90 := flowManager.indexStepsMap[90]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		next_to_step_action.NextToStepAction: step90,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	step54 := return_shipment_canceled_step.New([]states.IStep{step90}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step54.Index()] = step54
	flowManager.nameStepsMap[step54.Name()] = step54
}

func (flowManager *iFlowManagerImpl) createStep53() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step55 := flowManager.indexStepsMap[55]
	step54 := flowManager.indexStepsMap[54]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		operator_action.ReturnCanceledAction: step54,
		operator_action.ReturnedAction:       step55,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnCanceledAction, operator_action.ReturnedAction))
	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step53 := return_shipment_delivery_problem_step.New([]states.IStep{step54, step55}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step53.Index()] = step53
	flowManager.nameStepsMap[step53.Name()] = step53
}

func (flowManager *iFlowManagerImpl) createStep50() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step53 := flowManager.indexStepsMap[53]
	step55 := flowManager.indexStepsMap[55]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		seller_action.NeedSupportAction:     step53,
		seller_action.ApprovedAction:        step55,
		scheduler_action.AutoApprovedAction: step55,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.AutoApprovedAction))
	sellerApprovedActionState := seller_action_state.New("Return_Seller_Approval_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.ApprovedAction, seller_action.NeedSupportAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerApprovedActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step50 := return_shipment_delivered_step.New([]states.IStep{step53, step55}, emptyStep,
		notificationActionState, composeActorsActionState, sellerApprovedActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step50.Index()] = step50
	flowManager.nameStepsMap[step50.Name()] = step50
}

func (flowManager *iFlowManagerImpl) createStep52() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step50 := flowManager.indexStepsMap[50]
	step54 := flowManager.indexStepsMap[54]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		operator_action.ReturnDeliveredAction: step50,
		operator_action.ReturnCanceledAction:  step54,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnCanceledAction, operator_action.ReturnDeliveredAction))
	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step52 := return_shipment_delivery_delayed_step.New([]states.IStep{step54, step50}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step52.Index()] = step52
	flowManager.nameStepsMap[step52.Name()] = step52
}

// TODO schedulers need a config for timeout or auto approved
func (flowManager *iFlowManagerImpl) createStep51() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step50 := flowManager.indexStepsMap[50]
	step52 := flowManager.indexStepsMap[52]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		buyer_action.NeedSupportAction:                 step52,
		scheduler_action.NoActionForXDaysTimeoutAction: step50,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	buyerNeedSupportActionState := buyer_action_state.New("Buyer_Return_Shipment_Delivery_Pending_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.NeedSupportAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerNeedSupportActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step51 := return_shipment_delivery_pending_step.New([]states.IStep{step50, step52}, emptyStep,
		notificationActionState, composeActorsActionState, buyerNeedSupportActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step51.Index()] = step51
	flowManager.nameStepsMap[step51.Name()] = step51
}

func (flowManager *iFlowManagerImpl) createStep42() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step51 := flowManager.indexStepsMap[51]
	step50 := flowManager.indexStepsMap[50]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		buyer_action.DeliveredAction:                      step50,
		scheduler_action.WaitForShippingDaysTimeoutAction: step51,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.WaitForShippingDaysTimeoutAction))
	buyerReturnShipmentDeliveryActionState := buyer_action_state.New("Buyer_Return_Shipment_Delivery_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.DeliveredAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerReturnShipmentDeliveryActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step42 := return_shipped_step.New([]states.IStep{step50, step51}, emptyStep,
		notificationActionState, composeActorsActionState, buyerReturnShipmentDeliveryActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step42.Index()] = step42
	flowManager.nameStepsMap[step42.Name()] = step42
}

func (flowManager *iFlowManagerImpl) createStep40() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step90 := flowManager.indexStepsMap[90]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		next_to_step_action.NextToStepAction: step90,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	step40 := shipment_success_step.New([]states.IStep{step90}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step40.Index()] = step40
	flowManager.nameStepsMap[step40.Name()] = step40
}

func (flowManager *iFlowManagerImpl) createStep44() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step40 := flowManager.indexStepsMap[40]
	step42 := flowManager.indexStepsMap[42]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		buyer_action.EnterReturnShipmentDetailAction: step42,
		scheduler_action.WaitXDaysTimeoutAction:      step40,
	}

	nextToStepState := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStepState}, emptyState, scheduler_action.NewOf(scheduler_action.WaitXDaysTimeoutAction))
	buyerEnterReturnShipmentDetailDelayedActionState := buyer_action_state.New("Buyer_Enter_Return_Shipment_Detail_Delayed_Action_State", 2, []states_old.IState{nextToStepState}, emptyState, buyer_action.NewOf(buyer_action.EnterReturnShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerEnterReturnShipmentDetailDelayedActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step44 := return_shipment_detail_delayed_step.New([]states.IStep{step40, step42}, emptyStep,
		notificationState, composeActorsActionState, buyerEnterReturnShipmentDetailDelayedActionState,
		schedulerActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step44.Index()] = step44
	flowManager.nameStepsMap[step44.Name()] = step44
}

func (flowManager *iFlowManagerImpl) createStep41() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step42 := flowManager.indexStepsMap[42]
	step44 := flowManager.indexStepsMap[44]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		buyer_action.EnterReturnShipmentDetailAction:   step42,
		scheduler_action.NoActionForXDaysTimeoutAction: step44,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	BuyerEnterReturnShipmentDetailAction := buyer_action_state.New("Buyer_Enter_Return_Shipment_Detail_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.EnterReturnShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{BuyerEnterReturnShipmentDetailAction, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step41 := return_shipment_pending_step.New([]states.IStep{step42, step44}, emptyStep,
		notificationActionState, composeActorsActionState, BuyerEnterReturnShipmentDetailAction, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step41.Index()] = step41
	flowManager.nameStepsMap[step41.Name()] = step41
}

func (flowManager *iFlowManagerImpl) createStep43() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step41 := flowManager.indexStepsMap[41]
	step40 := flowManager.indexStepsMap[40]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		operator_action.ReturnedAction:       step41,
		operator_action.ReturnCanceledAction: step40,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnedAction, operator_action.ReturnCanceledAction))
	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step43 := shipment_delivery_problem_step.New([]states.IStep{step40, step41}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step43.Index()] = step43
	flowManager.nameStepsMap[step43.Name()] = step43
}

func (flowManager *iFlowManagerImpl) createStep36() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	step36 := shipment_canceled_step.New([]states.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step36.Index()] = step36
	flowManager.nameStepsMap[step36.Name()] = step36
}

func (flowManager *iFlowManagerImpl) createStep32() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step43 := flowManager.indexStepsMap[43]
	step41 := flowManager.indexStepsMap[41]
	step40 := flowManager.indexStepsMap[40]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		scheduler_action.AutoApprovedAction: step40,
		buyer_action.ApprovedAction:         step40,
		buyer_action.NeedSupportAction:      step43,
		buyer_action.ReturnIfPossibleAction: step41,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.AutoApprovedAction))
	buyerApprovalActionState := buyer_action_state.New("Buyer_Approval_Action_State", 2, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.ApprovedAction, buyer_action.NeedSupportAction, buyer_action.ReturnIfPossibleAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerApprovalActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step32 := shipment_delivered_step.New([]states.IStep{step43, step41, step40}, emptyStep,
		notificationActionState, composeActorsActionState, buyerApprovalActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step32.Index()] = step32
	flowManager.nameStepsMap[step32.Name()] = step32
}

func (flowManager *iFlowManagerImpl) createStep35() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step36 := flowManager.indexStepsMap[36]
	step32 := flowManager.indexStepsMap[32]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		operator_action.CanceledAction:  step36,
		operator_action.DeliveredAction: step32,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	operatorActionState := operator_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.DeliveredAction, operator_action.CanceledAction))
	notificationState := notification_state.New(0, []states_old.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step35 := shipment_delivery_delayed_step.New([]states.IStep{step32, step36}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step35.Index()] = step35
	flowManager.nameStepsMap[step35.Name()] = step35
}

func (flowManager *iFlowManagerImpl) createStep34() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step35 := flowManager.indexStepsMap[35]
	step32 := flowManager.indexStepsMap[32]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		seller_action.NeedSupportAction:                step35,
		scheduler_action.NoActionForXDaysTimeoutAction: step32,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	sellerShipmentDeliveryPendingActionState := seller_action_state.New("Seller_Shipment_Delivery_Pending_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.NeedSupportAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerShipmentDeliveryPendingActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step34 := shipment_delivery_pending_step.New([]states.IStep{step35, step32}, emptyStep,
		notificationActionState, composeActorsActionState, sellerShipmentDeliveryPendingActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step34.Index()] = step34
	flowManager.nameStepsMap[step34.Name()] = step34
}

func (flowManager *iFlowManagerImpl) createStep31() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step32 := flowManager.indexStepsMap[32]
	step34 := flowManager.indexStepsMap[34]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		seller_action.DeliveredAction:                     step32,
		scheduler_action.WaitForShippingDaysTimeoutAction: step34,
	}

	nextToStep := next_to_step_state.New(5, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(4, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.WaitForShippingDaysTimeoutAction))
	sellerShipmentDeliveryActionState := buyer_action_state.New("Seller_Shipment_Delivery_Action_State", 3, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.DeliveredAction))
	composeActorsActionState := system_action_state.New(2, []states_old.IState{sellerShipmentDeliveryActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(1, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))
	stockSettlementActionState := stock_action_state.New(0, []states_old.IState{notificationActionState}, emptyState, stock_action.NewOf(stock_action.SettlementAction))

	step31 := shipped_step.New([]states.IStep{step32, step34}, emptyStep,
		stockSettlementActionState, notificationActionState, composeActorsActionState,
		sellerShipmentDeliveryActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step31.Index()] = step31
	flowManager.nameStepsMap[step31.Name()] = step31
}

func (flowManager *iFlowManagerImpl) createStep33() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step31 := flowManager.indexStepsMap[31]
	step36 := flowManager.indexStepsMap[36]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		seller_action.EnterShipmentDetailAction: step31,
		buyer_action.CanceledAction:             step36,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	buyerWaitForSellerEnterShipmentDetailDelayedActionState := buyer_action_state.New("Buyer_Wait_For_Seller_Enter_Shipment_Detail_Delayed_Action_State", 3, []states_old.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.CanceledAction))
	sellerEnterShipmentDetailDelayedActionState := seller_action_state.New("Seller_Enter_Shipment_Detail_Delayed_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.EnterShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{buyerWaitForSellerEnterShipmentDetailDelayedActionState, sellerEnterShipmentDetailDelayedActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction, notification_action.BuyerNotificationAction))

	step33 := shipment_detail_delayed_step.New([]states.IStep{step31, step36}, emptyStep,
		notificationActionState, composeActorsActionState, sellerEnterShipmentDetailDelayedActionState, buyerWaitForSellerEnterShipmentDetailDelayedActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step33.Index()] = step33
	flowManager.nameStepsMap[step33.Name()] = step33
}

func (flowManager *iFlowManagerImpl) createStep30() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step31 := flowManager.indexStepsMap[31]
	step33 := flowManager.indexStepsMap[33]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		seller_action.EnterShipmentDetailAction:        step31,
		scheduler_action.NoActionForXDaysTimeoutAction: step33,
	}

	nextToStepState := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStepState}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	sellerEnterShipmentDetailActionState := seller_action_state.New("Seller_Enter_Shipment_Detail_Action_State", 2, []states_old.IState{nextToStepState}, emptyState, seller_action.NewOf(seller_action.EnterShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerEnterShipmentDetailActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step30 := shipment_pending_step.New([]states.IStep{step31, step33}, emptyStep,
		notificationState, composeActorsActionState, sellerEnterShipmentDetailActionState,
		schedulerActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step30.Index()] = step30
	flowManager.nameStepsMap[step30.Name()] = step30
}

func (flowManager *iFlowManagerImpl) createStep21() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	step21 := shipment_rejected_by_seller_step.New([]states.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step21.Index()] = step21
	flowManager.nameStepsMap[step21.Name()] = step21
}

func (flowManager *iFlowManagerImpl) createStep20() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step30 := flowManager.indexStepsMap[30]
	step21 := flowManager.indexStepsMap[21]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		seller_action.ApprovedAction:                   step30,
		seller_action.RejectAction:                     step21,
		scheduler_action.NoActionForXDaysTimeoutAction: step21,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states_old.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	sellerApprovalActionState := seller_action_state.New("Seller_Approval_Action_State", 2, []states_old.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.ApprovedAction, seller_action.RejectAction))
	composeActorsActionState := system_action_state.New(1, []states_old.IState{sellerApprovalActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states_old.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step20 := seller_approval_pending_step.New([]states.IStep{step30, step21}, emptyStep,
		notificationActionState, composeActorsActionState, sellerApprovalActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step20.Index()] = step20
	flowManager.nameStepsMap[step20.Name()] = step20
}

func (flowManager *iFlowManagerImpl) createStep14() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	step14 := payment_rejected_step.New([]states.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step14.Index()] = step14
	flowManager.nameStepsMap[step14.Name()] = step14
}

func (flowManager *iFlowManagerImpl) createStep11() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step14 := flowManager.indexStepsMap[14]
	step20 := flowManager.indexStepsMap[20]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		buyer_action.RejectAction:   step14,
		buyer_action.ApprovedAction: step20,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	buyerPaymentApprovalActionState := buyer_action_state.New("Buyer_Payment_Approval_Action_State", 1, []states_old.IState{nextToStepState}, emptyState, buyer_action.NewOf(buyer_action.ApprovedAction, buyer_action.RejectAction))
	notificationState := notification_state.New(0, []states_old.IState{buyerPaymentApprovalActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step11 := payment_success_step.New([]states.IStep{step14, step20}, emptyStep,
		notificationState, buyerPaymentApprovalActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step11.Index()] = step11
	flowManager.nameStepsMap[step11.Name()] = step11
}

func (flowManager *iFlowManagerImpl) createStep12() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(2, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.PaymentFailedFinalizeAction))

	// Create Notification State
	notificationState := notification_state.New(1, []states_old.IState{finalizeState}, emptyState,
		notification_action.NewOf(notification_action.BuyerNotificationAction))

	stockReleaseActionState := stock_action_state.New(0, []states_old.IState{notificationState, finalizeState}, emptyState, stock_action.NewOf(stock_action.ReleasedAction, stock_action.FailedAction))
	step12 := payment_failed_step.New(emptyStep, emptyStep,
		stockReleaseActionState, notificationState, finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step12.Index()] = step12
	flowManager.nameStepsMap[step12.Name()] = step12
}

func (flowManager *iFlowManagerImpl) createStep10() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step12 := flowManager.indexStepsMap[12]
	step11 := flowManager.indexStepsMap[11]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		order_payment_action.OrderPaymentFailedAction: step12,
		payment_action.FailedAction:                   step12,
		payment_action.SuccessAction:                  step11,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	paymentActionState := payment_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, payment_action.NewOf(payment_action.SuccessAction, payment_action.FailedAction))
	orderPaymentActionState := order_payment_action_state.New(0, []states_old.IState{paymentActionState, nextToStepState}, emptyState, order_payment_action.NewOf(order_payment_action.OrderPaymentAction, order_payment_action.OrderPaymentFailedAction))

	step10 := payment_pending_step.New([]states.IStep{step12, step11}, emptyStep,
		orderPaymentActionState, paymentActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step10.Index()] = step10
	flowManager.nameStepsMap[step10.Name()] = step10
}

func (flowManager *iFlowManagerImpl) createStep1() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(1, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.OrderFailedFinalizeAction))

	step1 := new_order_failed_step.New(emptyStep, emptyStep, finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step1.Index()] = step1
	flowManager.nameStepsMap[step1.Name()] = step1
}

func (flowManager *iFlowManagerImpl) createStep0() {
	var emptyState []states_old.IState
	var emptyStep []states.IStep

	step1 := flowManager.indexStepsMap[1]
	step10 := flowManager.indexStepsMap[10]

	actionStepMap := map[actions.IEnumAction]states.IStep{
		stock_action.FailedAction:   step1,
		stock_action.ReservedAction: step10,
	}

	nextToStepState := next_to_step_state.New(3, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	stockReservedActionState := stock_action_state.New(1, []states_old.IState{nextToStepState}, emptyState, stock_action.NewOf(stock_action.ReservedAction, stock_action.FailedAction))
	checkoutStateAction := checkout_action_state.New(0, []states_old.IState{stockReservedActionState, nextToStepState}, emptyState, checkout_action.NewOf(checkout_action.NewOrderAction))

	step0 := new_order_step.New([]states.IStep{step1, step10}, emptyStep,
		checkoutStateAction, stockReservedActionState, nextToStepState)
	// add to flowManager maps
	flowManager.indexStepsMap[step0.Index()] = step0
	flowManager.nameStepsMap[step0.Name()] = step0
}

// TODO Must be refactored
func (flowManager iFlowManagerImpl) MessageHandler(ctx context.Context, req *message.MessageRequest) promise.IPromise {
	// received New Order Request
	//if len(req.OrderId) == 0 {

	step0 := flowManager.indexStepsMap[0]
	return step0.ProcessMessage(ctx, req)

	// TODO must be implement
	//}
	//} else if len(req.ItemId) != 0 {

	//} else {
	//	order, err := global.Singletons.OrderRepository.FindById(req.OrderId)
	//	if err != nil {
	//		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
	//			req.OrderId, err)
	//		returnChannel := make(chan promise.FutureData, 1)
	//		returnChannel <- promise.FutureData{Data:nil, Ex:promise.FutureError{Code: promise.NotFound, Reason:"OrderId Not Found"}}
	//		defer close(returnChannel)
	//		return promise.NewPromise(returnChannel, 1, 1)
	//	}
	//}
}

func (flowManager iFlowManagerImpl) SellerApprovalPending(ctx context.Context, req *message.RequestSellerOrderAction) promise.IPromise {
	order, err := global.Singletons.OrderRepository.FindById(req.OrderId)
	if err != nil {
		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, orderId: %d, error: %s",
			req.OrderId, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "OrderId Not Found"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	itemsId := make([]uint64, 0, len(order.Items))
	for i := 0; i < len(order.Items); i++ {
		if order.Items[i].SellerInfo.SellerId == req.SellerId {
			itemsId = append(itemsId, order.Items[i].ItemId)
		}
	}

	if req.ActionType == "approved" {
		return flowManager.indexStepsMap[20].ProcessOrder(ctx, *order, itemsId, req)
	} else if req.ActionType == "shipped" {
		return flowManager.indexStepsMap[30].ProcessOrder(ctx, *order, itemsId, req)
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "ActionType Not Found"}}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (flowManager iFlowManagerImpl) BuyerApprovalPending(ctx context.Context, req *message.RequestBuyerOrderAction) promise.IPromise {
	order, err := global.Singletons.OrderRepository.FindById(req.OrderId)
	if err != nil {
		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, orderId: %d, error: %s",
			req.OrderId, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "OrderId Not Found"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	//itemsId := make([]string, 0, len(order.Items))
	//for i:= 0; i < len(order.Items); i++ {
	//	if order.Items[i].SellerInfo.SellerId == req.ItemId[i] {
	//		itemsId = append(itemsId,order.Items[i].ItemId)
	//	}
	//}

	if req.ActionType == "Approved" {
		return flowManager.indexStepsMap[32].ProcessOrder(ctx, *order, req.ItemsId, req)
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "ActionType Not Found"}}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (flowManager iFlowManagerImpl) PaymentGatewayResult(ctx context.Context, req *pg.PaygateHookRequest) promise.IPromise {
	orderId, err := strconv.Atoi(req.OrderID)
	if err != nil {
		logger.Err("PaymentGatewayResult() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
			req.OrderID, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.BadRequest, Reason: "OrderId Invalid"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	order, err := global.Singletons.OrderRepository.FindById(uint64(orderId))
	if err != nil {
		logger.Err("PaymentGatewayResult() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
			req.OrderID, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "OrderId Not Found"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	order.PaymentService[0].PaymentResult = &entities.PaymentResult{
		Result:      req.Result,
		Reason:      "",
		PaymentId:   req.PaymentId,
		InvoiceId:   req.InvoiceId,
		Amount:      uint64(req.Amount),
		ReqBody:     req.ReqBody,
		ResBody:     req.ResBody,
		CardNumMask: req.CardMask,
		CreatedAt:   time.Now().UTC(),
	}

	return flowManager.indexStepsMap[10].ProcessOrder(ctx, *order, nil, "OrderPayment")
}

func (flowManager iFlowManagerImpl) OperatorActionPending(ctx context.Context, req *message.RequestBackOfficeOrderAction) promise.IPromise {
	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
		return bson.D{{"items.itemId", req.ItemId}}
	})

	if err != nil {
		logger.Err("MessageHandler() => request itemId not found, OrderRepository.FindById failed, itemId: %d, error: %s",
			req.ItemId, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if len(orders) == 0 {
		logger.Err("MessageHandler() => request itemId not found, itemId: %d", req.ItemId)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "ItemId Not Found"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if len(orders) > 1 {
		logger.Err("MessageHandler() => request itemId found in multiple order, itemId: %d, error: %s",
			req.ItemId, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	itemsId := make([]uint64, 0, 1)
	for i := 0; i < len(orders[0].Items); i++ {
		if orders[0].Items[i].ItemId == req.ItemId {
			itemsId = append(itemsId, orders[0].Items[i].ItemId)
		}
	}

	if req.ActionType == "shipmentDelivered" {
		return flowManager.indexStepsMap[43].ProcessOrder(ctx, *orders[0], itemsId, req)
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "ActionType Not Found"}}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (flowManager iFlowManagerImpl) BackOfficeOrdersListView(ctx context.Context, req *message.RequestBackOfficeOrdersList) promise.IPromise {
	orders, total, err := global.Singletons.OrderRepository.FindAllWithPageAndSort(int64(req.Page), int64(req.PerPage), req.Sort, int(req.Direction))

	if err != nil {
		logger.Err("BackOfficeOrdersListView() => FindAllWithPageAndSort failed, error: %s", err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if len(orders) == 0 {
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "Orders Not Found"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	response := message.ResponseBackOfficeOrdersList{
		Total:  total,
		Orders: make([]*message.BackOfficeOrdersList, 0, len(orders)),
	}

	for _, order := range orders {
		backOfficeOrder := &message.BackOfficeOrdersList{
			OrderId:     order.OrderId,
			PurchasedOn: order.CreatedAt.Unix(),
			BasketSize:  0,
			BillTo:      order.BuyerInfo.FirstName + order.BuyerInfo.LastName,
			ShipTo:      order.BuyerInfo.ShippingAddress.FirstName + order.BuyerInfo.ShippingAddress.LastName,
			TotalAmount: int64(order.Invoice.Total),
			Status:      order.Status,
			LastUpdated: order.UpdatedAt.Unix(),
			Actions:     []string{"success", "cancel"},
		}

		itemsInventory := make(map[string]int, len(order.Items))
		for i := 0; i < len(order.Items); i++ {
			if _, ok := itemsInventory[order.Items[i].InventoryId]; !ok {
				backOfficeOrder.BasketSize += order.Items[i].Quantity
			}
		}

		if order.Invoice.Voucher != nil {
			backOfficeOrder.PaidAmount = int64(order.Invoice.Total - order.Invoice.Voucher.Amount)
			backOfficeOrder.Voucher = true
		} else {
			backOfficeOrder.Voucher = false
			backOfficeOrder.PaidAmount = 0
		}

		response.Orders = append(response.Orders, backOfficeOrder)
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: &response, Ex: nil}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

// TODO check payment length
func (flowManager iFlowManagerImpl) BackOfficeOrderDetailView(ctx context.Context, req *message.RequestIdentifier) promise.IPromise {

	orderId, err := strconv.Atoi(req.Id)
	if err != nil {
		logger.Err("BackOfficeOrderDetailView() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
			req.Id, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.BadRequest, Reason: "OrderId Invalid"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	order, err := global.Singletons.OrderRepository.FindById(uint64(orderId))
	if err != nil {
		logger.Err("BackOfficeOrderDetailView() => request orderId not found, OrderRepository.FindById failed, order: %s, error: %s",
			req.Id, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.NotFound, Reason: "OrderId Not Found"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	response := &message.ResponseOrderDetailView{
		OrderId:   order.OrderId,
		CreatedAt: order.CreatedAt.Unix(),
		Ip:        order.BuyerInfo.IP,
		Status:    order.Status,
		Payment: &message.PaymentInfo{
			PaymentMethod: order.Invoice.PaymentMethod,
			PaymentOption: order.Invoice.PaymentOption,
		},
		Billing: &message.BillingInfo{
			BuyerId:    order.BuyerInfo.BuyerId,
			FullName:   order.BuyerInfo.FirstName + order.BuyerInfo.LastName,
			Phone:      order.BuyerInfo.Phone,
			Mobile:     order.BuyerInfo.Mobile,
			NationalId: order.BuyerInfo.NationalId,
		},
		ShippingInfo: &message.ShippingInfo{
			FullName:     order.BuyerInfo.ShippingAddress.FirstName + order.BuyerInfo.ShippingAddress.LastName,
			Country:      order.BuyerInfo.ShippingAddress.Country,
			City:         order.BuyerInfo.ShippingAddress.City,
			Province:     order.BuyerInfo.ShippingAddress.Province,
			Neighborhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
			Address:      order.BuyerInfo.ShippingAddress.Address,
			ZipCode:      order.BuyerInfo.ShippingAddress.ZipCode,
		},
		Items: make([]*message.ItemInfo, 0, len(order.Items)),
	}

	for _, item := range order.Items {
		itemInfo := &message.ItemInfo{
			ItemId:      item.ItemId,
			SellerId:    item.SellerInfo.SellerId,
			InventoryId: item.InventoryId,
			Quantity:    item.Quantity,
			ItemStatus:  item.Status,
			Price: &message.PriceInfo{
				Unit:             item.Invoice.Unit,
				Total:            item.Invoice.Total,
				Original:         item.Invoice.Original,
				Special:          item.Invoice.Special,
				Discount:         item.Invoice.Discount,
				SellerCommission: item.Invoice.SellerCommission,
				Currency:         item.Invoice.Currency,
			},
			UpdatedAt: item.UpdatedAt.Unix(),
			Actions:   []string{"success", "cancel"},
		}

		lastStep := item.Progress.StepsHistory[len(item.Progress.StepsHistory)-1]

		if lastStep.ActionHistory != nil {
			lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
			itemInfo.StepStatus = lastAction.Name
		} else {
			itemInfo.StepStatus = "none"
			logger.Audit("BackOfficeOrderDetailView() => Action History is nil, orderId: %d, itemId: %d", order.OrderId, item.ItemId)
		}

		//lastAction := lastStep.ActionHistory[len(lastStep.ActionHistory)-1]
		//itemInfo.StepStatus = lastAction.Name

		response.Items = append(response.Items, itemInfo)
	}

	if order.PaymentService != nil && len(order.PaymentService) == 1 {
		if order.PaymentService[0].PaymentResult != nil {
			response.Payment.Result = order.PaymentService[0].PaymentResponse.Result
		} else {
			response.Payment.Result = false
		}
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: response, Ex: nil}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (flowManager iFlowManagerImpl) SellerReportOrders(req *message.RequestSellerReportOrders, srv message.OrderService_SellerReportOrdersServer) promise.IPromise {
	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
		return bson.D{{"createdAt",
			bson.D{{"$gte", time.Unix(int64(req.StartDateTime), 0).UTC()}}},
			{"items.status", req.Status}, {"items.sellerInfo.sellerId", req.SellerId}}
	})

	if err != nil {
		logger.Err("SellerReportOrders() => OrderRepository.FindByFilter failed, startDateTime: %v, status: %s, error: %s",
			req.StartDateTime, req.Status, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if orders == nil || len(orders) == 0 {
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: nil}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	reports := make([]*reports2.SellerExportOrders, 0, len(orders))

	for _, order := range orders {
		for _, item := range order.Items {
			if item.Status == req.Status {
				itemReport := &reports2.SellerExportOrders{
					OrderId:     order.OrderId,
					ItemId:      item.ItemId,
					ProductId:   item.InventoryId[0:8],
					InventoryId: item.InventoryId,
					PaidPrice:   item.Invoice.Total,
					Commission:  item.Invoice.SellerCommission,
					Category:    item.Category,
					Status:      item.Status,
				}

				//localTime := item.CreatedAt.Local()
				tempTime := time.Date(item.CreatedAt.Year(),
					item.CreatedAt.Month(),
					item.CreatedAt.Day(),
					item.CreatedAt.Hour(),
					item.CreatedAt.Minute(),
					item.CreatedAt.Second(),
					item.CreatedAt.Nanosecond(),
					ptime.Iran())

				pt := ptime.New(tempTime)
				itemReport.CreatedAt = pt.String()

				tempTime = time.Date(item.UpdatedAt.Year(),
					item.UpdatedAt.Month(),
					item.UpdatedAt.Day(),
					item.UpdatedAt.Hour(),
					item.UpdatedAt.Minute(),
					item.UpdatedAt.Second(),
					item.UpdatedAt.Nanosecond(),
					ptime.Iran())

				pt = ptime.New(tempTime)
				itemReport.UpdatedAt = pt.String()
				reports = append(reports, itemReport)
			}
		}
	}

	csvReports := make([][]string, 0, len(reports))
	csvHeadLines := []string{
		"OrderId", "ItemId", "ProductId", "InventoryId",
		"PaidPrice", "Commission", "Category", "Status", "CreatedAt", "UpdatedAt",
	}

	csvReports = append(csvReports, csvHeadLines)
	for _, itemReport := range reports {
		csvRecord := []string{
			strconv.Itoa(int(itemReport.OrderId)),
			strconv.Itoa(int(itemReport.ItemId)),
			itemReport.ProductId,
			itemReport.InventoryId,
			fmt.Sprint(itemReport.PaidPrice),
			fmt.Sprint(itemReport.Commission),
			itemReport.Category,
			itemReport.Status,
			itemReport.CreatedAt,
			itemReport.UpdatedAt,
		}
		csvReports = append(csvReports, csvRecord)
	}

	reportTime := time.Unix(int64(req.StartDateTime), 0)
	fileName := fmt.Sprintf("SellerReportOrders-%s.csv", fmt.Sprintf("%d", reportTime.UnixNano()))
	f, err := os.Create("/tmp/" + fileName)
	if err != nil {
		logger.Err("SellerReportOrders() => create file %s failed, startDateTime: %v, status: %s, error: %s",
			fileName, req.StartDateTime, req.Status, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	w := csv.NewWriter(f)
	// calls Flush internally
	if err := w.WriteAll(csvReports); err != nil {
		logger.Err("SellerReportOrders() => write csv to file failed, startDateTime: %v, : status: %s, error: %s",
			req.StartDateTime, req.Status, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if err := f.Close(); err != nil {
		logger.Err("SellerReportOrders() => file close failed, filename: %s, error: %s", fileName, err)
	}

	file, err := os.Open("/tmp/" + fileName)
	if err != nil {
		logger.Err("SellerReportOrders() => write csv to file failed, startDateTime: %v, : status: %s, error: %s",
			req.StartDateTime, req.Status, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
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
		logger.Err("SellerReportOrders() => file close failed, filename: %s, error: %s", fileName, err)
	}

	if err := os.Remove("/tmp/" + fileName); err != nil {
		logger.Err("SellerReportOrders() => remove file failed, filename: %s, error: %s", fileName, err)
	}

	if fileErr != nil {
		logger.Err("SellerReportOrders() => read csv from file failed, filename: %s, error: %s", fileName, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if grpcErr != nil {
		logger.Err("SellerReportOrders() => send cvs file failed, filename: %s, error: %s", fileName, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: nil, Ex: nil}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (flowManager iFlowManagerImpl) BackOfficeReportOrderItems(req *message.RequestBackOfficeReportOrderItems, srv message.OrderService_BackOfficeReportOrderItemsServer) promise.IPromise {
	orders, err := global.Singletons.OrderRepository.FindByFilter(func() interface{} {
		return bson.D{{"createdAt",
			bson.D{{"$gte", time.Unix(int64(req.StartDateTime), 0).UTC()},
				{"$lte", time.Unix(int64(req.EndDataTime), 0).UTC()}}}}
	})

	if err != nil {
		logger.Err("BackOfficeReportOrderItems() => request itemId not found, OrderRepository.FindById failed, startDateTime: %v, endDateTime: %v, error: %s",
			req.StartDateTime, req.EndDataTime, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if orders == nil || len(orders) == 0 {
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: nil}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	reports := make([]*reports2.BackOfficeExportItems, 0, len(orders))
	sellerProfileMap := make(map[uint64]entities.SellerProfile)

	for _, order := range orders {
		for _, item := range order.Items {
			itemReport := &reports2.BackOfficeExportItems{
				ItemId:      item.ItemId,
				InventoryId: item.InventoryId,
				ProductId:   item.InventoryId[0:8],
				BuyerId:     order.BuyerInfo.BuyerId,
				BuyerPhone:  order.BuyerInfo.Phone,
				SellerId:    item.SellerInfo.SellerId,
				SellerName:  "",
				Price:       item.Invoice.Total,
				Status:      item.Status,
			}

			tempTime := time.Date(item.CreatedAt.Year(),
				item.CreatedAt.Month(),
				item.CreatedAt.Day(),
				item.CreatedAt.Hour(),
				item.CreatedAt.Minute(),
				item.CreatedAt.Second(),
				item.CreatedAt.Nanosecond(),
				ptime.Iran())

			pt := ptime.New(tempTime)
			itemReport.CreatedAt = pt.String()

			tempTime = time.Date(item.UpdatedAt.Year(),
				item.UpdatedAt.Month(),
				item.UpdatedAt.Day(),
				item.UpdatedAt.Hour(),
				item.UpdatedAt.Minute(),
				item.UpdatedAt.Second(),
				item.UpdatedAt.Nanosecond(),
				ptime.Iran())

			pt = ptime.New(tempTime)
			itemReport.UpdatedAt = pt.String()
			reports = append(reports, itemReport)

			if sellerProfile, ok := sellerProfileMap[item.SellerInfo.SellerId]; !ok {
				userCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
				ipromise := global.Singletons.UserService.GetSellerProfile(userCtx, strconv.Itoa(int(item.SellerInfo.SellerId)))
				futureData := ipromise.Data()
				if futureData.Ex != nil {
					logger.Err("BackOfficeReportOrderItems() => get sellerProfile failed, orderId: %d, itemId: %d, sellerId: %d",
						order.OrderId, item.ItemId, item.SellerInfo.SellerId)
					continue
				}

				sellerInfo, ok := futureData.Data.(entities.SellerProfile)
				if ok != true {
					logger.Err("BackOfficeReportOrderItems() => get sellerProfile invalid, orderId: %d, itemId: %d, sellerId: %d",
						order.OrderId, item.ItemId, item.SellerInfo.SellerId)
					continue
				}

				sellerProfileMap[item.SellerInfo.SellerId] = sellerProfile
				itemReport.SellerName = sellerInfo.GeneralInfo.ShopDisplayName
			} else {
				itemReport.SellerName = sellerProfile.GeneralInfo.ShopDisplayName
			}
		}
	}

	csvReports := make([][]string, 0, len(reports))
	csvHeadLines := []string{
		"ItemId", "InventoryId", "ProductId", "BuyerId", "BuyerPhone", "SellerId",
		"SellerName", "ItemInvoice", "Status", "CreatedAt", "UpdatedAt",
	}

	csvReports = append(csvReports, csvHeadLines)
	for _, itemReport := range reports {
		csvRecord := []string{
			strconv.Itoa(int(itemReport.ItemId)),
			itemReport.InventoryId,
			itemReport.ProductId,
			strconv.Itoa(int(itemReport.BuyerId)),
			itemReport.BuyerPhone,
			strconv.Itoa(int(itemReport.SellerId)),
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
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	w := csv.NewWriter(f)
	// calls Flush internally
	if err := w.WriteAll(csvReports); err != nil {
		logger.Err("BackOfficeReportOrderItems() => write csv to file failed, startDateTime: %v, endDateTime: %v, error: %s",
			req.StartDateTime, req.EndDataTime, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if err := f.Close(); err != nil {
		logger.Err("BackOfficeReportOrderItems() => file close failed, filename: %s, error: %s", fileName, err)
	}

	file, err := os.Open("/tmp/" + fileName)
	if err != nil {
		logger.Err("BackOfficeReportOrderItems() => read csv from file failed, startDateTime: %v, endDateTime: %v, error: %s",
			req.StartDateTime, req.EndDataTime, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
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
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	if grpcErr != nil {
		logger.Err("BackOfficeReportOrderItems() => send cvs file failed, filename: %s, error: %s", fileName, err)
		returnChannel := make(chan promise.FutureData, 1)
		returnChannel <- promise.FutureData{Data: nil, Ex: promise.FutureError{Code: promise.InternalError, Reason: "Unknown Error"}}
		defer close(returnChannel)
		return promise.NewPromise(returnChannel, 1, 1)
	}

	returnChannel := make(chan promise.FutureData, 1)
	returnChannel <- promise.FutureData{Data: nil, Ex: nil}
	defer close(returnChannel)
	return promise.NewPromise(returnChannel, 1, 1)
}

func (flowManager iFlowManagerImpl) SchedulerEvents(event events.ISchedulerEvent) {
	order, err := global.Singletons.OrderRepository.FindById(event.OrderId)
	if err != nil {
		logger.Err("MessageHandler() => request orderId not found, OrderRepository.FindById failed, schedulerEvent: %v, error: %s",
			event, err)
		return
	}

	//itemsId := make([]string, 0, len(order.Items))
	//for i:= 0; i < len(event.ItemsId); i++ {
	//	if order.Items[i].ItemId == event.ItemsId[i] && order.Items[i].SellerInfo.SellerId == event.SellerId {
	//		itemsId = append(itemsId,order.Items[i].ItemId)
	//	}
	//}

	ctx, _ := context.WithCancel(context.Background())

	if event.Action == "ApprovalPending" {
		flowManager.indexStepsMap[20].ProcessOrder(ctx, *order, event.ItemsId, "actionExpired")

	} else if event.Action == "SellerShipmentPending" {
		flowManager.indexStepsMap[30].ProcessOrder(ctx, *order, event.ItemsId, "actionExpired")

	} else if event.Action == "ShipmentDeliveredPending" {
		flowManager.indexStepsMap[32].ProcessOrder(ctx, *order, event.ItemsId, "actionApproved")
	}

}
