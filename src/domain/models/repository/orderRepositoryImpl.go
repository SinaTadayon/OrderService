package repository

import (
	"context"
	"errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	databaseName  	string = "orderService"
	collectionName  string = "orders"
)

var errorTotalCountExceeded 	= errors.New("total count exceeded")
var errorPageNotAvailable    	= errors.New("page not available")
var errorDeleteFailed 			= errors.New("update deletedAt field failed")
var errorRemoveFailed 			= errors.New("remove order failed")
var errorUpdateFailed 			= errors.New("update order failed")

type iOrderRepositoryImpl struct {
	mongoAdapter  *mongoadapter.Mongo
}

func NewOrderRepository(config *configs.Cfg) (IOrderRepository, error) {

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:  time.Duration(config.Mongo.ConnectionTimeout),
		ReadTimeout:  time.Duration(config.Mongo.ReadTimeout),
		WriteTimeout: time.Duration(config.Mongo.WriteTimeout),
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("NewOrderRepository Mongo: %v", err.Error())
		return nil, err
	}
	_, err = mongoDriver.AddUniqueIndex(databaseName, collectionName, "orderId")
	if err != nil {
		logger.Err(err.Error())
		return nil, err
	}

	return &iOrderRepositoryImpl{mongoDriver}, nil
}

func (repo iOrderRepositoryImpl) Save(order entities.Order) (*entities.Order, error) {

	if len(order.OrderId) == 0 {
		order.OrderId = entities.GenerateOrderId()
		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()
		var insertOneResult, err = repo.mongoAdapter.InsertOne(databaseName, collectionName, order)
		if err != nil {
			return nil, err
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	} else {
		order.UpdatedAt = time.Now()
		var updateResult, err = repo.mongoAdapter.UpdateOne(databaseName, collectionName, bson.D{{"orderId", order.OrderId},{"deletedAt", nil}},
			bson.D{{"$set", order}})
		if err != nil {
			return nil, err
		}

		if updateResult.ModifiedCount != 1 {
			return nil, errorUpdateFailed
		}
	}

	return &order, nil
}

func (repo iOrderRepositoryImpl) SaveAll(orders []entities.Order) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) Insert(order entities.Order) (*entities.Order, error) {
	if len(order.OrderId) == 0 {
		order.OrderId = entities.GenerateOrderId()
	}

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	var insertOneResult, err= repo.mongoAdapter.InsertOne(databaseName, collectionName, order)
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
		total = 128
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
		total = 128
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

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount % perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, errorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, errorTotalCountExceeded
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

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount % perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, errorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, errorTotalCountExceeded
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

func (repo iOrderRepositoryImpl) FindAllById(ids ...string) ([]*entities.Order, error) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) FindById(orderId string) (*entities.Order, error) {
	var order entities.Order
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId},{"deletedAt", nil}})
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
		total = 128
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
		total = 128
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

	if totalCount % perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, errorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, errorTotalCountExceeded
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
	var totalCount, err = repo.CountWithFilter(func() interface{} {return filter})

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount % perPage != 0 {
		availablePages = (totalCount / perPage) + 1
	} else {
		availablePages = totalCount / perPage
	}

	if totalCount < perPage {
		availablePages = 1
	}

	if availablePages < page {
		return nil, availablePages, errorPageNotAvailable
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, availablePages, errorTotalCountExceeded
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

func (repo iOrderRepositoryImpl) ExistsById(orderId string) (bool, error) {
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId},{"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		if repo.mongoAdapter.NoDocument(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo iOrderRepositoryImpl) Count() (int64, error) {
	total, err := repo.mongoAdapter.Count(databaseName, collectionName, bson.D{{"deletedAt", nil }})
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

func (repo iOrderRepositoryImpl) DeleteById(orderId string) (*entities.Order, error) {
	var err error
	order, err := repo.FindById(orderId)
	if err != nil {
		return nil, err
	}

	deletedAt := time.Now().UTC()
	order.DeletedAt = &deletedAt
	updateResult, err := repo.mongoAdapter.UpdateOne(databaseName, collectionName,
					bson.D{{"orderId", order.OrderId},{"deletedAt", nil}},
					bson.D{{"$set", order}})
	if err != nil {
		return nil, err
	}

	if updateResult.ModifiedCount != 1 {
		return nil, errorDeleteFailed
	}

	return order, nil
}

func (repo iOrderRepositoryImpl) Delete(order entities.Order) (*entities.Order, error) {
	return repo.DeleteById(order.OrderId)
}

func (repo iOrderRepositoryImpl) DeleteAllWithOrders([]entities.Order) error {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) DeleteAll() error {
	 _, err := repo.mongoAdapter.UpdateMany(databaseName, collectionName, bson.D{{"deletedAt", nil}},
					bson.M{"$set": bson.M{"deletedAt": time.Now().UTC()}})
	 if err != nil {
	 	return err
	 }
	 return nil
}

func (repo iOrderRepositoryImpl) RemoveById(orderId string) error {
	 result, err := repo.mongoAdapter.DeleteOne(databaseName, collectionName, bson.M{"orderId": orderId})
	 if err != nil {
	 	return err
	 }

	 if result.DeletedCount != 1 {
	 	return errorRemoveFailed
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