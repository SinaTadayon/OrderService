package order_cmd_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
)

type IOrderCR interface {
	Save(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError)

	SaveAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError)

	Update(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError)

	UpdateAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError)

	Insert(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError)

	InsertAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError)

	// only set DeletedAt field
	DeleteById(ctx context.Context, orderId uint64) (*entities.Order, repository.IRepoError)

	Delete(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError)

	DeleteAllWithOrders(ctx context.Context, orders []*entities.Order) repository.IRepoError

	DeleteAll(ctx context.Context) repository.IRepoError

	// remove order from db
	RemoveById(ctx context.Context, orderId uint64) repository.IRepoError

	Remove(ctx context.Context, order *entities.Order) repository.IRepoError

	RemoveAllWithOrders(ctx context.Context, orders []*entities.Order) repository.IRepoError

	RemoveAll(ctx context.Context) repository.IRepoError

	Count(ctx context.Context) (int64, repository.IRepoError)
}
