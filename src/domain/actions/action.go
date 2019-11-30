package actions

type IAction interface {
	ActionType() ActionType
	ActionEnum() IEnumAction
}
