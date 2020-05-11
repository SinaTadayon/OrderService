package pkg_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
)

type IPkgItemRepository interface {
	Update(ctx context.Context, pkgItem entities.PackageItem) (*entities.PackageItem, repository.IRepoError)

	UpdateWithUpsert(ctx context.Context, pkgItem entities.PackageItem) (*entities.PackageItem, repository.IRepoError)

	FindPkgItmBuyinfById(ctx context.Context, orderId uint64, id uint64) (*entities.PackageItem, uint64, repository.IRepoError)

	FindById(ctx context.Context, orderId uint64, id uint64) (*entities.PackageItem, repository.IRepoError)

	FindPkgItmBuyinfById(ctx context.Context, orderId uint64, id uint64) (*entities.PackageItem, uint64, repository.IRepoError)

	FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.PackageItem, repository.IRepoError)

	ExistsById(ctx context.Context, orderId uint64, id uint64) (bool, repository.IRepoError)

	Count(ctx context.Context, id uint64) (int64, repository.IRepoError)

	CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError)
}
