package finance_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/reports"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"time"
)

type IFinanceReportRepository interface {
	FindAllWithPageAndSort(ctx context.Context, state string, startTimestamp, endTimestamp time.Time,
		page, perPage int64, fieldName string, direction int) ([]*reports.FinanceOrderItem, int64, repository.IRepoError)
}
