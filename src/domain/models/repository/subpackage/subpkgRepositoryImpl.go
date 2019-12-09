package subpackage

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
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
var ErrorRemoveFailed = errors.New("remove subpackage failed")
var ErrorUpdateFailed = errors.New("update subpackage failed")
var ErrorVersionUpdateFailed = errors.New("update subpackage version failed")

type iSubPkgRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
}

func NewSubPkgRepository(mongoDriver *mongoadapter.Mongo) ISubpackageRepository {
	return &iSubPkgRepositoryImpl{mongoDriver}
}

func (repo iSubPkgRepositoryImpl) findAndUpdate(ctx context.Context, subPkg *entities.Subpackage) error {
	subPkg.UpdatedAt = time.Now().UTC()
	currentVersion := subPkg.Version
	subPkg.Version += 1
	opt := options.FindOneAndUpdate()
	opt.SetUpsert(false)
	opt.SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"package.sellerId": subPkg.SellerId},
			bson.M{"subpackage.sid": subPkg.SId, "subpackage.version": currentVersion},
		},
	}).SetReturnDocument(options.After)

	singleResult := repo.mongoAdapter.GetConn().Database(databaseName).Collection(collectionName).FindOneAndUpdate(ctx,
		bson.D{{"orderId", subPkg.OrderId}},
		bson.D{{"$set", bson.D{{"packages.$[package].subpackages.$[subpackage]", subPkg}}}}, opt)

	if singleResult.Err() != nil {
		return errors.Wrap(singleResult.Err(), "findAndUpdate failed")
	}

	return nil
}

func (repo iSubPkgRepositoryImpl) Save(ctx context.Context, subPkg *entities.Subpackage) error {
	if subPkg.SId == 0 {
		var err error
		var updateResult *mongo.UpdateResult
		subPkg.CreatedAt = time.Now().UTC()
		subPkg.UpdatedAt = time.Now().UTC()

		for {
			random := strconv.Itoa(int(entities.GenerateRandomNumber()))
			sid, _ := strconv.Atoi(strconv.Itoa(int(subPkg.OrderId)) + random)
			subPkg.SId = uint64(sid)
			updateResult, err = repo.mongoAdapter.UpdateOne(databaseName, collectionName, bson.D{
				{"orderId", subPkg.OrderId},
				{"deletedAt", nil},
				{"packages.pid", subPkg.SellerId}},
				bson.D{{"$push", bson.D{{"packages.$.subpackages", subPkg}}}})

			if err != nil {
				if repo.mongoAdapter.IsDupError(err) {
					continue
				}
				return errors.Wrap(err, "Save Subpackage Failed")
			}

			break
		}

		if updateResult.ModifiedCount != 1 {
			return ErrorUpdateFailed
		}
	} else {
		updateResult, err := repo.mongoAdapter.UpdateOne(databaseName, collectionName, bson.D{
			{"orderId", subPkg.OrderId},
			{"deletedAt", nil},
			{"packages.pid", subPkg.SellerId}},
			bson.D{{"$push", bson.D{{"packages.$.subpackages", subPkg}}}})

		if err != nil {
			return errors.Wrap(err, "Save Subpackage Failed")
		}

		if updateResult.ModifiedCount != 1 {
			return ErrorUpdateFailed
		}
	}
	return nil
}

func (repo iSubPkgRepositoryImpl) SaveAll(ctx context.Context, subPkgList []*entities.Subpackage) error {
	//for _, subPkg := range subPkgList {
	//	if err := repo.Save(ctx, subPkg); err != nil {
	//		return err
	//	}
	//}
	//
	//return nil
	panic("must be implement")
}

func (repo iSubPkgRepositoryImpl) Update(ctx context.Context, subPkg entities.Subpackage) (*entities.Subpackage, error) {
	subPkg.UpdatedAt = time.Now().UTC()
	err := repo.findAndUpdate(ctx, &subPkg)
	if err != nil {
		return nil, err
	}

	return &subPkg, nil
}

func (repo iSubPkgRepositoryImpl) UpdateAll(ctx context.Context, subPkgList []entities.Subpackage) ([]*entities.Subpackage, error) {
	panic("must be implement")
}

func (repo iSubPkgRepositoryImpl) FindByItemId(ctx context.Context, sid uint64) (*entities.Subpackage, error) {
	panic("must be implement")
}

