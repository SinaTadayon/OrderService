package command_repository

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	order_cmd_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command/order"
	pkg_cmd_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command/pkg"
	subpkg_cmd_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/cqrs/command/subpkg"
	"sync"
)

const (
	defaultCommandStreamBuffer = 524288
)

type iCmdRepository struct {
	orderCmdRepo  order_cmd_repository.IOrderCR
	pkgCmdRepo    pkg_cmd_repository.IPkgCR
	subPkgCmdRepo subpkg_cmd_repository.ISubPkgCR
	//cmdStream     cqrs.CommandStream
}

func CmdRepoFactory(ctx context.Context, command *mongoadapter.Mongo, database, collection string) (ICmdRepository, repository.CommandReaderStream) {
	orderCmdRepo, orderCmdStream := order_cmd_repository.OrderCRFactory(command, database, collection)
	pkgCmdRepo, pkgCmdStream := pkg_cmd_repository.PkgCRFactory(command, database, collection)
	subPkgCmdRepo, subPkgCmdStream := subpkg_cmd_repository.SubPkgCRFactory(command, database, collection)
	stream := fanInCommandReaderStream(ctx, orderCmdStream, pkgCmdStream, subPkgCmdStream)

	return &iCmdRepository{
		orderCmdRepo:  orderCmdRepo,
		pkgCmdRepo:    pkgCmdRepo,
		subPkgCmdRepo: subPkgCmdRepo,
		//cmdStream:     stream,
	}, stream
}

func (cmd iCmdRepository) OrderCR() order_cmd_repository.IOrderCR {
	return cmd.orderCmdRepo
}

func (cmd iCmdRepository) PkgCR() pkg_cmd_repository.IPkgCR {
	return cmd.pkgCmdRepo
}

func (cmd iCmdRepository) SubPkgCR() subpkg_cmd_repository.ISubPkgCR {
	return cmd.subPkgCmdRepo
}

func fanInCommandReaderStream(ctx context.Context, channels ...repository.CommandReaderStream) repository.CommandReaderStream {
	var wg sync.WaitGroup
	multiplexedStream := make(chan *repository.CommandData, defaultCommandStreamBuffer)
	multiplex := func(commandStream repository.CommandReaderStream) {
		defer wg.Done()
		for commandData := range commandStream {
			select {
			case <-ctx.Done():
				return
			case multiplexedStream <- commandData:
			}
		}
	}
	// Select from all the channels
	wg.Add(len(channels))
	for _, channel := range channels {
		go multiplex(channel)
	}

	// Wait for all the reads to complete
	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}
