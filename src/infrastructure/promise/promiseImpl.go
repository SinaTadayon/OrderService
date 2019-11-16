package promise

type iPromiseImpl struct {
	channel  DataChan
	count    int
	capacity int
}

func NewPromise(channel DataChan, count int, capacity int) IPromise {
	return &iPromiseImpl{channel: channel, count: count, capacity: capacity}
}

func (promise iPromiseImpl) Data() *FutureData {
	futureData, ok := <-promise.channel
	if ok != true {
		return nil
	}
	return &futureData
}

func (promise iPromiseImpl) Channel() DataChan {
	return promise.channel
}

func (promise iPromiseImpl) Count() int {
	return promise.count
}

func (promise iPromiseImpl) Capacity() int {
	return promise.capacity
}
