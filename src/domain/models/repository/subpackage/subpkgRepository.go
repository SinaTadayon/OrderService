package subpackage

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

type ISubpackageRepository interface {
	Save(ctx context.Context, subPkg *entities.Subpackage) error

	SaveAll(ctx context.Context, subPkgList []*entities.Subpackage) error

	Update(ctx context.Context, subPkg entities.Subpackage) (*entities.Subpackage, error)

	UpdateAll(ctx context.Context, subPkgList []entities.Subpackage) ([]*entities.Subpackage, error)

	FindByItemId(ctx context.Context, itemId uint64) (*entities.Subpackage, error)

	FindByOrderAndItemId(ctx context.Context, orderId, itemId uint64) (*entities.Subpackage, error)

	FindByOrderAndSellerId(ctx context.Context, orderId, sellerId uint64) ([]*entities.Subpackage, error)

	FindAll(ctx context.Context, sellerId uint64) ([]*entities.Subpackage, error)

	FindAllWithSort(ctx context.Context, sellerId uint64, fieldName string, direction int) ([]*entities.Subpackage, error)

	FindAllWithPage(ctx context.Context, sellerId uint64, page, perPage int64) ([]*entities.Subpackage, int64, error)

	FindAllWithPageAndSort(ctx context.Context, sellerId uint64, page, perPage int64, fieldName string, direction int) ([]*entities.Subpackage, int64, error)

	FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.Subpackage, error)

	FindByFilterWithSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int)) ([]*entities.Order, error)

	FindByFilterWithPage(ctx context.Context, supplier func() (filter interface{}), page, perPage int64) ([]*entities.Order, int64, error)

	FindByFilterWithPageAndSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int), page, perPage int64) ([]*entities.Order, int64, error)

	ExistsById(ctx context.Context, orderId, itemId uint64) (bool, error)

	Count(ctx context.Context, sellerId uint64) (int64, error)

	CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, error)
}
