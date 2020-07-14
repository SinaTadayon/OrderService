package subpkg_query_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	subpkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/subpackage"
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
	subPkgQR          iSubPkgQRImpl
}

type pipelineInStream <-chan *streamPipeline

type iSubPkgQRImpl struct {
	subPkgRepo      subpkg_repository.ISubpackageRepository
	cmdStream       repository.CommandReaderStream
	mongoAdapter    *mongoadapter.Mongo
	database        string
	collection      string
	connectionCount int
	workerPool      worker_pool.IWorkerPool
}

func SubPkgQRFactory(ctx context.Context, query *mongoadapter.Mongo, database, collection string,
	commandStream repository.CommandReaderStream,
	connectionCount int, workerPool worker_pool.IWorkerPool) (ISubPkgQR, error) {
	subPkgRepository := subpkg_repository.NewSubPkgRepository(query, database, collection)

	subpkg := &iSubPkgQRImpl{
		subPkgRepo:      subPkgRepository,
		cmdStream:       commandStream,
		mongoAdapter:    query,
		database:        database,
		collection:      collection,
		connectionCount: connectionCount,
		workerPool:      workerPool,
	}

	pipeline, err := fanOutPipelines(ctx, commandStream, *subpkg)
	if err != nil {
		return nil, err
	}

	err = fanInPipelines(ctx, pipeline, *subpkg)
	if err != nil {
		return nil, err
	}

	return subpkg, nil
}

func fanOutPipelines(ctx context.Context, commandInStream repository.CommandReaderStream, subPkgQR iSubPkgQRImpl) (pipelineInStream, error) {
	commandWriterChannels := make([]repository.CommandWriterStream, 0, subPkgQR.connectionCount)
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
				pipelineStream <- &streamPipeline{commandDataStream: commandChannel, subPkgQR: subPkgQR}
				initIndex++
			}

			if index >= len(commandWriterChannels) {
				index = 0
			}

			commandWriterChannels[index] <- commandData
			index++
		}
	}

	if err := subPkgQR.workerPool.SubmitTask(fanOutTask); err != nil {
		applog.GLog.Logger.Error("WorkerPool.SubmitTask failed",
			"fn", "fanOutPipelines",
			"error", err)

		return nil, err
	}

	return pipelineStream, nil

}

