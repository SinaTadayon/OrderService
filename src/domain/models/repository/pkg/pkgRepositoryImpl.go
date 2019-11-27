package pkg_repository

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

type iPkgItemRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
}

func NewPkgItemRepository(mongoDriver *mongoadapter.Mongo) (IPkgItemRepository, error) {
	return &iPkgItemRepositoryImpl{mongoDriver}, nil
}

func (repo iPkgItemRepositoryImpl) findAndUpdate(ctx context.Context, pkgItem *entities.PackageItem) (*entities.PackageItem, error) {
	pkgItem.UpdatedAt = time.Now().UTC()
	currentVersion := pkgItem.Version
	pkgItem.Version += 1
	opt := options.FindOneAndUpdate()
	opt.SetUpsert(false)
	singleResult := repo.mongoAdapter.GetConn().Database(databaseName).Collection(collectionName).FindOneAndUpdate(ctx,
		bson.D{
			{"orderId", pkgItem.OrderId},
			{"packages", bson.D{
				{"$elemMatch", bson.D{
					{"sellerId", pkgItem.SellerId},
					{"version", currentVersion},
				}},
			}},
		},
		bson.D{{"$set", bson.D{{"packages.$", pkgItem}}}}, opt)
	//singleResult := repo.mongoAdapter.GetConn().Database(databaseName).Collection(collectionName).FindOneAndUpdate(ctx,
	//	bson.M{"$and": []bson.M{ // you can try this in []interface
	//		bson.M{"packages.id": pkgItem.Id},
	//		bson.M{"packages.deletedAt": nil},
	//		bson.M{"packages.version": currentVersion}}},
	//	bson.D{{"$set", bson.D{{"packages.1", pkgItem}}}}, opt)
	if singleResult.Err() != nil {
		return nil, errors.Wrap(singleResult.Err(), "findAndUpdate failed")
	}
	//{"$inc", bson.D{{"packages.$.version", 1}
	//var updatedPkgItem entities.PackageItem
	//if err := singleResult.Decode(&updatedPkgItem); err != nil {
	//	return nil, errors.Wrap(singleResult.Err(), "findAndUpdate pkgItem failed")
	//}

	return pkgItem, nil
}

func (repo iPkgItemRepositoryImpl) Update(ctx context.Context, pkgItem entities.PackageItem) (*entities.PackageItem, error) {

	pkgItem.UpdatedAt = time.Now().UTC()
	var updatedPkgItem, err = repo.findAndUpdate(ctx, &pkgItem)
	if err != nil {
		return nil, err
	}

	return updatedPkgItem, nil
}

func (repo iPkgItemRepositoryImpl) FindById(ctx context.Context, orderId uint64, id uint64) (*entities.PackageItem, error) {

	var PkgItem entities.PackageItem
	pipeline := []bson.M{
		{"$match": bson.M{"orderId": orderId, "deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.sellerId": id}},
		{"$project": bson.M{"_id": 0, "packages": 1}},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "Aggregate failed")
	}

	defer closeCursor(ctx, cursor)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&PkgItem); err != nil {
			return nil, errors.Wrap(err, "cursor.Decode failed")
		}
	}

	return &PkgItem, nil
}

func (repo iPkgItemRepositoryImpl) FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.PackageItem, error) {
	filter := supplier()

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, filter)
	if err != nil {
		return nil, errors.Wrap(err, "Aggregate failed")
	}

	defer closeCursor(ctx, cursor)
	pkgItems := make([]*entities.PackageItem, 0, defaultDocCount)

	// iterate through all documents
	for cursor.Next(ctx) {
		var packageItem entities.PackageItem
		// decode the document
		if err := cursor.Decode(&packageItem); err != nil {
			return nil, errors.Wrap(err, "cursor.Decode failed")
		}
		pkgItems = append(pkgItems, &packageItem)
	}

	return pkgItems, nil
}

func (repo iPkgItemRepositoryImpl) ExistsById(ctx context.Context, orderId uint64, id uint64) (bool, error) {
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId}, {"packages.sellerId", id}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		if repo.mongoAdapter.NoDocument(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "ExistsById failed")
	}
	return true, nil
}

func (repo iPkgItemRepositoryImpl) Count(ctx context.Context, id uint64) (int64, error) {
	total, err := repo.mongoAdapter.Count(databaseName, collectionName, bson.D{{"packages.sellerId", id},
		{"deletedAt", nil}})
	if err != nil {
		return 0, errors.Wrap(err, "Count failed")
	}
	return total, nil

}

func (repo iPkgItemRepositoryImpl) CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, error) {
	total, err := repo.mongoAdapter.Count(databaseName, collectionName, supplier())
	if err != nil {
		return 0, errors.Wrap(err, "CountWithFilter failed")
	}
	return total, nil
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		logger.Err("closeCursor() failed, error: %s", err)
	}
}
