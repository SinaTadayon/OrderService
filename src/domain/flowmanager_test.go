package domain

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/actions/actives"
	"gitlab.faza.io/order-project/order-service/domain/actions/actors"
	"gitlab.faza.io/order-project/order-service/domain/states"
	next_to_step_state "gitlab.faza.io/order-project/order-service/domain/states/launcher/nextstep"
	"gitlab.faza.io/order-project/order-service/domain/steps"
	"strconv"
	"testing"
)

func TestFlowManagerSteps(t *testing.T) {

	flowManager := iFlowManagerImpl{}
	flowManager.nameStepsMap = make(map[string]steps.IStep, 64)
	flowManager.indexStepsMap = make(map[int]steps.IStep, 64)

	assert.Nil(t, flowManager.setupFlowManager())

	if err := stepValidation(flowManager.GetIndexStepsMap()[0], 0, "0.New_Order", []int{1,10}); err != nil {
		t.Fatalf("validate step0 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[1], 1, "1.New_Order_Failed", []int{}); err != nil {
		t.Fatalf("validate step1 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[10], 10, "10.Payment_Pending", []int{11, 12}); err != nil {
		t.Fatalf("validate step10 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[12], 12, "12.Payment_Failed", []int{}); err != nil {
		t.Fatalf("validate step12 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[11], 11, "11.Payment_Success", []int{20,14}); err != nil {
		t.Fatalf("validate step11 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[14], 14, "14.Payment_Rejected", []int{80}); err != nil {
		t.Fatalf("validate step14 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[20], 20, "20.Seller_Approval_Pending", []int{30,21}); err != nil {
		t.Fatalf("validate step20 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[21], 21, "21.Shipment_Rejected_By_Seller", []int{80}); err != nil {
		t.Fatalf("validate step21 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[30], 30, "30.Shipment_Pending", []int{31,33}); err != nil {
		t.Fatalf("validate step30 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[31], 31, "31.Shipped", []int{32,34}); err != nil {
		t.Fatalf("validate step31 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[32], 32, "32.Shipment_Delivered", []int{40,41,43}); err != nil {
		t.Fatalf("validate step32 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[33], 33, "33.Shipment_Detail_Delayed", []int{31,36}); err != nil {
		t.Fatalf("validate step33 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[34], 34, "34.Shipment_Delivery_Pending", []int{32,35}); err != nil {
		t.Fatalf("validate step34 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[35], 35, "35.Shipment_Delivery_Delayed", []int{32,36}); err != nil {
		t.Fatalf("validate step35 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[36], 36, "36.Shipment_Canceled", []int{80}); err != nil {
		t.Fatalf("validate step36 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[40], 40, "40.Shipment_Success", []int{90}); err != nil {
		t.Fatalf("validate step40 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[41], 41, "41.Return_Shipment_Pending", []int{42,44}); err != nil {
		t.Fatalf("validate step41 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[42], 42, "42.Return_Shipped", []int{50,51}); err != nil {
		t.Fatalf("validate step42 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[43], 43, "43.Shipment_Delivery_Problem", []int{40,41}); err != nil {
		t.Fatalf("validate step43 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[44], 44, "44.Return_Shipment_Detail_Delayed", []int{40,42}); err != nil {
		t.Fatalf("validate step44 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[50], 50, "50.Return_Shipment_Delivered", []int{53,55}); err != nil {
		t.Fatalf("validate step50 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[51], 51, "51.Return_Shipment_Delivery_Pending", []int{50,52}); err != nil {
		t.Fatalf("validate step51 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[52], 52, "52.Return_Shipment_Delivery_Delayed", []int{50,54}); err != nil {
		t.Fatalf("validate step52 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[53], 53, "53.Return_Shipment_Delivery_Problem", []int{54,55}); err != nil {
		t.Fatalf("validate step53 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[54], 54, "54.Return_Shipment_Canceled", []int{90}); err != nil {
		t.Fatalf("validate step54 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[55], 55, "55.Return_Shipment_Success", []int{80}); err != nil {
		t.Fatalf("validate step55 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[80], 80, "80.Pay_To_Buyer", []int{81,82}); err != nil {
		t.Fatalf("validate step80 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[81], 81, "81.Pay_To_Buyer_Success", []int{}); err != nil {
		t.Fatalf("validate step81 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[82], 82, "82.Pay_To_Buyer_Failed", []int{81}); err != nil {
		t.Fatalf("validate step82 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[90], 90, "90.Pay_To_Seller", []int{91,92}); err != nil {
		t.Fatalf("validate step90 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[91], 91, "91.Pay_To_Seller_Success", []int{93}); err != nil {
		t.Fatalf("validate step91 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[92], 92, "92.Pay_To_Seller_Failed", []int{91}); err != nil {
		t.Fatalf("validate step92 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[93], 93, "93.Pay_To_Market", []int{94,95}); err != nil {
		t.Fatalf("validate step93 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[94], 94, "94.Pay_To_Market_Success", []int{}); err != nil {
		t.Fatalf("validate step94 failed: %s\n", err)
	}

	if err := stepValidation(flowManager.GetIndexStepsMap()[95], 95, "95.Pay_To_Market_Failed", []int{94}); err != nil {
		t.Fatalf("validate step95 failed: %s\n", err)
	}


	//keys := make([]int, 0, len(flowManager.GetIndexStepsMap()))
	//for k := range flowManager.GetIndexStepsMap() {
	//	keys = append(keys, k)
	//}
	//sort.Ints(keys)
	//
	//for _, k := range keys {
	//	step := flowManager.GetIndexStepsMap()[k]
	//	fmt.Printf("\n\n$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$\n")
	//	fmt.Printf("step.Name(): %s\nstep.Index(): %d\n", step.Name(), step.Index())
	//	fmt.Printf("step.Childes(): %s\n", step.Childes())
	//	fmt.Printf("step.Parents(): %s\n", step.Parents())
	//	fmt.Printf("step.States(): %s\n", step.States())
	//	traversState(step.States())
	//}
}

func traversState(states []states.IState) {
	for _, state := range states {
		fmt.Printf("################################################\n")
		fmt.Printf("************* => state.Name(): %s\n", state.Name())
		fmt.Printf("************* => state.Index(): %d\n", state.Index())
		fmt.Printf("************* => state.Action Type: %s\n", state.Actions().ActionType())

		if state.Actions().ActionType() == actions.ActorAction {
			actorAction := state.Actions().(actors.IActorAction)
			fmt.Printf("************* => actor type: %s\n", actorAction.ActorType())
			fmt.Printf("************* => actor enum actions: %s\n", actorAction.ActionEnums())
		} else {
			activeAction := state.Actions().(actives.IActiveAction)
			fmt.Printf("************* => active type: %s\n", activeAction.ActiveType())
			if activeAction.ActiveType() == actives.NextToStepAction {
				nextToStepState := state.(next_to_step_state.INextToStep)
				for action, step := range nextToStepState.ActionStepMap() {
					fmt.Printf("************* => ActionMap -> action: %s, stepIndex: %d\n",action, step.Index())
				}
			} else {
				fmt.Printf("************* => active enum actions: %s\n", activeAction.ActionEnums())
			}
		}
		fmt.Printf("************* => state.Parents(): %s\n", state.Parents())
		fmt.Printf("************* => state.Childes(): %s\n", state.Childes())
	}
}

func stepValidation(step steps.IStep, checkIndex int, checkName string, childesIndex []int) error {
	if checkIndex != step.Index() {
		return errors.New(step.Name() + " index invalid")
	}

	if checkName != step.Name() {
		return errors.New(step.Name() + " name invalid")
	}

	if len(step.Childes()) != len(childesIndex) {
		return errors.New(step.Name() + " invalid childes count")
	}

	for _, index := range childesIndex {
		var findIndex = false
		for _, childStep := range step.Childes() {
			if childStep.Index() == index {
				findIndex = true
				break
			}
		}
		if !findIndex {
			return errors.New(step.Name() + " required child with index " + strconv.Itoa(index) + " not found")
		}
	}

	return nil
}