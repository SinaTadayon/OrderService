package subpkg_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
)

type ISubpackageRepository interface {
	Save(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError

	SaveAll(ctx context.Context, subPkgList []*entities.Subpackage) repository.IRepoError

	Update(ctx context.Context, subPkg entities.Subpackage) (*entities.Subpackage, repository.IRepoError)

	UpdateAll(ctx context.Context, subPkgList []entities.Subpackage) ([]*entities.Subpackage, repository.IRepoError)

	//FindByItemId(ctx context.Context, sid uint64) (*entities.Subpackage, repository.IRepoError)

	FindByOrderAndItemId(ctx context.Context, orderId, sid uint64) (*entities.Subpackage, repository.IRepoError)

	FindByOrderAndSellerId(ctx context.Context, orderId, pid uint64) ([]*entities.Subpackage, repository.IRepoError)

	FindAll(ctx context.Context, pid uint64) ([]*entities.Subpackage, repository.IRepoError)

	FindAllWithSort(ctx context.Context, pid uint64, fieldName string, direction int) ([]*entities.Subpackage, repository.IRepoError)

	FindAllWithPage(ctx context.Context, pid uint64, page, perPage int64) ([]*entities.Subpackage, int64, repository.IRepoError)

	FindAllWithPageAndSort(ctx context.Context, pid uint64, page, perPage int64, fieldName string, direction int) ([]*entities.Subpackage, int64, repository.IRepoError)

	FindByFilter(ctx context.Context, totalSupplier func() (filter interface{}), supplier func() (filter interface{})) ([]*entities.Subpackage, repository.IRepoError)

	FindByFilterWithPage(ctx context.Context, totalSupplier func() (filter interface{}), supplier func() (filter interface{}), page, perPage int64) ([]*entities.Subpackage, int64, repository.IRepoError)

	ExistsById(ctx context.Context, sid uint64) (bool, repository.IRepoError)

	Count(ctx context.Context, pid uint64) (int64, repository.IRepoError)

	CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError)
}
