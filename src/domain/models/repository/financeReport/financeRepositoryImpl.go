package finance_repository

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/reports"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type iFinanceReportRepositoryImpl struct {
	mongoAdapter *mongoadapter.Mongo
	database     string
	collection   string
}

func NewFinanceReportRepository(mongoDriver *mongoadapter.Mongo, database, collection string) IFinanceReportRepository {
	return &iFinanceReportRepositoryImpl{mongoDriver, database, collection}
}

func (repo iFinanceReportRepositoryImpl) FindAllWithPageAndSort(ctx context.Context, state string, startTimestamp, endTimestamp time.Time,
	page, perPage int64, fieldName string, direction int) ([]*reports.FinanceOrderItem, int64, repository.IRepoError) {
	if page <= 0 || perPage <= 0 {
		return nil, 0, repository.ErrorFactory(repository.BadRequestErr, "Request Operation Failed", errors.New("neither offset nor start can be zero"))
	}

	var totalCount, err = repo.count(ctx, state, startTimestamp, endTimestamp)
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

	var pipeline []bson.M
	if fieldName != "" {
		pipeline = []bson.M{
			{"$match": bson.M{"packages.subpackages.updatedAt": bson.M{"$gte": startTimestamp, "$lte": endTimestamp}, "packages.subpackages.status": state, "packages.deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.subpackages.updatedAt": bson.M{"$gte": startTimestamp, "$lte": endTimestamp}, "packages.subpackages.status": state, "packages.deletedAt": nil}},
			{"$addFields": bson.M{
				"packages.subpackages.shipmentAmount":              "$packages.invoice.shipmentAmount",
				"packages.subpackages.rawSellerShippingNet":        "$packages.invoice.share.rawSellerShippingNet",
				"packages.subpackages.roundupSellerShippingNet":    "$packages.invoice.share.roundupSellerShippingNet",
				"packages.subpackages.orderCreatedAt":              "$createdAt",
				"packages.subpackages.items.invoice.sso.rate":      "$packages.invoice.sso.rate",
				"packages.subpackages.items.invoice.sso.isObliged": "$packages.invoice.sso.isObliged"},
			},
			{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
			{"$project": bson.M{
				"packages.subpackages.sid":                      1,
				"packages.subpackages.pid":                      1,
				"packages.subpackages.orderId":                  1,
				"packages.subpackages.items":                    1,
				"packages.subpackages.status":                   1,
				"packages.subpackages.createdAt":                1,
				"packages.subpackages.updatedAt":                1,
				"packages.subpackages.shipmentAmount":           1,
				"packages.subpackages.rawSellerShippingNet":     1,
				"packages.subpackages.roundupSellerShippingNet": 1,
				"packages.subpackages.orderCreatedAt":           1},
			},
			{"$replaceWith": "$packages.subpackages"},
			{"$sort": bson.M{fieldName: direction}},
			{"$skip": offset},
			{"$limit": perPage},
		}
	} else {
		pipeline = []bson.M{
			{"$match": bson.M{"packages.subpackages.updatedAt": bson.M{"$gte": startTimestamp, "$lte": endTimestamp}, "packages.subpackages.status": state, "packages.deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.subpackages.updatedAt": bson.M{"$gte": startTimestamp, "$lte": endTimestamp}, "packages.subpackages.status": state, "packages.deletedAt": nil}},
			{"$addFields": bson.M{
				"packages.subpackages.shipmentAmount":              "$packages.invoice.shipmentAmount",
				"packages.subpackages.rawSellerShippingNet":        bson.M{"$ifNull": bson.A{"$packages.invoice.share.rawSellerShippingNet", nil}},
				"packages.subpackages.roundupSellerShippingNet":    bson.M{"$ifNull": bson.A{"$packages.invoice.share.roundupSellerShippingNet", nil}},
				"packages.subpackages.orderCreatedAt":              "$createdAt",
				"packages.subpackages.items.invoice.sso.rate":      "$packages.invoice.sso.rate",
				"packages.subpackages.items.invoice.sso.isObliged": "$packages.invoice.sso.isObliged"},
			},
			{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
			{"$project": bson.M{
				"packages.subpackages.sid":                      1,
				"packages.subpackages.pid":                      1,
				"packages.subpackages.orderId":                  1,
				"packages.subpackages.items":                    1,
				"packages.subpackages.status":                   1,
				"packages.subpackages.createdAt":                1,
				"packages.subpackages.updatedAt":                1,
				"packages.subpackages.shipmentAmount":           1,
				"packages.subpackages.rawSellerShippingNet":     1,
				"packages.subpackages.roundupSellerShippingNet": 1,
				"packages.subpackages.orderCreatedAt":           1},
			},
			{"$replaceWith": "$packages.subpackages"},
			{"$skip": offset},
			{"$limit": perPage},
		}
	}

	cursor, e := repo.mongoAdapter.Aggregate(repo.database, repo.collection, pipeline)
	if e != nil {
		return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(e, "Aggregate Failed"))
	}

	defer closeCursor(ctx, cursor)

	finances := make([]*reports.FinanceOrderItem, 0, perPage)

	for cursor.Next(ctx) {
		var finance reports.FinanceOrderItem
		if err := cursor.Decode(&finance); err != nil {
			return nil, 0, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "cursor.Decode failed"))
		}

		finances = append(finances, &finance)
	}

	return finances, totalCount, nil
}

func (repo iFinanceReportRepositoryImpl) count(ctx context.Context, state string, startTimestamp, endTimestamp time.Time) (int64, repository.IRepoError) {
	var total struct {
		Count int
	}

	pipeline := []bson.M{
		{"$match": bson.M{"packages.subpackages.updatedAt": bson.M{"$gte": startTimestamp.UTC(), "$lte": endTimestamp.UTC()}, "packages.subpackages.status": state, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$unwind": "$packages.subpackages"},
		{"$match": bson.M{"packages.subpackages.updatedAt": bson.M{"$gte": startTimestamp.UTC(), "$lte": endTimestamp.UTC()}, "packages.subpackages.status": state, "packages.deletedAt": nil}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
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

func closeCursor(context context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(context)
	if err != nil {
		applog.GLog.Logger.Error("cursor.Close failed", "error", err)
	}
}
