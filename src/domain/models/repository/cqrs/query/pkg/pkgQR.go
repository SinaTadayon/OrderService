package pkg_query_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
)

type IPkgQR interface {
	FindById(ctx context.Context, oid, pid uint64) (*entities.PackageItem, repository.IRepoError)

	FindPkgItmBuyinfById(ctx context.Context, oid, pid uint64) (*entities.PackageItem, uint64, repository.IRepoError)

	FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.PackageItem, repository.IRepoError)

	ExistsById(ctx context.Context, oid, pid uint64) (bool, repository.IRepoError)

	Count(ctx context.Context, pid uint64) (int64, repository.IRepoError)

	CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError)
}
