package reason

import "gitlab.faza.io/order-project/order-service/domain/models/entities"

type ActionName string
type ReasonType int

const (
	Cancel       ReasonType = 0
	Return       ReasonType = 1
	CancelAction string     = "Cancel"
	ReturnAction string     = "SubmitReturnRequest"
)

//
// This function will extract reason from specified action type
func ExtractActionReason(history entities.Progress, actionName ActionName, reasonType ReasonType, stateName string) (reason *entities.Reason) {

	for _, state := range history.History {

		if stateName != "" && state.Name != stateName {
			continue
		}

		for _, action := range state.Actions {
			if action.Name != string(actionName) {
				continue
			}

			switch reasonType {
			case Cancel:
				reason = cancelReasons(action.Reasons)
			case Return:
				reason = returnReasons(action.Reasons)
			}
		}
	}

	return reason
}

func cancelReasons(reasons []entities.Reason) (reason *entities.Reason) {
	for _, reason := range reasons {
		if reason.Cancel {
			return &reason
		}
	}

	return nil
}

func returnReasons(reasons []entities.Reason) (reason *entities.Reason) {
	for _, reason := range reasons {
		if reason.Return {
			return &reason
		}
	}

	return nil
}
