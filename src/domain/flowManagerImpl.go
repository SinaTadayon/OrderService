package domain

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	finalize_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/finalize"
	manual_payment_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/manualpayment"
	new_order_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/neworder"
	next_to_step_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/nextstep"
	notification_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/notification"
	pay_to_buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/paytobuyer"
	pay_to_market_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/paytomarket"
	pay_to_seller_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/paytoseller"
	retry_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/retry"
	stock_action "gitlab.faza.io/order-project/order-service/domain/actions/actives/stock"
	buyer_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/buyer"
	operator_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/operator"
	payment_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/payment"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/scheduler"
	seller_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/seller"
	system_action "gitlab.faza.io/order-project/order-service/domain/actions/actors/system"
	"gitlab.faza.io/order-project/order-service/domain/states"
	finalize_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/finalize"
	manual_payment_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/manualpayment"
	new_order_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/neworder"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher/nextstep"
	notification_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/notification"
	pay_to_buyer_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/paytobuyer"
	pay_to_market_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/paytomarket"
	pay_to_seller_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/paytoseller"
	retry_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/retry"
	"gitlab.faza.io/order-project/order-service/domain/states/launcher/stock"
	buyer_action_state "gitlab.faza.io/order-project/order-service/domain/states/listener/buyer"
	operator_action_state "gitlab.faza.io/order-project/order-service/domain/states/listener/operator"
	payment_action_state "gitlab.faza.io/order-project/order-service/domain/states/listener/payment"
	scheduler_action_state "gitlab.faza.io/order-project/order-service/domain/states/listener/scheduler"
	seller_action_state "gitlab.faza.io/order-project/order-service/domain/states/listener/seller"
	system_action_state "gitlab.faza.io/order-project/order-service/domain/states/listener/system"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	new_order_step "gitlab.faza.io/order-project/order-service/domain/steps/step_0"
	new_order_failed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_1"
	payment_pending_step "gitlab.faza.io/order-project/order-service/domain/steps/step_10"
	payment_success_step "gitlab.faza.io/order-project/order-service/domain/steps/step_11"
	payment_failed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_12"
	payment_rejected_step "gitlab.faza.io/order-project/order-service/domain/steps/step_14"
	seller_approval_pending_step "gitlab.faza.io/order-project/order-service/domain/steps/step_20"
	shipment_rejected_by_seller_step "gitlab.faza.io/order-project/order-service/domain/steps/step_21"
	shipment_pending_step "gitlab.faza.io/order-project/order-service/domain/steps/step_30"
	shipped_step "gitlab.faza.io/order-project/order-service/domain/steps/step_31"
	shipment_delivered_step "gitlab.faza.io/order-project/order-service/domain/steps/step_32"
	shipment_detail_delayed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_33"
	shipment_delivery_pending_step "gitlab.faza.io/order-project/order-service/domain/steps/step_34"
	shipment_delivery_delayed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_35"
	shipment_canceled_step "gitlab.faza.io/order-project/order-service/domain/steps/step_36"
	shipment_success_step "gitlab.faza.io/order-project/order-service/domain/steps/step_40"
	return_shipment_pending_step "gitlab.faza.io/order-project/order-service/domain/steps/step_41"
	return_shipped_step "gitlab.faza.io/order-project/order-service/domain/steps/step_42"
	shipment_delivery_problem_step "gitlab.faza.io/order-project/order-service/domain/steps/step_43"
	return_shipment_detail_delayed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_44"
	return_shipment_delivered_step "gitlab.faza.io/order-project/order-service/domain/steps/step_50"
	return_shipment_delivery_pending_step "gitlab.faza.io/order-project/order-service/domain/steps/step_51"
	return_shipment_delivery_delayed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_52"
	return_shipment_delivery_problem_step "gitlab.faza.io/order-project/order-service/domain/steps/step_53"
	return_shipment_canceled_step "gitlab.faza.io/order-project/order-service/domain/steps/step_54"
	return_shipment_success_step "gitlab.faza.io/order-project/order-service/domain/steps/step_55"
	pay_to_buyer_step "gitlab.faza.io/order-project/order-service/domain/steps/step_80"
	pay_to_buyer_success_step "gitlab.faza.io/order-project/order-service/domain/steps/step_81"
	pay_to_buyer_failed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_82"
	pay_to_seller_step "gitlab.faza.io/order-project/order-service/domain/steps/step_90"
	pay_to_seller_success_step "gitlab.faza.io/order-project/order-service/domain/steps/step_91"
	pay_to_seller_failed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_92"
	pay_to_market_step "gitlab.faza.io/order-project/order-service/domain/steps/step_93"
	pay_to_market_success_step "gitlab.faza.io/order-project/order-service/domain/steps/step_94"
	pay_to_market_failed_step "gitlab.faza.io/order-project/order-service/domain/steps/step_95"
	message "gitlab.faza.io/protos/order/general"
)

