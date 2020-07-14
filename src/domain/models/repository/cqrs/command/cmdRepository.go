package command_repository

import (
	order_cmd_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command/order"
	pkg_cmd_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command/pkg"
	subpkg_cmd_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command/subpkg"
)

type ICmdRepository interface {
	OrderCR() order_cmd_repository.IOrderCR

	PkgCR() pkg_cmd_repository.IPkgCR

	SubPkgCR() subpkg_cmd_repository.ISubPkgCR
}
