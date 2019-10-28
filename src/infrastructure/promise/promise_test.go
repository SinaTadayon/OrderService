package promise

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncPromiseChannel(t *testing.T) {
	promiseTest := createPromise()
	futureData , ok := <- promiseTest.GetData()
	assert.True(t, ok, "channel is closed")
	assert.Equal(t, futureData.Data, "salam")
}

func TestAsyncPromiseChannel(t *testing.T) {
	promiseCall := func() (ipromise IPromise) {
		waitChannel := make(chan struct{})
		go func() {
			ipromise = createPromise()
			close(waitChannel)
		}()
		<- waitChannel
		return ipromise
	}
	promiseTest := promiseCall()
	futureData , ok := <- promiseTest.GetData()
	assert.True(t, ok, "channel is closed")
	assert.Equal(t, futureData.Data, "salam")
}

func createPromise() IPromise {
	returnChannel := make(chan FutureData, 1)
	returnChannel <- FutureData{Data:"salam", Error:FutureError{Code:int32(500), Reason:"Unknown Error"}}
	defer close(returnChannel)
	return NewPromise(returnChannel, 1, 1)
}