var flowManager *iFlowManagerImpl

type iFlowManagerImpl struct {
	nameStepsMap		map[string]steps.IStep
	indexStepsMap		map[int]steps.IStep
}

func init() {
	flowManager = new(iFlowManagerImpl)
	flowManager.nameStepsMap = make(map[string]steps.IStep, 64)
	flowManager.indexStepsMap = make(map[int]steps.IStep, 64)
}

func Get() IFlowManager {
	return flowManager
}

func GetNameStepsMap() map[string]steps.IStep {
	return flowManager.nameStepsMap
}

func GetIndexStepsMap() map[int]steps.IStep {
	return flowManager.indexStepsMap
}

func (flowManager *iFlowManagerImpl) SetupFlowManager() error {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	//////////////////////////////////////////////////////////////////
	// Pay To Market
	// create empty step93 that required for step95
	step93 := pay_to_market_step.New(emptyStep, emptyStep, emptyState...)
	baseStep93 := step93.(steps.IBaseStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step93.Index()] = step93
	flowManager.nameStepsMap[step93.Name()] = step93

	flowManager.createStep94()
	flowManager.createStep95()
	flowManager.createStep93(baseStep93)

	//////////////////////////////////////////////////////////////////
	// Pay To Seller
	// create empty step90 which is required for step92
	step90 := pay_to_seller_step.New(emptyStep, emptyStep, emptyState...)
	baseStep90 := step90.(steps.IBaseStep)

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
	baseStep80 := step80.(steps.IBaseStep)
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

	return nil
}

func (flowManager *iFlowManagerImpl) createStep94() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(1, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.MarketFinalizeAction))

	// Create Notification State
	notificationState := notification_state.New(0, []states.IState{finalizeState}, emptyState,
		notification_action.NewOf(notification_action.MarketNotificationAction))
	step94 := pay_to_market_success_step.New(emptyStep, emptyStep, notificationState, finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step94.Index()] = step94
	flowManager.nameStepsMap[step94.Name()] = step94
}

func (flowManager *iFlowManagerImpl) createStep95() {
	var emptyState 	[]states.IState

	step93 := flowManager.indexStepsMap[93]
	step94 := flowManager.indexStepsMap[94]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		retry_action.RetryAction: step93,
		manual_payment_action.ManualPaymentToMarketAction: step94,
	}

	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	manualPaymentState := manual_payment_state.New(2, []states.IState{nextToStep}, emptyState,
		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToMarketAction))
	operatorNotificationState := notification_state.New(1, []states.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
	retryState := retry_state.New(0, []states.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))

	step95 := pay_to_market_failed_step.New([]steps.IStep{step94}, []steps.IStep{step93},
				retryState, operatorNotificationState, manualPaymentState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step95.Index()] = step95
	flowManager.nameStepsMap[step95.Name()] = step95
}

