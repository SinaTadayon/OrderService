package order_query_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	worker_pool "gitlab.faza.io/order-project/order-service/infrastructure/workerPool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	defaultCommandStreamBufferSize = 81892
)

type streamPipeline struct {
	commandDataStream repository.CommandReaderStream
	orderQR           iOrderQRImpl
}

type pipelineInStream <-chan *streamPipeline

type iOrderQRImpl struct {
	orderRepo       order_repository.IOrderRepository
	cmdStream       repository.CommandReaderStream
	mongoAdapter    *mongoadapter.Mongo
	database        string
	collection      string
	connectionCount int
	workerPool      worker_pool.IWorkerPool
}

func OrderQRFactory(ctx context.Context, query *mongoadapter.Mongo,
	database, collection string, commandStream repository.CommandReaderStream,
	connectionCount int, workerPool worker_pool.IWorkerPool) (IOrderQR, error) {

	orderRepository := order_repository.NewOrderRepository(query, database, collection)

	orderQR := &iOrderQRImpl{
		orderRepo:       orderRepository,
		cmdStream:       commandStream,
		mongoAdapter:    query,
		database:        database,
		collection:      collection,
		connectionCount: connectionCount,
		workerPool:      workerPool,
	}

	pipeline, err := fanOutPipelines(ctx, commandStream, *orderQR)
	if err != nil {
		return nil, err
	}

	err = fanInPipelines(ctx, pipeline, *orderQR)
	if err != nil {
		return nil, err
	}

	return orderQR, nil
}

func fanOutPipelines(ctx context.Context, commandInStream repository.CommandReaderStream, orderQR iOrderQRImpl) (pipelineInStream, error) {
	commandWriterChannels := make([]repository.CommandWriterStream, 0, orderQR.connectionCount)
	pipelineStream := make(chan *streamPipeline)

	fanOutTask := func() {
		defer func() {
			for _, stream := range commandWriterChannels {
				close(stream)
			}

			close(pipelineStream)
		}()

		index := 0
		initIndex := 0
		for commandData := range commandInStream {
			//select {
			//case <-ctx.Done():
			//	return
			//default:
			//}

			if initIndex < cap(commandWriterChannels) {
				commandChannel := make(chan *repository.CommandData, defaultCommandStreamBufferSize)
				commandWriterChannels = append(commandWriterChannels, commandChannel)
				pipelineStream <- &streamPipeline{commandDataStream: commandChannel, orderQR: orderQR}
				initIndex++
			}

			if index >= len(commandWriterChannels) {
				index = 0
			}

			commandWriterChannels[index] <- commandData
			index++
		}
	}

	if err := orderQR.workerPool.SubmitTask(fanOutTask); err != nil {
		applog.GLog.Logger.Error("WorkerPool.SubmitTask failed",
			"fn", "fanOutPipelines",
			"error", err)

		return nil, err
	}

	return pipelineStream, nil

}

func fanInPipelines(ctx context.Context, pipelineStream pipelineInStream, orderQR iOrderQRImpl) error {

	fanInTask := func() {
		for pipeline := range pipelineStream {
			//select {
			//case <-ctx.Done():
			//	break
			//default:
			//}

			pipelineTask := pipeline.commandStreamHandler(ctx)

			if err := orderQR.workerPool.SubmitTask(pipelineTask); err != nil {
				applog.GLog.Logger.Error("submit pipelineTask to WorkerPool.SubmitTask failed",
					"fn", "fanInPipelines", "error", err)

				//applog.GLog.Logger.Warn("pipeline task launch without worker pool",
				//	"fn", "fanInPipelines")
				//go pipelineTask()
			}
		}
	}

	if err := orderQR.workerPool.SubmitTask(fanInTask); err != nil {
		applog.GLog.Logger.Error("WorkerPool.SubmitTask failed",
			"fn", "fanInPipelines",
			"error", err)

		return err
	}

	return nil
}

