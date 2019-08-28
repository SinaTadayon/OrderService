package main

const (
	PaymentPending                = "10.payment_pending"
	PaymentSuccess                = "11.payment_success"
	PaymentFailed                 = "12.payment_failed"
	PaymentControl                = "13.payment_control"
	PaymentRejected               = "14.payment_rejected"
	SellerApprovalPending         = "20.seller_approval_pending"
	ShipmentPending               = "30.shipment_pending"
	ShipmentDetailDelayed         = "33.shipment_detail_delayed"
	Shipped                       = "31.shipped"
	ShipmentDeliveryPending       = "34.shipment_delivery_pending"
	ShipmentDeliveryDelayed       = "35.shipment_delivery_delayed"
	ShipmentDelivered             = "32.shipment_delivered"
	ShipmentCanceled              = "36.shipment_canceled"
	ShipmentDeliveryProblem       = "43.shipment_delivery_problem"
	ReturnShipmentPending         = "41.return_shipment_pending"
	ReturnShipmentDetailDelayed   = "44.return_shipment_detail_delayed"
	ShipmentSuccess               = "40.shipment_success"
	ReturnShipped                 = "42.return_shipped"
	ReturnShipmentDeliveryPending = "51.return_shipment_delivery_pending"
	ReturnShipmentDeliveryDelayed = "52.return_shipment_delivery_delayed"
	ReturnShipmentDelivered       = "50.return_shipment_delivered"
	ReturnShipmentDeliveryProblem = "53.return_shipment_delivery_problem"
	ReturnShipmentCanceled        = "54.return_shipment_canceled"
	ReturnShipmentSuccess         = "55.return_shipment_success"
	ShipmentRejectedBySeller      = "21.shipment_rejected_by_seller"
	PayToBuyer                    = "80.pay_to_buyer"
	PayToSeller                   = "90.pay_to_seller"
	PayToSellerFailed             = "92.pay_to_seller_failed"
	PayToSellerSuccess            = "91.pay_to_seller_success"
	PayToBuyerFailed              = "82.pay_to_buyer_failed"
	PayToBuyerSuccess             = "81.pay_to_buyer_success"
)

type StateMachine struct {
	states []State
}

type State struct {
	title      string
	fromStates []string
	toStates   []string
}

func (sm *StateMachine) add(s State) {
	sm.states = append(sm.states, s)
}

func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

func CheckNextState(currentStep, nextStep string) bool {
	SM := generateSM()
	for _, state := range SM.states {
		for _, to := range state.toStates {
			if state.title == currentStep && to == nextStep {
				return true
			}
		}
	}
	return false
}

func generateSM() *StateMachine {
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