func (flowManager *iFlowManagerImpl) createStep93(baseStep93 steps.IBaseStep) {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step94 := flowManager.indexStepsMap[94]
	step95 := flowManager.indexStepsMap[95]

	payToMarketActions := pay_to_market_action.NewOf(pay_to_market_action.SuccessAction, pay_to_market_action.FailedAction)
	actionStepMap := map[actions.IEnumAction]steps.IStep{
		pay_to_market_action.SuccessAction: step94,
		pay_to_market_action.FailedAction: step95,
	}

	nextToStepState := next_to_step_state.New(1, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	payToMarketState := pay_to_market_state.New(0, []states.IState{nextToStepState}, emptyState, payToMarketActions)

	// TODO check change baseStep3 cause of change Step93 settings
	baseStep93.BaseStep().SetChildes([]steps.IStep{step94, step95})
	baseStep93.BaseStep().SetParents(emptyStep)
	baseStep93.BaseStep().SetStates(payToMarketState, nextToStepState)

	// add to flowManager maps
	//flowManager.indexStepsMap[step93.Index()] = step93
	//flowManager.nameStepsMap[step93.Name()] = step93
}

func (flowManager *iFlowManagerImpl) createStep91() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step93 := flowManager.indexStepsMap[93]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		next_to_step_action.NextToStepAction: step93,
	}

	nextToStep93 := next_to_step_state.New(1, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)

	// Create Notification State
	notificationState := notification_state.New(0, []states.IState{nextToStep93}, emptyState,
		notification_action.NewOf(notification_action.SellerNotificationAction))

	step91 := pay_to_seller_success_step.New([]steps.IStep{step93}, emptyStep, notificationState, nextToStep93)

	// add to flowManager maps
	flowManager.indexStepsMap[step91.Index()] = step91
	flowManager.nameStepsMap[step91.Name()] = step91
}

// TODO: checking flow and next to step states sequence
func (flowManager *iFlowManagerImpl) createStep92() {
	var emptyState 	[]states.IState

	step90 := flowManager.indexStepsMap[90]
	step91 := flowManager.indexStepsMap[91]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		retry_action.RetryAction: step90,
		manual_payment_action.ManualPaymentToSellerAction: step91,
	}

	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)

	manualPaymentState := manual_payment_state.New(2, []states.IState{nextToStep}, emptyState,
		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToSellerAction))

	operatorNotificationState := notification_state.New(1, []states.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
	retryState := retry_state.New(0, []states.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))

	step92 := pay_to_seller_failed_step.New([]steps.IStep{step91}, []steps.IStep{step90},
		retryState, operatorNotificationState, manualPaymentState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step92.Index()] = step92
	flowManager.nameStepsMap[step92.Name()] = step92
}

// TODO settlement stock must be call once as a result it must be save in db
func (flowManager *iFlowManagerImpl) createStep90(baseStep90 steps.IBaseStep) {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step91 := flowManager.indexStepsMap[91]
	step92 := flowManager.indexStepsMap[92]

	payToSellerActions := pay_to_seller_action.NewOf(pay_to_seller_action.SuccessAction, pay_to_seller_action.FailedAction)
	actionStepMap := map[actions.IEnumAction]steps.IStep{
		pay_to_seller_action.SuccessAction: step91,
		pay_to_seller_action.FailedAction: step92,
	}

	nextToStepState := next_to_step_state.New(1, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	payToSellerState := pay_to_seller_state.New(0, []states.IState{nextToStepState}, emptyState, payToSellerActions)

	// TODO check change baseStep90 cause of change Step90 settings
	baseStep90.BaseStep().SetChildes([]steps.IStep{step91, step92})
	baseStep90.BaseStep().SetParents(emptyStep)
	baseStep90.BaseStep().SetStates(payToSellerState, nextToStepState)

	// add to flowManager maps
	//flowManager.indexStepsMap[step90.Index()] = step90
	//flowManager.nameStepsMap[step90.Name()] = step90
}

func (flowManager *iFlowManagerImpl) createStep81() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(1, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.BuyerFinalizeAction))

	// Create Notification State
	notificationState := notification_state.New(0, []states.IState{finalizeState}, emptyState,
		notification_action.NewOf(notification_action.BuyerNotificationAction))
	step81 := pay_to_buyer_success_step.New(emptyStep, emptyStep, notificationState, finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step81.Index()] = step81
	flowManager.nameStepsMap[step81.Name()] = step81
}

// TODO: checking flow and next to step states sequence
func (flowManager *iFlowManagerImpl) createStep82() {
	var emptyState 	[]states.IState

	step80 := flowManager.indexStepsMap[80]
	step81 := flowManager.indexStepsMap[81]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		retry_action.RetryAction: step80,
		manual_payment_action.ManualPaymentToBuyerAction: step81,
	}

	nextToStep := next_to_step_state.New(3, emptyState, emptyState,
		next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)

	manualPaymentState := manual_payment_state.New(2, []states.IState{nextToStep}, emptyState,
		manual_payment_action.NewOf(manual_payment_action.ManualPaymentToBuyerAction))
	operatorNotificationState := notification_state.New(1, []states.IState{manualPaymentState, nextToStep}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))
	retryState := retry_state.New(0, []states.IState{operatorNotificationState, nextToStep}, emptyState, retry_action.NewOf(retry_action.RetryAction))

	step82 := pay_to_buyer_failed_step.New([]steps.IStep{step81}, []steps.IStep{step80},
		retryState,operatorNotificationState, manualPaymentState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step82.Index()] = step82
	flowManager.nameStepsMap[step82.Name()] = step82
}

