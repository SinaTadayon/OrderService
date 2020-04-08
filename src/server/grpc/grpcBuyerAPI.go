package grpc_server

import (
	"context"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	pb "gitlab.faza.io/protos/order"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// for internal use of buyer
const (
	DefaultBuyerOrderDetailListFilter FilterValue = "DefaultBuyerOrderDetailList"
)

func (server *Server) buyerGeneratePipelineFilter(ctx context.Context, filter FilterValue) []interface{} {

	newFilter := make([]interface{}, 2)

	if filter == AllOrdersFilter || filter == AllReportFilter {
		newFilter[0] = "$or"
		filterList := make([]interface{}, 0, 10)
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnRequestPendingFilter].queryPath: server.queryPathStates[ReturnRequestPendingFilter].state.StateName()})
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnRequestRejectedFilter].queryPath: server.queryPathStates[ReturnRequestRejectedFilter].state.StateName()})
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnShipmentPendingFilter].queryPath: server.queryPathStates[ReturnShipmentPendingFilter].state.StateName()})
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnShippedFilter].queryPath: server.queryPathStates[ReturnShippedFilter].state.StateName()})
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnDeliveryPendingFilter].queryPath: server.queryPathStates[ReturnDeliveryPendingFilter].state.StateName()})
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnDeliveryDelayedFilter].queryPath: server.queryPathStates[ReturnDeliveryDelayedFilter].state.StateName()})
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnDeliveryFailedFilter].queryPath: server.queryPathStates[ReturnDeliveryFailedFilter].state.StateName()})
		filterList = append(filterList, map[string]string{server.queryPathStates[ReturnDeliveredFilter].queryPath: server.queryPathStates[ReturnDeliveredFilter].state.StateName()})
		newFilter[1] = bson.A(filterList)
	} else if filter == DefaultBuyerOrderDetailListFilter {
		newFilter[0] = "$or"
		filterList := make([]interface{}, 0, 16)
		filterList = append(filterList, bson.D{{server.queryPathStates[NewOrderFilter].queryPath, server.queryPathStates[NewOrderFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[PaymentPendingFilter].queryPath, server.queryPathStates[PaymentPendingFilter].state.StateName()}})
		//filterList = append(filterList, bson.D{{server.queryPathStates[PaymentSuccessFilter].queryPath, server.queryPathStates[PaymentSuccessFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[PaymentFailedFilter].queryPath, server.queryPathStates[PaymentFailedFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[OrderVerificationPendingFilter].queryPath, server.queryPathStates[OrderVerificationPendingFilter].state.StateName()}})
		//filterList = append(filterList, bson.D{{server.queryPathStates[OrderVerificationSuccessFilter].queryPath, server.queryPathStates[OrderVerificationSuccessFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[OrderVerificationFailedFilter].queryPath, server.queryPathStates[OrderVerificationFailedFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[ApprovalPendingFilter].queryPath, server.queryPathStates[ApprovalPendingFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[ShipmentPendingFilter].queryPath, server.queryPathStates[ShipmentPendingFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[ShipmentDelayedFilter].queryPath, server.queryPathStates[ShipmentDelayedFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[ShippedFilter].queryPath, server.queryPathStates[ShippedFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[DeliveryPendingFilter].queryPath, server.queryPathStates[DeliveryPendingFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[DeliveryDelayedFilter].queryPath, server.queryPathStates[DeliveryDelayedFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[DeliveredFilter].queryPath, server.queryPathStates[DeliveredFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[DeliveryFailedFilter].queryPath, server.queryPathStates[DeliveryFailedFilter].state.StateName()}})
		filterList = append(filterList, bson.D{{server.queryPathStates[PayToBuyerFilter].queryPath, server.queryPathStates[PayToBuyerFilter].state.StateName()}})
		newFilter[1] = bson.A(filterList)
	} else {
		queryPathState := server.queryPathStates[filter]
		newFilter[0] = queryPathState.queryPath
		newFilter[1] = queryPathState.state.StateName()
	}

	//if filter == ApprovalPendingFilter {
	//	queryPathApprovalPendingState := server.queryPathStates[ApprovalPendingFilter]
	//	newFilter[0] = queryPathApprovalPendingState.queryPath
	//	newFilter[1] = queryPathApprovalPendingState.state.StateName()
	//} else if filter == ShipmentPendingFilter {
	//	queryPathShipmentPendingState := server.queryPathStates[ShipmentPendingFilter]
	//	queryPathShipmentDelayedState := server.queryPathStates[ShipmentDelayedFilter]
	//	newFilter[0] = "$or"
	//	newFilter[1] = bson.A{
	//		bson.M{queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()},
	//		bson.M{queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}}
	//} else if filter == ShippedFilter {
	//	queryPathShippedState := server.queryPathStates[ShippedFilter]
	//	newFilter[0] = queryPathShippedState.queryPath
	//	newFilter[1] = queryPathShippedState.state.StateName()
	//} else if filter == DeliveredFilter {
	//	queryPathDeliveryPendingState := server.queryPathStates[DeliveryPendingFilter]
	//	queryPathDeliveryDelayedState := server.queryPathStates[DeliveryDelayedFilter]
	//	queryPathDeliveredState := server.queryPathStates[DeliveredFilter]
	//	newFilter[0] = "$or"
	//	newFilter[1] = bson.A{
	//		bson.M{queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()},
	//		bson.M{queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()},
	//		bson.M{queryPathDeliveredState.queryPath: queryPathDeliveredState.state.StateName()}}
	//} else if filter == DeliveryFailedFilter {
	//	queryPathDeliveryFailedState := server.queryPathStates[DeliveryFailedFilter]
	//	newFilter[0] = queryPathDeliveryFailedState.queryPath
	//	newFilter[1] = queryPathDeliveryFailedState.state.StateName()
	//} else if filter == ReturnRequestPendingFilter {
	//	queryPathReturnRequestPendingState := server.queryPathStates[ReturnRequestPendingFilter]
	//	queryPathReturnRequestRejectedState := server.queryPathStates[ReturnRequestRejectedFilter]
	//	queryPathReturnCanceledState := server.queryPathStates[ReturnCanceledFilter]
	//	newFilter[0] = "$or"
	//	newFilter[1] = bson.A{
	//		bson.M{queryPathReturnRequestPendingState.queryPath: queryPathReturnRequestPendingState.state.StateName()},
	//		bson.M{queryPathReturnRequestRejectedState.queryPath: queryPathReturnRequestRejectedState.state.StateName()},
	//		bson.M{queryPathReturnCanceledState.queryPath: queryPathReturnCanceledState.state.StateName()}}
	//} else if filter == ReturnShipmentPendingFilter {
	//	queryPathReturnShipmentPendingState := server.queryPathStates[ReturnShipmentPendingFilter]
	//	newFilter[0] = queryPathReturnShipmentPendingState.queryPath
	//	newFilter[1] = queryPathReturnShipmentPendingState.state.StateName()
	//} else if filter == ReturnShippedFilter {
	//	queryPathReturnShippedFilterState := server.queryPathStates[ReturnShippedFilter]
	//	newFilter[0] = queryPathReturnShippedFilterState.queryPath
	//	newFilter[1] = queryPathReturnShippedFilterState.state.StateName()
	//} else if filter == ReturnDeliveredFilter {
	//	queryPathReturnDeliveryPendingState := server.queryPathStates[ReturnDeliveryPendingFilter]
	//	queryPathReturnDeliveryDelayedState := server.queryPathStates[ReturnDeliveryDelayedFilter]
	//	queryPathReturnDeliveredState := server.queryPathStates[ReturnDeliveredFilter]
	//	newFilter[0] = "$or"
	//	newFilter[1] = bson.A{
	//		bson.M{queryPathReturnDeliveryPendingState.queryPath: queryPathReturnDeliveryPendingState.state.StateName()},
	//		bson.M{queryPathReturnDeliveryDelayedState.queryPath: queryPathReturnDeliveryDelayedState.state.StateName()},
	//		bson.M{queryPathReturnDeliveredState.queryPath: queryPathReturnDeliveredState.state.StateName()}}
	//} else if filter == DeliveryFailedFilter {
	//	queryPathDeliveryFailedState := server.queryPathStates[DeliveryFailedFilter]
	//	newFilter[0] = queryPathDeliveryFailedState.queryPath
	//	newFilter[1] = queryPathDeliveryFailedState.state.StateName()
	//}
	return newFilter
}

