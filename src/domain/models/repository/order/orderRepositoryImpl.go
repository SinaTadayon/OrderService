package order_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	databaseName    string = "orderService"
	collectionName  string = "orders"
	defaultDocCount int    = 1024
)

var ErrorTotalCountExceeded = errors.New("total count exceeded")
var ErrorPageNotAvailable = errors.New("page not available")
var ErrorDeleteFailed = errors.New("update deletedAt field failed")
var ErrorRemoveFailed = errors.New("remove order failed")
var ErrorUpdateFailed = errors.New("update order failed")
var ErrorVersionUpdateFailed = errors.New("update order version failed")

type iOrderRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
}

func NewOrderRepository(mongoDriver *mongoadapter.Mongo) (IOrderRepository, error) {

	_, err := mongoDriver.AddUniqueIndex(databaseName, collectionName, "orderId")
	if err != nil {
		logger.Err("create orderId index failed, error: %s", err.Error())
		return nil, err
	}

	_, err = mongoDriver.AddUniqueIndex(databaseName, collectionName, "packages.pkgId")
	if err != nil {
		logger.Err("create packages.pkgId index failed, error: %s", err.Error())
		return nil, err
	}

	_, err = mongoDriver.AddUniqueIndex(databaseName, collectionName, "packages.subpackages.items.itemId")
	if err != nil {
		logger.Err("create packages.subpackages.items.itemId index failed, error: %s", err.Error())
		return nil, err
	}

	return &iOrderRepositoryImpl{mongoDriver}, nil
}

func (repo iOrderRepositoryImpl) Save(order entities.Order) (*entities.Order, error) {

	if order.OrderId == 0 {
		order.OrderId = entities.GenerateOrderId()
		mapItemIds := make(map[int]string, 64)
		mapInventoryIds := make(map[string]int, 64)

		for i := 0; i < len(order.Packages); i++ {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				for {
					random := int(entities.GenerateRandomNumber())
					if _, ok := mapItemIds[random]; ok {
						continue
					}
					mapItemIds[random] = order.Packages[i].Subpackages[j].Item.InventoryId
					mapInventoryIds[order.Packages[i].Subpackages[j].Item.InventoryId] = random
					break
				}
			}
		}

		for i := 0; i < len(order.Packages); i++ {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if value, ok := mapInventoryIds[order.Packages[i].Subpackages[j].Item.InventoryId]; ok {
					order.Packages[i].Subpackages[j].Id = order.OrderId + uint64(value)
					order.Packages[i].Subpackages[j].CreatedAt = time.Now().UTC()
					order.Packages[i].Subpackages[j].UpdatedAt = time.Now().UTC()
				}
			}
		}

		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, &order)
		if err != nil {
			if repo.mongoAdapter.IsDupError(err) {
				for repo.mongoAdapter.IsDupError(err) {
					insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, &order)
				}
			} else {
				return nil, err
			}
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	} else {
		var currentOrder, err = repo.FindById(order.OrderId)
		if err != nil {
			return nil, ErrorUpdateFailed
		}

		//order.UpdatedAt = time.Now().UTC()
		for i := 0; i < len(order.Packages); i++ {
			if currentOrder.Packages[i].Version == order.Packages[i].Version {
				order.Packages[i].Version += 1
				for j := 0; j < len(order.Packages[i].Subpackages); j++ {
					if currentOrder.Packages[i].Subpackages[j].Version == order.Packages[i].Subpackages[j].Version {
						order.Packages[i].Subpackages[j].Version += 1
					} else {
						logger.Err("Update order failed, subpackage version obsolete, "+
							"orderId: %d, subpackage Id: %d, last version: %d, update version: ",
							order.OrderId, order.Packages[i].Subpackages[j].Id,
							order.Packages[i].Subpackages[j].Version,
							currentOrder.Packages[i].Subpackages[j].Version)
						return nil, ErrorVersionUpdateFailed
					}
				}
			} else {
				logger.Err("Update order failed, package version obsolete, "+
					"orderId: %d, package Id: %d, last version: %d, update version: ",
					order.OrderId, order.Packages[i].Id,
					order.Packages[i].Version,
					currentOrder.Packages[i].Version)
				return nil, ErrorVersionUpdateFailed
			}
		}

		updateResult, err := repo.mongoAdapter.UpdateOne(databaseName, collectionName, bson.D{{"orderId", order.OrderId}, {"deletedAt", nil}},
			bson.D{{"$set", order}})
		if err != nil {
			return nil, err
		}

		if updateResult.ModifiedCount != 1 {
			return nil, ErrorUpdateFailed
		}
	}

	return &order, nil
}

