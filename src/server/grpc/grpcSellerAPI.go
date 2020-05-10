package grpc_server

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	pb "gitlab.faza.io/protos/order"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	//pg "gitlab.faza.io/protos/payment-gateway"
)

func (server *Server) sellerGeneratePipelineFilter(ctx context.Context, filter FilterValue) []interface{} {

	newFilter := make([]interface{}, 2)

	if filter == AllOrdersFilter {
		newFilter[0] = "$or"
		filterList := make([]interface{}, 0, 30)
		for filter, _ := range server.sellerFilterStates {
			//if filter != PayToSellerFilter {
			filterQueryState := server.queryPathStates[filter]
			filterList = append(filterList, map[string]string{filterQueryState.queryPath: filterQueryState.state.StateName()})
			//}
		}

		newFilter[1] = bson.A(filterList)
	} else {
		if filter == AllCanceledFilter {
			queryPathCanceledBySellerState := server.queryPathStates[CanceledBySellerFilter]
			queryPathCanceledByBuyerState := server.queryPathStates[CanceledByBuyerFilter]
			newFilter[0] = "$or"
			newFilter[1] = bson.A{
				bson.M{queryPathCanceledBySellerState.queryPath: queryPathCanceledBySellerState.state.StateName()},
				bson.M{queryPathCanceledByBuyerState.queryPath: queryPathCanceledByBuyerState.state.StateName()}}

		} else {
			queryPathState := server.queryPathStates[filter]
			newFilter[0] = queryPathState.queryPath
			newFilter[1] = queryPathState.state.StateName()
		}
	}

	return newFilter
}

