package actions

type IEnumAction interface {
	Name() string
	Ordinal() int
	Values() []string
}