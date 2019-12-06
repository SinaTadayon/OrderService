package future

import (
	"github.com/pkg/errors"
	"time"
)

type Builder struct {
	iFuture     *iFutureImpl
	dataFuture  *iDataFutureImpl
	errorFuture *iErrorFutureImpl
}

func Factory() Builder {
	return Builder{
		iFuture: &iFutureImpl{},
	}
}

func FactoryOf(future IFuture) Builder {
	return Builder{
		iFuture: future.(*iFutureImpl),
	}
}

func (builder Builder) SetCapacity(capacity int) Builder {
	builder.iFuture.capacity = capacity
	return builder
}

func (builder Builder) SetCount(count int) Builder {
	builder.iFuture.count = count
	return builder
}

func (builder Builder) SetData(data interface{}) Builder {

	if builder.dataFuture == nil {
		builder.dataFuture = &iDataFutureImpl{}
	}
	builder.dataFuture.data = data
	return builder
}

func (builder Builder) SetError(code ErrorCode, message string, reason error) Builder {
	builder.errorFuture = &iErrorFutureImpl{}
	builder.errorFuture.code = code
	builder.errorFuture.message = message
	builder.errorFuture.reason = reason

	if builder.dataFuture == nil {
		builder.dataFuture = &iDataFutureImpl{}
	}

	builder.dataFuture.futureError = builder.errorFuture
	return builder
}

func (builder Builder) SetErrorOf(errorFuture IErrorFuture) Builder {
	builder.errorFuture = &iErrorFutureImpl{}
	builder.errorFuture.code = errorFuture.Code()
	builder.errorFuture.message = errorFuture.Message()
	builder.errorFuture.reason = errorFuture.Reason()

	if builder.dataFuture == nil {
		builder.dataFuture = &iDataFutureImpl{}
	}

	builder.dataFuture.futureError = builder.errorFuture
	return builder
}

func (builder Builder) Send() {
	if builder.iFuture.channel == nil {
		builder.iFuture.channel = make(chan IDataFuture, builder.iFuture.capacity)
	}
	defer close(builder.iFuture.channel)
	builder.iFuture.channel <- builder.dataFuture
}

func (builder Builder) SendTimeout(duration time.Duration) error {
	if builder.iFuture.channel == nil {
		builder.iFuture.channel = make(chan IDataFuture, builder.iFuture.capacity)
	}
	defer close(builder.iFuture.channel)
	select {
	case builder.iFuture.channel <- builder.dataFuture:
		return nil
	case <-time.After(duration):
		return errors.New("Send Timeout")
	}
}

func (builder Builder) BuildAndSend() IFuture {
	if builder.iFuture.channel == nil {
		builder.iFuture.channel = make(chan IDataFuture, builder.iFuture.capacity)
	}
	defer close(builder.iFuture.channel)
	builder.iFuture.channel <- builder.dataFuture
	return builder.iFuture
}

func (builder Builder) BuildAndSendTimeout(duration time.Duration) (IFuture, error) {
	if builder.iFuture.channel == nil {
		builder.iFuture.channel = make(chan IDataFuture, builder.iFuture.capacity)
	}
	defer close(builder.iFuture.channel)
	select {
	case builder.iFuture.channel <- builder.dataFuture:
		return builder.iFuture, nil
	case <-time.After(duration):
		return nil, errors.New("Send Timeout")
	}
}

func (builder Builder) Build() IFuture {
	if builder.iFuture.channel == nil {
		builder.iFuture.channel = make(chan IDataFuture, builder.iFuture.capacity)
	}
	return builder.iFuture
}
