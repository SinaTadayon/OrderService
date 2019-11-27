package order_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

type IOrderRepository interface {
	Save(ctx context.Context, order entities.Order) (*entities.Order, error)

	SaveAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, error)

	Insert(ctx context.Context, order *entities.Order) error

	InsertAll(ctx context.Context, orders []*entities.Order) error

	FindAll(ctx context.Context) ([]*entities.Order, error)

	FindAllWithSort(ctx context.Context, fieldName string, direction int) ([]*entities.Order, error)

	FindAllWithPage(ctx context.Context, page, perPage int64) ([]*entities.Order, int64, error)

	FindAllWithPageAndSort(ctx context.Context, page, perPage int64, fieldName string, direction int) ([]*entities.Order, int64, error)

	FindAllById(ctx context.Context, ids ...uint64) ([]*entities.Order, error)

	FindById(ctx context.Context, orderId uint64) (*entities.Order, error)

	FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.Order, error)

	FindByFilterWithSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int)) ([]*entities.Order, error)

	FindByFilterWithPage(ctx context.Context, supplier func() (filter interface{}), page, perPage int64) ([]*entities.Order, int64, error)

	FindByFilterWithPageAndSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int), page, perPage int64) ([]*entities.Order, int64, error)

	ExistsById(ctx context.Context, orderId uint64) (bool, error)

	Count(ctx context.Context) (int64, error)

	CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, error)

	// only set DeletedAt field
	DeleteById(ctx context.Context, orderId uint64) (*entities.Order, error)

	Delete(ctx context.Context, order entities.Order) (*entities.Order, error)

	DeleteAllWithOrders(ctx context.Context, orders []entities.Order) error

	DeleteAll(ctx context.Context) error

	// remove order from db
	RemoveById(ctx context.Context, orderId uint64) error

	Remove(ctx context.Context, order entities.Order) error

	RemoveAllWithOrders(ctx context.Context, orders []entities.Order) error

	RemoveAll(ctx context.Context) error
}
