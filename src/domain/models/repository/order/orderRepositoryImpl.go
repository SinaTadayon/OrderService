package order_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

type iOrderRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
	database     string
	collection   string
}

func NewOrderRepository(mongoDriver *mongoadapter.Mongo, database, collection string) IOrderRepository {
	return &iOrderRepositoryImpl{mongoDriver, database, collection}
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

func (repo iOrderRepositoryImpl) Save(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError) {

	if order.OrderId == 0 {
		var newOrder *entities.Order
		for {
			newOrder = repo.generateAndSetId(ctx, order)
			var insertOneResult, err = repo.mongoAdapter.InsertOne(repo.database, repo.collection, newOrder)
			if err != nil {
				if repo.mongoAdapter.IsDupError(err) {
					continue
				} else {
					return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Save Order Failed"))
				}
			}
			newOrder.ID = insertOneResult.InsertedID.(primitive.ObjectID)
			break
		}
		return newOrder, nil
	} else {
		order.UpdatedAt = time.Now().UTC()
		currentVersion := order.Version
		order.Version += 1
		updateResult, e := repo.mongoAdapter.UpdateOne(repo.database, repo.collection, bson.D{{"orderId", order.OrderId}, {"deletedAt", nil}, {"version", currentVersion}},
			bson.D{{"$set", order}})
		if e != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "UpdateOne Failed"))
		}

		if updateResult.MatchedCount != 1 || updateResult.ModifiedCount != 1 {
			return nil, repository.ErrorFactory(repository.NotFoundErr, "Order Not Found", errors.New("Order Not Found"))
		}
	}

	return &order, nil
}

func (repo iOrderRepositoryImpl) SaveAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) UpdateStatus(ctx context.Context, order *entities.Order) repository.IRepoError {
	order.UpdatedAt = time.Now().UTC()
	currentVersion := order.Version
	order.Version += 1
	opt := options.FindOneAndUpdate()
	opt.SetUpsert(false)
	singleResult := repo.mongoAdapter.GetConn().Database(repo.database).Collection(repo.collection).FindOneAndUpdate(ctx,
		bson.D{
			{"orderId", order.OrderId},
			{"version", currentVersion},
		},
		bson.D{{"$set", bson.D{{"version", order.Version}, {"status", order.Status}, {"updateAt", order.UpdatedAt}}}}, opt)
	if singleResult.Err() != nil {
		if repo.mongoAdapter.NoDocument(singleResult.Err()) {
			return repository.ErrorFactory(repository.NotFoundErr, "Package Not Found", repository.ErrorUpdateFailed)
		}
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), ""))
	}

	return nil
}

func (repo iOrderRepositoryImpl) Insert(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError) {

	if order.OrderId == 0 {
		var newOrder *entities.Order
		for {
			newOrder = repo.generateAndSetId(ctx, order)
			var insertOneResult, err = repo.mongoAdapter.InsertOne(repo.database, repo.collection, newOrder)
			if err != nil {
				if repo.mongoAdapter.IsDupError(err) {
					continue
				} else {
					return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Insert Order Failed"))
				}
			}
			newOrder.ID = insertOneResult.InsertedID.(primitive.ObjectID)
			break
		}
		return newOrder, nil
	} else {
		var insertOneResult, err = repo.mongoAdapter.InsertOne(repo.database, repo.collection, &order)
		if err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Insert Order Failed"))
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	}
	return &order, nil
}

func (repo iOrderRepositoryImpl) InsertAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) FindAll(ctx context.Context) ([]*entities.Order, repository.IRepoError) {
	total, err := repo.Count(ctx)

	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
	}

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, bson.D{{"deletedAt", nil}})
	if e != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "FindMany Orders Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Order Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (repo iOrderRepositoryImpl) FindAllWithSort(ctx context.Context, fieldName string, direction int) ([]*entities.Order, repository.IRepoError) {
	total, err := repo.Count(ctx)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
	}

	sortMap := make(map[string]int)
	sortMap[fieldName] = direction

	optionFind := options.Find()
	optionFind.SetSort(sortMap)

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, bson.D{{"deletedAt", nil}}, optionFind)
	if e != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "FindMany Order Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Order Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (repo iOrderRepositoryImpl) FindAllWithPage(ctx context.Context, page, perPage int64) ([]*entities.Order, int64, repository.IRepoError) {
	if page <= 0 || perPage <= 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("neither offset nor start can be zero"))
	}

	var totalCount, err = repo.Count(ctx)
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
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorPageNotAvailable)
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorTotalCountExceeded)
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, bson.D{{"deletedAt", nil}}, optionFind)
	if e != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "FindMany Orders Failed"))
	} else if cursor.Err() != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "FindMany Orders Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Orders Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil
}