func (pipeline streamPipeline) commandStreamHandler(ctx context.Context) worker_pool.Task {

	return func() {
		for commandData := range pipeline.commandDataStream {
			switch commandData.Command {
			case repository.SaveCmd:
				err := pipeline.orderQR.Save(ctx, commandData.Data.(*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.Save of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.SaveAllCmd:
				err := pipeline.orderQR.SaveAll(ctx, commandData.Data.([]*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.SaveAll of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.UpdateCmd:
				err := pipeline.orderQR.Update(ctx, commandData.Data.(*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.Update of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.UpdateAllCmd:
				err := pipeline.orderQR.UpdateAll(ctx, commandData.Data.([]*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.UpdateAll of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.InsertCmd:
				err := pipeline.orderQR.Insert(ctx, commandData.Data.(*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.Insert of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.InsertAllCmd:
				err := pipeline.orderQR.InsertAll(ctx, commandData.Data.([]*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.InsertAll of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.DeleteCmd:
				err := pipeline.orderQR.Delete(ctx, commandData.Data.(*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.Delete of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.DeletePartialCmd:
				err := pipeline.orderQR.DeleteAllWithOrders(ctx, commandData.Data.([]*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.DeleteAllWithOrders of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.DeleteAllCmd:
				err := pipeline.orderQR.DeleteAll(ctx, commandData.Data.(time.Time))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.DeleteAll of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.RemoveCmd:
				err := pipeline.orderQR.RemoveById(ctx, commandData.Data.(uint64))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.RemoveById of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.RemoveEntityCmd:
				err := pipeline.orderQR.Remove(ctx, commandData.Data.(*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.Remove of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.RemovePartialCmd:
				err := pipeline.orderQR.RemoveAllWithOrders(ctx, commandData.Data.([]*entities.Order))
				if err != nil {
					applog.GLog.Logger.Error("orderQR.RemoveAllWithOrders of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.RemoveAllCmd:
				err := pipeline.orderQR.RemoveAll(ctx)
				if err != nil {
					applog.GLog.Logger.Error("orderQR.RemoveAll of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			default:
				applog.GLog.Logger.Error("received command not supported",
					"fn", "commandStreamHandler",
					"commandData", commandData)
			}
		}
	}
}

func (orderQR iOrderQRImpl) FindAll(ctx context.Context) ([]*entities.Order, repository.IRepoError) {
	return orderQR.orderRepo.FindAll(ctx)
}

func (orderQR iOrderQRImpl) FindAllWithSort(ctx context.Context, fieldName string, direction int) ([]*entities.Order, repository.IRepoError) {
	return orderQR.orderRepo.FindAllWithSort(ctx, fieldName, direction)
}

func (orderQR iOrderQRImpl) FindAllWithPage(ctx context.Context, page, perPage int64) ([]*entities.Order, int64, repository.IRepoError) {
	return orderQR.orderRepo.FindAllWithPage(ctx, page, perPage)
}

func (orderQR iOrderQRImpl) FindAllWithPageAndSort(ctx context.Context, page, perPage int64, fieldName string, direction int) ([]*entities.Order, int64, repository.IRepoError) {
	return orderQR.orderRepo.FindAllWithPageAndSort(ctx, page, perPage, fieldName, direction)
}

func (orderQR iOrderQRImpl) FindAllById(ctx context.Context, ids ...uint64) ([]*entities.Order, repository.IRepoError) {
	return orderQR.orderRepo.FindAllById(ctx, ids...)
}

func (orderQR iOrderQRImpl) FindById(ctx context.Context, orderId uint64) (*entities.Order, repository.IRepoError) {
	return orderQR.orderRepo.FindById(ctx, orderId)
}

func (orderQR iOrderQRImpl) FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.Order, repository.IRepoError) {
	return orderQR.orderRepo.FindByFilter(ctx, supplier)
}

func (orderQR iOrderQRImpl) FindByFilterWithSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int)) ([]*entities.Order, repository.IRepoError) {
	return orderQR.orderRepo.FindByFilterWithSort(ctx, supplier)
}

func (orderQR iOrderQRImpl) FindByFilterWithPage(ctx context.Context, supplier func() (filter interface{}), page, perPage int64) ([]*entities.Order, int64, repository.IRepoError) {
	return orderQR.orderRepo.FindByFilterWithPage(ctx, supplier, page, perPage)
}

func (orderQR iOrderQRImpl) FindByFilterWithPageAndSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int), page, perPage int64) ([]*entities.Order, int64, repository.IRepoError) {
	return orderQR.orderRepo.FindByFilterWithPageAndSort(ctx, supplier, page, perPage)
}

func (orderQR iOrderQRImpl) ExistsById(ctx context.Context, orderId uint64) (bool, repository.IRepoError) {
	return orderQR.orderRepo.ExistsById(ctx, orderId)
}

func (orderQR iOrderQRImpl) Count(ctx context.Context) (int64, repository.IRepoError) {
	return orderQR.orderRepo.Count(ctx)
}

func (orderQR iOrderQRImpl) CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError) {
	return orderQR.orderRepo.CountWithFilter(ctx, supplier)
}

func (orderQR iOrderQRImpl) Save(ctx context.Context, order *entities.Order) repository.IRepoError {
	return orderQR.Insert(ctx, order)
}

func (orderQR iOrderQRImpl) SaveAll(ctx context.Context, orders []*entities.Order) repository.IRepoError {
	panic("implementation required")
}

func (orderQR iOrderQRImpl) Update(ctx context.Context, order *entities.Order) repository.IRepoError {

	if order.OrderId == 0 {
		return repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("OrderId is zero"))
	}
	updateOptions := &options.UpdateOptions{}
	updateOptions.SetUpsert(true)
	updateResult, e := orderQR.mongoAdapter.UpdateOne(orderQR.database, orderQR.collection,
		bson.D{{"orderId", order.OrderId}, {"version", bson.D{{"$lt", order.Version}}}},
		bson.D{{"$set", order}}, updateOptions)
	if e != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "UpdateOne Failed"))
	}

	if updateResult.MatchedCount != 1 || updateResult.ModifiedCount != 1 {
		return repository.ErrorFactory(repository.NotFoundErr, "Order Not Found", errors.New("Order Not Found"))
	}

	return nil
}

func (orderQR iOrderQRImpl) UpdateAll(ctx context.Context, orders []*entities.Order) repository.IRepoError {
	panic("implementation required")
}

func (orderQR iOrderQRImpl) Insert(ctx context.Context, order *entities.Order) repository.IRepoError {

	_, err := orderQR.mongoAdapter.InsertOne(orderQR.database, orderQR.collection, order)
	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Insert Order Failed"))
	}

	return nil
}

func (orderQR iOrderQRImpl) InsertAll(ctx context.Context, orders []*entities.Order) repository.IRepoError {
	panic("implementation required")
}

func (orderQR iOrderQRImpl) Delete(ctx context.Context, order *entities.Order) repository.IRepoError {
	return orderQR.Update(ctx, order)
}

func (orderQR iOrderQRImpl) DeleteAllWithOrders(ctx context.Context, orders []*entities.Order) repository.IRepoError {
	panic("implementation required")
}

func (orderQR iOrderQRImpl) DeleteAll(ctx context.Context, timestamp time.Time) repository.IRepoError {
	_, err := orderQR.mongoAdapter.UpdateMany(orderQR.database, orderQR.collection,
		bson.D{{"deletedAt", nil}},
		bson.M{"$set": bson.M{"deletedAt": timestamp}})
	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "UpdateMany Order Failed"))
	}
	return nil
}

func (orderQR iOrderQRImpl) RemoveById(ctx context.Context, orderId uint64) repository.IRepoError {
	result, err := orderQR.mongoAdapter.DeleteOne(orderQR.database, orderQR.collection, bson.M{"orderId": orderId})
	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "RemoveById Order Failed"))
	}

	if result.DeletedCount != 1 {
		return repository.ErrorFactory(repository.NotFoundErr, "Order Not Found", repository.ErrorRemoveFailed)
	}
	return nil
}

func (orderQR iOrderQRImpl) Remove(ctx context.Context, order *entities.Order) repository.IRepoError {
	return orderQR.RemoveById(ctx, order.OrderId)
}

func (orderQR iOrderQRImpl) RemoveAllWithOrders(ctx context.Context, orders []*entities.Order) repository.IRepoError {
	panic("implementation required")
}

func (orderQR iOrderQRImpl) RemoveAll(ctx context.Context) repository.IRepoError {
	_, err := orderQR.mongoAdapter.DeleteMany(orderQR.database, orderQR.collection, bson.M{})
	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "DeleteMany Order Failed"))
	}
	return nil
}
