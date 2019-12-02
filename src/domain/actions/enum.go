package actions

type IEnumAction interface {
	ActionName() string
	ActionOrdinal() int
	Values() []string
}