func fanInPipelines(ctx context.Context, pipelineStream pipelineInStream, subPkgQR iSubPkgQRImpl) error {

	fanInTask := func() {
		for pipeline := range pipelineStream {
			//select {
			//case <-ctx.Done():
			//	break
			//default:
			//}

			pipelineTask := pipeline.commandStreamHandler(ctx)

			// TODO design error handling
			if err := subPkgQR.workerPool.SubmitTask(pipelineTask); err != nil {
				applog.GLog.Logger.Error("submit pipelineTask to WorkerPool.SubmitTask failed",
					"fn", "fanInPipelines", "error", err)

				//applog.GLog.Logger.Warn("pipeline task launch without worker pool",
				//	"fn", "fanInPipelines")
				//go pipelineTask()
			}
		}
	}

	if err := subPkgQR.workerPool.SubmitTask(fanInTask); err != nil {
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
				err := pipeline.subPkgQR.Save(ctx, commandData.Data.(*entities.Subpackage))
				if err != nil {
					applog.GLog.Logger.Error("subPkgQR.Save of mongoAdapter repository failed",
						"fn", "commandStreamHandler",
						"commandData", commandData)
				}
				break

			case repository.UpdateCmd:
				err := pipeline.subPkgQR.Update(ctx, commandData.Data.(*entities.Subpackage))
				if err != nil {
					applog.GLog.Logger.Error("subPkgQR.Update of mongoAdapter repository failed",
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

func (subPkgQR iSubPkgQRImpl) FindByOrderAndItemId(ctx context.Context, oid, sid uint64) (*entities.Subpackage, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindByOrderAndItemId(ctx, oid, sid)
}

func (subPkgQR iSubPkgQRImpl) FindByOrderAndSellerId(ctx context.Context, oid, pid uint64) ([]*entities.Subpackage, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindByOrderAndSellerId(ctx, oid, pid)
}

func (subPkgQR iSubPkgQRImpl) FindAll(ctx context.Context, pid uint64) ([]*entities.Subpackage, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindAll(ctx, pid)
}

func (subPkgQR iSubPkgQRImpl) FindAllWithSort(ctx context.Context, pid uint64, fieldName string, direction int) ([]*entities.Subpackage, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindAllWithSort(ctx, pid, fieldName, direction)
}

func (subPkgQR iSubPkgQRImpl) FindAllWithPage(ctx context.Context, pid uint64, page, perPage int64) ([]*entities.Subpackage, int64, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindAllWithPage(ctx, pid, page, perPage)
}

func (subPkgQR iSubPkgQRImpl) FindAllWithPageAndSort(ctx context.Context, pid uint64, page, perPage int64, fieldName string, direction int) ([]*entities.Subpackage, int64, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindAllWithPageAndSort(ctx, pid, page, perPage, fieldName, direction)
}

func (subPkgQR iSubPkgQRImpl) FindByFilter(ctx context.Context, totalSupplier func() (filter interface{}), supplier func() (filter interface{})) ([]*entities.Subpackage, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindByFilter(ctx, totalSupplier, supplier)
}

func (subPkgQR iSubPkgQRImpl) FindByFilterWithPage(ctx context.Context, totalSupplier func() (filter interface{}), supplier func() (filter interface{}), page, perPage int64) ([]*entities.Subpackage, int64, repository.IRepoError) {
	return subPkgQR.subPkgRepo.FindByFilterWithPage(ctx, totalSupplier, supplier, page, perPage)
}

func (subPkgQR iSubPkgQRImpl) ExistsById(ctx context.Context, sid uint64) (bool, repository.IRepoError) {
	return subPkgQR.subPkgRepo.ExistsById(ctx, sid)
}

func (subPkgQR iSubPkgQRImpl) Count(ctx context.Context, pid uint64) (int64, repository.IRepoError) {
	return subPkgQR.subPkgRepo.Count(ctx, pid)
}

func (subPkgQR iSubPkgQRImpl) CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError) {
	return subPkgQR.subPkgRepo.CountWithFilter(ctx, supplier)
}

func (subPkgQR iSubPkgQRImpl) findAndUpdate(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError {

	opt := options.FindOneAndUpdate()
	opt.SetUpsert(false)
	opt.SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"package.pid": subPkg.PId},
			bson.M{"subpackage.sid": subPkg.SId, "subpackage.version": bson.M{"$lt": subPkg.Version}},
		},
	}).SetReturnDocument(options.After)

	singleResult := subPkgQR.mongoAdapter.GetConn().Database(subPkgQR.database).Collection(subPkgQR.collection).FindOneAndUpdate(ctx,
		bson.D{{"orderId", subPkg.OrderId}},
		bson.D{{"$set", bson.D{{"packages.$[package].subpackages.$[subpackage]", subPkg}}}}, opt)

	if singleResult.Err() != nil {
		if subPkgQR.mongoAdapter.NoDocument(singleResult.Err()) {
			return repository.ErrorFactory(repository.NotFoundErr, "SubPackage Not Found", repository.ErrorUpdateFailed)
		}
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), ""))
	}
	return nil
}

func (subPkgQR iSubPkgQRImpl) Save(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError {
	updateResult, err := subPkgQR.mongoAdapter.UpdateOne(subPkgQR.database, subPkgQR.collection, bson.D{
		{"orderId", subPkg.OrderId},
		{"deletedAt", nil},
		{"packages.pid", subPkg.PId}},
		bson.D{{"$push", bson.D{{"packages.$.subpackages", subPkg}}}})

	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "UpdateOne Subpackage Failed"))
	}

	if updateResult.ModifiedCount != 1 {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", repository.ErrorUpdateFailed)
	}

	return nil
}

func (subPkgQR iSubPkgQRImpl) SaveAll(ctx context.Context, subPkgList []*entities.Subpackage) repository.IRepoError {
	panic("must be implement")
}

func (subPkgQR iSubPkgQRImpl) Update(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError {
	err := subPkgQR.findAndUpdate(ctx, subPkg)
	if err != nil {
		return err
	}

	return nil
}

func (subPkgQR iSubPkgQRImpl) UpdateAll(ctx context.Context, subPkgList []*entities.Subpackage) repository.IRepoError {
	panic("must be implement")
}