func (repo iOrderRepositoryImpl) FindAllWithPageAndSort(ctx context.Context, page, perPage int64, fieldName string, direction int) ([]*entities.Order, int64, repository.IRepoError) {
	if page <= 0 || perPage <= 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("neither offset nor start can be zero"))
	}

	var totalCount, err = repo.Count(ctx)
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
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorPageNotAvailable)
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorTotalCountExceeded)
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	if fieldName != "" {
		sortMap := make(map[string]int)
		sortMap[fieldName] = direction
		optionFind.SetSort(sortMap)
	}

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, bson.D{{"deletedAt", nil}}, optionFind)
	if e != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "FindMany Orders Failed"))
	} else if cursor.Err() != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "FindMany Orders Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Orders Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil
}

func (repo iOrderRepositoryImpl) FindAllById(ctx context.Context, ids ...uint64) ([]*entities.Order, repository.IRepoError) {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) FindById(ctx context.Context, orderId uint64) (*entities.Order, repository.IRepoError) {
	var order entities.Order
	singleResult := repo.mongoAdapter.FindOne(repo.database, repo.collection, bson.D{{"orderId", orderId}, {"deletedAt", nil}})
	if singleResult.Err() != nil {
		if repo.mongoAdapter.NoDocument(singleResult.Err()) {
			return nil, repository.ErrorFactory(repository.NotFoundErr, "Order Not Found", errors.Wrap(singleResult.Err(), "Order Not Found"))
		}

		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), "FindById failed"))
	}

	if err := singleResult.Decode(&order); err != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), "Decode Result failed"))
	}

	return &order, nil
}

func (repo iOrderRepositoryImpl) FindByFilter(ctx context.Context, supplier func() interface{}) ([]*entities.Order, repository.IRepoError) {
	filter := supplier()
	total, err := repo.CountWithFilter(ctx, supplier)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
	}

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, filter)
	if e != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "FindMany Order Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Order Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (repo iOrderRepositoryImpl) FindByFilterWithSort(ctx context.Context, supplier func() (interface{}, string, int)) ([]*entities.Order, repository.IRepoError) {
	filter, fieldName, direction := supplier()
	total, err := repo.CountWithFilter(ctx, func() interface{} { return filter })
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
	}

	sortMap := make(map[string]int)
	sortMap[fieldName] = direction

	optionFind := options.Find()
	optionFind.SetSort(sortMap)

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, filter, optionFind)
	if e != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "FindMany Orders Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Orders Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithPage(ctx context.Context, supplier func() interface{}, page, perPage int64) ([]*entities.Order, int64, repository.IRepoError) {
	if page <= 0 || perPage == 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("neither offset nor start can be zero"))
	}

	filter := supplier()
	var totalCount, err = repo.CountWithFilter(ctx, supplier)
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
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorPageNotAvailable)
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorTotalCountExceeded)
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, filter, optionFind)
	if e != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "FindMany Orders Failed"))
	} else if cursor.Err() != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "FindMany Orders Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Orders Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil

}

func (repo iOrderRepositoryImpl) FindByFilterWithPageAndSort(ctx context.Context, supplier func() (interface{}, string, int), page, perPage int64) ([]*entities.Order, int64, repository.IRepoError) {
	if page <= 0 || perPage == 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("Page/PerPage Invalid"))
	}
	filter, fieldName, direction := supplier()
	var totalCount, err = repo.CountWithFilter(ctx, func() interface{} { return filter })
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
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorPageNotAvailable)
	}

	var offset = (page - 1) * perPage
	if offset >= totalCount {
		return nil, totalCount, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", repository.ErrorTotalCountExceeded)
	}

	optionFind := options.Find()
	optionFind.SetLimit(perPage)
	optionFind.SetSkip(offset)

	if fieldName != "" {
		sortMap := make(map[string]int)
		sortMap[fieldName] = direction
		optionFind.SetSort(sortMap)
	}

	cursor, e := repo.mongoAdapter.FindMany(repo.database, repo.collection, filter, optionFind)
	if e != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "FindMany Orders Failed"))
	} else if cursor.Err() != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "FindMany Orders Failed"))
	}

	defer closeCursor(ctx, cursor)
	orders := make([]*entities.Order, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var order entities.Order
		// decode the document
		if err := cursor.Decode(&order); err != nil {
			return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Orders Failed"))
		}
		orders = append(orders, &order)
	}

	return orders, totalCount, nil
}

