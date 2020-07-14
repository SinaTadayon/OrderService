package pkg_cmd_repository

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
)

type IPkgCR interface {
	Update(ctx context.Context, pkgItem entities.PackageItem, upsert bool) (*entities.PackageItem, repository.IRepoError)
}
