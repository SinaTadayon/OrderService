package cqrs

import (
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command"
	"gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/query"
)

type ICQRSRepository interface {
	CmdR() command_repository.ICmdRepository

	QueryR() query_repository.IQueryRepository
}
