package states

type StateType int

type stateEnum struct {
	name  string
	index int
}

var stateTypeMap = map[int]stateEnum{
	0: {"NewOrder", 1},
	1: {"PaymentPending", 10},
	2: {"PaymentSuccess", 11},
	3: {"PaymentFailed", 12},

	4: {"OrderVerificationPending", 13},
	5: {"OrderVerificationSuccess", 14},
	6: {"OrderVerificationFail", 15},

	7: {"ApprovalPending", 20},
	8: {"CanceledBySeller", 21},
	9: {"CanceledByBuyer", 22},

	10: {"ShipmentPending", 30},
	11: {"Shipped", 31},
	12: {"Delivered", 32},
	13: {"ShipmentDelayed", 33},
	14: {"DeliveryPending", 34},
	15: {"DeliveryDelayed", 35},
	16: {"DeliveryFailed", 36},

	17: {"ReturnRequestPending", 40},
	18: {"ReturnRequestRejected", 41},

	19: {"ReturnShipmentPending", 50},
	20: {"ReturnShipped", 51},
	21: {"ReturnDelivered", 52},
	22: {"ReturnDeliveryPending", 53},
	23: {"ReturnDeliveryDelayed", 54},
	24: {"ReturnRejected", 55},
	25: {"ReturnDeliveryFailed", 56},

	26: {"PayToBuyer", 80},
	27: {"PayToSeller", 90},
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
	case "NewOrder":
		return NewOrder
	case "PaymentPending":
		return PaymentPending
	case "PaymentSuccess":
		return PaymentSuccess
	case "PaymentFailed":
		return PaymentFailed
	case "OrderVerificationPending":
		return OrderVerificationPending
	case "OrderVerificationSuccess":
		return OrderVerificationSuccess
	case "OrderVerificationFailed":
		return OrderVerificationFailed
	case "ApprovalPending":
		return ApprovalPending
	case "CanceledBySeller":
		return CanceledBySeller
	case "CanceledByBuyer":
		return CanceledByBuyer
	case "ShipmentPending":
		return ShipmentPending
	case "Shipped":
		return Shipped
	case "Delivered":
		return Delivered
	case "ShipmentDelayed":
		return ShipmentDelayed
	case "DeliveryPending":
		return DeliveryPending
	case "DeliveryDelayed":
		return DeliveryDelayed
	case "DeliveryFailed":
		return DeliveryFailed
	case "ReturnRequestPending":
		return ReturnRequestPending
	case "ReturnRequestRejected":
		return ReturnRequestRejected
	case "ReturnShipmentPending":
		return ReturnShipmentPending
	case "ReturnShipped":
		return ReturnShipped
	case "ReturnDelivered":
		return ReturnDelivered
	case "ReturnDeliveryPending":
		return ReturnDeliveryPending
	case "ReturnDeliveryDelayed":
		return ReturnDeliveryDelayed
	case "ReturnRejected":
		return ReturnRejected
	case "ReturnDeliveryFailed":
		return ReturnDeliveryFailed
	case "PayToBuyer":
		return PayToBuyer
	case "PayToSeller":
		return PayToSeller

	default:
		return nil
	}
}
