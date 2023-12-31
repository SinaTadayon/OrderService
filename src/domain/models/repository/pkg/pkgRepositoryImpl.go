package pkg_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	defaultDocCount int = 1024
)

type iPkgItemRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
	database     string
	collection   string
}

func NewPkgItemRepository(mongoDriver *mongoadapter.Mongo, database, collection string) IPkgItemRepository {
	return &iPkgItemRepositoryImpl{mongoDriver, database, collection}
}

func (repo iPkgItemRepositoryImpl) findAndUpdate(ctx context.Context, pkgItem *entities.PackageItem, upsert bool) (*entities.PackageItem, repository.IRepoError) {
	pkgItem.UpdatedAt = time.Now().UTC()
	currentVersion := pkgItem.Version
	pkgItem.Version += 1
	opt := options.FindOneAndUpdate()
	opt.SetUpsert(upsert)
	singleResult := repo.mongoAdapter.GetConn().Database(repo.database).Collection(repo.collection).FindOneAndUpdate(ctx,
		bson.D{
			{"orderId", pkgItem.OrderId},
			{"packages", bson.D{
				{"$elemMatch", bson.D{
					{"pid", pkgItem.PId},
					{"version", currentVersion},
				}},
			}},
		},
		bson.D{{"$set", bson.D{{"packages.$", pkgItem}}}}, opt)
	if singleResult.Err() != nil {
		if repo.mongoAdapter.NoDocument(singleResult.Err()) {
			return nil, repository.ErrorFactory(repository.NotFoundErr, "Package Not Found", repository.ErrorUpdateFailed)
		}
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), ""))
	}

	return pkgItem, nil
}

func (repo iPkgItemRepositoryImpl) Update(ctx context.Context, pkgItem entities.PackageItem, upsert bool) (*entities.PackageItem, repository.IRepoError) {

	pkgItem.UpdatedAt = time.Now().UTC()
	updatedPkgItem, err := repo.findAndUpdate(ctx, &pkgItem, upsert)
	if err != nil {
		return nil, err
	}

	return updatedPkgItem, nil
}

func (repo iPkgItemRepositoryImpl) FindPkgItmBuyinfById(ctx context.Context, orderId uint64, id uint64) (*entities.PackageItem, uint64, repository.IRepoError) {
	var temp struct {
		PkgItm []entities.PackageItem `bson:"pkgItm"`
		BuyId  []struct {
			BuyerId uint64 `bson:"buyerId"`
		} `bson:"buyId"`
	}

	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"pkgItm": []bson.M{ // find package item
					{"$match": bson.M{"orderId": orderId, "deletedAt": nil}},
					{"$unwind": "$packages"},
					{"$match": bson.M{"packages.pid": id}},
					{"$project": bson.M{"_id": 0, "packages": 1}},
					{"$replaceRoot": bson.M{"newRoot": "$packages"}},
				},
				"buyId": []bson.M{ // find buyerInfo
					{"$match": bson.M{"orderId": orderId, "deletedAt": nil}},
					{"$project": bson.M{
						"buyerId": "$buyerInfo.buyerId",
					}},
					{"$project": bson.M{
						"_id": 0,
					}},
				},
			},
		},
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if err != nil {
		return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&temp); err != nil {
			return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}
	}

	return &temp.PkgItm[0], temp.BuyId[0].BuyerId, nil
}

func (repo iPkgItemRepositoryImpl) FindById(ctx context.Context, orderId uint64, id uint64) (*entities.PackageItem, repository.IRepoError) {

	var PkgItem entities.PackageItem
	pipeline := []bson.M{
		{"$match": bson.M{"orderId": orderId, "deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": id}},
		{"$project": bson.M{"_id": 0, "packages": 1}},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if err != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&PkgItem); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}
	}

	if PkgItem.OrderId == 0 || PkgItem.PId == 0 {
		return nil, repository.ErrorFactory(repository.NotFoundErr, "Package Not Found", errors.New("Package Not Found"))
	}

	return &PkgItem, nil
}

func (repo iPkgItemRepositoryImpl) FindByFilter(ctx context.Context, supplier func() (filter interface{})) ([]*entities.PackageItem, repository.IRepoError) {
	filter := supplier()

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, filter)
	if err != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)
	pkgItems := make([]*entities.PackageItem, 0, defaultDocCount)

	// iterate through all documents
	for cursor.Next(ctx) {
		var packageItem entities.PackageItem
		// decode the document
		if err := cursor.Decode(&packageItem); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}
		pkgItems = append(pkgItems, &packageItem)
	}

	return pkgItems, nil
}

func (repo iPkgItemRepositoryImpl) ExistsById(ctx context.Context, orderId uint64, id uint64) (bool, repository.IRepoError) {
	singleResult := repo.mongoAdapter.FindOne(repo.database, repo.collection, bson.D{{"orderId", orderId}, {"packages.pid", id}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		if repo.mongoAdapter.NoDocument(err) {
			return false, nil
		}
		return false, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, ""))
	}
	return true, nil
}

func (repo iPkgItemRepositoryImpl) Count(ctx context.Context, id uint64) (int64, repository.IRepoError) {
	total, err := repo.mongoAdapter.Count(repo.database, repo.collection, bson.D{{"packages.pid", id},
		{"deletedAt", nil}})
	if err != nil {
		return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, ""))
	}
	return total, nil
}

func (repo iPkgItemRepositoryImpl) CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError) {

	var total struct {
		Count int
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, supplier())
	if err != nil {
		return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate failed"))
	}

	defer closeCursor(ctx, cursor)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&total); err != nil {
			return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}
	}

	return int64(total.Count), nil
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		applog.GLog.Logger.Error("cursor.Close failed", "error", err)
	}
}
