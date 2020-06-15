package query_repository

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/finance"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/order"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/pkg"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/subpkg"
	"gitlab.faza.io/order-project/order-service/infrastructure/logger"
	worker_pool "gitlab.faza.io/order-project/order-service/infrastructure/workerPool"
)

const (
	defaultCommandStreamBuffer = 262144
)

type iQueryRepositoryImpl struct {
	orderQueryRepo   order_query_repository.IOrderQR
	pkgQueryRepo     pkg_query_repository.IPkgQR
	subPkgQueryRepo  subpkg_query_repository.ISubPkgQR
	financeQueryRepo finance_query_repository.IFinanceQR
	cmdStream        repository.CommandReaderStream
	orderStream      repository.CommandStream
	pkgStream        repository.CommandStream
	subPkgStream     repository.CommandStream
}

func QueryRepoFactory(ctx context.Context, query *mongoadapter.Mongo, database, collection string,
	commandStream repository.CommandReaderStream,
	connectionCount int, workerPool worker_pool.IWorkerPool) (IQueryRepository, error) {

	orderStream := make(chan *repository.CommandData, defaultCommandStreamBuffer)
	pkgStream := make(chan *repository.CommandData, defaultCommandStreamBuffer)
	subPkgStream := make(chan *repository.CommandData, defaultCommandStreamBuffer)

	orderQR, err := order_query_repository.OrderQRFactory(ctx, query, database, collection, orderStream, connectionCount, workerPool)
	if err != nil {
		return nil, err
	}

	pkgQR, err := pkg_query_repository.PkgQRFactory(ctx, query, database, collection, pkgStream, connectionCount, workerPool)
	if err != nil {
		return nil, err
	}

	subPkgQR, err := subpkg_query_repository.SubPkgQRFactory(ctx, query, database, collection, subPkgStream, connectionCount, workerPool)
	if err != nil {
		return nil, err
	}

	financeQR := finance_query_repository.FinanceQRFactory(query, database, collection)

	iQuery := &iQueryRepositoryImpl{
		orderQueryRepo:   orderQR,
		pkgQueryRepo:     pkgQR,
		subPkgQueryRepo:  subPkgQR,
		financeQueryRepo: financeQR,
		cmdStream:        commandStream,
		orderStream:      orderStream,
		pkgStream:        pkgStream,
		subPkgStream:     subPkgStream,
	}

	iQuery.fanOutCommandStream()
	return iQuery, nil
}

func (query iQueryRepositoryImpl) OrderQR() order_query_repository.IOrderQR {
	return query.orderQueryRepo
}

func (query iQueryRepositoryImpl) PkgQR() pkg_query_repository.IPkgQR {
	return query.pkgQueryRepo
}

func (query iQueryRepositoryImpl) SubPkgQR() subpkg_query_repository.ISubPkgQR {
	return query.subPkgQueryRepo
}

func (query iQueryRepositoryImpl) FinanceQR() finance_query_repository.IFinanceQR {
	return query.financeQueryRepo
}

func (query iQueryRepositoryImpl) fanOutCommandStream() {
	go func() {
		for commandData := range query.cmdStream {
			switch commandData.Repository {
			case repository.OrderRepo:
				query.orderStream <- commandData
			case repository.PkgRepo:
				query.pkgStream <- commandData
			case repository.SubPkgRepo:
				query.subPkgStream <- commandData
			default:
				applog.GLog.Logger.Error("repository type of received command data invalid",
					"fn", "fanOutCommandStream",
					"commandData", commandData)
			}
		}
	}()
}