func (repo iOrderRepositoryImpl) SaveAll(orders []entities.Order) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) Insert(order entities.Order) (*entities.Order, error) {

	if order.OrderId == 0 {
		order.OrderId = entities.GenerateOrderId()
		mapItemIds := make(map[int]string, 64)
		mapInventoryIds := make(map[string]int, 64)

		for i := 0; i < len(order.Packages); i++ {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				for {
					random := int(entities.GenerateRandomNumber())
					if _, ok := mapItemIds[random]; ok {
						continue
					}
					mapItemIds[random] = order.Packages[i].Subpackages[j].Item.InventoryId
					mapInventoryIds[order.Packages[i].Subpackages[j].Item.InventoryId] = random
					break
				}
			}
		}

		for i := 0; i < len(order.Packages); i++ {
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if value, ok := mapInventoryIds[order.Packages[i].Subpackages[j].Item.InventoryId]; ok {
					order.Packages[i].Subpackages[j].Id = order.OrderId + uint64(value)
					order.Packages[i].Subpackages[j].CreatedAt = time.Now().UTC()
					order.Packages[i].Subpackages[j].UpdatedAt = time.Now().UTC()
				}
			}
		}

		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, &order)
		if err != nil {
			if repo.mongoAdapter.IsDupError(err) {
				for repo.mongoAdapter.IsDupError(err) {
					insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, &order)
				}
			} else {
				return nil, err
			}
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	}

	order.CreatedAt = time.Now().UTC()
	var insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, &order)
	if err != nil {
		return nil, err
	}
	order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	return &order, nil
}

func (repo iOrderRepositoryImpl) InsertAll(entities []entities.Order) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) FindAll() ([]*entities.Order, error) {
	total, err := repo.Count()

	if err != nil {
		logger.Err("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	if total == 0 {
		return nil, nil
	}

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (repo iOrderRepositoryImpl) FindAllWithSort(fieldName string, direction int) ([]*entities.Order, error) {
	total, err := repo.Count()
	if err != nil {
		logger.Err("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	if total == 0 {
		return nil, nil
	}

	sortMap := make(map[string]int)
	sortMap[fieldName] = direction

	optionFind := options.Find()
	optionFind.SetSort(sortMap)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}}, optionFind)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (repo iOrderRepositoryImpl) FindAllWithPage(page, perPage int64) ([]*entities.Order, int64, error) {
	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = repo.Count()
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return nil, 0, nil
	}

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}}, optionFind)
	if err != nil {
		return nil, availablePages, err
	} else if cursor.Err() != nil {
		return nil, availablePages, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, availablePages, err
		}
		orders = append(orders, &order)
	}

	return orders, availablePages, nil
}

func (repo iOrderRepositoryImpl) FindAllWithPageAndSort(page, perPage int64, fieldName string, direction int) ([]*entities.Order, int64, error) {
	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = repo.Count()
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return nil, 0, nil
	}

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	sortMap := make(map[string]int)
	sortMap[fieldName] = direction
	optionFind.SetSort(sortMap)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}}, optionFind)
	if err != nil {
		return nil, availablePages, err
	} else if cursor.Err() != nil {
		return nil, availablePages, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, availablePages, err
		}
		orders = append(orders, &order)
	}

	return orders, availablePages, nil
}

