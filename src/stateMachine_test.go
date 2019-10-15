package main

import (
	"errors"
	"fmt"
	"gitlab.faza.io/order-project/order-service/steps"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkJourney(path []string, checkEnd, debug bool) (int, error) {
	states := generateSM()
	stateMap := make(map[string]State)
	for _, s := range states.states {
		stateMap[s.title] = s
	}
	foundedRoutes := 0
	for i := range path {
		if i < len(path)-1 {
			if CheckNextState(path[i], path[i+1]) {
				foundedRoutes++
				if debug {
					fmt.Print(path[i], " --> ")
				}
			}
		} else {
			if len(stateMap[path[i]].toStates) == 0 {
				foundedRoutes++
				if debug {
					fmt.Println(path[i])
				}
			} else {
				if checkEnd {
					return 0, errors.New("not end of path")
				} else {
					foundedRoutes++
					if debug {
						fmt.Println(path[i])
					}
				}
			}
		}
	}
	return foundedRoutes, nil
}

func TestCheckNextStep_AssertTrue(t *testing.T) {
	currentStep := PaymentControl
	nextStep := SellerApprovalPending
	assert.True(t, CheckNextState(currentStep, nextStep))
}
func TestCheckNextStep_AssertFalse(t *testing.T) {
	currentStep := PaymentPending
	nextStep := ShipmentPending
	assert.False(t, CheckNextState(currentStep, nextStep))
}
func TestCheckHappyPath_shortestWithoutAnyIssue(t *testing.T) {
	path := []string{PaymentPending, PaymentSuccess, SellerApprovalPending, ShipmentPending, Shipped, ShipmentDelivered,
		ShipmentSuccess, PayToSeller, PayToSellerSuccess, PayToMarket, PayToMarketSuccess}
	foundedRoutes, err := checkJourney(path, true, false)
	assert.Nil(t, err)
	assert.Equal(t, len(path), foundedRoutes)
}
func TestCheckWorstCase_longestWithIssuePayBackToSeller(t *testing.T) {
	path := []string{PaymentPending, PaymentSuccess, PaymentControl, SellerApprovalPending, ShipmentPending,
		ShipmentDetailDelayed, Shipped, ShipmentDeliveryPending, ShipmentDeliveryDelayed, ShipmentDelivered,
		ShipmentDeliveryProblem, ReturnShipmentPending, ReturnShipmentDetailDelayed, ReturnShipped,
		ReturnShipmentDeliveryPending, ReturnShipmentDeliveryDelayed, ReturnShipmentDelivered,
		ReturnShipmentDeliveryProblem, ReturnShipmentCanceled, PayToSeller, PayToSellerFailed, PayToSeller,
		PayToSellerFailed, PayToSellerSuccess, PayToMarket, PayToMarketFailed, PayToMarket,
		PayToMarketFailed, PayToMarketSuccess}
	foundedRoutes, err := checkJourney(path, true, false)
	assert.Nil(t, err)
	assert.Equal(t, len(path), foundedRoutes)
}
func TestPayToBuyerHappyPath_AssertTrue(t *testing.T) {
	path := []string{PayToBuyer, PayToBuyerSuccess}
	foundedRoutes, err := checkJourney(path, true, false)
	assert.Nil(t, err)
	assert.Equal(t, len(path), foundedRoutes)
}
func TestPayToBuyerWithFailure_AssertTrue(t *testing.T) {
	path := []string{PayToBuyer, PayToBuyerFailed, PayToBuyer, PayToBuyerSuccess}
	foundedRoutes, err := checkJourney(path, true, false)
	assert.Nil(t, err)
	assert.Equal(t, len(path), foundedRoutes)
}
func TestPayToSellerHappyPath_AssertTrue(t *testing.T) {
	path := []string{PayToSeller, PayToSellerSuccess}
	foundedRoutes, err := checkJourney(path, false, false)
	assert.Nil(t, err)
	assert.Equal(t, len(path), foundedRoutes)
}
func TestPayToSellerWithFailure_AssertTrue(t *testing.T) {
	path := []string{PayToSeller, PayToSellerFailed, PayToSeller, PayToSellerSuccess}
	foundedRoutes, err := checkJourney(path, false, false)
	assert.Nil(t, err)
	assert.Equal(t, len(path), foundedRoutes)
}
func TestPayToSellerWithFailure_LongestAssertTrue(t *testing.T) {
	path := []string{PayToSeller, PayToSellerFailed, PayToSeller, PayToSellerFailed, PayToSellerSuccess}
	foundedRoutes, err := checkJourney(path, false, false)
	assert.Nil(t, err)
	assert.Equal(t, len(path), foundedRoutes)
}
func TestCheckPrevStep_AssertTrue(t *testing.T) {
	currentStep := PaymentControl
	prevStep := PaymentSuccess
	assert.True(t, CheckPrevState(currentStep, prevStep))
}
func TestCheckPrevStep_AssertFalse(t *testing.T) {
	currentStep := PaymentControl
	nextStep := PaymentPending
	assert.False(t, CheckNextState(currentStep, nextStep))
}
func TestNotifySellerForNewOrder(t *testing.T) {
	ppr := steps.PaymentPendingRequest{}
	item := steps.Item{
		Seller: steps.ItemSeller{
			Email:  "farzan.dalaee@gmail.com",
			Mobile: "+989121938710",
			Title:  "Faza.IO",
		},
	}
	ppr.Items = append(ppr.Items, item)
	err := NotifySellerForNewOrder(ppr)
	assert.Nil(t, err)
}

//func TestCreateConsumerFiles(t *testing.T) {
//	list := make(map[string]string)
//
//	//	list["PaymentPending"] = PaymentPending
//	//	list["PaymentSuccess"] = PaymentSuccess
//	list["PaymentFailed"] = PaymentFailed
//	list["PaymentControl"] = PaymentControl
//	list["PaymentRejected"] = PaymentRejected
//	list["SellerApprovalPending"] = SellerApprovalPending
//	list["ShipmentPending"] = ShipmentPending
//	list["ShipmentDetailDelayed"] = ShipmentDetailDelayed
//	list["Shipped"] = Shipped
//	list["ShipmentDeliveryPending"] = ShipmentDeliveryPending
//	list["ShipmentDeliveryDelayed"] = ShipmentDeliveryDelayed
//	list["ShipmentDelivered"] = ShipmentDelivered
//	list["ShipmentCanceled"] = ShipmentCanceled
//	list["ShipmentDeliveryProblem"] = ShipmentDeliveryProblem
//	list["ReturnShipmentPending"] = ReturnShipmentPending
//	list["ReturnShipmentDetailDelayed"] = ReturnShipmentDetailDelayed
//	list["ShipmentSuccess"] = ShipmentSuccess
//	list["ReturnShipped"] = ReturnShipped
//	list["ReturnShipmentDeliveryPending"] = ReturnShipmentDeliveryPending
//	list["ReturnShipmentDeliveryDelayed"] = ReturnShipmentDeliveryDelayed
//	list["ReturnShipmentDelivered"] = ReturnShipmentDelivered
//	list["ReturnShipmentDeliveryProblem"] = ReturnShipmentDeliveryProblem
//	list["ReturnShipmentCanceled"] = ReturnShipmentCanceled
//	list["ReturnShipmentSuccess"] = ReturnShipmentSuccess
//	list["ShipmentRejectedBySeller"] = ShipmentRejectedBySeller
//	list["PayToBuyer"] = PayToBuyer
//	list["PayToBuyerFailed"] = PayToBuyerFailed
//	list["PayToBuyerSuccess"] = PayToBuyerSuccess
//	list["PayToSeller"] = PayToSeller
//	list["PayToSellerFailed"] = PayToSellerFailed
//	list["PayToSellerSuccess"] = PayToSellerSuccess
//	list["PayToMarket"] = PayToMarket
//	list["PayToMarketFailed"] = PayToMarketFailed
//	list["PayToMarketSuccess"] = PayToMarketSuccess
//
//	for name, numbers := range list {
//		consumer, err := ioutil.ReadFile("./TmpConsumer.txt")
//		if err != nil {
//			os.Exit(1)
//		}
//		consumer = bytes.ReplaceAll(consumer, []byte("CLASSNAME"), []byte(name))
//
//		logic, err := ioutil.ReadFile("./TmpState.txt")
//		if err != nil {
//			os.Exit(1)
//		}
//		logic = bytes.ReplaceAll(logic, []byte("CLASSNAME"), []byte(name))
//
//		filenameConsumer := fmt.Sprintf("%s-%sConsumer.go", numbers[:2], name)
//		err = ioutil.WriteFile(filenameConsumer, consumer, os.ModePerm)
//		if err != nil {
//			fmt.Println(err)
//			os.Exit(1)
//		}
//
//		filenameLogic := fmt.Sprintf("%s-%s.go", numbers[:2], name)
//		err = ioutil.WriteFile(filenameLogic, logic, os.ModePerm)
//		if err != nil {
//			fmt.Println(err)
//			os.Exit(1)
//		}
//		fmt.Println(filenameConsumer, "created")
//		fmt.Println(filenameLogic, "created")
//	}
//}
