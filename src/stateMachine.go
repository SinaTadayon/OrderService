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