// TODO release stock must be call once as a result it must be save in db
func (flowManager *iFlowManagerImpl) createStep80(baseStep80 steps.IBaseStep) {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step81 := flowManager.indexStepsMap[81]
	step82 := flowManager.indexStepsMap[82]

	payToBuyerActions := pay_to_buyer_action.NewOf(pay_to_buyer_action.SuccessAction, pay_to_buyer_action.FailedAction)
	actionStepMap := map[actions.IEnumAction]steps.IStep{
		pay_to_buyer_action.SuccessAction: step81,
		pay_to_buyer_action.FailedAction: step82,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	payToBuyerState := pay_to_buyer_state.New(1, []states.IState{nextToStepState}, emptyState, payToBuyerActions)
	stockReleaseActionState := stock_action_state.New(0, []states.IState{payToBuyerState}, emptyState, stock_action.NewOf(stock_action.ReleasedAction))

	// TODO check change baseStep80 cause of change Step82 settings
	baseStep80.BaseStep().SetChildes([]steps.IStep{step81, step82})
	baseStep80.BaseStep().SetParents(emptyStep)
	baseStep80.BaseStep().SetStates(stockReleaseActionState, payToBuyerState, nextToStepState)

	// add to flowManager maps
	//flowManager.indexStepsMap[step93.Index()] = step93
	//flowManager.nameStepsMap[step93.Name()] = step93
}

func (flowManager *iFlowManagerImpl) createStep55() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)

	step55 := return_shipment_success_step.New([]steps.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step55.Index()] = step55
	flowManager.nameStepsMap[step55.Name()] = step55
}

func (flowManager *iFlowManagerImpl) createStep54() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step90 := flowManager.indexStepsMap[90]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		next_to_step_action.NextToStepAction: step90,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	step54 := return_shipment_canceled_step.New([]steps.IStep{step90}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step54.Index()] = step54
	flowManager.nameStepsMap[step54.Name()] = step54
}

func (flowManager *iFlowManagerImpl) createStep53() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step55 := flowManager.indexStepsMap[55]
	step54 := flowManager.indexStepsMap[54]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		operator_action.ReturnCanceledAction: step54,
		operator_action.ReturnedAction: step55,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	operatorActionState := operator_action_state.New(1, []states.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnCanceledAction, operator_action.ReturnedAction))
	notificationState := notification_state.New(0, []states.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step53 := return_shipment_delivery_problem_step.New([]steps.IStep{step54, step55}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step53.Index()] = step53
	flowManager.nameStepsMap[step53.Name()] = step53
}

func (flowManager *iFlowManagerImpl) createStep50() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step53 := flowManager.indexStepsMap[53]
	step55 := flowManager.indexStepsMap[55]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		seller_action.NeedSupportAction: step53,
		seller_action.ApprovedAction: step55,
		scheduler_action.AutoApprovedAction: step55,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.AutoApprovedAction))
	sellerApprovedActionState := seller_action_state.New("Return_Seller_Approval_Action_State", 2, []states.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.ApprovedAction, seller_action.NeedSupportAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{sellerApprovedActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step50 := return_shipment_delivered_step.New([]steps.IStep{step53, step55}, emptyStep,
		notificationActionState, composeActorsActionState, sellerApprovedActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step50.Index()] = step50
	flowManager.nameStepsMap[step50.Name()] = step50
}

func (flowManager *iFlowManagerImpl) createStep52() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step50 := flowManager.indexStepsMap[50]
	step54 := flowManager.indexStepsMap[54]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		operator_action.ReturnDeliveredAction: step50,
		operator_action.ReturnCanceledAction: step54,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	operatorActionState := operator_action_state.New(1, []states.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnCanceledAction, operator_action.ReturnDeliveredAction))
	notificationState := notification_state.New(0, []states.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step52 := return_shipment_delivery_delayed_step.New([]steps.IStep{step54, step50}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step52.Index()] = step52
	flowManager.nameStepsMap[step52.Name()] = step52
}

