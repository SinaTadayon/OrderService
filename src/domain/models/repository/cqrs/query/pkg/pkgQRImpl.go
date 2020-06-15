package pkg_query_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	worker_pool "gitlab.faza.io/order-project/order-service/infrastructure/workerPool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultCommandStreamBufferSize = 81892
)

type streamPipeline struct {
	commandDataStream repository.CommandReaderStream
	pkgQR             iPkgQRImpl
}

type pipelineInStream <-chan *streamPipeline

type iPkgQRImpl struct {
	pkgRepo         pkg_repository.IPkgItemRepository
	cmdStream       repository.CommandReaderStream
	mongoAdapter    *mongoadapter.Mongo
	database        string
	collection      string
	connectionCount int
	workerPool      worker_pool.IWorkerPool
}

func PkgQRFactory(ctx context.Context, query *mongoadapter.Mongo, database, collection string,
	commandStream repository.CommandReaderStream, connectionCount int, workerPool worker_pool.IWorkerPool) (IPkgQR, error) {
	pkgRepository := pkg_repository.NewPkgItemRepository(query, database, collection)

	pkgQR := &iPkgQRImpl{
		pkgRepo:         pkgRepository,
		cmdStream:       commandStream,
		mongoAdapter:    query,
		database:        database,
		collection:      collection,
		connectionCount: connectionCount,
		workerPool:      workerPool,
	}

	pipeline, err := fanOutPipelines(ctx, commandStream, *pkgQR)
	if err != nil {
		return nil, err
	}

	err = fanInPipelines(ctx, pipeline, *pkgQR)
	if err != nil {
		return nil, err
	}

	return pkgQR, nil
}

func fanOutPipelines(ctx context.Context, commandInStream repository.CommandReaderStream, pkgQR iPkgQRImpl) (pipelineInStream, error) {
	commandWriterChannels := make([]repository.CommandWriterStream, 0, pkgQR.connectionCount)
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
				pipelineStream <- &streamPipeline{commandDataStream: commandChannel, pkgQR: pkgQR}
				initIndex++
			}

			if index >= len(commandWriterChannels) {
				index = 0
			}

			commandWriterChannels[index] <- commandData
			index++
		}
	}

	if err := pkgQR.workerPool.SubmitTask(fanOutTask); err != nil {
		applog.GLog.Logger.Error("WorkerPool.SubmitTask failed",
			"fn", "fanOutPipelines",
			"error", err)

		return nil, err
	}

	return pipelineStream, nil

}

func fanInPipelines(ctx context.Context, pipelineStream pipelineInStream, pkgQR iPkgQRImpl) error {

	fanInTask := func() {
		for pipeline := range pipelineStream {
			//select {
			//case <-ctx.Done():
			//	break
			//default:
			//}

			pipelineTask := pipeline.commandStreamHandler(ctx)

			//// TODO design error handling
			if err := pkgQR.workerPool.SubmitTask(pipelineTask); err != nil {
				applog.GLog.Logger.Error("submit pipelineTask to WorkerPool.SubmitTask failed",
					"fn", "fanInPipelines", "error", err)

				//applog.GLog.Logger.Warn("pipeline task launch without worker pool",
				//	"fn", "fanInPipelines")
				//go pipelineTask()
			}
		}
	}

	if err := pkgQR.workerPool.SubmitTask(fanInTask); err != nil {
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
			case repository.UpdateCmd:
				err := pipeline.pkgQR.Update(ctx, commandData.Data.(*entities.PackageItem))
				if err != nil {
					applog.GLog.Logger.Error("pkgQR.Update of mongoAdapter repository failed",
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

func (pkgQR iPkgQRImpl) FindById(ctx context.Context, oid uint64, pid uint64) (*entities.PackageItem, repository.IRepoError) {
	return pkgQR.pkgRepo.FindById(ctx, oid, pid)
}

func (pkgQR iPkgQRImpl) FindPkgItmBuyinfById(ctx context.Context, oid uint64, pid uint64) (*entities.PackageItem, uint64, repository.IRepoError) {
	return pkgQR.pkgRepo.FindPkgItmBuyinfById(ctx, oid, pid)
}

func (pkgQR iPkgQRImpl) FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.PackageItem, repository.IRepoError) {
	return pkgQR.pkgRepo.FindByFilter(ctx, supplier)
}

func (pkgQR iPkgQRImpl) ExistsById(ctx context.Context, oid uint64, pid uint64) (bool, repository.IRepoError) {
	return pkgQR.pkgRepo.ExistsById(ctx, oid, pid)
}

func (pkgQR iPkgQRImpl) Count(ctx context.Context, pid uint64) (int64, repository.IRepoError) {
	return pkgQR.pkgRepo.Count(ctx, pid)
}

func (pkgQR iPkgQRImpl) CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError) {
	return pkgQR.pkgRepo.CountWithFilter(ctx, supplier)
}

// TODO test it
func (pkgQR iPkgQRImpl) findAndUpdate(ctx context.Context, pkgItem *entities.PackageItem, upsert bool) repository.IRepoError {

	opt := options.FindOneAndUpdate()
	opt.SetUpsert(upsert)
	singleResult := pkgQR.mongoAdapter.GetConn().Database(pkgQR.database).Collection(pkgQR.collection).FindOneAndUpdate(ctx,
		bson.D{
			{"orderId", pkgItem.OrderId},
			{"packages", bson.D{
				{"$elemMatch", bson.D{
					{"pid", pkgItem.PId},
					{"version", bson.D{{"$lt", pkgItem.Version}}},
				}},
			}},
		},
		bson.D{{"$set", bson.D{{"packages.$", pkgItem}}}}, opt)
	if singleResult.Err() != nil {
		if pkgQR.mongoAdapter.NoDocument(singleResult.Err()) {
			return repository.ErrorFactory(repository.NotFoundErr, "Package Not Found", repository.ErrorUpdateFailed)
		}
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), ""))
	}

	return nil
}

func (pkgQR iPkgQRImpl) Update(ctx context.Context, pkgItem *entities.PackageItem) repository.IRepoError {
	err := pkgQR.findAndUpdate(ctx, pkgItem, true)
	if err != nil {
		return err
	}

	return nil
}
