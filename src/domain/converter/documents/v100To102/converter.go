package v100To102

import (
	"context"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	databaseName    string = "orderService"
	collectionName  string = "orders"
	defaultDocCount int    = 1024
)

func SchedulerConvert() error {
	orders, err := app.Globals.OrderRepository.FindAll(context.Background())
	if err != nil {
		logger.Err("convert() => app.Globals.OrderRepository.FindAll failed, err: %v", err)
		return err
	}

	convertedOrders := make([]entities.Order, 0, len(orders))
	for i := 0; i < len(orders); i++ {
		newOrder := entities.Order{
			ID:            orders[i].ID,
			OrderId:       orders[i].OrderId,
			Version:       orders[i].Version,
			DocVersion:    entities.DocumentVersion,
			Platform:      orders[i].Platform,
			OrderPayment:  orders[i].OrderPayment,
			SystemPayment: orders[i].SystemPayment,
			Status:        orders[i].Status,
			BuyerInfo:     orders[i].BuyerInfo,
			Invoice:       orders[i].Invoice,
			Packages:      nil,
			CreatedAt:     orders[i].CreatedAt,
			UpdatedAt:     orders[i].UpdatedAt,
			DeletedAt:     orders[i].DeletedAt,
			Extended:      orders[i].Extended,
		}
		newOrder.Packages = make([]*entities.PackageItem, 0, len(orders[i].Packages))
		for j := 0; j < len(orders[i].Packages); j++ {
			newPackageItem := &entities.PackageItem{
				PId:             orders[i].Packages[j].PId,
				OrderId:         orders[i].Packages[j].OrderId,
				Version:         orders[i].Packages[j].Version,
				Invoice:         orders[i].Packages[j].Invoice,
				SellerInfo:      orders[i].Packages[j].SellerInfo,
				ShopName:        orders[i].Packages[j].ShopName,
				ShippingAddress: orders[i].Packages[j].ShippingAddress,
				ShipmentSpec:    orders[i].Packages[j].ShipmentSpec,
				PayToSeller:     orders[i].Packages[j].PayToSeller,
				Subpackages:     nil,
				Status:          orders[i].Packages[j].Status,
				CreatedAt:       orders[i].Packages[j].CreatedAt,
				UpdatedAt:       orders[i].Packages[j].UpdatedAt,
				DeletedAt:       orders[i].Packages[j].DeletedAt,
				Extended:        orders[i].Packages[j].Extended,
			}

			newOrder.Packages[j].Subpackages = make([]*entities.Subpackage, 0, len(orders[i].Packages[j].Subpackages))
			for t := 0; t < len(orders[i].Packages[j].Subpackages); t++ {
				newSubpackage := &entities.Subpackage{
					SId:       orders[i].Packages[j].Subpackages[t].SId,
					PId:       orders[i].Packages[j].Subpackages[t].PId,
					OrderId:   orders[i].Packages[j].Subpackages[t].OrderId,
					Version:   orders[i].Packages[j].Subpackages[t].Version,
					Items:     orders[i].Packages[j].Subpackages[t].Items,
					Shipments: orders[i].Packages[j].Subpackages[t].Shipments,
					Tracking: entities.Progress{
						State:    nil,
						Action:   orders[i].Packages[j].Subpackages[t].Tracking.Action,
						History:  nil,
						Extended: orders[i].Packages[j].Subpackages[t].Tracking.Extended,
					},
					Status:    orders[i].Packages[j].Subpackages[t].Status,
					CreatedAt: orders[i].Packages[j].Subpackages[t].CreatedAt,
					UpdatedAt: orders[i].Packages[j].Subpackages[t].UpdatedAt,
					DeletedAt: orders[i].Packages[j].Subpackages[t].DeletedAt,
					Extended:  orders[i].Packages[j].Subpackages[t].Extended,
				}

				if orders[i].Packages[j].Subpackages[t].Tracking.State != nil {
					newSubpackage.Tracking.State = &entities.State{
						Name:       orders[i].Packages[j].Subpackages[t].Tracking.State.Name,
						Index:      orders[i].Packages[j].Subpackages[t].Tracking.State.Index,
						Schedulers: nil,
						Data:       nil,
						Actions:    orders[i].Packages[j].Subpackages[t].Tracking.State.Actions,
						CreatedAt:  orders[i].Packages[j].Subpackages[t].Tracking.State.CreatedAt,
						Extended:   orders[i].Packages[j].Subpackages[t].Tracking.State.Extended,
					}

					if orders[i].Packages[j].Subpackages[t].Tracking.State.Data != nil {
						schedulerData := orders[i].Packages[j].Subpackages[t].Tracking.State.Data["scheduler"].(primitive.A)
						newSubpackage.Tracking.State.Schedulers = make([]*entities.SchedulerData, 0, len(schedulerData))

						for _, data := range schedulerData {
							scheduler := data.(map[string]interface{})
							schData := &entities.SchedulerData{
								Name:     states.SchedulerJobName,
								Group:    states.SchedulerGroupName,
								Action:   scheduler["action"].(string),
								Index:    scheduler["index"].(int32),
								Retry:    0,
								Cron:     "",
								Start:    nil,
								End:      nil,
								Type:     "",
								Mode:     "",
								Policy:   nil,
								Enabled:  scheduler["enabled"].(bool),
								Data:     scheduler["value"],
								Extended: nil,
							}

							if scheduler["name"].(string) == "expireAt" {
								schData.Type = string(states.SchedulerSubpackageStateExpire)
							} else {
								schData.Type = string(states.SchedulerSubpackageStateNotify)
							}

							newSubpackage.Tracking.State.Schedulers = append(newSubpackage.Tracking.State.Schedulers, schData)
						}
					}
				}

				newSubpackage.Tracking.History = make([]entities.State, 0, len(orders[i].Packages[j].Subpackages[t].Tracking.History))
				for z := 0; z < len(orders[i].Packages[j].Subpackages[t].Tracking.History); z++ {
					newState := &entities.State{
						Name:       orders[i].Packages[j].Subpackages[t].Tracking.History[z].Name,
						Index:      orders[i].Packages[j].Subpackages[t].Tracking.History[z].Index,
						Schedulers: nil,
						Data:       nil,
						Actions:    orders[i].Packages[j].Subpackages[t].Tracking.History[z].Actions,
						CreatedAt:  orders[i].Packages[j].Subpackages[t].Tracking.History[z].CreatedAt,
						Extended:   orders[i].Packages[j].Subpackages[t].Tracking.History[z].Extended,
					}

					if orders[i].Packages[j].Subpackages[t].Tracking.History[z].Data != nil {
						schedulerData := orders[i].Packages[j].Subpackages[t].Tracking.History[z].Data["scheduler"].(primitive.A)
						newState.Schedulers = make([]*entities.SchedulerData, 0, len(schedulerData))

						for _, data := range schedulerData {
							scheduler := data.(map[string]interface{})
							schData := &entities.SchedulerData{
								Name:     states.SchedulerJobName,
								Group:    states.SchedulerGroupName,
								Action:   scheduler["action"].(string),
								Index:    scheduler["index"].(int32),
								Retry:    0,
								Cron:     "",
								Start:    nil,
								End:      nil,
								Type:     "",
								Mode:     "",
								Policy:   nil,
								Enabled:  scheduler["enabled"].(bool),
								Data:     scheduler["value"],
								Extended: nil,
							}

							if scheduler["name"].(string) == "expireAt" {
								schData.Type = string(states.SchedulerSubpackageStateExpire)
							} else {
								schData.Type = string(states.SchedulerSubpackageStateNotify)
							}

							newState.Schedulers = append(newState.Schedulers, schData)
						}
					}
				}

				newOrder.Packages[j].Subpackages = append(newOrder.Packages[j].Subpackages, newSubpackage)
			}

			newOrder.Packages = append(newOrder.Packages, newPackageItem)
		}

		convertedOrders = append(convertedOrders, newOrder)
	}

	err = app.Globals.OrderRepository.RemoveAll(context.Background())
	if err != nil {
		logger.Err("convert() => app.Globals.OrderRepository.RemoveAll failed, err: %v", err)
		return err
	}

	for _, newOrder := range convertedOrders {
		_, err = app.Globals.OrderRepository.Save(context.Background(), newOrder)
		if err != nil {
			logger.Err("convert() => app.Globals.OrderRepository.RemoveAll failed, err: %v", err)
			return err
		}
	}

	return nil
}

//func FindAll(ctx context.Context, mongoAdapter *mongoadapter.Mongo) ([]*model_v100.Order, error) {
//
//	cursor, e := mongoAdapter.FindMany(databaseName, collectionName, bson.D{{"deletedAt", nil}})
//	if e != nil {
//		return nil, errors.Wrap(e, "FindMany Orders Failed")
//	}
//
//	defer closeCursor(ctx, cursor)
//	orders := make([]*model_v100.Order, 0, 1000)
//
//	// iterate through all documents
//	for cursor.Next(ctx) {
//		var order model_v100.Order
//		// decode the document
//		if err := cursor.Decode(&order); err != nil {
//			return nil, repository.ErrorFactory(repository.InternalErr, "Request Operation Failed", errors.Wrap(err, "Decode Order Failed"))
//		}
//		orders = append(orders, &order)
//	}
//
//	return orders, nil
//}
//
//func closeCursor(context context.Context, cursor *mongo.Cursor) {
//	err := cursor.Close(context)
//	if err != nil {
//		logger.Err("closeCursor() failed, err: %s", err)
//	}
//}
