package query_repository

import (
	finance_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/finance"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/order"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/pkg"
	subpkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query/subpkg"
)

type IQueryRepository interface {
	OrderQR() order_repository.IOrderQR

	PkgQR() pkg_repository.IPkgQR

	SubPkgQR() subpkg_repository.ISubPkgQR

	FinanceQR() finance_repository.IFinanceQR
}
