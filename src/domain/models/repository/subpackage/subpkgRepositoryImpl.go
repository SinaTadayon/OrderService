package subpkg_repository

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
	"strconv"
	"time"
)

const (
	defaultDocCount int = 1024
)

type iSubPkgRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
	database     string
	collection   string
}

func NewSubPkgRepository(mongoDriver *mongoadapter.Mongo, database, collection string) ISubpackageRepository {
	return &iSubPkgRepositoryImpl{mongoDriver, database, collection}
}

func (repo iSubPkgRepositoryImpl) findAndUpdate(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError {
	subPkg.UpdatedAt = time.Now().UTC()
	currentVersion := subPkg.Version
	subPkg.Version += 1
	opt := options.FindOneAndUpdate()
	opt.SetUpsert(false)
	opt.SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"package.pid": subPkg.PId},
			bson.M{"subpackage.sid": subPkg.SId, "subpackage.version": currentVersion},
		},
	}).SetReturnDocument(options.After)

	singleResult := repo.mongoAdapter.GetConn().Database(repo.database).Collection(repo.collection).FindOneAndUpdate(ctx,
		bson.D{{"orderId", subPkg.OrderId}},
		bson.D{{"$set", bson.D{{"packages.$[package].subpackages.$[subpackage]", subPkg}}}}, opt)

	if singleResult.Err() != nil {
		if repo.mongoAdapter.NoDocument(singleResult.Err()) {
			return repository.ErrorFactory(repository.NotFoundErr, "Package Not Found", repository.ErrorUpdateFailed)
		}
		return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(singleResult.Err(), ""))
	}
	return nil
}

func (repo iSubPkgRepositoryImpl) Save(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError {
	if subPkg.SId == 0 {
		var err error
		var updateResult *mongo.UpdateResult
		subPkg.CreatedAt = time.Now().UTC()
		subPkg.UpdatedAt = time.Now().UTC()

		for {
			random := strconv.Itoa(int(entities.GenerateRandomNumber()))
			sid, _ := strconv.Atoi(strconv.Itoa(int(subPkg.OrderId)) + random)
			subPkg.SId = uint64(sid)
			updateResult, err = repo.mongoAdapter.UpdateOne(repo.database, repo.collection, bson.D{
				{"orderId", subPkg.OrderId},
				{"deletedAt", nil},
				{"packages.pid", subPkg.PId}},
				bson.D{{"$push", bson.D{{"packages.$.subpackages", subPkg}}}})

			if err != nil {
				if repo.mongoAdapter.IsDupError(err) {
					continue
				}
				return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Save Subpackage Failed"))
			}

			break
		}

		if updateResult.ModifiedCount != 1 || updateResult.MatchedCount != 1 {
			return repository.ErrorFactory(repository.NotFoundErr, "Subpackage Not Found", repository.ErrorUpdateFailed)
		}
	} else {
		updateResult, err := repo.mongoAdapter.UpdateOne(repo.database, repo.collection, bson.D{
			{"orderId", subPkg.OrderId},
			{"deletedAt", nil},
			{"packages.pid", subPkg.PId}},
			bson.D{{"$push", bson.D{{"packages.$.subpackages", subPkg}}}})

		if err != nil {
			return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "UpdateOne Subpackage Failed"))
		}

		if updateResult.ModifiedCount != 1 {
			return repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", repository.ErrorUpdateFailed)
		}
	}
	return nil
}

func (repo iSubPkgRepositoryImpl) SaveAll(ctx context.Context, subPkgList []*entities.Subpackage) repository.IRepoError {
	panic("must be implement")
}

func (repo iSubPkgRepositoryImpl) Update(ctx context.Context, subPkg entities.Subpackage) (*entities.Subpackage, repository.IRepoError) {
	subPkg.UpdatedAt = time.Now().UTC()
	err := repo.findAndUpdate(ctx, &subPkg)
	if err != nil {
		return nil, err
	}

	return &subPkg, nil
}

func (repo iSubPkgRepositoryImpl) UpdateAll(ctx context.Context, subPkgList []entities.Subpackage) ([]*entities.Subpackage, repository.IRepoError) {
	panic("must be implement")
}

func (repo iSubPkgRepositoryImpl) FindByItemId(ctx context.Context, sid uint64) (*entities.Subpackage, repository.IRepoError) {
	panic("must be implement")
}

func (repo iSubPkgRepositoryImpl) FindByOrderAndItemId(ctx context.Context, orderId, sid uint64) (*entities.Subpackage, repository.IRepoError) {
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

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if err != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}
	}

	return &subpackage, nil
}

func (repo iSubPkgRepositoryImpl) FindByOrderAndSellerId(ctx context.Context, orderId, pid uint64) ([]*entities.Subpackage, repository.IRepoError) {

	pipeline := []bson.M{
		{"$match": bson.M{"orderId": orderId, "deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": pid}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if err != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, 16)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil
}

func (repo iSubPkgRepositoryImpl) FindAll(ctx context.Context, pid uint64) ([]*entities.Subpackage, repository.IRepoError) {
	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if err != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, 1024)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil
}