func (server *Server) buyerOrderDetailListHandler(ctx context.Context, oid, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	if oid > 0 {
		return server.buyerGetOrderDetailByIdHandler(ctx, oid)
	}

	if page <= 0 || perPage <= 0 {
		app.Globals.Logger.FromContext(ctx).Error("page or perPage invalid", "fn", "buyerOrderDetailListHandler", "uid", userId, "oid", oid, "page", page, "perPage", perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "Page/PerPage Invalid")
	}

	var buyerFilter bson.D
	genFilter := server.buyerGeneratePipelineFilter(ctx, DefaultBuyerOrderDetailListFilter)
	buyerFilter = make(bson.D, 0, 3)
	buyerFilter = append(buyerFilter, bson.E{Key: "buyerInfo.buyerId", Value: userId})
	buyerFilter = append(buyerFilter, bson.E{Key: "deletedAt", Value: nil})
	buyerFilter = append(buyerFilter, bson.E{Key: genFilter[0].(string), Value: genFilter[1]})

	var orderList []*entities.Order
	var total int64
	var err repository.IRepoError
	if sortName != "" {
		var sortDirect int
		if direction == "ASC" {
			sortDirect = 1
		} else {
			sortDirect = -1
		}

		orderFilter := func() (interface{}, string, int) {
			return buyerFilter, sortName, sortDirect
		}
		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "buyerOrderDetailListHandler", "oid", oid, "uid", userId, "page", page, "perPage", perPage, "error", err)
			return nil, status.Error(codes.Code(err.Code()), err.Message())
		}
	} else {
		orderFilter := func() interface{} {
			return buyerFilter
		}
		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "buyerOrderDetailListHandler", "oid", oid, "uid", userId, "page", page, "perPage", perPage, "error", err)
			return nil, status.Error(codes.Code(err.Code()), err.Message())
		}
	}

	if total == 0 || orderList == nil || len(orderList) == 0 {
		//app.Globals.Logger.FromContext(ctx).Info("oid not found", "fn", "buyerOrderDetailListHandler", "oid", oid, "uid", userId, "filter", filter)
		//return nil, status.Error(codes.Code(future.NotFound), "Order Not Found")

		buyerOrderDetailList := &pb.BuyerOrderDetailList{
			BuyerId:      userId,
			OrderDetails: nil,
		}

		serializedData, e := proto.Marshal(buyerOrderDetailList)
		if e != nil {
			app.Globals.Logger.FromContext(ctx).Error("could not marshal buyerOrderDetailList",
				"fn", "buyerOrderDetailListHandler",
				"oid", oid,
				"uid", userId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}

		response := &pb.MessageResponse{
			Entity: "buyerOrderDetailList",
			Meta: &pb.ResponseMetadata{
				Total:   0,
				Page:    page,
				PerPage: perPage,
			},
			Data: &any.Any{
				TypeUrl: "baman.io/" + proto.MessageName(buyerOrderDetailList),
				Value:   serializedData,
			},
		}

		return response, nil

	}

	orderDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		packageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package, 0, len(orderList[i].Packages))
		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				if filter == AllOrdersFilter {
					var statusName string
					if stateList, ok := server.buyerAllStatesMap[orderList[i].Packages[j].Subpackages[z].Status]; ok {
						if len(stateList) == 1 {
							statusName = stateList[0].StateName()
						} else {
							length := len(orderList[i].Packages[j].Subpackages[z].Tracking.History)
							for _, state := range stateList {
								if orderList[i].Packages[j].Subpackages[z].Tracking.History[length-2].Name == state.StateName() {
									statusName = state.StateName()
									break
								}
							}
						}

						itemPackageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package_Item, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
						for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
							itemPackageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package_Item{
								SID:                orderList[i].Packages[j].Subpackages[z].SId,
								Status:             statusName,
								SIdx:               int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
								IsCancelable:       false,
								IsReturnable:       false,
								IsReturnCancelable: false,
								InventoryId:        orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
								Title:              orderList[i].Packages[j].Subpackages[z].Items[t].Title,
								Brand:              orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
								Image:              orderList[i].Packages[j].Subpackages[z].Items[t].Image,
								Returnable:         orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
								Quantity:           orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
								Attributes:         nil,
								Invoice: &pb.BuyerOrderDetailList_OrderDetail_Package_Item_Invoice{
									Unit:     0,
									Total:    0,
									Original: 0,
									Special:  0,
									Discount: 0,
									Currency: "IRR",
								},
							}

							unit, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount)
							if err != nil {
								app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
									"fn", "buyerOrderDetailListHandler",
									"unit", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount,
									"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
									"pid", orderList[i].Packages[j].Subpackages[z].PId,
									"sid", orderList[i].Packages[j].Subpackages[z].SId,
									"error", err)
								return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
							}
							itemPackageDetail.Invoice.Unit = uint64(unit.IntPart())

							total, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount)
							if err != nil {
								app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
									"fn", "buyerOrderDetailListHandler",
									"total", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount,
									"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
									"pid", orderList[i].Packages[j].Subpackages[z].PId,
									"sid", orderList[i].Packages[j].Subpackages[z].SId,
									"error", err)
								return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
							}
							itemPackageDetail.Invoice.Total = uint64(total.IntPart())

							original, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount)
							if err != nil {
								app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
									"fn", "buyerOrderDetailListHandler",
									"original", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount,
									"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
									"pid", orderList[i].Packages[j].Subpackages[z].PId,
									"sid", orderList[i].Packages[j].Subpackages[z].SId,
									"error", err)
								return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
							}
							itemPackageDetail.Invoice.Original = uint64(original.IntPart())

							special, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount)
							if err != nil {
								app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
									"fn", "buyerOrderDetailListHandler",
									"special", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount,
									"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
									"pid", orderList[i].Packages[j].Subpackages[z].PId,
									"sid", orderList[i].Packages[j].Subpackages[z].SId,
									"error", err)
								return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
							}
							itemPackageDetail.Invoice.Special = uint64(special.IntPart())

							discount, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount)
							if err != nil {
								app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
									"fn", "buyerOrderDetailListHandler",
									"discount", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount,
									"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
									"pid", orderList[i].Packages[j].Subpackages[z].PId,
									"sid", orderList[i].Packages[j].Subpackages[z].SId,
									"error", err)
								return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
							}
							itemPackageDetail.Invoice.Discount = uint64(discount.IntPart())

							if itemPackageDetail.Status == states.ApprovalPending.StateName() ||
								itemPackageDetail.Status == states.ShipmentPending.StateName() ||
								itemPackageDetail.Status == states.ShipmentDelayed.StateName() {
								itemPackageDetail.IsCancelable = true

							} else if itemPackageDetail.Status == states.Delivered.StateName() {
								itemPackageDetail.IsReturnable = true

							} else if itemPackageDetail.Status == states.ReturnRequestPending.StateName() {
								itemPackageDetail.IsReturnCancelable = true
							}

							itemPackageDetailList = append(itemPackageDetailList, itemPackageDetail)
						}

						packageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package{
							PID:          orderList[i].Packages[j].PId,
							ShopName:     orderList[i].Packages[j].ShopName,
							Items:        itemPackageDetailList,
							ShipmentInfo: nil,
						}

						packageDetail.ShipmentInfo = &pb.BuyerOrderDetailList_OrderDetail_Package_Shipment{}
						if orderList[i].Packages[j].ShipmentSpec.ShippingCost != nil {
							shippingCost, err := decimal.NewFromString(orderList[i].Packages[j].ShipmentSpec.ShippingCost.Amount)
							if err != nil {
								app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, package ShippingCost.Amount invalid",
									"fn", "buyerOrderDetailListHandler",
									"ShippingCost", orderList[i].Packages[j].ShipmentSpec.ShippingCost,
									"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
									"pid", orderList[i].Packages[j].Subpackages[z].PId,
									"sid", orderList[i].Packages[j].Subpackages[z].SId,
									"error", err)
								return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
							}

							packageDetail.ShipmentInfo.ShipmentAmount = uint64(shippingCost.IntPart())
							packageDetail.ShipmentInfo.ReactionTime = uint32(orderList[i].Packages[j].ShipmentSpec.ReactionTime)
							// packageDetail.ShipmentInfo.
						}

						if orderList[i].Packages[j].Subpackages[z].Shipments != nil &&
							orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail != nil {
							packageDetail.ShipmentInfo.DeliveryAt = ""
							packageDetail.ShipmentInfo.ShippedAt = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601)
							packageDetail.ShipmentInfo.CarrierName = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.CourierName
							packageDetail.ShipmentInfo.TrackingNumber = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.TrackingNumber

							packageDetail.ShipmentInfo.DeliveryAt = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.
								Add(time.Duration(orderList[i].Packages[j].ShipmentSpec.ShippingTime) * time.Hour).Format(ISO8601)
						}

						packageDetailList = append(packageDetailList, packageDetail)
					}

				} else {
					for _, filterState := range server.buyerFilterStates[filter] {
						if orderList[i].Packages[j].Subpackages[i].Status == filterState.actualState.StateName() {
							var statusName string
							if len(filterState.expectedState) == 1 {
								statusName = filterState.expectedState[0].StateName()
							} else {
								length := len(orderList[i].Packages[j].Subpackages[i].Tracking.History)
								for _, state := range filterState.expectedState {
									if orderList[i].Packages[j].Subpackages[i].Tracking.History[length-2].Name == state.StateName() {
										statusName = state.StateName()
										break
									}
								}
							}

							itemPackageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package_Item, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
							for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
								itemPackageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package_Item{
									SID:                orderList[i].Packages[j].Subpackages[z].SId,
									Status:             statusName,
									SIdx:               int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
									IsCancelable:       false,
									IsReturnable:       false,
									IsReturnCancelable: false,
									InventoryId:        orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
									Title:              orderList[i].Packages[j].Subpackages[z].Items[t].Title,
									Brand:              orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
									Image:              orderList[i].Packages[j].Subpackages[z].Items[t].Image,
									Returnable:         orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
									Quantity:           orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
									Attributes:         nil,
									Invoice: &pb.BuyerOrderDetailList_OrderDetail_Package_Item_Invoice{
										Unit:     0,
										Total:    0,
										Original: 0,
										Special:  0,
										Discount: 0,
										Currency: "IRR",
									},
								}

								unit, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount)
								if err != nil {
									app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
										"fn", "buyerOrderDetailListHandler",
										"unit", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount,
										"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
										"pid", orderList[i].Packages[j].Subpackages[z].PId,
										"sid", orderList[i].Packages[j].Subpackages[z].SId,
										"error", err)
									return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
								}
								itemPackageDetail.Invoice.Unit = uint64(unit.IntPart())

								total, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount)
								if err != nil {
									app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
										"fn", "buyerOrderDetailListHandler",
										"total", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount,
										"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
										"pid", orderList[i].Packages[j].Subpackages[z].PId,
										"sid", orderList[i].Packages[j].Subpackages[z].SId,
										"error", err)
									return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
								}
								itemPackageDetail.Invoice.Total = uint64(total.IntPart())

								original, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount)
								if err != nil {
									app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
										"fn", "buyerOrderDetailListHandler",
										"original", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount,
										"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
										"pid", orderList[i].Packages[j].Subpackages[z].PId,
										"sid", orderList[i].Packages[j].Subpackages[z].SId,
										"error", err)
									return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
								}
								itemPackageDetail.Invoice.Original = uint64(original.IntPart())

								special, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount)
								if err != nil {
									app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
										"fn", "buyerOrderDetailListHandler",
										"special", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount,
										"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
										"pid", orderList[i].Packages[j].Subpackages[z].PId,
										"sid", orderList[i].Packages[j].Subpackages[z].SId,
										"error", err)
									return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
								}
								itemPackageDetail.Invoice.Special = uint64(special.IntPart())

								discount, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount)
								if err != nil {
									app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
										"fn", "buyerOrderDetailListHandler",
										"discount", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount,
										"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
										"pid", orderList[i].Packages[j].Subpackages[z].PId,
										"sid", orderList[i].Packages[j].Subpackages[z].SId,
										"error", err)
									return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
								}
								itemPackageDetail.Invoice.Discount = uint64(discount.IntPart())

								if itemPackageDetail.Status == states.ApprovalPending.StateName() ||
									itemPackageDetail.Status == states.ShipmentPending.StateName() ||
									itemPackageDetail.Status == states.ShipmentDelayed.StateName() {
									itemPackageDetail.IsCancelable = true

								} else if itemPackageDetail.Status == states.Delivered.StateName() {
									itemPackageDetail.IsReturnable = true

								} else if itemPackageDetail.Status == states.ReturnRequestPending.StateName() {
									itemPackageDetail.IsReturnCancelable = true
								}

								itemPackageDetailList = append(itemPackageDetailList, itemPackageDetail)
							}

							packageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package{
								PID:          orderList[i].Packages[j].PId,
								ShopName:     orderList[i].Packages[j].ShopName,
								Items:        itemPackageDetailList,
								ShipmentInfo: nil,
							}

							packageDetail.ShipmentInfo = &pb.BuyerOrderDetailList_OrderDetail_Package_Shipment{}
							if orderList[i].Packages[j].ShipmentSpec.ShippingCost != nil {
								shippingCost, err := decimal.NewFromString(orderList[i].Packages[j].ShipmentSpec.ShippingCost.Amount)
								if err != nil {
									app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, package ShippingCost.Amount invalid",
										"fn", "buyerOrderDetailListHandler",
										"ShippingCost", orderList[i].Packages[j].ShipmentSpec.ShippingCost,
										"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
										"pid", orderList[i].Packages[j].Subpackages[z].PId,
										"sid", orderList[i].Packages[j].Subpackages[z].SId,
										"error", err)
									return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
								}

								packageDetail.ShipmentInfo.ShipmentAmount = uint64(shippingCost.IntPart())
								packageDetail.ShipmentInfo.ReactionTime = uint32(orderList[i].Packages[j].ShipmentSpec.ReactionTime)
								// packageDetail.ShipmentInfo.
							}

							if orderList[i].Packages[j].Subpackages[z].Shipments != nil &&
								orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail != nil {
								packageDetail.ShipmentInfo.DeliveryAt = ""
								packageDetail.ShipmentInfo.ShippedAt = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601)
								packageDetail.ShipmentInfo.CarrierName = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.CourierName
								packageDetail.ShipmentInfo.TrackingNumber = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.TrackingNumber

								packageDetail.ShipmentInfo.DeliveryAt = orderList[i].Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.
									Add(time.Duration(orderList[i].Packages[j].ShipmentSpec.ShippingTime) * time.Hour).Format(ISO8601)
							}

							packageDetailList = append(packageDetailList, packageDetail)
						}
					}
				}
			}
		}

		if len(packageDetailList) <= 0 {
			continue
		}

		orderDetail := &pb.BuyerOrderDetailList_OrderDetail{
			OrderId:          orderList[i].OrderId,
			PackageCount:     int32(len(orderList[i].Packages)),
			TotalAmount:      0,
			PayableAmount:    0,
			Discounts:        0,
			ShipmentAmount:   0,
			IsPaymentSuccess: false,
			RequestAt:        orderList[i].CreatedAt.Format(ISO8601),
			Address: &pb.BuyerOrderDetailList_OrderDetail_BuyerAddress{
				FirstName:     orderList[i].BuyerInfo.ShippingAddress.FirstName,
				LastName:      orderList[i].BuyerInfo.ShippingAddress.LastName,
				Address:       orderList[i].BuyerInfo.ShippingAddress.Address,
				Phone:         orderList[i].BuyerInfo.ShippingAddress.Phone,
				Mobile:        orderList[i].BuyerInfo.ShippingAddress.Mobile,
				Country:       orderList[i].BuyerInfo.ShippingAddress.Country,
				City:          orderList[i].BuyerInfo.ShippingAddress.City,
				Province:      orderList[i].BuyerInfo.ShippingAddress.Province,
				Neighbourhood: orderList[i].BuyerInfo.ShippingAddress.Neighbourhood,
				Lat:           "",
				Long:          "",
				ZipCode:       orderList[i].BuyerInfo.ShippingAddress.ZipCode,
			},
			Packages: packageDetailList,
		}

		grandTotal, err := decimal.NewFromString(orderList[i].Invoice.GrandTotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, GrandTotal invalid",
				"fn", "buyerOrderDetailListHandler",
				"grandTotal", orderList[i].Invoice.GrandTotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		orderDetail.PayableAmount = uint64(grandTotal.IntPart())

		subtotal, err := decimal.NewFromString(orderList[i].Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
				"fn", "buyerOrderDetailListHandler",
				"subtotal", orderList[i].Invoice.Subtotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		orderDetail.TotalAmount = uint64(subtotal.IntPart())

		shipmentTotal, err := decimal.NewFromString(orderList[i].Invoice.ShipmentTotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, shipmentTotal invalid",
				"fn", "buyerOrderDetailListHandler",
				"shipmentTotal", orderList[i].Invoice.ShipmentTotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		orderDetail.ShipmentAmount = uint64(shipmentTotal.IntPart())

		discount, err := decimal.NewFromString(orderList[i].Invoice.Discount.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, discount invalid",
				"fn", "buyerOrderDetailListHandler",
				"discount", orderList[i].Invoice.Discount.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		orderDetail.Discounts = uint64(discount.IntPart())

		if orderList[i].BuyerInfo.ShippingAddress.Location != nil {
			orderDetail.Address.Lat = strconv.Itoa(int(orderList[i].BuyerInfo.ShippingAddress.Location.Coordinates[0]))
			orderDetail.Address.Long = strconv.Itoa(int(orderList[i].BuyerInfo.ShippingAddress.Location.Coordinates[1]))
		}

		if orderList[i].OrderPayment != nil && orderList[i].OrderPayment[0].PaymentResult != nil {
			orderDetail.IsPaymentSuccess = orderList[i].OrderPayment[0].PaymentResult.Result
		}

		orderDetailList = append(orderDetailList, orderDetail)
	}

	buyerOrderDetailList := &pb.BuyerOrderDetailList{
		BuyerId:      userId,
		OrderDetails: orderDetailList,
	}

	serializedData, e := proto.Marshal(buyerOrderDetailList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("could not marshal buyerOrderDetailList",
			"fn", "buyerOrderDetailListHandler",
			"oid", oid,
			"uid", userId,
			"error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "buyerOrderDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(total),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerOrderDetailList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerGetOrderDetailByIdHandler(ctx context.Context, oid uint64) (*pb.MessageResponse, error) {

	order, err := app.Globals.OrderRepository.FindById(ctx, oid)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("FindById failed", "fn", "buyerGetOrderDetailByIdHandler", "oid", oid, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	orderDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail, 0, 1)

	packageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package, 0, len(order.Packages))
	for j := 0; j < len(order.Packages); j++ {
		for z := 0; z < len(order.Packages[j].Subpackages); z++ {
			itemPackageDetailList := make([]*pb.BuyerOrderDetailList_OrderDetail_Package_Item, 0, len(order.Packages[j].Subpackages[z].Items))
			for t := 0; t < len(order.Packages[j].Subpackages[z].Items); t++ {
				itemPackageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package_Item{
					SID:                order.Packages[j].Subpackages[z].SId,
					Status:             order.Packages[j].Subpackages[z].Status,
					SIdx:               int32(states.FromString(order.Packages[j].Subpackages[z].Status).StateIndex()),
					IsCancelable:       false,
					IsReturnable:       false,
					IsReturnCancelable: false,
					InventoryId:        order.Packages[j].Subpackages[z].Items[t].InventoryId,
					Title:              order.Packages[j].Subpackages[z].Items[t].Title,
					Brand:              order.Packages[j].Subpackages[z].Items[t].Brand,
					Image:              order.Packages[j].Subpackages[z].Items[t].Image,
					Returnable:         order.Packages[j].Subpackages[z].Items[t].Returnable,
					Quantity:           order.Packages[j].Subpackages[z].Items[t].Quantity,
					Attributes:         nil,
					Invoice: &pb.BuyerOrderDetailList_OrderDetail_Package_Item_Invoice{
						Unit:     0,
						Total:    0,
						Original: 0,
						Special:  0,
						Discount: 0,
						Currency: "IRR",
					},
				}

				if order.Packages[j].Subpackages[z].Items[t].Attributes != nil {
					itemPackageDetail.Attributes = make(map[string]*pb.BuyerOrderDetailList_OrderDetail_Package_Item_Attribute, len(order.Packages[j].Subpackages[z].Items[t].Attributes))
					for attrKey, attribute := range order.Packages[j].Subpackages[z].Items[t].Attributes {
						keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
						for keyTran, value := range attribute.KeyTranslate {
							keyTranslates[keyTran] = value
						}
						valTranslates := make(map[string]string, len(attribute.ValueTranslate))
						for valTran, value := range attribute.ValueTranslate {
							valTranslates[valTran] = value
						}
						itemPackageDetail.Attributes[attrKey] = &pb.BuyerOrderDetailList_OrderDetail_Package_Item_Attribute{
							KeyTranslates:   keyTranslates,
							ValueTranslates: valTranslates,
						}
					}
				}

				unit, err := decimal.NewFromString(order.Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
						"fn", "buyerGetOrderDetailByIdHandler",
						"unit", order.Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount,
						"oid", order.Packages[j].Subpackages[z].OrderId,
						"pid", order.Packages[j].Subpackages[z].PId,
						"sid", order.Packages[j].Subpackages[z].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemPackageDetail.Invoice.Unit = uint64(unit.IntPart())

				total, err := decimal.NewFromString(order.Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
						"fn", "buyerGetOrderDetailByIdHandler",
						"total", order.Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount,
						"oid", order.Packages[j].Subpackages[z].OrderId,
						"pid", order.Packages[j].Subpackages[z].PId,
						"sid", order.Packages[j].Subpackages[z].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemPackageDetail.Invoice.Total = uint64(total.IntPart())

				original, err := decimal.NewFromString(order.Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("buyerGetOrderDetailByIdHandler() => decimal.NewFromString failed, subpackage Invoice.Original invalid",
						"fn", "buyerGetOrderDetailByIdHandler",
						"original", order.Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount,
						"oid", order.Packages[j].Subpackages[z].OrderId,
						"pid", order.Packages[j].Subpackages[z].PId,
						"sid", order.Packages[j].Subpackages[z].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemPackageDetail.Invoice.Original = uint64(original.IntPart())

				special, err := decimal.NewFromString(order.Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
						"fn", "buyerGetOrderDetailByIdHandler",
						"special", order.Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount,
						"oid", order.Packages[j].Subpackages[z].OrderId,
						"pid", order.Packages[j].Subpackages[z].PId,
						"sid", order.Packages[j].Subpackages[z].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemPackageDetail.Invoice.Special = uint64(special.IntPart())

				discount, err := decimal.NewFromString(order.Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
						"fn", "buyerGetOrderDetailByIdHandler",
						"discount", order.Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount,
						"oid", order.Packages[j].Subpackages[z].OrderId,
						"pid", order.Packages[j].Subpackages[z].PId,
						"sid", order.Packages[j].Subpackages[z].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemPackageDetail.Invoice.Discount = uint64(discount.IntPart())

				if itemPackageDetail.Status == states.ApprovalPending.StateName() ||
					itemPackageDetail.Status == states.ShipmentPending.StateName() ||
					itemPackageDetail.Status == states.ShipmentDelayed.StateName() {
					itemPackageDetail.IsCancelable = true

				} else if itemPackageDetail.Status == states.Delivered.StateName() {
					itemPackageDetail.IsReturnable = true

				} else if itemPackageDetail.Status == states.ReturnRequestPending.StateName() {
					itemPackageDetail.IsReturnCancelable = true
				}

				itemPackageDetailList = append(itemPackageDetailList, itemPackageDetail)
			}

			packageDetail := &pb.BuyerOrderDetailList_OrderDetail_Package{
				PID:          order.Packages[j].PId,
				ShopName:     order.Packages[j].ShopName,
				Items:        itemPackageDetailList,
				ShipmentInfo: nil,
			}

			packageDetail.ShipmentInfo = &pb.BuyerOrderDetailList_OrderDetail_Package_Shipment{}
			if order.Packages[j].ShipmentSpec.ShippingCost != nil {
				shippingCost, err := decimal.NewFromString(order.Packages[j].ShipmentSpec.ShippingCost.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, package ShippingCost.Amount invalid",
						"fn", "buyerGetOrderDetailByIdHandler",
						"ShippingCost", order.Packages[j].ShipmentSpec.ShippingCost,
						"oid", order.Packages[j].Subpackages[z].OrderId,
						"pid", order.Packages[j].Subpackages[z].PId,
						"sid", order.Packages[j].Subpackages[z].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}

				packageDetail.ShipmentInfo.ShipmentAmount = uint64(shippingCost.IntPart())
				packageDetail.ShipmentInfo.ReactionTime = uint32(order.Packages[j].ShipmentSpec.ReactionTime)
			}

			if order.Packages[j].Subpackages[z].Shipments != nil &&
				order.Packages[j].Subpackages[z].Shipments.ShipmentDetail != nil {
				packageDetail.ShipmentInfo.ShippedAt = order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601)
				packageDetail.ShipmentInfo.CarrierName = order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.CourierName
				packageDetail.ShipmentInfo.TrackingNumber = order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.TrackingNumber

				packageDetail.ShipmentInfo.DeliveryAt = order.Packages[j].Subpackages[z].Shipments.ShipmentDetail.ShippedAt.
					Add(time.Duration(order.Packages[j].ShipmentSpec.ShippingTime) * time.Hour).Format(ISO8601)
			}

			packageDetailList = append(packageDetailList, packageDetail)
		}
	}

	orderDetail := &pb.BuyerOrderDetailList_OrderDetail{
		OrderId:          order.OrderId,
		PackageCount:     int32(len(order.Packages)),
		TotalAmount:      0,
		PayableAmount:    0,
		Discounts:        0,
		ShipmentAmount:   0,
		IsPaymentSuccess: false,
		RequestAt:        order.CreatedAt.Format(ISO8601),
		Address: &pb.BuyerOrderDetailList_OrderDetail_BuyerAddress{
			FirstName:     order.BuyerInfo.ShippingAddress.FirstName,
			LastName:      order.BuyerInfo.ShippingAddress.LastName,
			Address:       order.BuyerInfo.ShippingAddress.Address,
			Phone:         order.BuyerInfo.ShippingAddress.Phone,
			Mobile:        order.BuyerInfo.ShippingAddress.Mobile,
			Country:       order.BuyerInfo.ShippingAddress.Country,
			City:          order.BuyerInfo.ShippingAddress.City,
			Province:      order.BuyerInfo.ShippingAddress.Province,
			Neighbourhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
			Lat:           "",
			Long:          "",
			ZipCode:       order.BuyerInfo.ShippingAddress.ZipCode,
		},
		Packages: packageDetailList,
	}

	grandTotal, e := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, GrandTotal invalid",
			"fn", "buyerGetOrderDetailByIdHandler",
			"grandTotal", order.Invoice.GrandTotal.Amount,
			"oid", order.OrderId,
			"error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.PayableAmount = uint64(grandTotal.IntPart())

	subtotal, e := decimal.NewFromString(order.Invoice.Subtotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
			"fn", "buyerGetOrderDetailByIdHandler",
			"subtotal", order.Invoice.Subtotal.Amount,
			"oid", order.OrderId,
			"error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.TotalAmount = uint64(subtotal.IntPart())

	shipmentTotal, e := decimal.NewFromString(order.Invoice.ShipmentTotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, shipmentTotal invalid",
			"fn", "buyerGetOrderDetailByIdHandler",
			"shipmentTotal", order.Invoice.ShipmentTotal.Amount,
			"oid", order.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.ShipmentAmount = uint64(shipmentTotal.IntPart())

	discount, e := decimal.NewFromString(order.Invoice.Discount.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, discount invalid",
			"fn", "buyerGetOrderDetailByIdHandler",
			"discount", order.Invoice.Discount.Amount,
			"oid", order.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.Discounts = uint64(discount.IntPart())

	if order.BuyerInfo.ShippingAddress.Location != nil {
		orderDetail.Address.Lat = strconv.Itoa(int(order.BuyerInfo.ShippingAddress.Location.Coordinates[0]))
		orderDetail.Address.Long = strconv.Itoa(int(order.BuyerInfo.ShippingAddress.Location.Coordinates[1]))
	}

	if order.OrderPayment != nil && order.OrderPayment[0].PaymentResult != nil {
		orderDetail.IsPaymentSuccess = order.OrderPayment[0].PaymentResult.Result
	}

	orderDetailList = append(orderDetailList, orderDetail)

	buyerOrderDetailList := &pb.BuyerOrderDetailList{
		BuyerId:      order.BuyerInfo.BuyerId,
		OrderDetails: orderDetailList,
	}

	serializedData, e := proto.Marshal(buyerOrderDetailList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal buyerOrderDetailList failed", "fn", "buyerGetOrderDetailByIdHandler", "oid", oid, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "BuyerOrderDetailList",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerOrderDetailList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerAllOrderReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	returnFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"$or": server.buyerGeneratePipelineFilter(ctx, AllReportFilter)[1]}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	orderFilter := func() interface{} {
		return bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil, "$or": server.buyerGeneratePipelineFilter(ctx, DefaultBuyerOrderDetailListFilter)[1]}
	}

	returnOrdersCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnFilter failed", "fn", "buyerAllOrderReportsHandler", "uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	OrdersCount, err := app.Globals.OrderRepository.CountWithFilter(ctx, orderFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for orderFilter failed", "fn", "buyerAllOrderReportsHandler", "uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	buyerAllOrderReports := &pb.BuyerAllOrderReports{
		BuyerId:      userId,
		Orders:       int32(OrdersCount),
		ReturnOrders: int32(returnOrdersCount),
	}

	serializedData, e := proto.Marshal(buyerAllOrderReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal buyerAllOrderReports failed", "fn", "buyerAllOrderReportsHandler", "uid", userId, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}

	response := &pb.MessageResponse{
		Entity: "BuyerAllOrderReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerAllOrderReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerReturnOrderReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathReturnRequestPendingState := server.queryPathStates[ReturnRequestPendingFilter]
	queryPathReturnRequestRejectedState := server.queryPathStates[ReturnRequestRejectedFilter]
	returnRequestPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"$or": bson.A{
				bson.M{queryPathReturnRequestPendingState.queryPath: queryPathReturnRequestPendingState.state.StateName()},
				bson.M{queryPathReturnRequestRejectedState.queryPath: queryPathReturnRequestRejectedState.state.StateName()}},
				"packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnShipmentPendingState := server.queryPathStates[ReturnShipmentPendingFilter]
	returnShipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{queryPathReturnShipmentPendingState.queryPath: queryPathReturnShipmentPendingState.state.StateName(), "packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnShippedState := server.queryPathStates[ReturnShippedFilter]
	returnShippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{queryPathReturnShippedState.queryPath: queryPathReturnShippedState.state.StateName(), "packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveredState := server.queryPathStates[ReturnDeliveredFilter]
	queryPathReturnDeliveryDelayedState := server.queryPathStates[ReturnDeliveryDelayedFilter]
	queryPathReturnDeliveryPendingState := server.queryPathStates[ReturnDeliveryPendingFilter]
	returnDeliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"$or": bson.A{
				bson.M{queryPathReturnDeliveredState.queryPath: queryPathReturnDeliveredState.state.StateName()},
				bson.M{queryPathReturnDeliveryDelayedState.queryPath: queryPathReturnDeliveryDelayedState.state.StateName()},
				bson.M{queryPathReturnDeliveryPendingState.queryPath: queryPathReturnDeliveryPendingState.state.StateName()}},
				"packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveryFailedState := server.queryPathStates[ReturnDeliveryFailedFilter]
	returnDeliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"buyerInfo.buyerId": userId, "deletedAt": nil}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{queryPathReturnDeliveryFailedState.queryPath: queryPathReturnDeliveryFailedState.state.StateName(), "packages.deletedAt": nil}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnRequestPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnRequestPendingFilter failed", "fn", "buyerReturnOrderReportsHandler", "uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnShipmentPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnShipmentPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnShipmentPendingFilter failed", "fn", "buyerReturnOrderReportsHandler", "uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnShippedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnShippedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnShippedFilter failed", "fn", "buyerReturnOrderReportsHandler", "uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveredCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveredFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveredFilter failed", "fn", "buyerReturnOrderReportsHandler", "uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveryFailedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveryFailedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveryFailedFilter failed", "fn", "buyerReturnOrderReportsHandler", "uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	buyerReturnOrderReports := &pb.BuyerReturnOrderReports{
		BuyerId:               userId,
		ReturnRequestPending:  int32(returnRequestPendingCount),
		ReturnShipmentPending: int32(returnShipmentPendingCount),
		ReturnShipped:         int32(returnShippedCount),
		ReturnDelivered:       int32(returnDeliveredCount),
		ReturnDeliveryFailed:  int32(returnDeliveryFailedCount),
	}

	serializedData, e := proto.Marshal(buyerReturnOrderReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal buyerReturnOrderReports failed", "fn", "buyerReturnOrderReportsHandler", "uid", userId, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}

	response := &pb.MessageResponse{
		Entity: "BuyerReturnOrderReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerReturnOrderReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) buyerAllReturnOrderItemsHandler(ctx context.Context, userId uint64, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		app.Globals.Logger.FromContext(ctx).Error("page or perPage invalid", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "page", page, "perPage", perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "Page/PerPage Invalid")
	}

	var returnFilter bson.M
	genFilter := server.buyerGeneratePipelineFilter(ctx, AllOrdersFilter)
	returnFilter = make(bson.M, 3)
	returnFilter["buyerInfo.buyerId"] = userId
	returnFilter["deletedAt"] = nil
	returnFilter[genFilter[0].(string)] = genFilter[1]

	var orderList []*entities.Order
	var total int64
	var err repository.IRepoError

	if sortName != "" {
		var sortDirect int
		if direction == "ASC" {
			sortDirect = 1
		} else {
			sortDirect = -1
		}

		orderFilter := func() (interface{}, string, int) {
			return returnFilter, sortName, sortDirect
		}

		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "page", page, "perPage", perPage, "error", err)
			return nil, status.Error(codes.Code(err.Code()), err.Message())
		}
	} else {
		orderFilter := func() interface{} {
			return returnFilter
		}

		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "page", page, "perPage", perPage, "error", err)
			return nil, status.Error(codes.Code(err.Code()), err.Message())
		}
	}

	if total == 0 || orderList == nil || len(orderList) == 0 {
		//app.Globals.Logger.FromContext(ctx).Info("order not found", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId)
		//return nil, status.Error(codes.Code(future.NotFound), "Order Not Found")

		BuyerReturnOrderItemDetailList := &pb.BuyerReturnOrderItemDetailList{
			BuyerId:                   userId,
			ReturnOrderItemDetailList: nil,
		}

		serializedData, e := proto.Marshal(BuyerReturnOrderItemDetailList)
		if e != nil {
			app.Globals.Logger.FromContext(ctx).Error("marshal BuyerReturnOrderItemDetailList failed", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "error", e)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}

		response := &pb.MessageResponse{
			Entity: "BuyerReturnOrderItemDetailList",
			Meta: &pb.ResponseMetadata{
				Total:   0,
				Page:    page,
				PerPage: perPage,
			},
			Data: &any.Any{
				TypeUrl: "baman.io/" + proto.MessageName(BuyerReturnOrderItemDetailList),
				Value:   serializedData,
			},
		}

		return response, nil
	}

	returnOrderItemDetailList := make([]*pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail, 0, len(orderList)*2)
	for i := 0; i < len(orderList); i++ {
		for j := 0; j < len(orderList[i].Packages); j++ {
			if orderList[i].Packages[j].SellerInfo == nil {
				futureData := app.Globals.UserService.GetSellerProfile(ctx, strconv.Itoa(int(orderList[i].Packages[j].PId))).Get()
				if futureData.Error() != nil {
					app.Globals.Logger.FromContext(ctx).Warn("UserService.GetSellerProfile failed",
						"fn", "buyerAllReturnOrderItemsHandler",
						"oid", orderList[i].Packages[j].OrderId,
						"pid", orderList[i].Packages[j].PId,
						"error", futureData.Error().Reason())
				} else {
					if futureData.Data() != nil {
						orderList[i].Packages[j].SellerInfo = futureData.Data().(*entities.SellerProfile)
					}
				}
			}

			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				var statusName string
				if stateList, ok := server.buyerReturnStatesMap[orderList[i].Packages[j].Subpackages[z].Status]; ok {
					if len(stateList) == 1 {
						statusName = stateList[0].StateName()
					} else {
						length := len(orderList[i].Packages[j].Subpackages[z].Tracking.History)
						for _, state := range stateList {
							if orderList[i].Packages[j].Subpackages[z].Tracking.History[length-2].Name == state.StateName() {
								statusName = state.StateName()
								break
							}
						}
					}

					for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
						returnOrderItemDetail := &pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail{
							OID:         orderList[i].OrderId,
							CreatedAt:   orderList[i].CreatedAt.Format(ISO8601),
							TotalAmount: 0,
							SellerInfo:  nil,
							Item: &pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail_Item{
								SID:             orderList[i].Packages[j].Subpackages[z].SId,
								Status:          statusName,
								SIdx:            int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
								IsCancelable:    false,
								IsAccepted:      false,
								InventoryId:     orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
								Title:           orderList[i].Packages[j].Subpackages[z].Items[t].Title,
								Brand:           orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
								Image:           orderList[i].Packages[j].Subpackages[z].Items[t].Image,
								Returnable:      orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
								Quantity:        orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
								Attributes:      nil,
								Reason:          nil,
								ReturnRequestAt: "",
								Invoice: &pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail_Item_Invoice{
									Unit:     0,
									Total:    0,
									Original: 0,
									Special:  0,
									Discount: 0,
									Currency: "IRR",
								},
							},
						}

						grandTotal, err := decimal.NewFromString(orderList[i].Invoice.GrandTotal.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
								"fn", "buyerAllReturnOrderItemsHandler",
								"grandTotal", orderList[i].Invoice.GrandTotal.Amount,
								"oid", orderList[i].OrderId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						returnOrderItemDetail.TotalAmount = uint64(grandTotal.IntPart())

						if orderList[i].Packages[j].Subpackages[z].Items[t].Attributes != nil {
							returnOrderItemDetail.Item.Attributes = make(map[string]*pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail_Item_Attribute, len(orderList[i].Packages[j].Subpackages[z].Items[t].Attributes))
							for attrKey, attribute := range orderList[i].Packages[j].Subpackages[z].Items[t].Attributes {
								keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
								for keyTran, value := range attribute.KeyTranslate {
									keyTranslates[keyTran] = value
								}
								valTranslates := make(map[string]string, len(attribute.ValueTranslate))
								for valTran, value := range attribute.ValueTranslate {
									valTranslates[valTran] = value
								}
								returnOrderItemDetail.Item.Attributes[attrKey] = &pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail_Item_Attribute{
									KeyTranslates:   keyTranslates,
									ValueTranslates: valTranslates,
								}
							}
						}

						unit, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
								"fn", "buyerAllReturnOrderItemsHandler",
								"unit", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount,
								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
								"pid", orderList[i].Packages[j].Subpackages[z].PId,
								"sid", orderList[i].Packages[j].Subpackages[z].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						returnOrderItemDetail.Item.Invoice.Unit = uint64(unit.IntPart())

						total, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
								"fn", "buyerAllReturnOrderItemsHandler",
								"total", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount,
								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
								"pid", orderList[i].Packages[j].Subpackages[z].PId,
								"sid", orderList[i].Packages[j].Subpackages[z].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						returnOrderItemDetail.Item.Invoice.Total = uint64(total.IntPart())

						original, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
								"fn", "buyerAllReturnOrderItemsHandler",
								"original", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount,
								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
								"pid", orderList[i].Packages[j].Subpackages[z].PId,
								"sid", orderList[i].Packages[j].Subpackages[z].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						returnOrderItemDetail.Item.Invoice.Original = uint64(original.IntPart())

						special, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
								"fn", "buyerAllReturnOrderItemsHandler",
								"special", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount,
								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
								"pid", orderList[i].Packages[j].Subpackages[z].PId,
								"sid", orderList[i].Packages[j].Subpackages[z].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						returnOrderItemDetail.Item.Invoice.Special = uint64(special.IntPart())

						discount, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
								"fn", "buyerAllReturnOrderItemsHandler",
								"discount", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount,
								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
								"pid", orderList[i].Packages[j].Subpackages[z].PId,
								"sid", orderList[i].Packages[j].Subpackages[z].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						returnOrderItemDetail.Item.Invoice.Discount = uint64(discount.IntPart())

						if orderList[i].Packages[j].Subpackages[z].Items[t].Reasons != nil {
							returnOrderItemDetail.Item.Reason = orderList[i].Packages[j].Subpackages[z].Items[t].Reasons[0].ToRPC()
						}

						if orderList[i].Packages[j].Subpackages[z].Shipments != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt != nil {
							returnOrderItemDetail.Item.ReturnRequestAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
						}

						if returnOrderItemDetail.Item.Status == states.ReturnRequestPending.StateName() {
							returnOrderItemDetail.Item.IsCancelable = true

						} else if returnOrderItemDetail.Item.Status == states.ReturnShipmentPending.StateName() {
							returnOrderItemDetail.Item.IsAccepted = true
						}

						returnOrderItemDetail.SellerInfo = &pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail_SellerInfo{
							PID:          orderList[i].Packages[j].PId,
							ShopName:     orderList[i].Packages[j].ShopName,
							Mobile:       "",
							Phone:        "",
							ShipmentInfo: nil,
						}

						if orderList[i].Packages[j].SellerInfo != nil {
							if orderList[i].Packages[j].SellerInfo.ReturnInfo != nil {
								returnOrderItemDetail.SellerInfo.ShipmentInfo = &pb.BuyerReturnOrderItemDetailList_ReturnOrderItemDetail_SellerInfo_ReturnShipmentInfo{
									Country:      orderList[i].Packages[j].SellerInfo.ReturnInfo.Country,
									Province:     orderList[i].Packages[j].SellerInfo.ReturnInfo.Province,
									City:         orderList[i].Packages[j].SellerInfo.ReturnInfo.City,
									Neighborhood: orderList[i].Packages[j].SellerInfo.ReturnInfo.Neighborhood,
									Address:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalAddress,
									ZipCode:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalCode}
							}
							if orderList[i].Packages[j].SellerInfo.GeneralInfo != nil {
								returnOrderItemDetail.SellerInfo.Mobile = orderList[i].Packages[j].SellerInfo.GeneralInfo.MobilePhone
								returnOrderItemDetail.SellerInfo.Phone = orderList[i].Packages[j].SellerInfo.GeneralInfo.LandPhone
							}
						}

						returnOrderItemDetailList = append(returnOrderItemDetailList, returnOrderItemDetail)
					}
				}
			}
		}
	}

	BuyerReturnOrderItemDetailList := &pb.BuyerReturnOrderItemDetailList{
		BuyerId:                   userId,
		ReturnOrderItemDetailList: returnOrderItemDetailList,
	}

	serializedData, e := proto.Marshal(BuyerReturnOrderItemDetailList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal BuyerReturnOrderItemDetailList failed", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "BuyerReturnOrderItemDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(total),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(BuyerReturnOrderItemDetailList),
			Value:   serializedData,
		},
	}

	return response, nil
}

//func (server *Server) buyerAllReturnOrderItemsHandler(ctx context.Context, userId uint64, page, perPage uint32,
//	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
//	if page <= 0 || perPage <= 0 {
//		app.Globals.Logger.FromContext(ctx).Error("page or perPage invalid", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "page", page, "perPage", perPage)
//		return nil, status.Error(codes.Code(future.BadRequest), "Page/PerPage Invalid")
//	}
//
//	var returnFilter bson.M
//	genFilter := server.buyerGeneratePipelineFilter(ctx, AllReturnOrdersFilter)
//	returnFilter = make(bson.M, 3)
//	returnFilter["buyerInfo.buyerId"] = userId
//	returnFilter["deletedAt"] = nil
//	returnFilter[genFilter[0].(string)] = genFilter[1]
//
//	//var returnFilter bson.D
//	//returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"$or", bson.A{
//	//	bson.D{{server.queryPathStates[ReturnRequestPendingFilter].queryPath, server.queryPathStates[ReturnRequestPendingFilter].state.StateName()}},
//	//	bson.D{{server.queryPathStates[ReturnRequestRejectedFilter].queryPath, server.queryPathStates[ReturnRequestRejectedFilter].state.StateName()}},
//	//	bson.D{{server.queryPathStates[ReturnShipmentPendingFilter].queryPath, server.queryPathStates[ReturnShipmentPendingFilter].state.StateName()}},
//	//	bson.D{{server.queryPathStates[ReturnShippedFilter].queryPath, server.queryPathStates[ReturnShippedFilter].state.StateName()}},
//	//	bson.D{{server.queryPathStates[ReturnDeliveryPendingFilter].queryPath, server.queryPathStates[ReturnDeliveryPendingFilter].state.StateName()}},
//	//	bson.D{{server.queryPathStates[ReturnDeliveryDelayedFilter].queryPath, server.queryPathStates[ReturnDeliveryDelayedFilter].state.StateName()}},
//	//	bson.D{{server.queryPathStates[ReturnDeliveryFailedFilter].queryPath, server.queryPathStates[ReturnDeliveryFailedFilter].state.StateName()}},
//	//	bson.D{{server.queryPathStates[ReturnDeliveredFilter].queryPath, server.queryPathStates[ReturnDeliveredFilter].state.StateName()}}}}}
//
//	var orderList []*entities.Order
//	var total int64
//	var err repository.IRepoError
//
//	if sortName != "" {
//		var sortDirect int
//		if direction == "ASC" {
//			sortDirect = 1
//		} else {
//			sortDirect = -1
//		}
//
//		//genFilter := server.buyerGeneratePipelineFilter(ctx, filter)
//		//filters := make(bson.M, 3)
//		//filters["buyerInfo.buyerId"] = userId
//		//filters["deletedAt"] = nil
//		//filters[genFilter[0].(string)] = genFilter[1]
//		orderFilter := func() (interface{}, string, int) {
//			return returnFilter, sortName, sortDirect
//		}
//
//		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
//		if err != nil {
//			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "page", page, "perPage", perPage, "error", err)
//			return nil, status.Error(codes.Code(err.Code()), err.Message())
//		}
//	} else {
//		orderFilter := func() interface{} {
//			return returnFilter
//		}
//
//		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
//		if err != nil {
//			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "page", page, "perPage", perPage, "error", err)
//			return nil, status.Error(codes.Code(err.Code()), err.Message())
//		}
//	}
//
//	if total == 0 || orderList == nil || len(orderList) == 0 {
//		app.Globals.Logger.FromContext(ctx).Info("order not found", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId)
//		return nil, status.Error(codes.Code(future.NotFound), "Order Not Found")
//	}
//
//	returnOrderDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail, 0, len(orderList))
//	for i := 0; i < len(orderList); i++ {
//		returnPackageDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail, 0, len(orderList[i].Packages))
//		for j := 0; j < len(orderList[i].Packages); j++ {
//			if orderList[i].Packages[j].SellerInfo == nil {
//				futureData := app.Globals.UserService.GetSellerProfile(ctx, strconv.Itoa(int(orderList[i].Packages[j].PId))).Get()
//				if futureData.Error() != nil {
//					app.Globals.Logger.FromContext(ctx).Warn("UserService.GetSellerProfile failed",
//						"fn", "buyerAllReturnOrderItemsHandler",
//						"oid", orderList[i].Packages[j].OrderId,
//						"pid", orderList[i].Packages[j].PId,
//						"error", futureData.Error().Reason())
//				} else {
//					if futureData.Data() != nil {
//						orderList[i].Packages[j].SellerInfo = futureData.Data().(*entities.SellerProfile)
//					}
//				}
//			}
//
//			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
//				var statusName string
//				if stateList, ok := server.buyerReturnStatesMap[orderList[i].Packages[j].Subpackages[z].Status]; ok {
//					if len(stateList) == 1 {
//						statusName = stateList[0].StateName()
//					} else {
//						length := len(orderList[i].Packages[j].Subpackages[z].Tracking.History)
//						for _, state := range stateList {
//							if orderList[i].Packages[j].Subpackages[z].Tracking.History[length-2].Name == state.StateName() {
//								statusName = state.StateName()
//								break
//							}
//						}
//					}
//
//					returnItemPackageDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
//					for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
//						returnItemPackageDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item{
//							SID:             orderList[i].Packages[j].Subpackages[z].SId,
//							Status:          statusName,
//							SIdx:            int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
//							IsCancelable:    false,
//							IsAccepted:      false,
//							InventoryId:     orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
//							Title:           orderList[i].Packages[j].Subpackages[z].Items[t].Title,
//							Brand:           orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
//							Image:           orderList[i].Packages[j].Subpackages[z].Items[t].Image,
//							Returnable:      orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
//							Quantity:        orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
//							Attributes:      nil,
//							Reason:          nil,
//							ReturnRequestAt: "",
//							Invoice: &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item_Invoice{
//								Unit:     0,
//								Total:    0,
//								Original: 0,
//								Special:  0,
//								Discount: 0,
//								Currency: "IRR",
//							},
//						}
//
//						if orderList[i].Packages[j].Subpackages[z].Items[t].Attributes != nil {
//							returnItemPackageDetail.Attributes = make(map[string]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item_Attribute, len(orderList[i].Packages[j].Subpackages[z].Items[t].Attributes))
//							for attrKey, attribute := range orderList[i].Packages[j].Subpackages[z].Items[t].Attributes {
//								keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
//								for keyTran, value := range attribute.KeyTranslate {
//									keyTranslates[keyTran] = value
//								}
//								valTranslates := make(map[string]string, len(attribute.ValueTranslate))
//								for valTran, value := range attribute.ValueTranslate {
//									valTranslates[valTran] = value
//								}
//								returnItemPackageDetail.Attributes[attrKey] = &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item_Attribute{
//									KeyTranslates:   keyTranslates,
//									ValueTranslates: valTranslates,
//								}
//							}
//						}
//
//						unit, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount)
//						if err != nil {
//							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
//								"fn", "buyerAllReturnOrderItemsHandler",
//								"unit", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount,
//								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
//								"pid", orderList[i].Packages[j].Subpackages[z].PId,
//								"sid", orderList[i].Packages[j].Subpackages[z].SId,
//								"error", err)
//							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//						}
//						returnItemPackageDetail.Invoice.Unit = uint64(unit.IntPart())
//
//						total, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount)
//						if err != nil {
//							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
//								"fn", "buyerAllReturnOrderItemsHandler",
//								"total", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount,
//								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
//								"pid", orderList[i].Packages[j].Subpackages[z].PId,
//								"sid", orderList[i].Packages[j].Subpackages[z].SId,
//								"error", err)
//							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//						}
//						returnItemPackageDetail.Invoice.Total = uint64(total.IntPart())
//
//						original, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount)
//						if err != nil {
//							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
//								"fn", "buyerAllReturnOrderItemsHandler",
//								"original", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount,
//								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
//								"pid", orderList[i].Packages[j].Subpackages[z].PId,
//								"sid", orderList[i].Packages[j].Subpackages[z].SId,
//								"error", err)
//							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//						}
//						returnItemPackageDetail.Invoice.Original = uint64(original.IntPart())
//
//						special, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount)
//						if err != nil {
//							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
//								"fn", "buyerAllReturnOrderItemsHandler",
//								"special", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount,
//								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
//								"pid", orderList[i].Packages[j].Subpackages[z].PId,
//								"sid", orderList[i].Packages[j].Subpackages[z].SId,
//								"error", err)
//							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//						}
//						returnItemPackageDetail.Invoice.Special = uint64(special.IntPart())
//
//						discount, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount)
//						if err != nil {
//							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
//								"fn", "buyerAllReturnOrderItemsHandler",
//								"discount", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount,
//								"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
//								"pid", orderList[i].Packages[j].Subpackages[z].PId,
//								"sid", orderList[i].Packages[j].Subpackages[z].SId,
//								"error", err)
//							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//						}
//						returnItemPackageDetail.Invoice.Discount = uint64(discount.IntPart())
//
//						if orderList[i].Packages[j].Subpackages[z].Items[t].Reasons != nil {
//							returnItemPackageDetail.Reason = orderList[i].Packages[j].Subpackages[z].Items[t].Reasons[0].ToRPC()
//						}
//
//						if orderList[i].Packages[j].Subpackages[z].Shipments != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt != nil {
//							returnItemPackageDetail.ReturnRequestAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
//						}
//
//						if returnItemPackageDetail.Status == states.ReturnRequestPending.StateName() {
//							returnItemPackageDetail.IsCancelable = true
//
//						} else if returnItemPackageDetail.Status == states.ReturnShipmentPending.StateName() {
//							returnItemPackageDetail.IsAccepted = true
//
//						}
//
//						returnItemPackageDetailList = append(returnItemPackageDetailList, returnItemPackageDetail)
//					}
//
//					returnPackageDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail{
//						PID:      orderList[i].Packages[j].PId,
//						ShopName: orderList[i].Packages[j].ShopName,
//						Mobile:   "",
//						Phone:    "",
//						Shipment: nil,
//						Items:    returnItemPackageDetailList,
//					}
//
//					if orderList[i].Packages[j].SellerInfo != nil {
//						if orderList[i].Packages[j].SellerInfo.ReturnInfo != nil {
//							returnPackageDetail.Shipment = &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_SellerReturnShipment{
//								Country:      orderList[i].Packages[j].SellerInfo.ReturnInfo.Country,
//								Province:     orderList[i].Packages[j].SellerInfo.ReturnInfo.Province,
//								City:         orderList[i].Packages[j].SellerInfo.ReturnInfo.City,
//								Neighborhood: orderList[i].Packages[j].SellerInfo.ReturnInfo.Neighborhood,
//								Address:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalAddress,
//								ZipCode:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalCode}
//						}
//						if orderList[i].Packages[j].SellerInfo.GeneralInfo != nil {
//							returnPackageDetail.Mobile = orderList[i].Packages[j].SellerInfo.GeneralInfo.MobilePhone
//							returnPackageDetail.Phone = orderList[i].Packages[j].SellerInfo.GeneralInfo.LandPhone
//						}
//					}
//
//					returnPackageDetailList = append(returnPackageDetailList, returnPackageDetail)
//				}
//			}
//		}
//
//		returnOrderDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail{
//			OID:                 orderList[i].OrderId,
//			CreatedAt:           orderList[i].CreatedAt.Format(ISO8601),
//			TotalAmount:         0,
//			ReturnPackageDetail: returnPackageDetailList,
//		}
//
//		subtotal, err := decimal.NewFromString(orderList[i].Invoice.Subtotal.Amount)
//		if err != nil {
//			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
//				"fn", "buyerAllReturnOrderItemsHandler",
//				"subtotal", orderList[i].Invoice.Subtotal.Amount,
//				"oid", orderList[i].OrderId,
//				"error", err)
//			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//		}
//		returnOrderDetail.TotalAmount = uint64(subtotal.IntPart())
//
//		returnOrderDetailList = append(returnOrderDetailList, returnOrderDetail)
//	}
//
//	buyerReturnOrderDetailList := &pb.BuyerReturnOrderDetailList{
//		BuyerId:           userId,
//		ReturnOrderDetail: returnOrderDetailList,
//	}
//
//	serializedData, e := proto.Marshal(buyerReturnOrderDetailList)
//	if e != nil {
//		app.Globals.Logger.FromContext(ctx).Error("marshal buyerReturnOrderDetailList failed", "fn", "buyerAllReturnOrderItemsHandler", "uid", userId, "error", e)
//		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
//	}
//
//	response := &pb.MessageResponse{
//		Entity: "BuyerReturnOrderDetailList",
//		Meta: &pb.ResponseMetadata{
//			Total:   uint32(total),
//			Page:    page,
//			PerPage: perPage,
//		},
//		Data: &any.Any{
//			TypeUrl: "baman.io/" + proto.MessageName(buyerReturnOrderDetailList),
//			Value:   serializedData,
//		},
//	}
//
//	return response, nil
//}

func (server *Server) buyerReturnOrderDetailListHandler(ctx context.Context, userId uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		app.Globals.Logger.FromContext(ctx).Error("page or perPage invalid", "fn", "buyerReturnOrderDetailListHandler", "uid", userId, "filter", filter, "page", page, "perPage", perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "Page/PerPage invalid")
	}

	if filter == AllOrdersFilter {
		return server.buyerAllReturnOrderItemsHandler(ctx, userId, page, perPage, sortName, direction)
	}

	var returnFilter bson.D
	if filter == ReturnDeliveredFilter {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"$or", bson.A{
			bson.D{{server.queryPathStates[ReturnDeliveryPendingFilter].queryPath, server.queryPathStates[ReturnDeliveryPendingFilter].state.StateName()}},
			bson.D{{server.queryPathStates[ReturnDeliveryDelayedFilter].queryPath, server.queryPathStates[ReturnDeliveryDelayedFilter].state.StateName()}},
			bson.D{{server.queryPathStates[ReturnDeliveredFilter].queryPath, server.queryPathStates[ReturnDeliveredFilter].state.StateName()}}}}}
	} else if filter == DeliveryFailedFilter {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {server.queryPathStates[DeliveryFailedFilter].queryPath, server.queryPathStates[DeliveryFailedFilter].state.StateName()}}
	} else if filter == ReturnRequestPendingFilter {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {"$or", bson.A{
			bson.D{{server.queryPathStates[ReturnRequestPendingFilter].queryPath, server.queryPathStates[ReturnRequestPendingFilter].state.StateName()}},
			bson.D{{server.queryPathStates[ReturnRequestRejectedFilter].queryPath, server.queryPathStates[ReturnRequestRejectedFilter].state.StateName()}}}}}
	} else {
		returnFilter = bson.D{{"buyerInfo.buyerId", userId}, {"deletedAt", nil}, {server.queryPathStates[filter].queryPath, server.queryPathStates[filter].state.StateName()}}
	}

	var orderList []*entities.Order
	var total int64
	var err repository.IRepoError
	if sortName != "" {
		var sortDirect int
		if direction == "ASC" {
			sortDirect = 1
		} else {
			sortDirect = -1
		}

		//genFilter := server.buyerGeneratePipelineFilter(ctx, filter)
		//filters := make(bson.M, 3)
		//filters["buyerInfo.buyerId"] = userId
		//filters["deletedAt"] = nil
		//filters[genFilter[0].(string)] = genFilter[1]
		orderFilter := func() (interface{}, string, int) {
			return returnFilter, sortName, sortDirect
		}

		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "buyerReturnOrderDetailListHandler", "uid", userId, "filter", filter, "page", page, "perPage", perPage, "error", err)
			return nil, status.Error(codes.Code(err.Code()), err.Message())
		}
	} else {
		orderFilter := func() interface{} {
			return returnFilter
		}

		orderList, total, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "buyerReturnOrderDetailListHandler", "uid", userId, "filter", filter, "page", page, "perPage", perPage, "error", err)
			return nil, status.Error(codes.Code(err.Code()), err.Message())
		}
	}

	if total == 0 || orderList == nil || len(orderList) == 0 {
		//app.Globals.Logger.FromContext(ctx).Error("oid not found, userId: %d, filter:%s", userId, filter)
		//return nil, status.Error(codes.Code(future.NotFound), "Order Not Found")

		buyerReturnOrderDetailList := &pb.BuyerReturnOrderDetailList{
			BuyerId:           userId,
			ReturnOrderDetail: nil,
		}

		serializedData, e := proto.Marshal(buyerReturnOrderDetailList)
		if e != nil {
			app.Globals.Logger.FromContext(ctx).Error("marshal buyerReturnOrderDetailList failed", "fn", "buyerReturnOrderDetailListHandler", "uid", userId, "filter", filter, "error", e)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}

		response := &pb.MessageResponse{
			Entity: "buyerReturnOrderDetailList",
			Meta: &pb.ResponseMetadata{
				Total:   0,
				Page:    page,
				PerPage: perPage,
			},
			Data: &any.Any{
				TypeUrl: "baman.io/" + proto.MessageName(buyerReturnOrderDetailList),
				Value:   serializedData,
			},
		}

		return response, nil
	}

	returnOrderDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		returnPackageDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail, 0, len(orderList[i].Packages))
		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				returnItemPackageDetailList := make([]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item, 0, len(orderList[i].Packages[j].Subpackages[z].Items))
				for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
					returnItemPackageDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item{
						SID:             orderList[i].Packages[j].Subpackages[z].SId,
						Status:          orderList[i].Packages[j].Subpackages[z].Status,
						SIdx:            int32(states.FromString(orderList[i].Packages[j].Subpackages[z].Status).StateIndex()),
						IsCancelable:    false,
						IsAccepted:      false,
						InventoryId:     orderList[i].Packages[j].Subpackages[z].Items[t].InventoryId,
						Title:           orderList[i].Packages[j].Subpackages[z].Items[t].Title,
						Brand:           orderList[i].Packages[j].Subpackages[z].Items[t].Brand,
						Image:           orderList[i].Packages[j].Subpackages[z].Items[t].Image,
						Returnable:      orderList[i].Packages[j].Subpackages[z].Items[t].Returnable,
						Quantity:        orderList[i].Packages[j].Subpackages[z].Items[t].Quantity,
						Attributes:      nil,
						Reason:          nil,
						ReturnRequestAt: "",
						Invoice: &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item_Invoice{
							Unit:     0,
							Total:    0,
							Original: 0,
							Special:  0,
							Discount: 0,
							Currency: "IRR",
						},
					}

					if orderList[i].Packages[j].Subpackages[z].Items[t].Attributes != nil {
						returnItemPackageDetail.Attributes = make(map[string]*pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item_Attribute, len(orderList[i].Packages[j].Subpackages[z].Items[t].Attributes))
						for attrKey, attribute := range orderList[i].Packages[j].Subpackages[z].Items[t].Attributes {
							keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
							for keyTran, value := range attribute.KeyTranslate {
								keyTranslates[keyTran] = value
							}
							valTranslates := make(map[string]string, len(attribute.ValueTranslate))
							for valTran, value := range attribute.ValueTranslate {
								valTranslates[valTran] = value
							}
							returnItemPackageDetail.Attributes[attrKey] = &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_Item_Attribute{
								KeyTranslates:   keyTranslates,
								ValueTranslates: valTranslates,
							}
						}
					}

					unit, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
							"fn", "buyerReturnOrderDetailListHandler",
							"unit", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Unit.Amount,
							"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
							"pid", orderList[i].Packages[j].Subpackages[z].PId,
							"sid", orderList[i].Packages[j].Subpackages[z].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					returnItemPackageDetail.Invoice.Unit = uint64(unit.IntPart())

					total, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
							"fn", "buyerReturnOrderDetailListHandler",
							"total", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Total.Amount,
							"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
							"pid", orderList[i].Packages[j].Subpackages[z].PId,
							"sid", orderList[i].Packages[j].Subpackages[z].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					returnItemPackageDetail.Invoice.Total = uint64(total.IntPart())

					original, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
							"fn", "buyerReturnOrderDetailListHandler",
							"original", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Original.Amount,
							"orderId", orderList[i].Packages[j].Subpackages[z].OrderId,
							"pid", orderList[i].Packages[j].Subpackages[z].PId,
							"sid", orderList[i].Packages[j].Subpackages[z].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					returnItemPackageDetail.Invoice.Original = uint64(original.IntPart())

					special, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
							"fn", "buyerReturnOrderDetailListHandler",
							"special", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Special.Amount,
							"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
							"pid", orderList[i].Packages[j].Subpackages[z].PId,
							"sid", orderList[i].Packages[j].Subpackages[z].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					returnItemPackageDetail.Invoice.Special = uint64(special.IntPart())

					discount, err := decimal.NewFromString(orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
							"fn", "buyerReturnOrderDetailListHandler",
							"discount", orderList[i].Packages[j].Subpackages[z].Items[t].Invoice.Discount.Amount,
							"oid", orderList[i].Packages[j].Subpackages[z].OrderId,
							"pid", orderList[i].Packages[j].Subpackages[z].PId,
							"sid", orderList[i].Packages[j].Subpackages[z].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					returnItemPackageDetail.Invoice.Discount = uint64(discount.IntPart())

					if orderList[i].Packages[j].Subpackages[z].Items[t].Reasons != nil {
						returnItemPackageDetail.Reason = orderList[i].Packages[j].Subpackages[z].Items[t].Reasons[0].ToRPC()
					}

					if orderList[i].Packages[j].Subpackages[z].Shipments != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail != nil && orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt != nil {
						returnItemPackageDetail.ReturnRequestAt = orderList[i].Packages[j].Subpackages[z].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
					}

					if returnItemPackageDetail.Status == states.ReturnRequestPending.StateName() {
						returnItemPackageDetail.IsCancelable = true

					} else if returnItemPackageDetail.Status == states.ReturnShipmentPending.StateName() {
						returnItemPackageDetail.IsAccepted = true

					}

					returnItemPackageDetailList = append(returnItemPackageDetailList, returnItemPackageDetail)
				}

				returnPackageDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail{
					PID:      orderList[i].Packages[j].PId,
					ShopName: orderList[i].Packages[j].ShopName,
					Mobile:   "",
					Phone:    "",
					Shipment: nil,
					Items:    returnItemPackageDetailList,
				}

				if orderList[i].Packages[j].SellerInfo != nil {
					if orderList[i].Packages[j].SellerInfo.ReturnInfo != nil {
						returnPackageDetail.Shipment = &pb.BuyerReturnOrderDetailList_ReturnOrderDetail_ReturnPackageDetail_SellerReturnShipment{
							Country:      orderList[i].Packages[j].SellerInfo.ReturnInfo.Country,
							Province:     orderList[i].Packages[j].SellerInfo.ReturnInfo.Province,
							City:         orderList[i].Packages[j].SellerInfo.ReturnInfo.City,
							Neighborhood: orderList[i].Packages[j].SellerInfo.ReturnInfo.Neighborhood,
							Address:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalAddress,
							ZipCode:      orderList[i].Packages[j].SellerInfo.ReturnInfo.PostalCode}
					}
					if orderList[i].Packages[j].SellerInfo.GeneralInfo != nil {
						returnPackageDetail.Mobile = orderList[i].Packages[j].SellerInfo.GeneralInfo.MobilePhone
						returnPackageDetail.Phone = orderList[i].Packages[j].SellerInfo.GeneralInfo.LandPhone
					}
				}

				returnPackageDetailList = append(returnPackageDetailList, returnPackageDetail)
			}
		}

		returnOrderDetail := &pb.BuyerReturnOrderDetailList_ReturnOrderDetail{
			OID:                 orderList[i].OrderId,
			CreatedAt:           orderList[i].CreatedAt.Format(ISO8601),
			TotalAmount:         0,
			ReturnPackageDetail: returnPackageDetailList,
		}

		subtotal, err := decimal.NewFromString(orderList[i].Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
				"fn", "buyerReturnOrderDetailListHandler",
				"subtotal", orderList[i].Invoice.Subtotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		returnOrderDetail.TotalAmount = uint64(subtotal.IntPart())

		returnOrderDetailList = append(returnOrderDetailList, returnOrderDetail)
	}

	buyerReturnOrderDetailList := &pb.BuyerReturnOrderDetailList{
		BuyerId:           userId,
		ReturnOrderDetail: returnOrderDetailList,
	}

	serializedData, e := proto.Marshal(buyerReturnOrderDetailList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal buyerReturnOrderDetailList failed", "fn", "buyerReturnOrderDetailListHandler", "uid", userId, "filter", filter, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "buyerReturnOrderDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(total),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(buyerReturnOrderDetailList),
			Value:   serializedData,
		},
	}

	return response, nil

}