func (repo iOrderRepositoryImpl) FindAllById(ids ...uint64) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) FindById(orderId uint64) (*entities.Order, error) {
	var order entities.Order
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		return nil, err
	}

	if err := singleResult.Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (repo iOrderRepositoryImpl) FindByFilter(supplier func() interface{}) ([]*entities.Order, error) {
	filter := supplier()
	total, err := repo.CountWithFilter(supplier)
	if err != nil {
		logger.Err("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	if total == 0 {
		return nil, nil
	}

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, filter)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithSort(supplier func() (interface{}, string, int)) ([]*entities.Order, error) {
	filter, fieldName, direction := supplier()
	total, err := repo.CountWithFilter(func() interface{} { return filter })
	if err != nil {
		logger.Err("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	if total == 0 {
		return nil, nil
	}

	sortMap := make(map[string]int)
	sortMap[fieldName] = direction

	optionFind := options.Find()
	optionFind.SetSort(sortMap)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, filter, optionFind)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithPage(supplier func() interface{}, page, perPage int64) ([]*entities.Order, int64, error) {
	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	filter := supplier()
	var totalCount, err = repo.CountWithFilter(supplier)

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, filter, optionFind)
	if err != nil {
		return nil, availablePages, err
	} else if cursor.Err() != nil {
		return nil, availablePages, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, availablePages, err
		}
		orders = append(orders, &order)
	}

	return orders, availablePages, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithPageAndSort(supplier func() (interface{}, string, int), page, perPage int64) ([]*entities.Order, int64, error) {
	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	filter, fieldName, direction := supplier()
	var totalCount, err = repo.CountWithFilter(func() interface{} { return filter })

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	sortMap := make(map[string]int)
	sortMap[fieldName] = direction
	optionFind.SetSort(sortMap)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, filter, optionFind)
	if err != nil {
		return nil, availablePages, err
	} else if cursor.Err() != nil {
		return nil, availablePages, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, availablePages, err
		}
		orders = append(orders, &order)
	}

	return orders, availablePages, nil
}

func (repo iOrderRepositoryImpl) ExistsById(orderId uint64) (bool, error) {
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		if repo.mongoAdapter.NoDocument(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo iOrderRepositoryImpl) Count() (int64, error) {
	total, err := repo.mongoAdapter.Count(databaseName, collectionName, bson.D{{"deletedAt", nil}})
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (repo iOrderRepositoryImpl) CountWithFilter(supplier func() interface{}) (int64, error) {
	total, err := repo.mongoAdapter.Count(databaseName, collectionName, supplier())
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (repo iOrderRepositoryImpl) DeleteById(orderId uint64) (*entities.Order, error) {
	var err error
	order, err := repo.FindById(orderId)
	if err != nil {
		return nil, err
	}

	deletedAt := time.Now().UTC()
	order.DeletedAt = &deletedAt

	updateResult, err := repo.mongoAdapter.UpdateOne(databaseName, collectionName,
		bson.D{{"orderId", order.OrderId}, {"deletedAt", nil}},
		bson.D{{"$set", order}})
	if err != nil {
		return nil, err
	}

	if updateResult.ModifiedCount != 1 {
		return nil, ErrorDeleteFailed
	}

	return order, nil
}

func (repo iOrderRepositoryImpl) Delete(order entities.Order) (*entities.Order, error) {
	return repo.DeleteById(order.OrderId)
}

func (repo iOrderRepositoryImpl) DeleteAllWithOrders([]entities.Order) error {
	panic("implementation required")
}

// TODO items.$.deleteAt must be checked if nil
func (repo iOrderRepositoryImpl) DeleteAll() error {
	_, err := repo.mongoAdapter.UpdateMany(databaseName, collectionName,
		bson.D{{"deletedAt", nil}, {"items.$.deletedAt", nil}},
		bson.M{"$set": bson.M{"deletedAt": time.Now().UTC(),
			"items.$.deletedAt": time.Now().UTC()}})
	if err != nil {
		return err
	}
	return nil
}

func (repo iOrderRepositoryImpl) RemoveById(orderId uint64) error {
	result, err := repo.mongoAdapter.DeleteOne(databaseName, collectionName, bson.M{"orderId": orderId})
	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return ErrorRemoveFailed
	}
	return nil
}

func (repo iOrderRepositoryImpl) Remove(order entities.Order) error {
	return repo.RemoveById(order.OrderId)
}

func (repo iOrderRepositoryImpl) RemoveAllWithOrders([]entities.Order) error {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) RemoveAll() error {
	_, err := repo.mongoAdapter.DeleteMany(databaseName, collectionName, bson.M{})
	return err
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		logger.Err("closeCursor() failed, err: %s", err)
	}
}