func (repo iSubPkgRepositoryImpl) FindAllWithSort(ctx context.Context, pid uint64, fieldName string, direction int) ([]*entities.Subpackage, repository.IRepoError) {
	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$sort": bson.M{"packages.subpackages." + fieldName: direction}},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if err != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, 1024)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil
}

func (repo iSubPkgRepositoryImpl) FindAllWithPage(ctx context.Context, pid uint64, page, perPage int64) ([]*entities.Subpackage, int64, repository.IRepoError) {

	if page <= 0 || perPage <= 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("neither offset nor start can be zero"))
	}

	var totalCount, err = repo.Count(ctx, pid)
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

	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$skip": offset},
		{"$limit": perPage},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, e := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if e != nil {
		return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, perPage)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, totalCount, nil
}

func (repo iSubPkgRepositoryImpl) FindAllWithPageAndSort(ctx context.Context, pid uint64, page, perPage int64, fieldName string, direction int) ([]*entities.Subpackage, int64, repository.IRepoError) {
	if page <= 0 || perPage <= 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("neither offset nor start can be zero"))
	}

	var totalCount, err = repo.Count(ctx, pid)
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

	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$sort": bson.M{"packages.subpackages." + fieldName: direction}},
		{"$skip": offset},
		{"$limit": perPage},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	cursor, e := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if e != nil {
		return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	subpackages := make([]*entities.Subpackage, 0, perPage)

	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}

		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, totalCount, nil

}

func (repo iSubPkgRepositoryImpl) FindByFilter(ctx context.Context, totalSupplier, supplier func() (filter interface{})) ([]*entities.Subpackage, repository.IRepoError) {
	filter := supplier()
	total, err := repo.CountWithFilter(ctx, totalSupplier)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
	}

	cursor, e := repo.mongoAdapter.Aggregate(repo.database, repo.collection, filter)
	if e != nil {
		return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "Aggregate Subpackage Failed"))
	}

	defer closeCursor(ctx, cursor)
	subpackages := make([]*entities.Subpackage, 0, total)

	// iterate through all documents
	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		// decode the document
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Subpackage Failed"))
		}
		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, nil

}

func (repo iSubPkgRepositoryImpl) FindByFilterWithPage(ctx context.Context, totalSupplier, supplier func() (filter interface{}), page, perPage int64) ([]*entities.Subpackage, int64, repository.IRepoError) {
	if page <= 0 || perPage <= 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("neither offset nor start can be zero"))
	}

	filter := supplier()
	var totalCount, err = repo.CountWithFilter(ctx, totalSupplier)
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

	cursor, e := repo.mongoAdapter.Aggregate(repo.database, repo.collection, filter)
	if e != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "Aggregate Subpackages Failed"))
	} else if cursor.Err() != nil {
		return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "Aggregate Subpackages Failed"))
	}

	defer closeCursor(ctx, cursor)
	subpackages := make([]*entities.Subpackage, 0, perPage)

	// iterate through all documents
	for cursor.Next(ctx) {
		var subpackage entities.Subpackage
		// decode the document
		if err := cursor.Decode(&subpackage); err != nil {
			return nil, totalCount, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Subpackage Failed"))
		}
		subpackages = append(subpackages, &subpackage)
	}

	return subpackages, totalCount, nil
}

func (repo iSubPkgRepositoryImpl) ExistsById(ctx context.Context, sid uint64) (bool, repository.IRepoError) {
	singleResult := repo.mongoAdapter.FindOne(repo.database, repo.collection, bson.D{{"packages.subpackages.sid", sid}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		if repo.mongoAdapter.NoDocument(err) {
			return false, nil
		}
		return false, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "ExistsById failed"))
	}
	return true, nil
}

func (repo iSubPkgRepositoryImpl) Count(ctx context.Context, pid uint64) (int64, repository.IRepoError) {
	var total struct {
		Count int
	}

	pipeline := []bson.M{
		{"$match": bson.M{"packages.pid": pid, "packages.deletedAt": nil}},
		{"$project": bson.M{"subSize": bson.M{"$size": "$packages.subpackages"}}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
		{"$project": bson.M{"_id": 0, "count": 1}},
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if err != nil {
		return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&total); err != nil {
			return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}
	}

	return int64(total.Count), nil
}

func (repo iSubPkgRepositoryImpl) CountWithFilter(ctx context.Context, supplier func() (filter interface{})) (int64, repository.IRepoError) {
	var total struct {
		Count int
	}

	cursor, err := repo.mongoAdapter.Aggregate(repo.database, repo.collection, supplier())
	if err != nil {
		return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&total); err != nil {
			return 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}
	}

	return int64(total.Count), nil
}

func (repo iSubPkgRepositoryImpl) GenerateUniqSid(ctx context.Context, oid uint64) (uint64, repository.IRepoError) {

	for {
		random := strconv.Itoa(int(entities.GenerateRandomNumber()))
		sid, _ := strconv.Atoi(strconv.Itoa(int(oid)) + random)
		if result, err := repo.ExistsById(ctx, uint64(sid)); err != nil {
			return 0, err
		} else {
			if !result {
				return uint64(sid), nil
			}
		}
	}

}

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		applog.GLog.Logger.Error("cursor.Close failed", "error", err)
	}
}