// TODO schedulers need a config for timeout or auto approved
func (flowManager *iFlowManagerImpl) createStep51() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step50 := flowManager.indexStepsMap[50]
	step52 := flowManager.indexStepsMap[52]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		buyer_action.NeedSupportAction: step52,
		scheduler_action.NoActionForXDaysTimeoutAction: step50,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	buyerNeedSupportActionState := buyer_action_state.New("Buyer_Return_Shipment_Delivery_Pending_Action_State", 2, []states.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.NeedSupportAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{buyerNeedSupportActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step51 := return_shipment_delivery_pending_step.New([]steps.IStep{step50, step52}, emptyStep,
		notificationActionState, composeActorsActionState, buyerNeedSupportActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step51.Index()] = step51
	flowManager.nameStepsMap[step51.Name()] = step51
}

func (flowManager *iFlowManagerImpl) createStep42() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step51 := flowManager.indexStepsMap[51]
	step50 := flowManager.indexStepsMap[50]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		buyer_action.DeliveredAction: step50,
		scheduler_action.WaitForShippingDaysTimeoutAction: step51,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.WaitForShippingDaysTimeoutAction))
	buyerReturnShipmentDeliveryActionState := buyer_action_state.New("Buyer_Return_Shipment_Delivery_Action_State", 2, []states.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.DeliveredAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{buyerReturnShipmentDeliveryActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step42 := return_shipped_step.New([]steps.IStep{step50, step51}, emptyStep,
		notificationActionState, composeActorsActionState, buyerReturnShipmentDeliveryActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step42.Index()] = step42
	flowManager.nameStepsMap[step42.Name()] = step42
}

func (flowManager *iFlowManagerImpl) createStep40() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step90 := flowManager.indexStepsMap[90]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		next_to_step_action.NextToStepAction: step90,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	step40 := shipment_success_step.New([]steps.IStep{step90}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step40.Index()] = step40
	flowManager.nameStepsMap[step40.Name()] = step40
}

