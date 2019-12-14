package states

type IEnumState interface {
	StateName() string
	StateIndex() int
	Ordinal() int
	Values() []string
}
