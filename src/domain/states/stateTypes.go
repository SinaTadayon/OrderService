package states

import "github.com/pkg/errors"

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
	19: {"ReturnCanceled", 42},

	20: {"ReturnShipmentPending", 50},
	21: {"ReturnShipped", 51},
	22: {"ReturnDelivered", 52},
	23: {"ReturnDeliveryPending", 53},
	24: {"ReturnDeliveryDelayed", 54},
	25: {"ReturnRejected", 55},
	26: {"ReturnDeliveryFailed", 56},

	27: {"PayToBuyer", 80},
	28: {"PayToSeller", 90},
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

func FromString(stateType string) (StateType, error) {
	switch stateType {
	case "NewOrder":
		return NewOrder, nil
	case "PaymentPending":
		return PaymentPending, nil
	case "PaymentSuccess":
		return PaymentSuccess, nil
	case "PaymentFailed":
		return PaymentFailed, nil
	case "OrderVerificationPending":
		return OrderVerificationPending, nil
	case "OrderVerificationSuccess":
		return OrderVerificationSuccess, nil
	case "OrderVerificationFailed":
		return OrderVerificationFailed, nil
	case "ApprovalPending":
		return ApprovalPending, nil
	case "CanceledBySeller":
		return CanceledBySeller, nil
	case "CanceledByBuyer":
		return CanceledByBuyer, nil
	case "ShipmentPending":
		return ShipmentPending, nil
	case "Shipped":
		return Shipped, nil
	case "Delivered":
		return Delivered, nil
	case "ShipmentDelayed":
		return ShipmentDelayed, nil
	case "DeliveryPending":
		return DeliveryPending, nil
	case "DeliveryDelayed":
		return DeliveryDelayed, nil
	case "DeliveryFailed":
		return DeliveryFailed, nil
	case "ReturnRequestPending":
		return ReturnRequestPending, nil
	case "ReturnRequestRejected":
		return ReturnRequestRejected, nil
	case "ReturnCanceled":
		return ReturnCanceled, nil
	case "ReturnShipmentPending":
		return ReturnShipmentPending, nil
	case "ReturnShipped":
		return ReturnShipped, nil
	case "ReturnDelivered":
		return ReturnDelivered, nil
	case "ReturnDeliveryPending":
		return ReturnDeliveryPending, nil
	case "ReturnDeliveryDelayed":
		return ReturnDeliveryDelayed, nil
	case "ReturnRejected":
		return ReturnRejected, nil
	case "ReturnDeliveryFailed":
		return ReturnDeliveryFailed, nil
	case "PayToBuyer":
		return PayToBuyer, nil
	case "PayToSeller":
		return PayToSeller, nil

	default:
		return -1, errors.New("invalid stateType string")
	}
}