func (flowManager *iFlowManagerImpl) createStep44() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step40 := flowManager.indexStepsMap[40]
	step42 := flowManager.indexStepsMap[42]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		buyer_action.EnterReturnShipmentDetailAction: step42,
		scheduler_action.WaitXDaysTimeoutAction:      step40,
	}

	nextToStepState := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStepState}, emptyState, scheduler_action.NewOf(scheduler_action.WaitXDaysTimeoutAction))
	buyerEnterReturnShipmentDetailDelayedActionState := buyer_action_state.New("Buyer_Enter_Return_Shipment_Detail_Delayed_Action_State", 2, []states.IState{nextToStepState}, emptyState, buyer_action.NewOf(buyer_action.EnterReturnShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{buyerEnterReturnShipmentDetailDelayedActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step44 := return_shipment_detail_delayed_step.New([]steps.IStep{step40, step42}, emptyStep,
		notificationState, composeActorsActionState, buyerEnterReturnShipmentDetailDelayedActionState,
		schedulerActionState,  nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step44.Index()] = step44
	flowManager.nameStepsMap[step44.Name()] = step44
}

func (flowManager *iFlowManagerImpl) createStep41() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step42 := flowManager.indexStepsMap[42]
	step44 := flowManager.indexStepsMap[44]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		buyer_action.EnterReturnShipmentDetailAction:   step42,
		scheduler_action.NoActionForXDaysTimeoutAction: step44,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	BuyerEnterReturnShipmentDetailAction := buyer_action_state.New("Buyer_Enter_Return_Shipment_Detail_Action_State", 2, []states.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.EnterReturnShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{BuyerEnterReturnShipmentDetailAction, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step41 := return_shipment_pending_step.New([]steps.IStep{step42, step44}, emptyStep,
		notificationActionState, composeActorsActionState, BuyerEnterReturnShipmentDetailAction, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step41.Index()] = step41
	flowManager.nameStepsMap[step41.Name()] = step41
}

func (flowManager *iFlowManagerImpl) createStep43() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step41 := flowManager.indexStepsMap[41]
	step40 := flowManager.indexStepsMap[40]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		operator_action.ReturnedAction: step41,
		operator_action.ReturnCanceledAction: step40,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	operatorActionState := operator_action_state.New(1, []states.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.ReturnedAction, operator_action.ReturnCanceledAction))
	notificationState := notification_state.New(0, []states.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step43 := shipment_delivery_problem_step.New([]steps.IStep{step40, step41}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step43.Index()] = step43
	flowManager.nameStepsMap[step43.Name()] = step43
}

func (flowManager *iFlowManagerImpl) createStep36() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	step36 := shipment_canceled_step.New([]steps.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step36.Index()] = step36
	flowManager.nameStepsMap[step36.Name()] = step36
}

func (flowManager *iFlowManagerImpl) createStep32() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step43 := flowManager.indexStepsMap[43]
	step41 := flowManager.indexStepsMap[41]
	step40 := flowManager.indexStepsMap[40]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		scheduler_action.AutoApprovedAction: step40,
		buyer_action.ApprovedAction: step40,
		buyer_action.NeedSupportAction: step43,
		buyer_action.ReturnIfPossibleAction: step41,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.AutoApprovedAction))
	buyerApprovalActionState := buyer_action_state.New("Buyer_Approval_Action_State", 2, []states.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.ApprovedAction, buyer_action.NeedSupportAction, buyer_action.ReturnIfPossibleAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{buyerApprovalActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step32 := shipment_delivered_step.New([]steps.IStep{step43, step41, step40}, emptyStep,
		notificationActionState, composeActorsActionState, buyerApprovalActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step32.Index()] = step32
	flowManager.nameStepsMap[step32.Name()] = step32
}

func (flowManager *iFlowManagerImpl) createStep35() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step36 := flowManager.indexStepsMap[36]
	step32 := flowManager.indexStepsMap[32]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		operator_action.CanceledAction: step36,
		operator_action.DeliveredAction: step32,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	operatorActionState := operator_action_state.New(1, []states.IState{nextToStepState}, emptyState, operator_action.NewOf(operator_action.DeliveredAction, operator_action.CanceledAction))
	notificationState := notification_state.New(0, []states.IState{operatorActionState}, emptyState, notification_action.NewOf(notification_action.OperatorNotificationAction))

	step35 := shipment_delivery_delayed_step.New([]steps.IStep{step32, step36}, emptyStep,
		notificationState, operatorActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step35.Index()] = step35
	flowManager.nameStepsMap[step35.Name()] = step35
}

func (flowManager *iFlowManagerImpl) createStep34() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step35 := flowManager.indexStepsMap[35]
	step32 := flowManager.indexStepsMap[32]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		seller_action.NeedSupportAction: step35,
		scheduler_action.NoActionForXDaysTimeoutAction: step32,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	sellerShipmentDeliveryPendingActionState := seller_action_state.New("Seller_Shipment_Delivery_Pending_Action_State", 2, []states.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.NeedSupportAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{sellerShipmentDeliveryPendingActionState,schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step34 := shipment_delivery_pending_step.New([]steps.IStep{step35, step32}, emptyStep,
		notificationActionState, composeActorsActionState, sellerShipmentDeliveryPendingActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step34.Index()] = step34
	flowManager.nameStepsMap[step34.Name()] = step34
}

func (flowManager *iFlowManagerImpl) createStep31() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step32 := flowManager.indexStepsMap[32]
	step34 := flowManager.indexStepsMap[34]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		seller_action.DeliveredAction: step32,
		scheduler_action.WaitForShippingDaysTimeoutAction: step34,
	}

	nextToStep := next_to_step_state.New(5, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(4, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.WaitForShippingDaysTimeoutAction))
	sellerShipmentDeliveryActionState := buyer_action_state.New("Seller_Shipment_Delivery_Action_State", 3, []states.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.DeliveredAction))
	composeActorsActionState := system_action_state.New(2, []states.IState{sellerShipmentDeliveryActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(1, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))
	stockSettlementActionState := stock_action_state.New(0, []states.IState{notificationActionState}, emptyState, stock_action.NewOf(stock_action.SettlementAction))

	step31 := shipped_step.New([]steps.IStep{step32, step34}, emptyStep,
		stockSettlementActionState, notificationActionState, composeActorsActionState,
		sellerShipmentDeliveryActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step31.Index()] = step31
	flowManager.nameStepsMap[step31.Name()] = step31
}