func (repo iSubPkgRepositoryImpl) FindByOrderAndItemId(ctx context.Context, orderId, sid uint64) (*entities.Subpackage, error) {
	var subpackage entities.Subpackage
	pipeline := []bson.M{
		{"$match": bson.M{"orderId": orderId, "deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$unwind": "$packages.subpackages"},
		{"$match": bson.M{"packages.subpackages.sid": sid}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, errors.Wrap(err, "cursor.Decode failed")
		}
	}

	return &subpackage, nil

}

func (repo iSubPkgRepositoryImpl) FindByOrderAndSellerId(ctx context.Context, orderId, sellerId uint64) ([]*entities.Subpackage, error) {

	pipeline := []bson.M{
		{"$match": bson.M{"orderId": orderId, "deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": sellerId}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, 16)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, errors.Wrap(err, "cursor.Decode failed")
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil
}

func (repo iSubPkgRepositoryImpl) FindAll(ctx context.Context, sellerId uint64) ([]*entities.Subpackage, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, 1024)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, errors.Wrap(err, "cursor.Decode failed")
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil
}

func (repo iSubPkgRepositoryImpl) FindAllWithSort(ctx context.Context, sellerId uint64, fieldName string, direction int) ([]*entities.Subpackage, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$sort": bson.M{"packages.subpackages." + fieldName: direction}},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, 1024)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, errors.Wrap(err, "cursor.Decode failed")
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil
}

func (repo iSubPkgRepositoryImpl) FindAllWithPage(ctx context.Context, sellerId uint64, page, perPage int64) ([]*entities.Subpackage, int64, error) {

	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = repo.Count(ctx, sellerId)
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindAllWithPage Subpackage Failed")
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

	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$skip": offset},
		{"$limit": perPage},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, perPage)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, 0, errors.Wrap(err, "cursor.Decode failed")
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, availablePages, nil
}

func (repo iSubPkgRepositoryImpl) FindAllWithPageAndSort(ctx context.Context, sellerId uint64, page, perPage int64, fieldName string, direction int) ([]*entities.Subpackage, int64, error) {
	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	var totalCount, err = repo.Count(ctx, sellerId)
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindAllWithPageAndSort Subpackage Failed")
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

	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$sort": bson.M{"packages.subpackages." + fieldName: direction}},
		{"$skip": offset},
		{"$limit": perPage},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, perPage)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, 0, errors.Wrap(err, "cursor.Decode failed")
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, availablePages, nil

}

func (repo iSubPkgRepositoryImpl) FindByFilter(ctx context.Context, totalSupplier, supplier func() (filter interface{})) ([]*entities.Subpackage, error) {
	filter := supplier()
	total, err := repo.CountWithFilter(ctx, totalSupplier)
	if err != nil {
		logger.Err("repo.Count() failed, %s", err)
		total = int64(defaultDocCount)
	}

	if total == 0 {
		return nil, nil
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, filter)
	if err != nil {
		return nil, errors.Wrap(err, "FindByFilter Subpackage Failed")
	}

	defer closeCursor(ctx, cursor)
	subpackages := make([]*entities.Subpackage, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		// decode the document
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, errors.Wrap(err, "FindByFilter Subpackage Failed")
		}
		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil

}

func (repo iSubPkgRepositoryImpl) FindByFilterWithPage(ctx context.Context, totalSupplier, supplier func() (filter interface{}), page, perPage int64) ([]*entities.Subpackage, int64, error) {
	if page < 0 || perPage == 0 {
		return nil, 0, errors.New("neither offset nor start can be zero")
	}

	filter := supplier()
	var totalCount, err = repo.CountWithFilter(ctx, totalSupplier)
	if err != nil {
		return nil, 0, errors.Wrap(err, "FindByFilterWithPage Subpackages Failed")
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

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, filter)
	if err != nil {
		return nil, availablePages, errors.Wrap(err, "FindByFilterWithPage Subpackages Failed")
	} else if cursor.Err() != nil {
		return nil, availablePages, errors.Wrap(err, "FindByFilterWithPage Subpackages Failed")
	}

	defer closeCursor(ctx, cursor)
	subpackages := make([]*entities.Subpackage, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		// decode the document
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, availablePages, errors.Wrap(err, "FindByFilter Subpackage Failed")
		}
		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, availablePages, nil
}

func (repo iSubPkgRepositoryImpl) ExistsById(ctx context.Context, sid uint64) (bool, error) {
	singleResult := repo.mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"packages.subpackages.sid", sid}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		if repo.mongoAdapter.NoDocument(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "ExistsById failed")
	}
	return true, nil
}

func (repo iSubPkgRepositoryImpl) Count(ctx context.Context, sellerId uint64) (int64, error) {
	var total struct {
		Count int
	}

	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": sellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"subSize": bson.M{"$size": "$packages.subpackages"}}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
		{"$project": bson.M{"_id": 0, "count": 1}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, pipeline)
	if err != nil {
		return 0, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&total); err != nil {
			return 0, errors.Wrap(err, "cursor.Decode failed")
		}
	}

	return int64(total.Count), nil
}

func (repo iSubPkgRepositoryImpl) CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, error) {
	var total struct {
		Count int
	}

	cursor, err := repo.mongoAdapter.Aggregate(databaseName, collectionName, supplier())
	if err != nil {
		return 0, errors.Wrap(err, "Aggregate Failed")
	}

	defer closeCursor(ctx, cursor)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&total); err != nil {
			return 0, errors.Wrap(err, "cursor.Decode failed")
		}
	}

	return int64(total.Count), nil
}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		logger.Err("closeCursor() failed, error: %s", err)
	}
}
