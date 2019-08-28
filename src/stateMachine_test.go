package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

//func TestCreateConsumerFiles(t *testing.T) {
//	list := make(map[string]string)
//
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
//	list["PayToSeller"] = PayToSeller
//	list["PayToSellerFailed"] = PayToSellerFailed
//	list["PayToSellerSuccess"] = PayToSellerSuccess
//	list["PayToBuyerFailed"] = PayToBuyerFailed
//	list["PayToBuyerSuccess"] = PayToBuyerSuccess
//
//	for name, numbers := range list {
//		consumer, err := ioutil.ReadFile("./TmpConsumer.go")
//		if err != nil {
//			os.Exit(1)
//		}
//		consumer = bytes.ReplaceAll(consumer, []byte("CLASSNAME"), []byte(name))
//
//		logic, err := ioutil.ReadFile("./TmpState.go")
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
