package item_repository

import "gitlab.faza.io/order-project/order-service/domain/models/entities"

type IItemRepository interface {
	Update(item entities.Item) (*entities.Item, error)

	UpdateAll(items []entities.Item) ([]*entities.Item, error)

	Insert(item entities.Item) (*entities.Item, error)

	InsertAll(entities []entities.Item) ([]*entities.Item, error)

	FindAll() ([]*entities.Item, error)

	FindAllWithSort(fieldName string, direction int) ([]*entities.Item, error)

	FindAllWithPage(page, perPage int64) ([]*entities.Item, int64, error)

	FindAllWithPageAndSort(page, perPage int64, fieldName string, direction int) ([]*entities.Item, int64, error)

	FindAllById(ids ...string) ([]*entities.Item, error)

	FindById(itemId string) (*entities.Item, error)

	FindByFilter(supplier func() (filter interface{})) ([]*entities.Item, error)

	FindByFilterWithSort(supplier func() (filter interface{}, fieldName string, direction int)) ([]*entities.Item, error)

	FindByFilterWithPage(supplier func() (filter interface{}), page, perPage int64) ([]*entities.Item, int64, error)

	FindByFilterWithPageAndSort(supplier func() (filter interface{}, fieldName string, direction int), page, perPage int64) ([]*entities.Item, int64, error)

	ExistsById(itemId uint64) (bool, error)

	Count() (int64, error)

	CountWithFilter(supplier func() (filter interface{})) (int64, error)

	// only set DeletedAt field
	DeleteById(itemId uint64) (*entities.Item, error)

	Delete(item entities.Item) (*entities.Item, error)

	DeleteAllWithItems([]entities.Item) error

	DeleteAll() error

	// remove item from db
	RemoveById(itemId uint64) error

	Remove(item entities.Item) error

	RemoveAllWithItems([]entities.Item) error

	RemoveAll() error
}
