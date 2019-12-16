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
	"strconv"
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

func NewOrderRepository(mongoDriver *mongoadapter.Mongo) IOrderRepository {
	return &iOrderRepositoryImpl{mongoDriver}
}

func (repo iOrderRepositoryImpl) generateAndSetId(ctx context.Context, order entities.Order) *entities.Order {
	order.OrderId = entities.GenerateOrderId()
	mapItemIds := make(map[int]uint64, 64)

	order.CreatedAt = time.Now().UTC()
	order.UpdatedAt = time.Now().UTC()
	for i := 0; i < len(order.Packages); i++ {
		order.Packages[i].OrderId = order.OrderId
		order.Packages[i].CreatedAt = time.Now().UTC()
		order.Packages[i].UpdatedAt = time.Now().UTC()
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			for {
				random := int(entities.GenerateRandomNumber())
				if _, ok := mapItemIds[random]; ok {
					continue
				}
				mapItemIds[random] = order.Packages[i].PId
				sid, _ := strconv.Atoi(strconv.Itoa(int(order.OrderId)) + strconv.Itoa(random))
				order.Packages[i].Subpackages[j].SId = uint64(sid)
				order.Packages[i].Subpackages[j].PId = order.Packages[i].PId
				order.Packages[i].Subpackages[j].OrderId = order.OrderId
				order.Packages[i].Subpackages[j].CreatedAt = time.Now().UTC()
				order.Packages[i].Subpackages[j].UpdatedAt = time.Now().UTC()
				break
			}
		}
	}

	return &order
}

func (repo iOrderRepositoryImpl) Save(ctx context.Context, order entities.Order) (*entities.Order, error) {

	if order.OrderId == 0 {
		var newOrder *entities.Order
		for {
			newOrder = repo.generateAndSetId(ctx, order)
			var insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, newOrder)
			if err != nil {
				if repo.mongoAdapter.IsDupError(err) {
					continue
				} else {
					return nil, errors.Wrap(err, "Save Order Failed")
				}
			}
			newOrder.ID = insertOneResult.InsertedID.(primitive.ObjectID)
			break
		}
		return newOrder, nil
	} else {
		var currentOrder, err = repo.FindById(ctx, order.OrderId)
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
							"orderId: %d, sid: %d, last version: %d, update version: %d",
							order.OrderId, order.Packages[i].Subpackages[j].SId,
							order.Packages[i].Subpackages[j].Version,
							currentOrder.Packages[i].Subpackages[j].Version)
						return nil, ErrorVersionUpdateFailed
					}
				}
			} else {
				logger.Err("Update order failed, package version obsolete, "+
					"orderId: %d, sellerId: %d, last version: %d, update version: %d",
					order.OrderId, order.Packages[i].PId,
					order.Packages[i].Version,
					currentOrder.Packages[i].Version)
				return nil, ErrorVersionUpdateFailed
			}
		}

		updateResult, err := repo.mongoAdapter.UpdateOne(databaseName, collectionName, bson.D{{"orderId", order.OrderId}, {"deletedAt", nil}},
			bson.D{{"$set", order}})
		if err != nil {
			return nil, errors.Wrap(err, "Save Order Failed")
		}

		if updateResult.ModifiedCount != 1 {
			return nil, ErrorUpdateFailed
		}
	}

	return &order, nil
}

func (repo iOrderRepositoryImpl) SaveAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) Insert(ctx context.Context, order entities.Order) (*entities.Order, error) {

	if order.OrderId == 0 {
		var newOrder *entities.Order
		for {
			newOrder = repo.generateAndSetId(ctx, order)
			var insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, newOrder)
			if err != nil {
				if repo.mongoAdapter.IsDupError(err) {
					continue
				} else {
					return nil, errors.Wrap(err, "Insert Order Failed")
				}
			}
			newOrder.ID = insertOneResult.InsertedID.(primitive.ObjectID)
			break
		}
		return newOrder, nil
	} else {
		var insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, &order)
		if err != nil {
			return nil, errors.Wrap(err, "Insert Order Failed")
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	}
	return &order, nil
}

