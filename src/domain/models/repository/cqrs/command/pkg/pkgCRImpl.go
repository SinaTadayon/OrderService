package pkg_cmd_repository

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	pkg_repo "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
)

const (
	defaultCommandStreamBuffer = 262144
)

type iPkgCRImpl struct {
	pkgRepo   pkg_repo.IPkgItemRepository
	cmdStream repository.CommandStream
}

func PkgCRFactory(command *mongoadapter.Mongo, database, collection string) (IPkgCR, repository.CommandReaderStream) {
	pkgRepository := pkg_repo.NewPkgItemRepository(command, database, collection)
	stream := make(chan *repository.CommandData, defaultCommandStreamBuffer)

	return &iPkgCRImpl{
		pkgRepo:   pkgRepository,
		cmdStream: stream,
	}, stream
}

func (pkgCR iPkgCRImpl) Update(ctx context.Context, pkgItem entities.PackageItem, upsert bool) (*entities.PackageItem, repository.IRepoError) {
	updatePkg, err := pkgCR.pkgRepo.Update(ctx, pkgItem, upsert)
	if err == nil {
		pkgCR.cmdStream <- &repository.CommandData{
			Repository: repository.PkgRepo,
			Command:    repository.UpdateCmd,
			Data:       updatePkg,
		}
	}

	return updatePkg, err
}
