package order_cmd_repository

import (
	"context"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	order_repo "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	"time"
)

const (
	defaultCommandStreamBuffer = 262144
)

type iOrderCRImpl struct {
	orderRepo order_repo.IOrderRepository
	cmdStream repository.CommandStream
}

func OrderCRFactory(command *mongoadapter.Mongo, database, collection string) (IOrderCR, repository.CommandReaderStream) {
	orderRepository := order_repo.NewOrderRepository(command, database, collection)
	stream := make(chan *repository.CommandData, defaultCommandStreamBuffer)

	return &iOrderCRImpl{
		orderRepo: orderRepository,
		cmdStream: stream,
	}, stream
}

func (orderCR iOrderCRImpl) Save(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError) {

	newOrder, err := orderCR.orderRepo.Save(ctx, order)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.SaveCmd,
			Data:       newOrder,
		}
	}
	return newOrder, err
}

func (orderCR iOrderCRImpl) SaveAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError) {
	newOrders, err := orderCR.orderRepo.SaveAll(ctx, orders)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.SaveAllCmd,
			Data:       newOrders,
		}
	}

	return newOrders, err
}

func (orderCR iOrderCRImpl) Update(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError) {
	updatedOrder, err := orderCR.orderRepo.Update(ctx, order)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.UpdateCmd,
			Data:       updatedOrder,
		}
	}

	return updatedOrder, err
}

func (orderCR iOrderCRImpl) UpdateAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError) {
	updateOrders, err := orderCR.orderRepo.UpdateAll(ctx, orders)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.UpdateAllCmd,
			Data:       updateOrders,
		}
	}

	return updateOrders, err
}

func (orderCR iOrderCRImpl) Insert(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError) {
	newOrder, err := orderCR.orderRepo.Insert(ctx, order)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.InsertCmd,
			Data:       newOrder,
		}
	}

	return newOrder, err
}

func (orderCR iOrderCRImpl) InsertAll(ctx context.Context, orders []entities.Order) ([]*entities.Order, repository.IRepoError) {
	newOrders, err := orderCR.orderRepo.InsertAll(ctx, orders)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.InsertAllCmd,
			Data:       newOrders,
		}
	}

	return newOrders, err
}

// only set DeletedAt field
func (orderCR iOrderCRImpl) DeleteById(ctx context.Context, orderId uint64) (*entities.Order, repository.IRepoError) {
	order, err := orderCR.orderRepo.DeleteById(ctx, orderId)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.DeleteCmd,
			Data:       order,
		}
	}

	return order, err
}

func (orderCR iOrderCRImpl) Delete(ctx context.Context, order entities.Order) (*entities.Order, repository.IRepoError) {
	deleteOrder, err := orderCR.orderRepo.Delete(ctx, order)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.DeleteCmd,
			Data:       deleteOrder,
		}
	}

	return deleteOrder, err
}

func (orderCR iOrderCRImpl) DeleteAllWithOrders(ctx context.Context, orders []*entities.Order) repository.IRepoError {
	err := orderCR.orderRepo.DeleteAllWithOrders(ctx, orders)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.DeletePartialCmd,
			Data:       orders,
		}
	}

	return err
}

func (orderCR iOrderCRImpl) DeleteAll(ctx context.Context) repository.IRepoError {
	err := orderCR.orderRepo.DeleteAll(ctx)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.DeleteAllCmd,
			Data:       time.Now().UTC(),
		}
	}

	return err
}

// remove order from db
func (orderCR iOrderCRImpl) RemoveById(ctx context.Context, orderId uint64) repository.IRepoError {
	err := orderCR.orderRepo.RemoveById(ctx, orderId)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.RemoveCmd,
			Data:       orderId,
		}
	}

	return err
}

func (orderCR iOrderCRImpl) Remove(ctx context.Context, order *entities.Order) repository.IRepoError {
	err := orderCR.orderRepo.Remove(ctx, order)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.RemoveEntityCmd,
			Data:       order,
		}
	}

	return err
}

func (orderCR iOrderCRImpl) RemoveAllWithOrders(ctx context.Context, orders []*entities.Order) repository.IRepoError {
	err := orderCR.orderRepo.RemoveAllWithOrders(ctx, orders)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.RemovePartialCmd,
			Data:       orders,
		}
	}

	return err
}

func (orderCR iOrderCRImpl) RemoveAll(ctx context.Context) repository.IRepoError {
	err := orderCR.orderRepo.RemoveAll(ctx)
	if err == nil {
		orderCR.cmdStream <- &repository.CommandData{
			Repository: repository.OrderRepo,
			Command:    repository.RemoveAllCmd,
			Data:       nil,
		}
	}

	return err
}

func (orderCR iOrderCRImpl) Count(ctx context.Context) (int64, repository.IRepoError) {
	return orderCR.orderRepo.Count(ctx)
}