func (flowManager *iFlowManagerImpl) createStep33() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step31 := flowManager.indexStepsMap[31]
	step36 := flowManager.indexStepsMap[36]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		seller_action.EnterShipmentDetailAction: step31,
		buyer_action.CanceledAction: step36,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	buyerWaitForSellerEnterShipmentDetailDelayedActionState := buyer_action_state.New("Buyer_Wait_For_Seller_Enter_Shipment_Detail_Delayed_Action_State", 3, []states.IState{nextToStep}, emptyState, buyer_action.NewOf(buyer_action.CanceledAction))
	sellerEnterShipmentDetailDelayedActionState := seller_action_state.New("Seller_Enter_Shipment_Detail_Delayed_Action_State", 2, []states.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.EnterShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{buyerWaitForSellerEnterShipmentDetailDelayedActionState, sellerEnterShipmentDetailDelayedActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction, notification_action.BuyerNotificationAction))

	step33 := shipment_detail_delayed_step.New([]steps.IStep{step31, step36}, emptyStep,
		notificationActionState, composeActorsActionState, sellerEnterShipmentDetailDelayedActionState, buyerWaitForSellerEnterShipmentDetailDelayedActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step33.Index()] = step33
	flowManager.nameStepsMap[step33.Name()] = step33
}

func (flowManager *iFlowManagerImpl) createStep30() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step31 := flowManager.indexStepsMap[31]
	step33 := flowManager.indexStepsMap[33]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		seller_action.EnterShipmentDetailAction: step31,
		scheduler_action.NoActionForXDaysTimeoutAction: step33,
	}

	nextToStepState := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStepState}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	sellerEnterShipmentDetailActionState := seller_action_state.New("Seller_Enter_Shipment_Detail_Action_State", 2, []states.IState{nextToStepState}, emptyState, seller_action.NewOf(seller_action.EnterShipmentDetailAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{sellerEnterShipmentDetailActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step30 := shipment_pending_step.New([]steps.IStep{step31, step33}, emptyStep,
		notificationState, composeActorsActionState, sellerEnterShipmentDetailActionState,
		schedulerActionState,  nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step30.Index()] = step30
	flowManager.nameStepsMap[step30.Name()] = step30
}

func (flowManager *iFlowManagerImpl) createStep21() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	step21 := shipment_rejected_by_seller_step.New([]steps.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step21.Index()] = step21
	flowManager.nameStepsMap[step21.Name()] = step21
}

func (flowManager *iFlowManagerImpl) createStep20() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step30 := flowManager.indexStepsMap[30]
	step21 := flowManager.indexStepsMap[21]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		seller_action.ApprovedAction: step30,
		seller_action.RejectAction: step21,
		scheduler_action.NoActionForXDaysTimeoutAction: step21,
	}

	nextToStep := next_to_step_state.New(4, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction), actionStepMap)
	schedulerActionState := scheduler_action_state.New(3, []states.IState{nextToStep}, emptyState, scheduler_action.NewOf(scheduler_action.NoActionForXDaysTimeoutAction))
	sellerApprovalActionState := seller_action_state.New("Seller_Approval_Action_State", 2, []states.IState{nextToStep}, emptyState, seller_action.NewOf(seller_action.ApprovedAction, seller_action.RejectAction))
	composeActorsActionState := system_action_state.New(1, []states.IState{sellerApprovalActionState, schedulerActionState}, emptyState, system_action.NewOf(system_action.ComposeActorsAction))
	notificationActionState := notification_state.New(0, []states.IState{composeActorsActionState}, emptyState, notification_action.NewOf(notification_action.SellerNotificationAction))

	step20 := seller_approval_pending_step.New([]steps.IStep{step30, step21}, emptyStep,
		notificationActionState, composeActorsActionState, sellerApprovalActionState, schedulerActionState, nextToStep)

	// add to flowManager maps
	flowManager.indexStepsMap[step20.Index()] = step20
	flowManager.nameStepsMap[step20.Name()] = step20
}

func (flowManager *iFlowManagerImpl) createStep14() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step80 := flowManager.indexStepsMap[80]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		next_to_step_action.NextToStepAction: step80,
	}

	nextToStepState := next_to_step_state.New(0, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	step14 := payment_rejected_step.New([]steps.IStep{step80}, emptyStep, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step14.Index()] = step14
	flowManager.nameStepsMap[step14.Name()] = step14
}