func (server *Server) sellerGetOrderByIdHandler(ctx context.Context, oid uint64, pid uint64, filter FilterValue) (*pb.MessageResponse, error) {
	genFilter := server.sellerGeneratePipelineFilter(ctx, filter)
	filters := make(bson.M, 4)
	filters["packages.orderId"] = oid
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil
	filters[genFilter[0].(string)] = genFilter[1]
	findFilter := func() interface{} {
		return []bson.M{
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$project": bson.M{"_id": 0, "packages": 1}},
			{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		}
	}

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, findFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("FindByFilter failed",
			"fn", "sellerGetOrderByIdHandler",
			"oid", oid,
			"pid", pid,
			"filterValue", filter,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	if pkgList == nil || len(pkgList) == 0 {
		app.Globals.Logger.FromContext(ctx).Error("Order not found",
			"fn", "sellerGetOrderByIdHandler",
			"oid", oid,
			"pid", pid,
			"filter", filter)
		return nil, status.Error(codes.Code(future.NotFound), "Order Not Found")
	}

	itemList := make([]*pb.SellerOrderList_ItemList, 0, 1)

	for _, pkgItem := range pkgList {
		sellerOrderItem := &pb.SellerOrderList_ItemList{
			OID:       pkgItem.OrderId,
			RequestAt: pkgItem.CreatedAt.Format(ISO8601),
			Amount:    0,
		}

		subtotal, err := decimal.NewFromString(pkgItem.Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid",
				"fn", "sellerGetOrderByIdHandler",
				"subtotal", pkgItem.Invoice.Subtotal.Amount,
				"oid", pkgItem.OrderId,
				"pid", pkgItem.PId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		sellerOrderItem.Amount = uint64(subtotal.IntPart())
		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		PID:   pid,
		Items: itemList,
	}

	serializedData, e := proto.Marshal(sellerOrderList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderList failed", "fn", "sellerGetOrderByIdHandler", "sellerOrderList", sellerOrderList, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderList",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderListHandler(ctx context.Context, oid, pid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	if oid > 0 {
		return server.sellerGetOrderByIdHandler(ctx, oid, pid, filter)
	}

	if page <= 0 || perPage <= 0 {
		app.Globals.Logger.FromContext(ctx).Error("page or perPage invalid",
			"fn", "sellerOrderListHandler",
			"oid", oid,
			"pid", pid,
			"filterValue", filter,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "Page/PerPage Invalid")
	}

	if filter == AllOrdersFilter {
		return server.sellerAllOrdersHandler(ctx, pid, page, perPage, sortName, direction)
	}

	genFilter := server.sellerGeneratePipelineFilter(ctx, filter)
	filters := make(bson.M, 3)
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil
	filters[genFilter[0].(string)] = genFilter[1]
	countFilter := func() interface{} {
		return []bson.M{
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	var totalCount, err = app.Globals.PkgItemRepository.CountWithFilter(ctx, countFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter failed",
			"fn", "sellerOrderListHandler",
			"oid", oid,
			"pid", pid,
			"filterValue", filter,
			"page", page,
			"perPage", perPage,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	if totalCount == 0 {
		app.Globals.Logger.FromContext(ctx).Info("total count is zero",
			"fn", "sellerOrderListHandler",
			"oid", oid,
			"pid", pid,
			"filterValue", filter,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%int64(perPage) != 0 {
		availablePages = (totalCount / int64(perPage)) + 1
	} else {
		availablePages = totalCount / int64(perPage)
	}

	if totalCount < int64(perPage) {
		availablePages = 1
	}

	if availablePages < int64(page) {
		app.Globals.Logger.FromContext(ctx).Error("availablePages less than page",
			"fn", "sellerOrderListHandler",
			"availablePages", availablePages,
			"oid", oid,
			"pid", pid,
			"filterValue", filter,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Package Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		app.Globals.Logger.FromContext(ctx).Error("offset invalid",
			"fn", "sellerOrderListHandler",
			"offset", offset,
			"oid", oid,
			"pid", pid,
			"filterValue", filter,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var pkgFilter func() interface{}

	if sortName != "" {
		var sortDirect int
		if direction == "ASC" {
			sortDirect = 1
		} else {
			sortDirect = -1
		}

		pkgFilter = func() interface{} {
			return []bson.M{
				{"$match": filters},
				{"$unwind": "$packages"},
				{"$match": filters},
				{"$project": bson.M{"_id": 0, "packages": 1}},
				{"$sort": bson.M{"packages.subpackages." + sortName: sortDirect}},
				{"$skip": offset},
				{"$limit": perPage},
				{"$replaceRoot": bson.M{"newRoot": "$packages"}},
			}
		}
	} else {
		pkgFilter = func() interface{} {
			return []bson.M{
				{"$match": filters},
				{"$unwind": "$packages"},
				{"$match": filters},
				{"$project": bson.M{"_id": 0, "packages": 1}},
				{"$skip": offset},
				{"$limit": perPage},
				{"$replaceRoot": bson.M{"newRoot": "$packages"}},
			}
		}

	}

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("FindByFilter failed",
			"oid", oid,
			"pid", pid,
			"filterValue", filter,
			"page", page,
			"perPage", perPage,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	itemList := make([]*pb.SellerOrderList_ItemList, 0, perPage)

	for _, pkgItem := range pkgList {
		sellerOrderItem := &pb.SellerOrderList_ItemList{
			OID:       pkgItem.OrderId,
			RequestAt: pkgItem.CreatedAt.Format(ISO8601),
			Amount:    0,
		}
		subtotal, err := decimal.NewFromString(pkgItem.Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid",
				"fn", "sellerOrderListHandler",
				"subtotal", pkgItem.Invoice.Subtotal.Amount,
				"oid", pkgItem.OrderId,
				"pid", pkgItem.PId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

		}
		sellerOrderItem.Amount = uint64(subtotal.IntPart())

		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		PID:   pid,
		Items: itemList,
	}

	serializedData, e := proto.Marshal(sellerOrderList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderList failed",
			"fn", "sellerOrderListHandler",
			"sellerOrderList", sellerOrderList, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(totalCount),
		Page:    page,
		PerPage: perPage,
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderList",
		Meta:   meta,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerAllOrdersHandler(ctx context.Context, pid uint64, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	if page <= 0 || perPage <= 0 {
		app.Globals.Logger.FromContext(ctx).Error("page or perPage invalid",
			"fn", "sellerAllOrdersHandler",
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "Page/PerPage Invalid")
	}

	filters := make(bson.M, 3)
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil

	var criteria = make([]interface{}, 0, len(server.sellerFilterStates))
	for filter, _ := range server.sellerFilterStates {
		//if filter == DeliveryPendingFilter {
		//	criteria = append(criteria, map[string]string{
		//		server.queryPathStates[DeliveryPendingFilter].queryPath: server.queryPathStates[DeliveryPendingFilter].state.StateName(),
		//	})
		//	criteria = append(criteria, map[string]string{
		//		server.queryPathStates[DeliveryDelayedFilter].queryPath: server.queryPathStates[DeliveryDelayedFilter].state.StateName(),
		//	})
		//} else if filter != AllCanceledFilter {
		if filter != AllCanceledFilter {
			criteria = append(criteria, map[string]string{
				server.queryPathStates[filter].queryPath: server.queryPathStates[filter].state.StateName(),
			})
		}
	}
	filters["$or"] = bson.A(criteria)
	countFilter := func() interface{} {
		return []bson.M{
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	var totalCount, err = app.Globals.PkgItemRepository.CountWithFilter(ctx, countFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter failed",
			"fn", "sellerAllOrdersHandler",
			"pid", pid,
			"page", page,
			"perPage", perPage,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	if totalCount == 0 {
		app.Globals.Logger.FromContext(ctx).Info("total count is zero",
			"fn", "sellerAllOrdersHandler",
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%int64(perPage) != 0 {
		availablePages = (totalCount / int64(perPage)) + 1
	} else {
		availablePages = totalCount / int64(perPage)
	}

	if totalCount < int64(perPage) {
		availablePages = 1
	}

	if availablePages < int64(page) {
		app.Globals.Logger.FromContext(ctx).Error("availablePages less than page",
			"fn", "sellerAllOrdersHandler",
			"availablePages", availablePages,
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Package Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		app.Globals.Logger.FromContext(ctx).Error("offset invalid",
			"fn", "sellerAllOrdersHandler",
			"offset", offset,
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var pkgFilter func() interface{}
	if sortName != "" {
		var sortDirect int
		if direction == "ASC" {
			sortDirect = 1
		} else {
			sortDirect = -1
		}

		pkgFilter = func() interface{} {
			return []bson.M{
				{"$match": filters},
				{"$unwind": "$packages"},
				{"$match": filters},
				{"$project": bson.M{"_id": 0, "packages": 1}},
				{"$sort": bson.M{"packages.subpackages." + sortName: sortDirect}},
				{"$skip": offset},
				{"$limit": perPage},
				{"$replaceRoot": bson.M{"newRoot": "$packages"}},
			}
		}
	} else {
		pkgFilter = func() interface{} {
			return []bson.M{
				{"$match": filters},
				{"$unwind": "$packages"},
				{"$match": filters},
				{"$project": bson.M{"_id": 0, "packages": 1}},
				{"$skip": offset},
				{"$limit": perPage},
				{"$replaceRoot": bson.M{"newRoot": "$packages"}},
			}
		}
	}

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("FindByFilter failed",
			"fn", "sellerAllOrdersHandler",
			"pid", pid,
			"page", page,
			"perPage", perPage,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	itemList := make([]*pb.SellerOrderList_ItemList, 0, perPage)

	for _, pkgItem := range pkgList {
		sellerOrderItem := &pb.SellerOrderList_ItemList{
			OID:       pkgItem.OrderId,
			RequestAt: pkgItem.CreatedAt.Format(ISO8601),
			Amount:    0,
		}
		subtotal, err := decimal.NewFromString(pkgItem.Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid",
				"fn", "sellerAllOrdersHandler",
				"subtotal", pkgItem.Invoice.Subtotal.Amount,
				"oid", pkgItem.OrderId,
				"pid", pkgItem.PId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

		}
		sellerOrderItem.Amount = uint64(subtotal.IntPart())

		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		PID:   pid,
		Items: itemList,
	}

	serializedData, e := proto.Marshal(sellerOrderList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderList failed",
			"fn", "sellerAllOrdersHandler",
			"sellerOrderList", sellerOrderList, "error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(totalCount),
		Page:    page,
		PerPage: perPage,
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderList",
		Meta:   meta,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderDetailHandler(ctx context.Context, pid, oid uint64, filter FilterValue) (*pb.MessageResponse, error) {

	pkgItem, buyerId, err := app.Globals.PkgItemRepository.FindPkgItmBuyinfById(ctx, oid, pid)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("PkgItemRepository.FindById failed",
			"fn", "sellerOrderDetailHandler",
			"oid", oid,
			"pid", pid,
			"filter", filter,
			"error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderDetailItems := make([]*pb.SellerOrderDetail_ItemDetail, 0, 32)
	for i := 0; i < len(pkgItem.Subpackages); i++ {
		if filter == AllOrdersFilter {
			for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
				var statusName string
				if stateList, ok := server.sellerStatesMap[pkgItem.Subpackages[i].Status]; ok {
					if len(stateList) == 1 {
						statusName = stateList[0].StateName()
					} else {
						length := len(pkgItem.Subpackages[i].Tracking.History)
						for _, state := range stateList {
							if pkgItem.Subpackages[i].Tracking.History[length-2].Name == state.StateName() {
								statusName = state.StateName()
								break
							}
						}
					}
				}
				itemDetail := &pb.SellerOrderDetail_ItemDetail{
					SID:         pkgItem.Subpackages[i].SId,
					Sku:         pkgItem.Subpackages[i].Items[j].SKU,
					Status:      statusName,
					SIdx:        int32(states.FromString(pkgItem.Subpackages[i].Status).StateIndex()),
					InventoryId: pkgItem.Subpackages[i].Items[j].InventoryId,
					Title:       pkgItem.Subpackages[i].Items[j].Title,
					Brand:       pkgItem.Subpackages[i].Items[j].Brand,
					Category:    pkgItem.Subpackages[i].Items[j].Category,
					Guaranty:    pkgItem.Subpackages[i].Items[j].Guaranty,
					Image:       pkgItem.Subpackages[i].Items[j].Image,
					Returnable:  pkgItem.Subpackages[i].Items[j].Returnable,
					Quantity:    pkgItem.Subpackages[i].Items[j].Quantity,
					Attributes:  nil,
					Invoice: &pb.SellerOrderDetail_ItemDetail_Invoice{
						Unit:             0,
						Total:            0,
						Original:         0,
						Special:          0,
						Discount:         0,
						SellerCommission: pkgItem.Subpackages[i].Items[j].Invoice.SellerCommission,
					},
					ShipmentDetail: nil,
				}

				if pkgItem.Subpackages[i].Shipments != nil && pkgItem.Subpackages[i].Shipments.ShipmentDetail != nil {
					itemDetail.ShipmentDetail = &pb.SellerOrderDetail_ItemDetail_ShipmentInfo{
						CourierName:    pkgItem.Subpackages[i].Shipments.ShipmentDetail.CourierName,
						ShippingMethod: pkgItem.Subpackages[i].Shipments.ShipmentDetail.ShippingMethod,
						TrackingNumber: pkgItem.Subpackages[i].Shipments.ShipmentDetail.TrackingNumber,
						Image:          pkgItem.Subpackages[i].Shipments.ShipmentDetail.Image,
						Description:    pkgItem.Subpackages[i].Shipments.ShipmentDetail.Description,
						ShippedAt:      "",
						CreatedAt:      pkgItem.Subpackages[i].Shipments.ShipmentDetail.CreatedAt.Format(ISO8601),
						UpdatedAt:      "",
					}

					if pkgItem.Subpackages[i].Shipments.ShipmentDetail.ShippedAt != nil {
						itemDetail.ShipmentDetail.ShippedAt = pkgItem.Subpackages[i].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601)
					}

					if pkgItem.Subpackages[i].Shipments.ShipmentDetail.UpdatedAt != nil {
						itemDetail.ShipmentDetail.UpdatedAt = pkgItem.Subpackages[i].Shipments.ShipmentDetail.UpdatedAt.Format(ISO8601)
					}
				}

				if pkgItem.Subpackages[i].Items[j].Attributes != nil {
					itemDetail.Attributes = make(map[string]*pb.SellerOrderDetail_ItemDetail_Attribute, len(pkgItem.Subpackages[i].Items[j].Attributes))
					for attrKey, attribute := range pkgItem.Subpackages[i].Items[j].Attributes {
						keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
						for keyTran, value := range attribute.KeyTranslate {
							keyTranslates[keyTran] = value
						}
						valTranslates := make(map[string]string, len(attribute.ValueTranslate))
						for valTran, value := range attribute.ValueTranslate {
							valTranslates[valTran] = value
						}
						itemDetail.Attributes[attrKey] = &pb.SellerOrderDetail_ItemDetail_Attribute{
							KeyTranslates:   keyTranslates,
							ValueTranslates: valTranslates,
						}
					}
				}

				unit, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
						"fn", "sellerOrderDetailHandler",
						"unit", pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount,
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", pkgItem.Subpackages[i].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemDetail.Invoice.Unit = uint64(unit.IntPart())

				total, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
						"fn", "sellerOrderDetailHandler",
						"total", pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount,
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", pkgItem.Subpackages[i].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemDetail.Invoice.Total = uint64(total.IntPart())

				original, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
						"fn", "sellerOrderDetailHandler",
						"original", pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount,
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", pkgItem.Subpackages[i].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemDetail.Invoice.Original = uint64(original.IntPart())

				special, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
						"fn", "sellerOrderDetailHandler",
						"special", pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount,
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", pkgItem.Subpackages[i].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemDetail.Invoice.Special = uint64(special.IntPart())

				discount, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
						"fn", "sellerOrderDetailHandler",
						"discount", pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount,
						"oid", pkgItem.OrderId,
						"pid", pkgItem.PId,
						"sid", pkgItem.Subpackages[i].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				itemDetail.Invoice.Discount = uint64(discount.IntPart())

				sellerOrderDetailItems = append(sellerOrderDetailItems, itemDetail)
			}
		} else {
			for _, filterState := range server.sellerFilterStates[filter] {
				if pkgItem.Subpackages[i].Status == filterState.actualState.StateName() {
					var statusName string
					if len(filterState.expectedState) == 1 {
						statusName = filterState.expectedState[0].StateName()
					} else {
						length := len(pkgItem.Subpackages[i].Tracking.History)
						for _, state := range filterState.expectedState {
							if pkgItem.Subpackages[i].Tracking.History[length-2].Name == state.StateName() {
								statusName = state.StateName()
								break
							}
						}
					}

					for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
						itemDetail := &pb.SellerOrderDetail_ItemDetail{
							SID:         pkgItem.Subpackages[i].SId,
							Sku:         pkgItem.Subpackages[i].Items[j].SKU,
							Status:      statusName,
							SIdx:        int32(states.FromString(pkgItem.Subpackages[i].Status).StateIndex()),
							InventoryId: pkgItem.Subpackages[i].Items[j].InventoryId,
							Title:       pkgItem.Subpackages[i].Items[j].Title,
							Brand:       pkgItem.Subpackages[i].Items[j].Brand,
							Category:    pkgItem.Subpackages[i].Items[j].Category,
							Guaranty:    pkgItem.Subpackages[i].Items[j].Guaranty,
							Image:       pkgItem.Subpackages[i].Items[j].Image,
							Returnable:  pkgItem.Subpackages[i].Items[j].Returnable,
							Quantity:    pkgItem.Subpackages[i].Items[j].Quantity,
							Attributes:  nil,
							Invoice: &pb.SellerOrderDetail_ItemDetail_Invoice{
								Unit:             0,
								Total:            0,
								Original:         0,
								Special:          0,
								Discount:         0,
								SellerCommission: pkgItem.Subpackages[i].Items[j].Invoice.SellerCommission,
							},
							ShipmentDetail: nil,
						}

						if pkgItem.Subpackages[i].Shipments != nil && pkgItem.Subpackages[i].Shipments.ShipmentDetail != nil {
							itemDetail.ShipmentDetail = &pb.SellerOrderDetail_ItemDetail_ShipmentInfo{
								CourierName:    pkgItem.Subpackages[i].Shipments.ShipmentDetail.CourierName,
								ShippingMethod: pkgItem.Subpackages[i].Shipments.ShipmentDetail.ShippingMethod,
								TrackingNumber: pkgItem.Subpackages[i].Shipments.ShipmentDetail.TrackingNumber,
								Image:          pkgItem.Subpackages[i].Shipments.ShipmentDetail.Image,
								Description:    pkgItem.Subpackages[i].Shipments.ShipmentDetail.Description,
								ShippedAt:      "",
								CreatedAt:      pkgItem.Subpackages[i].Shipments.ShipmentDetail.CreatedAt.Format(ISO8601),
								UpdatedAt:      "",
							}

							if pkgItem.Subpackages[i].Shipments.ShipmentDetail.ShippedAt != nil {
								itemDetail.ShipmentDetail.ShippedAt = pkgItem.Subpackages[i].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601)
							}

							if pkgItem.Subpackages[i].Shipments.ShipmentDetail.UpdatedAt != nil {
								itemDetail.ShipmentDetail.UpdatedAt = pkgItem.Subpackages[i].Shipments.ShipmentDetail.UpdatedAt.Format(ISO8601)
							}
						}

						if pkgItem.Subpackages[i].Items[j].Attributes != nil {
							itemDetail.Attributes = make(map[string]*pb.SellerOrderDetail_ItemDetail_Attribute, len(pkgItem.Subpackages[i].Items[j].Attributes))
							for attrKey, attribute := range pkgItem.Subpackages[i].Items[j].Attributes {
								keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
								for keyTran, value := range attribute.KeyTranslate {
									keyTranslates[keyTran] = value
								}
								valTranslates := make(map[string]string, len(attribute.ValueTranslate))
								for valTran, value := range attribute.ValueTranslate {
									valTranslates[valTran] = value
								}
								itemDetail.Attributes[attrKey] = &pb.SellerOrderDetail_ItemDetail_Attribute{
									KeyTranslates:   keyTranslates,
									ValueTranslates: valTranslates,
								}
							}
						}

						unit, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
								"fn", "sellerOrderDetailHandler",
								"unit", pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount,
								"oid", pkgItem.OrderId,
								"pid", pkgItem.PId,
								"sid", pkgItem.Subpackages[i].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemDetail.Invoice.Unit = uint64(unit.IntPart())

						total, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
								"fn", "sellerOrderDetailHandler",
								"total", pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount,
								"oid", pkgItem.OrderId,
								"pid", pkgItem.PId,
								"sid", pkgItem.Subpackages[i].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemDetail.Invoice.Total = uint64(total.IntPart())

						original, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
								"fn", "sellerOrderDetailHandler",
								"original", pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount,
								"oid", pkgItem.OrderId,
								"pid", pkgItem.PId,
								"sid", pkgItem.Subpackages[i].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemDetail.Invoice.Original = uint64(original.IntPart())

						special, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
								"fn", "sellerOrderDetailHandler",
								"special", pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount,
								"oid", pkgItem.OrderId,
								"pid", pkgItem.PId,
								"sid", pkgItem.Subpackages[i].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemDetail.Invoice.Special = uint64(special.IntPart())

						discount, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount)
						if err != nil {
							app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
								"fn", "sellerOrderDetailHandler",
								"discount", pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount,
								"oid", pkgItem.OrderId,
								"pid", pkgItem.PId,
								"sid", pkgItem.Subpackages[i].SId,
								"error", err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemDetail.Invoice.Discount = uint64(discount.IntPart())

						sellerOrderDetailItems = append(sellerOrderDetailItems, itemDetail)
					}
				}
			}
		}
	}

	if len(sellerOrderDetailItems) == 0 {
		return nil, status.Error(codes.Code(future.NotFound), "Order Item Not Found")
	}

	sellerOrderDetail := &pb.SellerOrderDetail{
		OID:       oid,
		PID:       pid,
		Amount:    0,
		RequestAt: pkgItem.CreatedAt.Format(ISO8601),
		Address: &pb.SellerOrderDetail_ShipmentAddress{
			FirstName:     pkgItem.ShippingAddress.FirstName,
			LastName:      pkgItem.ShippingAddress.LastName,
			Address:       pkgItem.ShippingAddress.Address,
			Phone:         pkgItem.ShippingAddress.Phone,
			Mobile:        pkgItem.ShippingAddress.Mobile,
			Country:       pkgItem.ShippingAddress.Country,
			City:          pkgItem.ShippingAddress.City,
			Province:      pkgItem.ShippingAddress.Province,
			Neighbourhood: pkgItem.ShippingAddress.Neighbourhood,
			Lat:           "",
			Long:          "",
			ZipCode:       pkgItem.ShippingAddress.ZipCode,
			BuyerId: 	   buyerId,
		},
		Items: sellerOrderDetailItems,
	}

	subtotal, e := decimal.NewFromString(pkgItem.Invoice.Subtotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid",
			"fn", "sellerOrderDetailHandler",
			"subtotal", pkgItem.Invoice.Subtotal.Amount,
			"oid", pkgItem.OrderId,
			"pid", pkgItem.PId,
			"error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}
	sellerOrderDetail.Amount = uint64(subtotal.IntPart())

	if pkgItem.ShippingAddress.Location != nil {
		sellerOrderDetail.Address.Lat = strconv.Itoa(int(pkgItem.ShippingAddress.Location.Coordinates[0]))
		sellerOrderDetail.Address.Long = strconv.Itoa(int(pkgItem.ShippingAddress.Location.Coordinates[1]))
	}

	serializedData, e := proto.Marshal(sellerOrderDetail)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderDetail failed",
			"fn", "sellerOrderDetailHandler",
			"sellerOrderDetail", sellerOrderDetail, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderDetail",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderDetail),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerReturnOrderListHandler(ctx context.Context, pid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		app.Globals.Logger.FromContext(ctx).Error("page or perPage invalid",
			"fn", "sellerReturnOrderListHandler",
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "Page/PerPage Invalid")
	}

	genFilter := server.sellerGeneratePipelineFilter(ctx, filter)
	filters := make(bson.M, 3)
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil
	filters[genFilter[0].(string)] = genFilter[1]
	countFilter := func() interface{} {
		return []bson.M{
			{"$match": filters},
			{"$unwind": "$packages"},
			{"$match": filters},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	var totalCount, err = app.Globals.PkgItemRepository.CountWithFilter(ctx, countFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter failed",
			"fn", "sellerReturnOrderListHandler",
			"pid", pid,
			"page", page,
			"perPage", perPage,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	if totalCount == 0 {
		app.Globals.Logger.FromContext(ctx).Info("total count is zero",
			"fn", "sellerReturnOrderListHandler",
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	// total 160 page=6 perPage=30
	var availablePages int64

	if totalCount%int64(perPage) != 0 {
		availablePages = (totalCount / int64(perPage)) + 1
	} else {
		availablePages = totalCount / int64(perPage)
	}

	if totalCount < int64(perPage) {
		availablePages = 1
	}

	if availablePages < int64(page) {
		app.Globals.Logger.FromContext(ctx).Error("availablePages less than page",
			"fn", "sellerReturnOrderListHandler",
			"availablePages", availablePages,
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Package Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		app.Globals.Logger.FromContext(ctx).Error("offset invalid",
			"fn", "sellerReturnOrderListHandler",
			"offset", offset,
			"pid", pid,
			"page", page,
			"perPage", perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var pkgFilter func() interface{}
	if sortName != "" {
		var sortDirect int
		if direction == "ASC" {
			sortDirect = 1
		} else {
			sortDirect = -1
		}

		pkgFilter = func() interface{} {
			return []bson.M{
				{"$match": filters},
				{"$unwind": "$packages"},
				{"$match": filters},
				{"$project": bson.M{"_id": 0, "packages": 1}},
				{"$sort": bson.M{"packages.subpackages." + sortName: sortDirect}},
				{"$skip": offset},
				{"$limit": perPage},
				{"$replaceRoot": bson.M{"newRoot": "$packages"}},
			}
		}
	} else {
		pkgFilter = func() interface{} {
			return []bson.M{
				{"$match": filters},
				{"$unwind": "$packages"},
				{"$match": filters},
				{"$project": bson.M{"_id": 0, "packages": 1}},
				{"$skip": offset},
				{"$limit": perPage},
				{"$replaceRoot": bson.M{"newRoot": "$packages"}},
			}
		}
	}

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("FindByFilter failed",
			"fn", "sellerReturnOrderListHandler",
			"pid", pid,
			"filterValue", filter,
			"page", page,
			"perPage", perPage,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	itemList := make([]*pb.SellerReturnOrderList_ItemList, 0, len(pkgList))

	for i := 0; i < len(pkgList); i++ {
		//for j := 0; j < len(pkgList[i].Subpackages); j++ {
		//	for _, filterState := range server.sellerFilterStates[filter] {
		//		if pkgList[i].Subpackages[j].Status == filterState.actualState.StateName() {
		//			for z := 0; z < len(pkgList[i].Subpackages[j].Items); z++ {
		item := &pb.SellerReturnOrderList_ItemList{
			OID: pkgList[i].OrderId,
			//SID: pkgList[i].Subpackages[j].SId,
			//Sku:             pkgList[i].Subpackages[j].Items[z].SKU,
			//InventoryId:     pkgList[i].Subpackages[j].Items[z].InventoryId,
			RequestAt: pkgList[i].CreatedAt.Format(ISO8601),
			//ReturnRequestAt: "",
			Amount: 0,
			//Title:           pkgList[i].Subpackages[j].Items[z].Title,
			//Brand:           pkgList[i].Subpackages[j].Items[z].Brand,
			//Category:        pkgList[i].Subpackages[j].Items[z].Category,
			//Quantity:        pkgList[i].Subpackages[j].Items[z].Quantity,
			//Address: &pb.SellerReturnOrderList_ItemList_ShipmentAddress{
			//	FirstName:     pkgList[i].ShippingAddress.FirstName,
			//	LastName:      pkgList[i].ShippingAddress.LastName,
			//	Address:       pkgList[i].ShippingAddress.Address,
			//	Phone:         pkgList[i].ShippingAddress.Phone,
			//	Mobile:        pkgList[i].ShippingAddress.Mobile,
			//	Country:       pkgList[i].ShippingAddress.Country,
			//	City:          pkgList[i].ShippingAddress.City,
			//	Province:      pkgList[i].ShippingAddress.Province,
			//	Neighbourhood: pkgList[i].ShippingAddress.Neighbourhood,
			//}
		}

		amount, err := decimal.NewFromString(pkgList[i].Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, package Invoice.Subtotal invalid",
				"fn", "sellerReturnOrderListHandler",
				"subtotal", pkgList[i].Invoice.Subtotal.Amount,
				"oid", pkgList[i].OrderId,
				"pid", pkgList[i].PId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		item.Amount = uint64(amount.IntPart())
		itemList = append(itemList, item)
		//		}
		//	}
		//}
		//}

		//if len(itemList) == 0 {
		//	app.Globals.Logger.FromContext(ctx).Error("get item from orderList failed",
		//		"fn", "sellerReturnOrderListHandler",
		//		"oid", pkgList[i].OrderId,
		//		"pid", pid,
		//		"filterValue", filter,
		//		"page", page,
		//		"perPage", perPage)
		//	return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		//}
	}

	sellerReturnOrderDetail := &pb.SellerReturnOrderList{
		PID:   pid,
		Items: itemList,
	}

	serializedData, e := proto.Marshal(sellerReturnOrderDetail)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerReturnOrderDetail failed",
			"fn", "sellerReturnOrderListHandler",
			"ReturnOrderDetail", sellerReturnOrderDetail, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerReturnOrderList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(totalCount),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerReturnOrderDetail),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerReturnOrderDetailHandler(ctx context.Context, pid, oid uint64, filter FilterValue) (*pb.MessageResponse, error) {
	pkgItem, buyerId, err := app.Globals.PkgItemRepository.FindPkgItmBuyinfById(ctx, oid, pid)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("PkgItemRepository.FindById failed",
			"fn", "sellerReturnOrderDetailHandler",
			"oid", oid,
			"pid", pid,
			"filter", filter,
			"error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerReturnOrderDetailItems := make([]*pb.SellerReturnOrderDetail_Item, 0, 32)
	for i := 0; i < len(pkgItem.Subpackages); i++ {
		for _, filterState := range server.sellerFilterStates[filter] {
			if pkgItem.Subpackages[i].Status == filterState.actualState.StateName() {
				var statusName string
				if len(filterState.expectedState) == 1 {
					statusName = filterState.expectedState[0].StateName()
				} else {
					length := len(pkgItem.Subpackages[i].Tracking.History)
					for _, state := range filterState.expectedState {
						if pkgItem.Subpackages[i].Tracking.History[length-2].Name == state.StateName() {
							statusName = state.StateName()
							break
						}
					}
				}

				for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
					item := &pb.SellerReturnOrderDetail_Item{
						SID:    pkgItem.Subpackages[i].SId,
						Sku:    pkgItem.Subpackages[i].Items[j].SKU,
						Status: statusName,
						SIdx:   int32(states.FromString(pkgItem.Subpackages[i].Status).StateIndex()),
						Detail: &pb.SellerReturnOrderDetail_Item_Detail{
							InventoryId: pkgItem.Subpackages[i].Items[j].InventoryId,
							Title:       pkgItem.Subpackages[i].Items[j].Title,
							Brand:       pkgItem.Subpackages[i].Items[j].Brand,
							Category:    pkgItem.Subpackages[i].Items[j].Category,
							Guaranty:    pkgItem.Subpackages[i].Items[j].Guaranty,
							Image:       pkgItem.Subpackages[i].Items[j].Image,
							Returnable:  pkgItem.Subpackages[i].Items[j].Returnable,
							Quantity:    pkgItem.Subpackages[i].Items[j].Quantity,
							Attributes:  nil,
							Invoice: &pb.SellerReturnOrderDetail_Item_Detail_Invoice{
								Unit:             0,
								Total:            0,
								Original:         0,
								Special:          0,
								Discount:         0,
								SellerCommission: pkgItem.Subpackages[i].Items[j].Invoice.SellerCommission,
								Currency:         "IRR",
							},
							ReturnShipmentDetail: nil,
						},
					}

					if pkgItem.Subpackages[i].Shipments != nil && pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail != nil {
						item.Detail.ReturnShipmentDetail = &pb.SellerReturnOrderDetail_Item_Detail_ReturnShipmentInfo{
							CourierName:    pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.CourierName,
							ShippingMethod: pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.ShippingMethod,
							TrackingNumber: pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.TrackingNumber,
							Image:          pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.Image,
							Description:    pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.Description,
							ShippedAt:      "",
							RequestedAt:    "",
							CreatedAt:      pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.CreatedAt.Format(ISO8601),
							UpdatedAt:      "",
						}

						if pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.ShippedAt != nil {
							item.Detail.ReturnShipmentDetail.ShippedAt = pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.ShippedAt.Format(ISO8601)
						}

						if pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.RequestedAt != nil {
							item.Detail.ReturnShipmentDetail.ShippedAt = pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
						}

						if pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.UpdatedAt != nil {
							item.Detail.ReturnShipmentDetail.UpdatedAt = pkgItem.Subpackages[i].Shipments.ReturnShipmentDetail.UpdatedAt.Format(ISO8601)
						}
					}

					if pkgItem.Subpackages[i].Items[j].Attributes != nil {
						item.Detail.Attributes = make(map[string]*pb.SellerReturnOrderDetail_Item_Detail_Attribute, len(pkgItem.Subpackages[i].Items[j].Attributes))
						for attrKey, attribute := range pkgItem.Subpackages[i].Items[j].Attributes {
							keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
							for keyTran, value := range attribute.KeyTranslate {
								keyTranslates[keyTran] = value
							}
							valTranslates := make(map[string]string, len(attribute.ValueTranslate))
							for valTran, value := range attribute.ValueTranslate {
								valTranslates[valTran] = value
							}
							item.Detail.Attributes[attrKey] = &pb.SellerReturnOrderDetail_Item_Detail_Attribute{
								KeyTranslates:   keyTranslates,
								ValueTranslates: valTranslates,
							}
						}
					}

					unit, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
							"fn", "sellerReturnOrderDetailHandler",
							"unit", pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount,
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", pkgItem.Subpackages[i].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					item.Detail.Invoice.Unit = uint64(unit.IntPart())

					total, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
							"fn", "sellerReturnOrderDetailHandler",
							"total", pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount,
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", pkgItem.Subpackages[i].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					item.Detail.Invoice.Total = uint64(total.IntPart())

					original, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
							"fn", "sellerReturnOrderDetailHandler",
							"original", pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount,
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", pkgItem.Subpackages[i].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					item.Detail.Invoice.Original = uint64(original.IntPart())

					special, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
							"fn", "sellerReturnOrderDetailHandler",
							"special", pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount,
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", pkgItem.Subpackages[i].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					item.Detail.Invoice.Special = uint64(special.IntPart())

					discount, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
							"fn", "sellerReturnOrderDetailHandler",
							"discount", pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount,
							"oid", pkgItem.OrderId,
							"pid", pkgItem.PId,
							"sid", pkgItem.Subpackages[i].SId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					item.Detail.Invoice.Discount = uint64(discount.IntPart())

					sellerReturnOrderDetailItems = append(sellerReturnOrderDetailItems, item)
				}
			}
		}
	}

	if len(sellerReturnOrderDetailItems) == 0 {
		return nil, status.Error(codes.Code(future.NotFound), "Return order Item Not Found")
	}

	sellerReturnOrderDetail := &pb.SellerReturnOrderDetail{
		PID:       pid,
		OID:       oid,
		Amount:    0,
		RequestAt: pkgItem.CreatedAt.Format(ISO8601),
		Address: &pb.SellerReturnOrderDetail_ShipmentAddress{
			FirstName:     pkgItem.ShippingAddress.FirstName,
			LastName:      pkgItem.ShippingAddress.LastName,
			Address:       pkgItem.ShippingAddress.Address,
			Phone:         pkgItem.ShippingAddress.Phone,
			Mobile:        pkgItem.ShippingAddress.Mobile,
			Country:       pkgItem.ShippingAddress.Country,
			City:          pkgItem.ShippingAddress.City,
			Province:      pkgItem.ShippingAddress.Province,
			Neighbourhood: pkgItem.ShippingAddress.Neighbourhood,
			Lat:           "",
			Long:          "",
			ZipCode:       pkgItem.ShippingAddress.ZipCode,
			BuyerId:       buyerId,
		},
		Items: sellerReturnOrderDetailItems,
	}

	subtotal, e := decimal.NewFromString(pkgItem.Invoice.Subtotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid",
			"fn", "sellerReturnOrderDetailHandler",
			"subtotal", pkgItem.Invoice.Subtotal.Amount,
			"oid", pkgItem.OrderId,
			"pid", pkgItem.PId,
			"error", err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}
	sellerReturnOrderDetail.Amount = uint64(subtotal.IntPart())

	if pkgItem.ShippingAddress.Location != nil {
		sellerReturnOrderDetail.Address.Lat = strconv.Itoa(int(pkgItem.ShippingAddress.Location.Coordinates[0]))
		sellerReturnOrderDetail.Address.Long = strconv.Itoa(int(pkgItem.ShippingAddress.Location.Coordinates[1]))
	}

	serializedData, e := proto.Marshal(sellerReturnOrderDetail)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerReturnOrderDetail failed",
			"fn", "sellerReturnOrderDetailHandler",
			"sellerReturnOrderDetail", sellerReturnOrderDetail, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerReturnOrderDetail",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerReturnOrderDetail),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderDashboardReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathApprovalPendingState := server.queryPathStates[ApprovalPendingFilter]
	approvalPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathApprovalPendingState.queryPath: queryPathApprovalPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathApprovalPendingState.queryPath: queryPathApprovalPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShipmentPendingState := server.queryPathStates[ShipmentPendingFilter]
	shipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShipmentDelayedState := server.queryPathStates[ShipmentDelayedFilter]
	shipmentDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnRequestPendingState := server.queryPathStates[ReturnRequestPendingFilter]
	returnRequestPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnRequestPendingState.queryPath: queryPathReturnRequestPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnRequestPendingState.queryPath: queryPathReturnRequestPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	approvalPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, approvalPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for approvalPendingFilter failed",
			"fn", "sellerOrderDashboardReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	shipmentPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shipmentPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shipmentPendingFilter failed",
			"fn", "sellerOrderDashboardReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	shipmentDelayedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shipmentDelayedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shipmentDelayedFilter failed",
			"fn", "sellerOrderDashboardReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnRequestPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnRequestPendingFilter failed",
			"fn", "sellerOrderDashboardReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	sellerOrderDashboardReports := &pb.SellerOrderDashboardReports{
		SellerId:             userId,
		ApprovalPending:      uint32(approvalPendingCount),
		ShipmentPending:      uint32(shipmentPendingCount),
		ShipmentDelayed:      uint32(shipmentDelayedCount),
		ReturnRequestPending: uint32(returnRequestPendingCount),
	}

	serializedData, e := proto.Marshal(sellerOrderDashboardReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderDashboardReportsHandler failed", "fn", "sellerOrderDashboardReportsHandler", "uid", userId, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderDashboardReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderDashboardReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderShipmentReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathShipmentPendingState := server.queryPathStates[ShipmentPendingFilter]
	shipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShipmentDelayedState := server.queryPathStates[ShipmentDelayedFilter]
	shipmentDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShippedState := server.queryPathStates[ShippedFilter]
	shippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShippedState.queryPath: queryPathShippedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShippedState.queryPath: queryPathShippedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	shipmentPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shipmentPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shipmentPendingFilter failed",
			"fn", "sellerOrderShipmentReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	shipmentDelayedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shipmentDelayedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shipmentDelayedFilter failed",
			"fn", "sellerOrderShipmentReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	shippedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shippedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shippedFilter failed",
			"fn", "sellerOrderShipmentReportsHandler",
			"uid", userId,
			"error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	sellerOrderShipmentReports := &pb.SellerOrderShipmentReports{
		SellerId:        userId,
		ShipmentPending: uint32(shipmentPendingCount),
		ShipmentDelayed: uint32(shipmentDelayedCount),
		Shipped:         uint32(shippedCount),
	}

	serializedData, e := proto.Marshal(sellerOrderShipmentReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderShipmentReports failed", "fn", "sellerOrderShipmentReportsHandler", "uid", userId, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderShipmentReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderShipmentReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderReturnReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathRequestPendingState := server.queryPathStates[ReturnRequestPendingFilter]
	returnRequestPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestPendingState.queryPath: queryPathRequestPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestPendingState.queryPath: queryPathRequestPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathRequestRejectedState := server.queryPathStates[ReturnRequestRejectedFilter]
	returnRequestRejectedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestRejectedState.queryPath: queryPathRequestRejectedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestRejectedState.queryPath: queryPathRequestRejectedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnShipmentPendingState := server.queryPathStates[ReturnShipmentPendingFilter]
	returnShipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShipmentPendingState.queryPath: queryPathReturnShipmentPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShipmentPendingState.queryPath: queryPathReturnShipmentPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnShippedState := server.queryPathStates[ReturnShippedFilter]
	returnShippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShippedState.queryPath: queryPathReturnShippedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShippedState.queryPath: queryPathReturnShippedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveredState := server.queryPathStates[ReturnDeliveredFilter]
	returnDeliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveredState.queryPath: queryPathReturnDeliveredState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveredState.queryPath: queryPathReturnDeliveredState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveryPendingState := server.queryPathStates[ReturnDeliveryPendingFilter]
	returnDeliveryPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryPendingState.queryPath: queryPathReturnDeliveryPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryPendingState.queryPath: queryPathReturnDeliveryPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveryDelayedState := server.queryPathStates[ReturnDeliveryDelayedFilter]
	returnDeliveryDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryDelayedState.queryPath: queryPathReturnDeliveryDelayedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryDelayedState.queryPath: queryPathReturnDeliveryDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveryFailedState := server.queryPathStates[ReturnDeliveryFailedFilter]
	returnDeliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryFailedState.queryPath: queryPathReturnDeliveryFailedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryFailedState.queryPath: queryPathReturnDeliveryFailedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnRequestPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnRequestPendingFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnShipmentPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnShipmentPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnShipmentPendingFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnShippedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnShippedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnShippedFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveredCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveredFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveredFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveryFailedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveryFailedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveryFailedFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnRequestRejectedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnRequestRejectedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnRequestRejectedFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveryPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveryPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveryPendingFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveryDelayedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveryDelayedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveryDelayedFilter failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	sellerOrderReturnReports := &pb.SellerOrderReturnReports{
		SellerId:              userId,
		ReturnRequestPending:  uint32(returnRequestPendingCount),
		ReturnShipmentPending: uint32(returnShipmentPendingCount),
		ReturnShipped:         uint32(returnShippedCount),
		ReturnDeliveryPending: uint32(returnDeliveryPendingCount),
		ReturnDeliveryDelayed: uint32(returnDeliveryDelayedCount),
		ReturnDelivered:       uint32(returnDeliveredCount),
		ReturnRequestRejected: uint32(returnRequestRejectedCount),
		ReturnDeliveryFailed:  uint32(returnDeliveryFailedCount),
	}

	serializedData, e := proto.Marshal(sellerOrderReturnReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderReturnReports failed",
			"fn", "sellerOrderReturnReportsHandler",
			"uid", userId, "error")
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderReturnReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderReturnReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderDeliveredReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathDeliveryPendingState := server.queryPathStates[DeliveryPendingFilter]
	deliveryPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveryDelayedState := server.queryPathStates[DeliveryDelayedFilter]
	deliveryDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveredState := server.queryPathStates[DeliveredFilter]
	deliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveredState.queryPath: queryPathDeliveredState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveredState.queryPath: queryPathDeliveredState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveryFailedState := server.queryPathStates[DeliveryFailedFilter]
	deliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryFailedState.queryPath: queryPathDeliveryFailedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryFailedState.queryPath: queryPathDeliveryFailedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	deliveryPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveryPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveryPendingFilter failed",
			"fn", "sellerOrderDeliveredReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	deliveryDelayedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveryDelayedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveryDelayedFilter failed",
			"fn", "sellerOrderDeliveredReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	deliveredCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveredFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveredFilter failed",
			"fn", "sellerOrderDeliveredReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	deliveryFailedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveryFailedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveryFailedFilter failed",
			"fn", "sellerOrderDeliveredReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	sellerOrderDeliveredReports := &pb.SellerOrderDeliveredReports{
		SellerId:        userId,
		DeliveryPending: uint32(deliveryPendingCount),
		DeliveryDelayed: uint32(deliveryDelayedCount),
		Delivered:       uint32(deliveredCount),
		DeliveryFailed:  uint32(deliveryFailedCount),
	}

	serializedData, e := proto.Marshal(sellerOrderDeliveredReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderDeliveredReports failed",
			"fn", "sellerOrderDeliveredReportsHandler",
			"uid", userId, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderDeliveredReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderDeliveredReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerOrderCancelReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathCanceledByBuyerState := server.queryPathStates[CanceledByBuyerFilter]
	cancelByBuyerFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledByBuyerState.queryPath: queryPathCanceledByBuyerState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledByBuyerState.queryPath: queryPathCanceledByBuyerState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathCanceledBySellerState := server.queryPathStates[CanceledBySellerFilter]
	cancelBySellerFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledBySellerState.queryPath: queryPathCanceledBySellerState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledBySellerState.queryPath: queryPathCanceledBySellerState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	cancelByBuyerCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, cancelByBuyerFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for cancelByBuyerFilter failed",
			"fn", "sellerOrderCancelReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	cancelBySellerCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, cancelBySellerFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for cancelBySellerFilter failed",
			"fn", "sellerOrderCancelReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	sellerOrderCancelReports := &pb.SellerOrderCancelReports{
		SellerId:         userId,
		CanceledBySeller: uint32(cancelBySellerCount),
		CanceledByBuyer:  uint32(cancelByBuyerCount),
	}

	serializedData, e := proto.Marshal(sellerOrderCancelReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerOrderCancelReports failed",
			"fn", "sellerOrderCancelReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())

	}

	response := &pb.MessageResponse{
		Entity: "SellerOrderCancelReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerOrderCancelReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerAllOrderReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathApprovalPendingFilterState := server.queryPathStates[ApprovalPendingFilter]
	approvalPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathApprovalPendingFilterState.queryPath: queryPathApprovalPendingFilterState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathApprovalPendingFilterState.queryPath: queryPathApprovalPendingFilterState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShipmentPendingState := server.queryPathStates[ShipmentPendingFilter]
	shipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShipmentDelayedState := server.queryPathStates[ShipmentDelayedFilter]
	shipmentDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShippedState := server.queryPathStates[ShippedFilter]
	shippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShippedState.queryPath: queryPathShippedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShippedState.queryPath: queryPathShippedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathRequestPendingState := server.queryPathStates[ReturnRequestPendingFilter]
	returnRequestPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestPendingState.queryPath: queryPathRequestPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestPendingState.queryPath: queryPathRequestPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathRequestRejectedState := server.queryPathStates[ReturnRequestRejectedFilter]
	returnRequestRejectedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestRejectedState.queryPath: queryPathRequestRejectedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathRequestRejectedState.queryPath: queryPathRequestRejectedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnShipmentPendingState := server.queryPathStates[ReturnShipmentPendingFilter]
	returnShipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShipmentPendingState.queryPath: queryPathReturnShipmentPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShipmentPendingState.queryPath: queryPathReturnShipmentPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnShippedState := server.queryPathStates[ReturnShippedFilter]
	returnShippedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShippedState.queryPath: queryPathReturnShippedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnShippedState.queryPath: queryPathReturnShippedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveredState := server.queryPathStates[ReturnDeliveredFilter]
	returnDeliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveredState.queryPath: queryPathReturnDeliveredState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveredState.queryPath: queryPathReturnDeliveredState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveryPendingState := server.queryPathStates[ReturnDeliveryPendingFilter]
	returnDeliveryPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryPendingState.queryPath: queryPathReturnDeliveryPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryPendingState.queryPath: queryPathReturnDeliveryPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveryDelayedState := server.queryPathStates[ReturnDeliveryDelayedFilter]
	returnDeliveryDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryDelayedState.queryPath: queryPathReturnDeliveryDelayedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryDelayedState.queryPath: queryPathReturnDeliveryDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnDeliveryFailedState := server.queryPathStates[ReturnDeliveryFailedFilter]
	returnDeliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryFailedState.queryPath: queryPathReturnDeliveryFailedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryFailedState.queryPath: queryPathReturnDeliveryFailedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveryPendingState := server.queryPathStates[DeliveryPendingFilter]
	deliveryPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveryDelayedState := server.queryPathStates[DeliveryDelayedFilter]
	deliveryDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveredState := server.queryPathStates[DeliveredFilter]
	deliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveredState.queryPath: queryPathDeliveredState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveredState.queryPath: queryPathDeliveredState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveryFailedState := server.queryPathStates[DeliveryFailedFilter]
	deliveryFailedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryFailedState.queryPath: queryPathDeliveryFailedState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryFailedState.queryPath: queryPathDeliveryFailedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathCanceledByBuyerState := server.queryPathStates[CanceledByBuyerFilter]
	cancelByBuyerFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledByBuyerState.queryPath: queryPathCanceledByBuyerState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledByBuyerState.queryPath: queryPathCanceledByBuyerState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathCanceledBySellerState := server.queryPathStates[CanceledBySellerFilter]
	cancelBySellerFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledBySellerState.queryPath: queryPathCanceledBySellerState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledBySellerState.queryPath: queryPathCanceledBySellerState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	approvalPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, approvalPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for cancelByBuyerFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	shipmentPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shipmentPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shipmentPendingFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	shipmentDelayedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shipmentDelayedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shipmentDelayedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	shippedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shippedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for shippedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnRequestPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnRequestPendingFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnShipmentPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnShipmentPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnShipmentPendingFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnShippedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnShippedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnShippedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveredCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveredFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveredFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveryFailedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveryFailedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveryFailedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnRequestRejectedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnRequestRejectedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnRequestRejectedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveryPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveryPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveryPendingFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	returnDeliveryDelayedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, returnDeliveryDelayedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for returnDeliveryDelayedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	deliveryPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveryPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveryPendingFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	deliveryDelayedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveryDelayedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveryDelayedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	deliveredCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveredFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveredFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	deliveryFailedCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, deliveryFailedFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for deliveryFailedFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	canceledByBuyerCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, cancelByBuyerFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for cancelByBuyerFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	canceledBySellerCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, cancelBySellerFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for cancelBySellerFilter failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	sellerAllOrderReports := &pb.SellerAllOrderReports{
		SellerId:        userId,
		ApprovalPending: uint32(approvalPendingCount),
		ShipmentReports: &pb.SellerAllOrderReports_ShipmentReport{
			ShipmentPending: uint32(shipmentPendingCount),
			ShipmentDelayed: uint32(shipmentDelayedCount),
			Shipped:         uint32(shippedCount),
		},
		DeliverReports: &pb.SellerAllOrderReports_DeliverReport{
			DeliveryPending: uint32(deliveryPendingCount),
			DeliveryDelayed: uint32(deliveryDelayedCount),
			Delivered:       uint32(deliveredCount),
			DeliveryFailed:  uint32(deliveryFailedCount),
		},
		ReturnReports: &pb.SellerAllOrderReports_ReturnReport{
			ReturnRequestPending:  uint32(returnRequestPendingCount),
			ReturnShipmentPending: uint32(returnShipmentPendingCount),
			ReturnShipped:         uint32(returnShippedCount),
			ReturnDeliveryPending: uint32(returnDeliveryPendingCount),
			ReturnDeliveryDelayed: uint32(returnDeliveryDelayedCount),
			ReturnDelivered:       uint32(returnDeliveredCount),
			ReturnRequestRejected: uint32(returnRequestRejectedCount),
			ReturnDeliveryFailed:  uint32(returnDeliveryFailedCount),
		},
		CancelReport: &pb.SellerAllOrderReports_CancelReport{
			CanceledBySeller: uint32(canceledBySellerCount),
			CanceledByBuyer:  uint32(canceledByBuyerCount),
		},
	}

	serializedData, e := proto.Marshal(sellerAllOrderReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerAllOrderReports failed",
			"fn", "sellerAllOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())

	}

	response := &pb.MessageResponse{
		Entity: "SellerAllOrderReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerAllOrderReports),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) sellerApprovalPendingOrderReportsHandler(ctx context.Context, userId uint64) (*pb.MessageResponse, error) {

	queryPathApprovalPendingFilterState := server.queryPathStates[ApprovalPendingFilter]
	approvalPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathApprovalPendingFilterState.queryPath: queryPathApprovalPendingFilterState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathApprovalPendingFilterState.queryPath: queryPathApprovalPendingFilterState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	approvalPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, approvalPendingFilter)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("CountWithFilter for cancelByBuyerFilter failed",
			"fn", "sellerApprovalPendingOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	sellerApprovalPendingReports := &pb.SellerApprovalPendingReports{
		SellerId:        userId,
		ApprovalPending: uint32(approvalPendingCount),
	}

	serializedData, e := proto.Marshal(sellerApprovalPendingReports)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal sellerApprovalPendingReports failed",
			"fn", "sellerApprovalPendingOrderReportsHandler",
			"uid", userId, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())

	}

	response := &pb.MessageResponse{
		Entity: "SellerApprovalPendingReports",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerApprovalPendingReports),
			Value:   serializedData,
		},
	}

	return response, nil
}
