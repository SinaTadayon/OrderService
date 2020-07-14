package subpkg_cmd_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
)

type ISubPkgCR interface {
	Save(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError

	SaveAll(ctx context.Context, subPkgList []*entities.Subpackage) repository.IRepoError

	Update(ctx context.Context, subPkg entities.Subpackage) (*entities.Subpackage, repository.IRepoError)

	UpdateAll(ctx context.Context, subPkgList []entities.Subpackage) ([]*entities.Subpackage, repository.IRepoError)

	GenerateUniqSid(ctx context.Context, oid uint64) (uint64, repository.IRepoError)
}