func (repo iOrderRepositoryImpl) InsertAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) FindAll(ctx context.Context) ([]*entities.Order, error) {
	total, err := repo.Count(ctx)

	if err != nil {
		logger.Err("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	if total == 0 {
		return nil, nil
	}

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}})
	if err != nil {
		return nil, errors.Wrap(err, "FindAll Orders Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, errors.Wrap(err, "FindAll Orders Failed")
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (repo iOrderRepositoryImpl) FindAllWithSort(ctx context.Context, fieldName string, direction int) ([]*entities.Order, error) {
	total, err := repo.Count(ctx)
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
		return nil, errors.Wrap(err, "FindAllWithSort Orders Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, errors.Wrap(err, "FindAllWithSort Orders Failed")
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (repo iOrderRepositoryImpl) FindAllWithPage(ctx context.Context, page, perPage int64) ([]*entities.Order, int64, error) {
	if page <= 0 || perPage <= 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = repo.Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindAllWithPage Orders Failed")
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
		return nil, totalCount, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}}, optionFind)
	if err != nil {
		return nil, totalCount, errors.Wrap(err, "FindAllWithPage Orders Failed")
	} else if cursor.Err() != nil {
		return nil, totalCount, errors.Wrap(err, "FindAllWithPage Orders Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, errors.Wrap(err, "FindAllWithPage Orders Failed")
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil
}

func (repo iOrderRepositoryImpl) FindAllWithPageAndSort(ctx context.Context, page, perPage int64, fieldName string, direction int) ([]*entities.Order, int64, error) {
	if page <= 0 || perPage <= 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = repo.Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindAllWithPageAndSort Orders Failed")
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
		return nil, totalCount, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	if fieldName != "" {
		sortMap := make(map[string]int)
		sortMap[fieldName] = direction
		optionFind.SetSort(sortMap)
	}

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}}, optionFind)
	if err != nil {
		return nil, totalCount, errors.Wrap(err, "FindAllWithPageAndSort Orders Failed")
	} else if cursor.Err() != nil {
		return nil, totalCount, errors.Wrap(err, "FindAllWithPageAndSort Orders Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, errors.Wrap(err, "FindAllWithPageAndSort Orders Failed")
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil
}

func (repo iOrderRepositoryImpl) FindAllById(ctx context.Context, ids ...uint64) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) FindById(ctx context.Context, orderId uint64) (*entities.Order, error) {
	var order entities.Order
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		return nil, errors.Wrap(err, "FindById Order Failed")
	}

	if err := singleResult.Decode(&order); err != nil {
		return nil, errors.Wrap(err, "FindById Order Failed")
	}

	return &order, nil
}