func (flowManager *iFlowManagerImpl) createStep11() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step14 := flowManager.indexStepsMap[14]
	step20 := flowManager.indexStepsMap[20]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		buyer_action.RejectAction: step14,
		buyer_action.ApprovedAction: step20,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	buyerPaymentApprovalActionState := buyer_action_state.New("Buyer_Payment_Approval_Action_State",1, []states.IState{nextToStepState}, emptyState, buyer_action.NewOf(buyer_action.ApprovedAction, buyer_action.RejectAction))
	notificationState := notification_state.New(0, []states.IState{buyerPaymentApprovalActionState}, emptyState, notification_action.NewOf(notification_action.BuyerNotificationAction))

	step11 := payment_success_step.New([]steps.IStep{step14, step20}, emptyStep,
		notificationState, buyerPaymentApprovalActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step11.Index()] = step11
	flowManager.nameStepsMap[step11.Name()] = step11
}

func (flowManager *iFlowManagerImpl) createStep12() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(2, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.PaymentFailedFinalizeAction))

	// Create Notification State
	notificationState := notification_state.New(1, []states.IState{finalizeState}, emptyState,
		notification_action.NewOf(notification_action.BuyerNotificationAction))

	stockReleaseActionState := stock_action_state.New(0, []states.IState{notificationState}, emptyState, stock_action.NewOf(stock_action.ReleasedAction))
	step12 := payment_failed_step.New(emptyStep, emptyStep, stockReleaseActionState, notificationState, finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step12.Index()] = step12
	flowManager.nameStepsMap[step12.Name()] = step12
}

func (flowManager *iFlowManagerImpl) createStep10() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step12 := flowManager.indexStepsMap[12]
	step11 := flowManager.indexStepsMap[11]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		payment_action.FailedAction: step12,
		payment_action.SuccessAction: step11,
	}

	nextToStepState := next_to_step_state.New(1, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	paymentActionState := payment_action_state.New(0, []states.IState{nextToStepState}, emptyState, payment_action.NewOf(payment_action.SuccessAction, payment_action.FailedAction))

	step10 := payment_pending_step.New([]steps.IStep{step12, step11}, emptyStep,
		paymentActionState, nextToStepState)

	// add to flowManager maps
	flowManager.indexStepsMap[step10.Index()] = step10
	flowManager.nameStepsMap[step10.Name()] = step10
}

func (flowManager *iFlowManagerImpl) createStep1() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	// Create Finalize State
	finalizeState := finalize_state.New(1, emptyState, emptyState,
		finalize_action.NewOf(finalize_action.OrderFailedFinalizeAction))

	step1 := new_order_failed_step.New(emptyStep, emptyStep,finalizeState)

	// add to flowManager maps
	flowManager.indexStepsMap[step1.Index()] = step1
	flowManager.nameStepsMap[step1.Name()] = step1
}

func (flowManager *iFlowManagerImpl) createStep0() {
	var emptyState 	[]states.IState
	var emptyStep 	[]steps.IStep

	step1 := flowManager.indexStepsMap[1]
	step10 := flowManager.indexStepsMap[10]

	actionStepMap := map[actions.IEnumAction]steps.IStep{
		new_order_action.FailedAction:  step1,
		new_order_action.SuccessAction: step10,
	}

	nextToStepState := next_to_step_state.New(2, emptyState, emptyState, next_to_step_action.NewOf(next_to_step_action.NextToStepAction) ,actionStepMap)
	stockReservedActionState := stock_action_state.New(1, []states.IState{nextToStepState}, emptyState, stock_action.NewOf(stock_action.ReservedAction))
	newOrderProcessAction := new_order_state.New(0, []states.IState{stockReservedActionState, nextToStepState}, emptyState, new_order_action.NewOf(new_order_action.SuccessAction, new_order_action.FailedAction))

	step0 := new_order_step.New([]steps.IStep{step1, step10}, emptyStep,
		newOrderProcessAction, stockReservedActionState, nextToStepState)
	// add to flowManager maps
	flowManager.indexStepsMap[step0.Index()] = step0
	flowManager.nameStepsMap[step0.Name()] = step0
}

func (flowManager iFlowManagerImpl) MessageHandler(ctx context.Context, req *message.Request) (*message.Response, error) {
	panic("implementation required")
}