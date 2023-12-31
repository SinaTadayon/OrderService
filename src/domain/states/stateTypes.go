package states

type StateType int

type stateEnum struct {
	name  string
	index int
}

var stateTypeMap = map[int]stateEnum{
	0: {"New_Order", 1},
	1: {"Payment_Pending", 10},
	2: {"Payment_Success", 11},
	3: {"Payment_Failed", 12},

	4: {"Order_Verification_Pending", 13},
	5: {"Order_Verification_Success", 14},
	6: {"Order_Verification_Fail", 15},

	7: {"Approval_Pending", 20},
	8: {"Canceled_By_Seller", 21},
	9: {"Canceled_By_Buyer", 22},

	10: {"Shipment_Pending", 30},
	11: {"Shipped", 31},
	12: {"Delivered", 32},
	13: {"Shipment_Delayed", 33},
	14: {"Delivery_Pending", 34},
	15: {"Delivery_Delayed", 35},
	16: {"Delivery_Failed", 36},

	17: {"Return_Request_Pending", 40},
	18: {"Return_Request_Rejected", 41},
	19: {"Return_Canceled", 42},

	20: {"Return_Shipment_Pending", 50},
	21: {"Return_Shipped", 51},
	22: {"Return_Delivered", 52},
	23: {"Return_Delivery_Pending", 53},
	24: {"Return_Delivery_Delayed", 54},
	25: {"Return_Rejected", 55},
	26: {"Return_Delivery_Failed", 56},

	27: {"Pay_To_Buyer", 80},
	28: {"Pay_To_Seller", 90},
}

const (
	NewOrder StateType = iota
	PaymentPending
	PaymentSuccess
	PaymentFailed

	OrderVerificationPending
	OrderVerificationSuccess
	OrderVerificationFailed

	ApprovalPending
	CanceledBySeller
	CanceledByBuyer

	ShipmentPending
	Shipped
	Delivered
	ShipmentDelayed
	DeliveryPending
	DeliveryDelayed
	DeliveryFailed

	ReturnRequestPending
	ReturnRequestRejected
	ReturnCanceled

	ReturnShipmentPending
	ReturnShipped
	ReturnDelivered
	ReturnDeliveryPending
	ReturnDeliveryDelayed
	ReturnRejected
	ReturnDeliveryFailed

	PayToBuyer
	PayToSeller
)

func (stateType StateType) StateName() string {
	return stateType.String()
}

func (stateType StateType) StateIndex() int {
	return stateTypeMap[stateType.Ordinal()].index
}

func (stateType StateType) Ordinal() int {
	if stateType < NewOrder || stateType > PayToSeller {
		return -1
	}
	return int(stateType)
}

func (stateType StateType) Values() []string {
	keys := make([]string, 0, len(stateTypeMap))
	for _, value := range stateTypeMap {
		keys = append(keys, value.name)
	}

	return keys
}

func (stateType StateType) String() string {
	if stateType < NewOrder || stateType > PayToSeller {
		return ""
	}

	return stateTypeMap[stateType.Ordinal()].name
}

func FromIndex(index int32) IEnumState {
	switch index {
	case 1:
		return NewOrder
	case 10:
		return PaymentPending
	case 11:
		return PaymentSuccess
	case 12:
		return PaymentFailed
	case 13:
		return OrderVerificationPending
	case 14:
		return OrderVerificationSuccess
	case 15:
		return OrderVerificationFailed
	case 20:
		return ApprovalPending
	case 21:
		return CanceledBySeller
	case 22:
		return CanceledByBuyer
	case 30:
		return ShipmentPending
	case 31:
		return Shipped
	case 32:
		return Delivered
	case 33:
		return ShipmentDelayed
	case 34:
		return DeliveryPending
	case 35:
		return DeliveryDelayed
	case 36:
		return DeliveryFailed
	case 40:
		return ReturnRequestPending
	case 41:
		return ReturnRequestRejected
	case 42:
		return ReturnCanceled
	case 50:
		return ReturnShipmentPending
	case 51:
		return ReturnShipped
	case 52:
		return ReturnDelivered
	case 53:
		return ReturnDeliveryPending
	case 54:
		return ReturnDeliveryDelayed
	case 55:
		return ReturnRejected
	case 56:
		return ReturnDeliveryFailed
	case 80:
		return PayToBuyer
	case 90:
		return PayToSeller

	default:
		return nil
	}
}

func FromString(stateType string) IEnumState {
	switch stateType {
	case "New_Order":
		return NewOrder
	case "Payment_Pending":
		return PaymentPending
	case "Payment_Success":
		return PaymentSuccess
	case "Payment_Failed":
		return PaymentFailed
	case "Order_Verification_Pending":
		return OrderVerificationPending
	case "Order_Verification_Success":
		return OrderVerificationSuccess
	case "Order_Verification_Failed":
		return OrderVerificationFailed
	case "Approval_Pending":
		return ApprovalPending
	case "Canceled_By_Seller":
		return CanceledBySeller
	case "Canceled_By_Buyer":
		return CanceledByBuyer
	case "Shipment_Pending":
		return ShipmentPending
	case "Shipped":
		return Shipped
	case "Delivered":
		return Delivered
	case "Shipment_Delayed":
		return ShipmentDelayed
	case "Delivery_Pending":
		return DeliveryPending
	case "Delivery_Delayed":
		return DeliveryDelayed
	case "Delivery_Failed":
		return DeliveryFailed
	case "Return_Request_Pending":
		return ReturnRequestPending
	case "Return_Request_Rejected":
		return ReturnRequestRejected
	case "Return_Canceled":
		return ReturnCanceled
	case "Return_Shipment_Pending":
		return ReturnShipmentPending
	case "Return_Shipped":
		return ReturnShipped
	case "Return_Delivered":
		return ReturnDelivered
	case "Return_Delivery_Pending":
		return ReturnDeliveryPending
	case "Return_Delivery_Delayed":
		return ReturnDeliveryDelayed
	case "Return_Rejected":
		return ReturnRejected
	case "Return_Delivery_Failed":
		return ReturnDeliveryFailed
	case "Pay_To_Buyer":
		return PayToBuyer
	case "Pay_To_Seller":
		return PayToSeller

	default:
		return nil
	}
}
