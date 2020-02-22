package converter

import "context"

type IConverter interface {
	Map(ctx context.Context, in interface{}, out interface{}) (interface{}, error)
}