func (repo iOrderRepositoryImpl) FindByFilter(ctx context.Context, supplier func() interface{}) ([]*entities.Order, error) {
	filter := supplier()
	total, err := repo.CountWithFilter(ctx, supplier)
	if err != nil {
		logger.Err("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	if total == 0 {
		return nil, nil
	}

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, filter)
	if err != nil {
		return nil, errors.Wrap(err, "FindByFilter Order Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, errors.Wrap(err, "FindByFilter Order Failed")
		}
		orders = append(orders, &order)
	}

	return orders, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithSort(ctx context.Context, supplier func() (interface{}, string, int)) ([]*entities.Order, error) {
	filter, fieldName, direction := supplier()
	total, err := repo.CountWithFilter(ctx, func() interface{} { return filter })
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
		return nil, errors.Wrap(err, "FindByFilterWithSort Orders Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, errors.Wrap(err, "FindByFilterWithSort Orders Failed")
		}
		orders = append(orders, &order)
	}

	return orders, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithPage(ctx context.Context, supplier func() interface{}, page, perPage int64) ([]*entities.Order, int64, error) {
	if page <= 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	filter := supplier()
	var totalCount, err = repo.CountWithFilter(ctx, supplier)
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindByFilterWithPage Orders Failed")
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
		return nil, totalCount, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, filter, optionFind)
	if err != nil {
		return nil, totalCount, errors.Wrap(err, "FindByFilterWithPage Orders Failed")
	} else if cursor.Err() != nil {
		return nil, totalCount, errors.Wrap(err, "FindByFilterWithPage Orders Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, errors.Wrap(err, "FindByFilterWithPage Orders Failed")
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithPageAndSort(ctx context.Context, supplier func() (interface{}, string, int), page, perPage int64) ([]*entities.Order, int64, error) {
	if page <= 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	filter, fieldName, direction := supplier()
	var totalCount, err = repo.CountWithFilter(ctx, func() interface{} { return filter })
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindByFilterWithPageAndSort Orders Failed")
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
		return nil, totalCount, ErrorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, ErrorTotalCountExceeded
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	if fieldName != "" {
		sortMap := make(map[string]int)
		sortMap[fieldName] = direction
		optionFind.SetSort(sortMap)
	}

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName, filter, optionFind)
	if err != nil {
		return nil, totalCount, errors.Wrap(err, "FindByFilterWithPageAndSort Orders Failed")
	} else if cursor.Err() != nil {
		return nil, totalCount, errors.Wrap(err, "FindByFilterWithPageAndSort Orders Failed")
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, errors.Wrap(err, "FindByFilterWithPageAndSort Orders Failed")
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil
}

func (repo iOrderRepositoryImpl) ExistsById(ctx context.Context, orderId uint64) (bool, error) {
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		if repo.mongoAdapter.NoDocument(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "ExistsById Order Failed")
	}
	return true, nil
}

func (repo iOrderRepositoryImpl) Count(ctx context.Context) (int64, error) {
	total, err := repo.mongoAdapter.Count(databaseName, collectionName, bson.D{{"deletedAt", nil}})
	if err != nil {
		return 0, errors.Wrap(err, "Count Orders Failed")
	}
	return total, nil
}

func (repo iOrderRepositoryImpl) CountWithFilter(ctx context.Context, supplier func() interface{}) (int64, error) {
	total, err := repo.mongoAdapter.Count(databaseName, collectionName, supplier())
	if err != nil {
		return 0, errors.Wrap(err, "CountWithFilter Orders Failed")
	}
	return total, nil
}

func (repo iOrderRepositoryImpl) DeleteById(ctx context.Context, orderId uint64) (*entities.Order, error) {
	var err error
	order, err := repo.FindById(ctx, orderId)
	if err != nil {
		return nil, err
	}

	deletedAt := time.Now().UTC()
	order.DeletedAt = &deletedAt

	updateResult, err := repo.mongoAdapter.UpdateOne(databaseName, collectionName,
		bson.D{{"orderId", order.OrderId}, {"deletedAt", nil}},
		bson.D{{"$set", order}})
	if err != nil {
		return nil, errors.Wrap(err, "DeleteById Order Failed")
	}

	if updateResult.ModifiedCount != 1 {
		return nil, ErrorDeleteFailed
	}

	return order, nil
}

func (repo iOrderRepositoryImpl) Delete(ctx context.Context, order entities.Order) (*entities.Order, error) {
	return repo.DeleteById(ctx, order.OrderId)
}

func (repo iOrderRepositoryImpl) DeleteAllWithOrders(ctx context.Context, orders []entities.Order) error {
	panic("implementation required")
}

// TODO cascade delete in packages and subpackages
func (repo iOrderRepositoryImpl) DeleteAll(ctx context.Context) error {
	_, err := repo.mongoAdapter.UpdateMany(databaseName, collectionName,
		bson.D{{"deletedAt", nil}},
		bson.M{"$set": bson.M{"deletedAt": time.Now().UTC()}})
	if err != nil {
		return errors.Wrap(err, "DeleteAll Order Failed")
	}
	return nil
}

func (repo iOrderRepositoryImpl) RemoveById(ctx context.Context, orderId uint64) error {
	result, err := repo.mongoAdapter.DeleteOne(databaseName, collectionName, bson.M{"orderId": orderId})
	if err != nil {
		return errors.Wrap(err, "RemoveById Order Failed")
	}

	if result.DeletedCount != 1 {
		return ErrorRemoveFailed
	}
	return nil
}

func (repo iOrderRepositoryImpl) Remove(ctx context.Context, order entities.Order) error {
	return repo.RemoveById(ctx, order.OrderId)
}

func (repo iOrderRepositoryImpl) RemoveAllWithOrders(ctx context.Context, orders []entities.Order) error {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) RemoveAll(ctx context.Context) error {
	_, err := repo.mongoAdapter.DeleteMany(databaseName, collectionName, bson.M{})
	return err
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		logger.Err("closeCursor() failed, err: %s", err)
	}
}
