package cqrs

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query"
	worker_pool "gitlab.faza.io/order-project/order-service/infrastructure/workerPool"
)

type iCQRSRepositoryImpl struct {
	cmdRepo   command_repository.ICmdRepository
	queryRepo query_repository.IQueryRepository
}

func CQRSFactory(ctx context.Context, command, query *mongoadapter.Mongo, cmdDatabase, cmdCollection,
	queryDatabase, queryCollection string,
	queryConnectionCount int, workerPool worker_pool.IWorkerPool) (ICQRSRepository, error) {

	commandRepo, cmdStream := command_repository.CmdRepoFactory(ctx, command, cmdDatabase, cmdCollection)
	queryRepo, err := query_repository.QueryRepoFactory(ctx, query, queryDatabase, queryCollection, cmdStream, queryConnectionCount, workerPool)
	if err != nil {
		return nil, err
	}

	cqrsRepo := &iCQRSRepositoryImpl{
		cmdRepo:   commandRepo,
		queryRepo: queryRepo,
	}

	return cqrsRepo, nil
}

func (cqrs iCQRSRepositoryImpl) CmdR() command_repository.ICmdRepository {
	return cqrs.cmdRepo
}

func (cqrs iCQRSRepositoryImpl) QueryR() query_repository.IQueryRepository {
	return cqrs.queryRepo
}
