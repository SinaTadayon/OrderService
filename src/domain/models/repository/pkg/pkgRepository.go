package pkg_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

type IPkgItemRepository interface {
	Update(ctx context.Context, pkgItem entities.PackageItem) (*entities.PackageItem, error)

	FindById(ctx context.Context, orderId uint64, id uint64) (*entities.PackageItem, error)

	FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.PackageItem, error)

	FindByFilterWithPage(ctx context.Context, totalSupplier, supplier func() (filter interface{}), page, perPage int64) ([]*entities.PackageItem, int64, error)

	ExistsById(ctx context.Context, orderId uint64, id uint64) (bool, error)

	Count(ctx context.Context, id uint64) (int64, error)

	CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, error)
}
