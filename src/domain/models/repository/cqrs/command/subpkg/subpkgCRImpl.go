package subpkg_cmd_repository

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	subpkg_repo "gitlab.faza.io/order-project/order-service/domain/models/repository/subpackage"
)

const (
	defaultCommandStreamBuffer = 262144
)

type iSubPkgCRImpl struct {
	subPkgRepo subpkg_repo.ISubpackageRepository
	cmdStream  repository.CommandStream
}

func SubPkgCRFactory(command *mongoadapter.Mongo, database, collection string) (ISubPkgCR, repository.CommandReaderStream) {
	subPkgRepository := subpkg_repo.NewSubPkgRepository(command, database, collection)
	stream := make(chan *repository.CommandData, defaultCommandStreamBuffer)

	return &iSubPkgCRImpl{
		subPkgRepo: subPkgRepository,
		cmdStream:  stream,
	}, stream
}

func (subPkgCR iSubPkgCRImpl) Save(ctx context.Context, subPkg *entities.Subpackage) repository.IRepoError {
	err := subPkgCR.subPkgRepo.Save(ctx, subPkg)
	if err == nil {
		subPkgCR.cmdStream <- &repository.CommandData{
			Repository: repository.SubPkgRepo,
			Command:    repository.SaveCmd,
			Data:       subPkg,
		}
	}

	return err
}

func (subPkgCR iSubPkgCRImpl) SaveAll(ctx context.Context, subPkgList []*entities.Subpackage) repository.IRepoError {
	err := subPkgCR.subPkgRepo.SaveAll(ctx, subPkgList)
	if err == nil {
		subPkgCR.cmdStream <- &repository.CommandData{
			Repository: repository.SubPkgRepo,
			Command:    repository.SaveAllCmd,
			Data:       subPkgList,
		}
	}

	return err
}

func (subPkgCR iSubPkgCRImpl) Update(ctx context.Context, subPkg entities.Subpackage) (*entities.Subpackage, repository.IRepoError) {
	updateSubPkg, err := subPkgCR.subPkgRepo.Update(ctx, subPkg)
	if err == nil {
		subPkgCR.cmdStream <- &repository.CommandData{
			Repository: repository.SubPkgRepo,
			Command:    repository.UpdateCmd,
			Data:       updateSubPkg,
		}
	}

	return updateSubPkg, err
}

func (subPkgCR iSubPkgCRImpl) UpdateAll(ctx context.Context, subPkgList []entities.Subpackage) ([]*entities.Subpackage, repository.IRepoError) {
	updateSubPkges, err := subPkgCR.subPkgRepo.UpdateAll(ctx, subPkgList)
	if err == nil {
		subPkgCR.cmdStream <- &repository.CommandData{
			Repository: repository.SubPkgRepo,
			Command:    repository.UpdateAllCmd,
			Data:       updateSubPkges,
		}
	}

	return updateSubPkges, err
}

func (subPkgCR iSubPkgCRImpl) GenerateUniqSid(ctx context.Context, oid uint64) (uint64, repository.IRepoError) {
	return subPkgCR.subPkgRepo.GenerateUniqSid(ctx, oid)
}
