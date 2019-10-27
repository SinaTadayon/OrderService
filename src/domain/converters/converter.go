package converters

type IConverter interface {
	Map(in interface{}, out interface{}) (interface{}, error)
}
