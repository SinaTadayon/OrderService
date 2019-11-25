package item_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	databaseName    string = "orderService"
	collectionName  string = "orders"
	defaultDocCount int    = 1024
)

var errorTotalCountExceeded = errors.New("total count exceeded")
var errorPageNotAvailable = errors.New("page not available")
var errorDeleteFailed = errors.New("update deletedAt field failed")
var errorRemoveFailed = errors.New("remove order failed")
var errorUpdateFailed = errors.New("update order failed")

type iItemRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
}

func NewItemRepository(mongoDriver *mongoadapter.Mongo) (IItemRepository, error) {

	_, err := mongoDriver.AddUniqueIndex(databaseName, collectionName, "items.itemId")
	if err != nil {
		logger.Err(err.Error())
		return nil, err
	}

	return &iItemRepositoryImpl{mongoDriver}, nil
}

func (repo iItemRepositoryImpl) Update(item entities.Item) (*entities.Item, error) {

	item.UpdatedAt = time.Now().UTC()
	var updateResult, err = repo.mongoAdapter.UpdateOne(databaseName, collectionName, bson.D{{"items.itemId", item.ItemId}, {"items.deletedAt", nil}},
		bson.D{{"$set", item}})
	if err != nil {
		return nil, err
	}

	if updateResult.ModifiedCount != 1 {
		return nil, errorUpdateFailed
	}

	return &item, nil
}

func (repo iItemRepositoryImpl) UpdateAll(items []entities.Item) ([]*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) Insert(item entities.Item) (*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) InsertAll(items []entities.Item) ([]*entities.Item, error) {
	panic("implementation required")
}

// TODO bug in fetch data
func (repo iItemRepositoryImpl) FindAll() ([]*entities.Item, error) {
	total, err := repo.Count()

	if err != nil {
		logger.Audit("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	projection := bson.D{
		{"items", 1},
	}

	cursor, err := repo.mongoAdapter.FindMany(databaseName, collectionName,
		bson.D{{"items.deletedAt", nil}}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)
	items := make([]*entities.Item, 0, total)

	//iterate through all documents
	for cursor.Next(ctx) {
		var item entities.Item
		// decode the document
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}

func (repo iItemRepositoryImpl) FindAllWithSort(fieldName string, direction int) ([]*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindAllWithPage(page, perPage int64) ([]*entities.Item, int64, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindAllWithPageAndSort(page, perPage int64, fieldName string, direction int) ([]*entities.Item, int64, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindAllById(ids ...string) ([]*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindById(orderId string) (*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindByFilter(supplier func() interface{}) ([]*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindByFilterWithSort(supplier func() (interface{}, string, int)) ([]*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindByFilterWithPage(supplier func() interface{}, page, perPage int64) ([]*entities.Item, int64, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) FindByFilterWithPageAndSort(supplier func() (interface{}, string, int), page, perPage int64) ([]*entities.Item, int64, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) ExistsById(itemId uint64) (bool, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) Count() (int64, error) {
	var result struct {
		Count int64
	}

	pipeline := []bson.M{{"$match": bson.M{"items.deletedAt": nil}},
		{"$unwind": "$items"},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
		{"$project": bson.M{"count": 1, "_id": 0}}}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	defer closeCursor(ctx, cursor)

	// iterate through all documents
	for cursor.Next(ctx) {

		// decode the document
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
	}
	return result.Count, nil
}

func (repo iItemRepositoryImpl) CountWithFilter(supplier func() interface{}) (int64, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) DeleteById(itemId uint64) (*entities.Item, error) {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) Delete(item entities.Item) (*entities.Item, error) {
	return repo.DeleteById(item.ItemId)
}

func (repo iItemRepositoryImpl) DeleteAllWithItems([]entities.Item) error {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) DeleteAll() error {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) RemoveById(itemId uint64) error {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) Remove(item entities.Item) error {
	return repo.RemoveById(item.ItemId)
}

func (repo iItemRepositoryImpl) RemoveAllWithItems([]entities.Item) error {
	panic("implementation required")
}

func (repo iItemRepositoryImpl) RemoveAll() error {
	panic("implementation required")
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		logger.Err("closeCursor() failed, err: %s", err)
	}
}
