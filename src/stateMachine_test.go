package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func genrateSM() *StateMachine {
	SM := NewStateMachine()

	paymentPending := State{
		title:      PaymentPending,
		fromStates: []string{PaymentPending},
		toStates:   []string{PaymentSuccess, PaymentFailed},
	}
	SM.add(paymentPending)

	paymentFailed := State{
		title:      PaymentFailed,
		fromStates: []string{PaymentPending},
		toStates:   []string{},
	}
	SM.add(paymentFailed)

	paymentSuccess := State{
		title:      PaymentSuccess,
		fromStates: []string{PaymentPending},
		toStates:   []string{PaymentControl, PaymentRejected, SellerApprovalPending},
	}
	SM.add(paymentSuccess)

	paymentControl := State{
		title:      PaymentControl,
		fromStates: []string{PaymentSuccess},
		toStates:   []string{PaymentRejected, SellerApprovalPending},
	}
	SM.add(paymentControl)

	sellerApprovalPending := State{
		title:      SellerApprovalPending,
		fromStates: []string{PaymentSuccess, PaymentControl},
		toStates:   []string{ShipmentPending, ShipmentRejectedBySeller},
	}
	SM.add(sellerApprovalPending)

	paymentRejected := State{
		title:      PaymentRejected,
		fromStates: []string{PaymentControl, PaymentSuccess},
		toStates:   []string{PayToBuyer},
	}
	SM.add(paymentRejected)

	shipmentPending := State{
		title:      ShipmentPending,
		fromStates: []string{SellerApprovalPending},
		toStates:   []string{Shipped, ShipmentDetailDelayed},
	}
	SM.add(shipmentPending)

	shipmentDetailDelayed := State{
		title:      ShipmentDetailDelayed,
		fromStates: []string{ShipmentPending},
		toStates:   []string{Shipped, ShipmentCanceled},
	}
	SM.add(shipmentDetailDelayed)

	shipped := State{
		title:      Shipped,
		fromStates: []string{ShipmentPending, ShipmentDetailDelayed},
		toStates:   []string{ShipmentDelivered, ShipmentDeliveryPending},
	}
	SM.add(shipped)

	shipmentDeliveryPending := State{
		title:      ShipmentDeliveryPending,
		fromStates: []string{Shipped},
		toStates:   []string{ShipmentDelivered, ShipmentDeliveryDelayed},
	}
	SM.add(shipmentDeliveryPending)

	shipmentDeliveryDelayed := State{
		title:      ShipmentDeliveryDelayed,
		fromStates: []string{ShipmentDeliveryPending},
		toStates:   []string{ShipmentDelivered, ShipmentCanceled},
	}
	SM.add(shipmentDeliveryDelayed)

	shipmentCanceled := State{
		title:      ShipmentCanceled,
		fromStates: []string{ShipmentDeliveryDelayed, ShipmentDetailDelayed},
		toStates:   []string{PayToBuyer},
	}
	SM.add(shipmentCanceled)

	shipmentDelivered := State{
		title:      ShipmentDelivered,
		fromStates: []string{Shipped, ShipmentDeliveryDelayed, ShipmentDeliveryPending},
		toStates:   []string{ShipmentSuccess, ShipmentDeliveryProblem, ReturnShipmentPending},
	}
	SM.add(shipmentDelivered)

	shipmentDeliveryProblem := State{
		title:      ShipmentDeliveryProblem,
		fromStates: []string{ShipmentDelivered},
		toStates:   []string{ShipmentSuccess, ReturnShipmentPending},
	}
	SM.add(shipmentDeliveryProblem)

	returnShipmentPending := State{
		title:      ReturnShipmentPending,
		fromStates: []string{ShipmentDelivered, ShipmentDeliveryProblem},
		toStates:   []string{ReturnShipmentDetailDelayed, ReturnShipped},
	}
	SM.add(returnShipmentPending)

	returnShipmentDetailDelayed := State{
		title:      ReturnShipmentDetailDelayed,
		fromStates: []string{ReturnShipmentPending},
		toStates:   []string{ShipmentSuccess, ReturnShipped},
	}
	SM.add(returnShipmentDetailDelayed)

	shipmentSuccess := State{
		title:      ShipmentSuccess,
		fromStates: []string{ReturnShipmentDetailDelayed, ShipmentDeliveryProblem, ShipmentDelivered},
		toStates:   []string{PayToSeller},
	}
	SM.add(shipmentSuccess)

	returnShipped := State{
		title:      ReturnShipped,
		fromStates: []string{ReturnShipmentDetailDelayed, ReturnShipmentPending},
		toStates:   []string{ReturnShipmentDeliveryPending, ReturnShipmentDelivered},
	}
	SM.add(returnShipped)

	returnShipmentDeliveryPending := State{
		title:      ReturnShipmentDeliveryPending,
		fromStates: []string{ReturnShipped},
		toStates:   []string{ReturnShipmentDelivered, ReturnShipmentDeliveryDelayed},
	}
	SM.add(returnShipmentDeliveryPending)

	returnShipmentDeliveryDelayed := State{
		title:      ReturnShipmentDeliveryDelayed,
		fromStates: []string{ReturnShipmentDeliveryPending},
		toStates:   []string{ReturnShipmentDelivered, ReturnShipmentCanceled},
	}
	SM.add(returnShipmentDeliveryDelayed)

	returnShipmentDelivered := State{
		title:      ReturnShipmentDelivered,
		fromStates: []string{ReturnShipmentDeliveryDelayed, ReturnShipmentDeliveryPending, ReturnShipped},
		toStates:   []string{ReturnShipmentSuccess, ReturnShipmentDeliveryProblem},
	}
	SM.add(returnShipmentDelivered)

	returnShipmentDeliveryProblem := State{
		title:      ReturnShipmentDeliveryProblem,
		fromStates: []string{ReturnShipmentDelivered},
		toStates:   []string{ReturnShipmentSuccess, ReturnShipmentCanceled},
	}
	SM.add(returnShipmentDeliveryProblem)

	returnShipmentCanceled := State{
		title:      ReturnShipmentCanceled,
		fromStates: []string{ReturnShipmentDeliveryDelayed, ReturnShipmentDeliveryProblem},
		toStates:   []string{PayToSeller},
	}
	SM.add(returnShipmentCanceled)

	returnShipmentSuccess := State{
		title:      ReturnShipmentSuccess,
		fromStates: []string{ReturnShipmentDeliveryProblem, ReturnShipmentDelivered},
		toStates:   []string{PayToBuyer},
	}
	SM.add(returnShipmentSuccess)

	shipmentRejectedBySeller := State{
		title:      ShipmentRejectedBySeller,
		fromStates: []string{SellerApprovalPending},
		toStates:   []string{PayToBuyer},
	}
	SM.add(shipmentRejectedBySeller)

	payToSeller := State{
		title:      PayToSeller,
		fromStates: []string{ReturnShipmentCanceled, ShipmentSuccess},
		toStates:   []string{PayToSellerSuccess, PayToSellerFailed},
	}
	SM.add(payToSeller)

	payToBuyer := State{
		title:      PayToBuyer,
		fromStates: []string{ReturnShipmentSuccess, ShipmentRejectedBySeller, ShipmentCanceled, PaymentRejected},
		toStates:   []string{PayToBuyerSuccess, PayToBuyerFailed},
	}
	SM.add(payToBuyer)

	payToSellerFailed := State{
		title:      PayToSellerFailed,
		fromStates: []string{PayToSeller},
		toStates:   []string{PayToSellerSuccess},
	}
	SM.add(payToSellerFailed)

	payToSellerSuccess := State{
		title:      PayToSellerSuccess,
		fromStates: []string{PayToSeller, PayToSellerFailed},
		toStates:   []string{},
	}
	SM.add(payToSellerSuccess)

	payToBuyerFailed := State{
		title:      PayToBuyerFailed,
		fromStates: []string{PayToBuyer},
		toStates:   []string{PayToBuyerSuccess},
	}
	SM.add(payToBuyerFailed)

	payToBuyerSuccess := State{
		title:      PayToBuyerSuccess,
		fromStates: []string{PayToBuyer, PayToBuyerFailed},
		toStates:   []string{},
	}
	SM.add(payToBuyerSuccess)

	return SM
}

func TestCheckNextStep_AssertTrue(t *testing.T) {
	SM := genrateSM()
	currentStep := PaymentPending
	nextStep := PaymentFailed
	status := false

	for _, state := range SM.states {
		for _, from := range state.fromStates {
			for _, to := range state.toStates {
				if from == currentStep && to == nextStep {
					status = true
					return
				}
			}
		}
	}
	assert.True(t, status)
}

func TestCheckNextStep_AssertFalse(t *testing.T) {
	SM := genrateSM()
	currentStep := PaymentPending
	nextStep := SellerApprovalPending
	status := false

	for _, state := range SM.states {
		for _, from := range state.fromStates {
			for _, to := range state.toStates {
				if from == currentStep && to == nextStep {
					status = true
					return
				}
			}
		}
	}
	assert.False(t, status)
}
