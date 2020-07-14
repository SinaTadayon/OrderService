package finance_query_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/reports"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"time"
)

type IFinanceQR interface {
	FindAllWithPageAndSort(ctx context.Context, state string, startAt, endAt time.Time,
		page, perPage int64, fieldName string, direction int) ([]*reports.FinanceOrderItem, int64, repository.IRepoError)
}