func (repo iOrderRepositoryImpl) ExistsById(ctx context.Context, orderId uint64) (bool, repository.IRepoError) {
	singleResult := repo.mongoAdapter.FindOne(repo.database, repo.collection, bson.D{{"orderId", orderId}, {"deletedAt", nil}})
	if singleResult.Err() != nil {
		if repo.mongoAdapter.NoDocument(singleResult.Err()) {
			return false, nil
		}
		return false, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), "ExistsById Order Failed"))
	}
	return true, nil
}

func (repo iOrderRepositoryImpl) Count(ctx context.Context) (int64, repository.IRepoError) {
	total, err := repo.mongoAdapter.Count(repo.database, repo.collection, bson.D{{"deletedAt", nil}})
	if err != nil {
		return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Quantity Orders Failed"))
	}
	return total, nil
}

func (repo iOrderRepositoryImpl) CountWithFilter(ctx context.Context, supplier func() interface{}) (int64, repository.IRepoError) {
	total, err := repo.mongoAdapter.Count(repo.database, repo.collection, supplier())
	if err != nil {
		return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "CountWithFilter Orders Failed"))
	}
	return total, nil
}

func (repo iOrderRepositoryImpl) DeleteById(ctx context.Context, orderId uint64) (*entities.Order, repository.IRepoError) {
	var err repository.IRepoError
	order, err := repo.FindById(ctx, orderId)
	if err != nil {
		return nil, err
	}

	deletedAt := time.Now().UTC()
	order.DeletedAt = &deletedAt

	updateResult, e := repo.mongoAdapter.UpdateOne(repo.database, repo.collection,
		bson.D{{"orderId", order.OrderId}, {"deletedAt", nil}},
		bson.D{{"$set", order}})
	if e != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "UpdateOne Order Failed"))
	}

	if updateResult.ModifiedCount != 1 || updateResult.MatchedCount != 1 {
		return nil, repository.ErrorFactory(repository.NotFoundErr, "Order Not Found", errors.Wrap(e, "UpdateOne Order Failed"))
	}

	return order, nil
}

func (repo iOrderRepositoryImpl) Delete(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError) {
	return repo.DeleteById(ctx, order.OrderId)
}

func (repo iOrderRepositoryImpl) DeleteAllWithOrders(ctx context.Context, orders []entities.Order) repository.IRepoError {
	panic("implementation required")
}

// TODO cascade delete in packages and subpackages
func (repo iOrderRepositoryImpl) DeleteAll(ctx context.Context) repository.IRepoError {
	_, err := repo.mongoAdapter.UpdateMany(repo.database, repo.collection,
		bson.D{{"deletedAt", nil}},
		bson.M{"$set": bson.M{"deletedAt": time.Now().UTC()}})
	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "UpdateMany Order Failed"))
	}
	return nil
}

func (repo iOrderRepositoryImpl) RemoveById(ctx context.Context, orderId uint64) repository.IRepoError {
	result, err := repo.mongoAdapter.DeleteOne(repo.database, repo.collection, bson.M{"orderId": orderId})
	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "RemoveById Order Failed"))
	}

	if result.DeletedCount != 1 {
		return repository.ErrorFactory(repository.NotFoundErr, "Order Not Found", repository.ErrorRemoveFailed)
	}
	return nil
}

func (repo iOrderRepositoryImpl) Remove(ctx context.Context, order entities.Order) repository.IRepoError {
	return repo.RemoveById(ctx, order.OrderId)
}

func (repo iOrderRepositoryImpl) RemoveAllWithOrders(ctx context.Context, orders []entities.Order) repository.IRepoError {
	panic("implementation required")
}

func (repo iOrderRepositoryImpl) RemoveAll(ctx context.Context) repository.IRepoError {
	_, err := repo.mongoAdapter.DeleteMany(repo.database, repo.collection, bson.M{})
	if err != nil {
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "DeleteMany Order Failed"))
	}
	return nil
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		applog.GLog.Logger.Error("cursor.Close failed", "error", err)
	}
}
