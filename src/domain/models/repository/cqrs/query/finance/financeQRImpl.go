package finance_query_repository

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/reports"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	finance_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/financeReport"
	"time"
)

type iFinanceQRImpl struct {
	financeRepo finance_repository.IFinanceReportRepository
}

func FinanceQRFactory(command *mongoadapter.Mongo, database, collection string) IFinanceQR {
	financeRepository := finance_repository.NewFinanceReportRepository(command, database, collection)

	return &iFinanceQRImpl{
		financeRepo: financeRepository,
	}
}

func (financeQR iFinanceQRImpl) FindAllWithPageAndSort(ctx context.Context, state string, startAt, endAt time.Time,
	page, perPage int64, fieldName string, direction int) ([]*reports.FinanceOrderItem, int64, repository.IRepoError) {
	return financeQR.financeRepo.FindAllWithPageAndSort(ctx, state, startAt, endAt, page, perPage, fieldName, direction)
}
