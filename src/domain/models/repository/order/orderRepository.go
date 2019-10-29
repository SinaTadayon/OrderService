package order_repository

import "gitlab.faza.io/order-project/order-service/domain/models/entities"

type IOrderRepository interface {

	Save(order entities.Order) (*entities.Order, error)

	SaveAll(orders []entities.Order) ([]*entities.Order, error)

	Insert(order entities.Order) (*entities.Order, error)

	InsertAll(entities []entities.Order) ([]*entities.Order, error)

	FindAll() ([]*entities.Order, error)

	FindAllWithSort(fieldName string, direction int) ([]*entities.Order, error)

	FindAllWithPage(page, perPage int64) ([]*entities.Order, int64, error)

	FindAllWithPageAndSort(page, perPage int64, fieldName string, direction int) ([]*entities.Order, int64, error)

	FindAllById(ids ...string) ([]*entities.Order, error)

	FindById(orderId string) (*entities.Order, error)

	FindByFilter(supplier func() (filter interface{})) ([]*entities.Order, error)

	FindByFilterWithSort(supplier func() (filter interface{}, fieldName string, direction int)) ([]*entities.Order, error)

	FindByFilterWithPage(supplier func() (filter interface{}), page, perPage int64) ([]*entities.Order, int64, error)

	FindByFilterWithPageAndSort(supplier func() (filter interface{}, fieldName string, direction int) , page, perPage int64) ([]*entities.Order, int64, error)

	ExistsById(orderId string) (bool, error)

	Count() (int64, error)

	CountWithFilter(supplier func() (filter interface{})) (int64, error)

	// only set DeletedAt field
	DeleteById(orderId string) (*entities.Order, error)

	Delete(order entities.Order) (*entities.Order, error)

	DeleteAllWithOrders([]entities.Order) error

	DeleteAll() error

	// remove order from db
	RemoveById(orderId string) error

	Remove(order entities.Order) error

	RemoveAllWithOrders([]entities.Order) error

	RemoveAll() error
}
