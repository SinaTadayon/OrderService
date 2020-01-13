package order_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
)

type IOrderRepository interface {
	Save(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError)

	SaveAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError)

	UpdateStatus(ctx context.Context, order *entities.Order) repository.IRepoError

	Insert(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError)

	InsertAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError)

	FindAll(ctx context.Context) ([]*entities.Order, repository.IRepoError)

	FindAllWithSort(ctx context.Context, fieldName string, direction int) ([]*entities.Order, repository.IRepoError)

	FindAllWithPage(ctx context.Context, page, perPage int64) ([]*entities.Order, int64, repository.IRepoError)

	FindAllWithPageAndSort(ctx context.Context, page, perPage int64, fieldName string, direction int) ([]*entities.Order, int64, repository.IRepoError)

	FindAllById(ctx context.Context, ids ...uint64) ([]*entities.Order, repository.IRepoError)

	FindById(ctx context.Context, orderId uint64) (*entities.Order, repository.IRepoError)

	FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.Order, repository.IRepoError)

	FindByFilterWithSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int)) ([]*entities.Order, repository.IRepoError)

	FindByFilterWithPage(ctx context.Context, supplier func() (filter interface{}), page, perPage int64) ([]*entities.Order, int64, repository.IRepoError)

	FindByFilterWithPageAndSort(ctx context.Context, supplier func() (filter interface{}, fieldName string, direction int), page, perPage int64) ([]*entities.Order, int64, repository.IRepoError)

	ExistsById(ctx context.Context, orderId uint64) (bool, repository.IRepoError)

	Count(ctx context.Context) (int64, repository.IRepoError)

	CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError)

	// only set DeletedAt field
	DeleteById(ctx context.Context, orderId uint64) (*entities.Order, repository.IRepoError)

	Delete(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError)

	DeleteAllWithOrders(ctx context.Context, orders []entities.Order) repository.IRepoError

	DeleteAll(ctx context.Context) repository.IRepoError

	// remove order from db
	RemoveById(ctx context.Context, orderId uint64) repository.IRepoError

	Remove(ctx context.Context, order entities.Order) repository.IRepoError

	RemoveAllWithOrders(ctx context.Context, orders []entities.Order) repository.IRepoError

	RemoveAll(ctx context.Context) repository.IRepoError
}
