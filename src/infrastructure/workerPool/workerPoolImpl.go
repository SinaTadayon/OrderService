package worker_pool

import (
	"gitlab.faza.io/order-project/order-service/infrastructure/workerPool/internal"
	"time"
)

const (
	defaultWorkerPoolSize int = 16384
)

type iWorkerPoolImpl struct {
	workerPool *internal.Pool
}

func Factory() (IWorkerPool, error) {
	workerPool, err := internal.NewPool(defaultWorkerPoolSize)
	if err != nil {
		return nil, err
	}

	return &iWorkerPoolImpl{workerPool}, nil
}

func FactoryOf(capacity int, expiration time.Duration) (IWorkerPool, error) {
	workerPool, err := internal.NewPool(capacity, internal.WithExpiryDuration(expiration))
	if err != nil {
		return nil, err
	}

	return &iWorkerPoolImpl{workerPool}, nil
}

func (iWorker iWorkerPoolImpl) SubmitTask(task Task) error {
	return iWorker.workerPool.Submit(task)
}

func (iWorker iWorkerPoolImpl) Running() int {
	return iWorker.workerPool.Running()
}

func (iWorker iWorkerPoolImpl) Capability() int {
	return iWorker.workerPool.Cap()
}

func (iWorker iWorkerPoolImpl) Available() int {
	return iWorker.workerPool.Free()
}

func (iWorker iWorkerPoolImpl) Resize(size int) {
	iWorker.workerPool.Tune(size)
}

func (iWorker iWorkerPoolImpl) Restart() {
	iWorker.workerPool.Reboot()
}

func (iWorker iWorkerPoolImpl) Shutdown() {
	iWorker.workerPool.Release()
}
