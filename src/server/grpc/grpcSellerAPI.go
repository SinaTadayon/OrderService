package grpc_server

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/go-framework/logger"
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

	if filter == DeliveryPendingFilter {
		queryPathDeliveryPendingState := server.queryPathStates[DeliveryPendingFilter]
		queryPathDeliveryDelayedState := server.queryPathStates[DeliveryDelayedFilter]
		newFilter[0] = "$or"
		newFilter[1] = bson.A{
			bson.M{queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()},
			bson.M{queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()}}

	} else if filter == AllCanceledFilter {
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

	return newFilter
}

func (server *Server) sellerGetOrderByIdHandler(ctx context.Context, oid uint64, pid uint64, filter FilterValue) (*pb.MessageResponse, error) {
	genFilter := server.sellerGeneratePipelineFilter(ctx, filter)
	filters := make(bson.M, 3)
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
		logger.Err("sellerGetOrderByIdHandler() => FindByFilter failed, pid: %d, filterValue: %s, error: %s", pid, filter, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if pkgList == nil || len(pkgList) == 0 {
		logger.Err("sellerGetOrderByIdHandler() => pid not found, orderId: %d, pid: %d, filter:%s", oid, pid, filter)
		return nil, status.Error(codes.Code(future.NotFound), "Pid Not Found")
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
			logger.Err("sellerGetOrderByIdHandler() => decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid, subtotal: %s, orderId: %d, pid: %d, error: %s",
				pkgItem.Invoice.Subtotal.Amount, pkgItem.OrderId, pkgItem.PId, err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		sellerOrderItem.Amount = uint64(subtotal.IntPart())
		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		PID:   pid,
		Items: itemList,
	}

	serializedData, err := proto.Marshal(sellerOrderList)
	if err != nil {
		logger.Err("sellerGetOrderByIdHandler() => could not serialize sellerOrderList, sellerOrderList: %v, error:%s", sellerOrderList, err)
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
		logger.Err("sellerOrderListHandler() => page or perPage invalid, pid: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
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
		logger.Err("sellerOrderListHandler() => CountWithFilter failed,  pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if totalCount == 0 {
		logger.Err("sellerOrderListHandler() => total count is zero,  pid: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
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
		logger.Err("sellerOrderListHandler() => availablePages less than page, availablePages: %d, pid: %d, filterValue: %s, page: %d, perPage: %d", availablePages, pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		logger.Err("sellerOrderListHandler() => offset invalid, offset: %d, pid: %d, filterValue: %s, page: %d, perPage: %d", offset, pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	pkgFilter := func() interface{} {
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

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		logger.Err("sellerOrderListHandler() => FindByFilter failed, pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
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
			logger.Err("sellerOrderListHandler() => decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid, subtotal: %s, orderId: %d, pid: %d, error: %s",
				pkgItem.Invoice.Subtotal.Amount, pkgItem.OrderId, pkgItem.PId, err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

		}
		sellerOrderItem.Amount = uint64(subtotal.IntPart())

		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		PID:   pid,
		Items: itemList,
	}

	serializedData, err := proto.Marshal(sellerOrderList)
	if err != nil {
		logger.Err("sellerOrderListHandler() => could not serialize sellerOrderList, sellerOrderList: %v, error:%s", sellerOrderList, err)
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
		logger.Err("sellerAllOrdersHandler() => page or perPage invalid, pid: %d, page: %d, perPage: %d", pid, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
	}

	filters := make(bson.M, 3)
	filters["packages.pid"] = pid
	filters["packages.deletedAt"] = nil

	var criteria = make([]interface{}, 0, len(server.sellerFilterStates))
	for filter, _ := range server.sellerFilterStates {
		if filter == DeliveryPendingFilter {
			criteria = append(criteria, map[string]string{
				server.queryPathStates[DeliveryPendingFilter].queryPath: server.queryPathStates[DeliveryPendingFilter].state.StateName(),
			})
			criteria = append(criteria, map[string]string{
				server.queryPathStates[DeliveryDelayedFilter].queryPath: server.queryPathStates[DeliveryDelayedFilter].state.StateName(),
			})
		} else if filter != AllCanceledFilter {
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
		logger.Err("sellerAllOrdersHandler() => CountWithFilter failed,  pid: %d, page: %d, perPage: %d, error: %s", pid, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if totalCount == 0 {
		logger.Err("sellerAllOrdersHandler() => total count is zero,  pid: %d, page: %d, perPage: %d", pid, page, perPage)
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
		logger.Err("sellerAllOrdersHandler() => availablePages less than page, availablePages: %d, pid: %d, page: %d, perPage: %d", availablePages, pid, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		logger.Err("sellerAllOrdersHandler() => offset invalid, offset: %d, pid: %d, page: %d, perPage: %d", offset, pid, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	pkgFilter := func() interface{} {
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

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		logger.Err("sellerAllOrdersHandler() => FindByFilter failed, pid: %d, page: %d, perPage: %d, error: %s", pid, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
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
			logger.Err("sellerAllOrdersHandler() => decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid, subtotal: %s, orderId: %d, pid: %d, error: %s",
				pkgItem.Invoice.Subtotal.Amount, pkgItem.OrderId, pkgItem.PId, err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

		}
		sellerOrderItem.Amount = uint64(subtotal.IntPart())

		itemList = append(itemList, sellerOrderItem)
	}

	sellerOrderList := &pb.SellerOrderList{
		PID:   pid,
		Items: itemList,
	}

	serializedData, err := proto.Marshal(sellerOrderList)
	if err != nil {
		logger.Err("sellerAllOrdersHandler() => could not serialize sellerOrderList, sellerOrderList: %v, error:%s", sellerOrderList, err)
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

func (server *Server) sellerOrderDetailHandler(ctx context.Context, pid, orderId uint64, filter FilterValue) (*pb.MessageResponse, error) {

	pkgItem, err := app.Globals.PkgItemRepository.FindById(ctx, orderId, pid)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => PkgItemRepository.FindById failed, orderId: %d, pid: %d, filter:%s , error: %s", orderId, pid, filter, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderDetailItems := make([]*pb.SellerOrderDetail_ItemDetail, 0, 32)
	for i := 0; i < len(pkgItem.Subpackages); i++ {
		for _, filterState := range server.sellerFilterStates[filter] {
			if pkgItem.Subpackages[i].Status == filterState.actualState.StateName() {
				for j := 0; j < len(pkgItem.Subpackages[i].Items); j++ {
					itemDetail := &pb.SellerOrderDetail_ItemDetail{
						SID:         pkgItem.Subpackages[i].SId,
						Sku:         pkgItem.Subpackages[i].Items[j].SKU,
						Status:      filterState.expectedState.StateName(),
						SIdx:        int32(states.FromString(pkgItem.Subpackages[i].Status).StateIndex()),
						InventoryId: pkgItem.Subpackages[i].Items[j].InventoryId,
						Title:       pkgItem.Subpackages[i].Items[j].Title,
						Brand:       pkgItem.Subpackages[i].Items[j].Brand,
						Category:    pkgItem.Subpackages[i].Items[j].Category,
						Guaranty:    pkgItem.Subpackages[i].Items[j].Guaranty,
						Image:       pkgItem.Subpackages[i].Items[j].Image,
						Returnable:  pkgItem.Subpackages[i].Items[j].Returnable,
						Quantity:    pkgItem.Subpackages[i].Items[j].Quantity,
						Attributes:  pkgItem.Subpackages[i].Items[j].Attributes,
						Invoice: &pb.SellerOrderDetail_ItemDetail_Invoice{
							Unit:             0,
							Total:            0,
							Original:         0,
							Special:          0,
							Discount:         0,
							SellerCommission: pkgItem.Subpackages[i].Items[j].Invoice.SellerCommission,
						},
					}

					unit, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount)
					if err != nil {
						logger.Err("sellerOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Unit invalid, unit: %s, orderId: %d, pid: %d, sid: %d, error: %s",
							pkgItem.Subpackages[i].Items[j].Invoice.Unit.Amount, pkgItem.OrderId, pkgItem.PId, pkgItem.Subpackages[i].SId, err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					itemDetail.Invoice.Unit = uint64(unit.IntPart())

					total, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount)
					if err != nil {
						logger.Err("sellerOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Total invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
							pkgItem.Subpackages[i].Items[j].Invoice.Total.Amount, pkgItem.OrderId, pkgItem.PId, pkgItem.Subpackages[i].SId, err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					itemDetail.Invoice.Total = uint64(total.IntPart())

					original, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount)
					if err != nil {
						logger.Err("sellerOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Original invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
							pkgItem.Subpackages[i].Items[j].Invoice.Original.Amount, pkgItem.OrderId, pkgItem.PId, pkgItem.Subpackages[i].SId, err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					itemDetail.Invoice.Original = uint64(original.IntPart())

					special, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount)
					if err != nil {
						logger.Err("sellerOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Special invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
							pkgItem.Subpackages[i].Items[j].Invoice.Special.Amount, pkgItem.OrderId, pkgItem.PId, pkgItem.Subpackages[i].SId, err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					itemDetail.Invoice.Special = uint64(special.IntPart())

					discount, err := decimal.NewFromString(pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount)
					if err != nil {
						logger.Err("sellerOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Discount invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
							pkgItem.Subpackages[i].Items[j].Invoice.Discount.Amount, pkgItem.OrderId, pkgItem.PId, pkgItem.Subpackages[i].SId, err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
					itemDetail.Invoice.Discount = uint64(discount.IntPart())

					sellerOrderDetailItems = append(sellerOrderDetailItems, itemDetail)
				}
			}
		}
	}

	sellerOrderDetail := &pb.SellerOrderDetail{
		OID:       orderId,
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
		},
		Items: sellerOrderDetailItems,
	}

	subtotal, err := decimal.NewFromString(pkgItem.Invoice.Subtotal.Amount)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid, subtotal: %s, orderId: %d, pid: %d, error: %s",
			pkgItem.Invoice.Subtotal.Amount, pkgItem.OrderId, pkgItem.PId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

	}
	sellerOrderDetail.Amount = uint64(subtotal.IntPart())

	if pkgItem.ShippingAddress.Location != nil {
		sellerOrderDetail.Address.Lat = strconv.Itoa(int(pkgItem.ShippingAddress.Location.Coordinates[0]))
		sellerOrderDetail.Address.Long = strconv.Itoa(int(pkgItem.ShippingAddress.Location.Coordinates[1]))
	}

	serializedData, err := proto.Marshal(sellerOrderDetail)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => could not serialize sellerOrderDetail, sellerOrderDetail: %v, error:%s", sellerOrderDetail, err)
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

func (server *Server) sellerOrderReturnDetailListHandler(ctx context.Context, pid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	if page <= 0 || perPage <= 0 {
		logger.Err("sellerOrderReturnDetailListHandler() => page or perPage invalid, pid: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.BadRequest), "neither offset nor start can be zero")
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
		logger.Err("sellerOrderReturnDetailListHandler() => CountWithFilter failed,  pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if totalCount == 0 {
		logger.Err("sellerOrderReturnDetailListHandler() => total count is zero,  pid: %d, filterValue: %s, page: %d, perPage: %d", pid, filter, page, perPage)
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
		logger.Err("sellerOrderReturnDetailListHandler() => availablePages less than page, availablePages: %d, pid: %d, filterValue: %s, page: %d, perPage: %d", availablePages, pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var offset = (page - 1) * perPage
	if int64(offset) >= totalCount {
		logger.Err("sellerOrderReturnDetailListHandler() => offset invalid, offset: %d, pid: %d, filterValue: %s, page: %d, perPage: %d", offset, pid, filter, page, perPage)
		return nil, status.Error(codes.Code(future.NotFound), "Not Found")
	}

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	pkgFilter := func() interface{} {
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

	pkgList, err := app.Globals.PkgItemRepository.FindByFilter(ctx, pkgFilter)
	if err != nil {
		logger.Err("sellerOrderReturnDetailListHandler() => FindByFilter failed, pid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", pid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerReturnOrderList := make([]*pb.SellerReturnOrderDetailList_ReturnOrderDetail, 0, len(pkgList))
	var itemDetailList []*pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item

	for i := 0; i < len(pkgList); i++ {
		itemDetailList = nil
		for j := 0; j < len(pkgList[i].Subpackages); j++ {
			for _, filterState := range server.sellerFilterStates[filter] {
				if pkgList[i].Subpackages[i].Status == filterState.actualState.StateName() {
					itemDetailList = make([]*pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item, 0, len(pkgList[i].Subpackages[j].Items))
					for z := 0; z < len(pkgList[i].Subpackages[j].Items); z++ {
						itemOrder := &pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item{
							SID:    pkgList[i].Subpackages[j].SId,
							Sku:    pkgList[i].Subpackages[j].Items[z].SKU,
							Status: filterState.expectedState.StateName(),
							SIdx:   int32(states.FromString(pkgList[i].Subpackages[j].Status).StateIndex()),
							Detail: &pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item_Detail{
								InventoryId:     pkgList[i].Subpackages[j].Items[z].InventoryId,
								Title:           pkgList[i].Subpackages[j].Items[z].Title,
								Brand:           pkgList[i].Subpackages[j].Items[z].Brand,
								Category:        pkgList[i].Subpackages[j].Items[z].Category,
								Guaranty:        pkgList[i].Subpackages[j].Items[z].Guaranty,
								Image:           pkgList[i].Subpackages[j].Items[z].Image,
								Returnable:      pkgList[i].Subpackages[j].Items[z].Returnable,
								Quantity:        pkgList[i].Subpackages[j].Items[z].Quantity,
								Attributes:      pkgList[i].Subpackages[j].Items[z].Attributes,
								ReturnRequestAt: "",
								ReturnShippedAt: "",
								Invoice: &pb.SellerReturnOrderDetailList_ReturnOrderDetail_Item_Detail_Invoice{
									Unit:             0,
									Total:            0,
									Original:         0,
									Special:          0,
									Discount:         0,
									SellerCommission: pkgList[i].Subpackages[j].Items[z].Invoice.SellerCommission,
									Currency:         "IRR",
								},
							},
						}

						unit, err := decimal.NewFromString(pkgList[i].Subpackages[j].Items[z].Invoice.Unit.Amount)
						if err != nil {
							logger.Err("sellerOrderReturnDetailListHandler() => decimal.NewFromString failed, subpackage Invoice.Unit invalid, unit: %s, orderId: %d, pid: %d, sid: %d, error: %s",
								pkgList[i].Subpackages[j].Items[z].Invoice.Unit.Amount, pkgList[i].OrderId, pkgList[i].PId, pkgList[i].Subpackages[j].SId, err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemOrder.Detail.Invoice.Unit = uint64(unit.IntPart())

						total, err := decimal.NewFromString(pkgList[i].Subpackages[j].Items[z].Invoice.Total.Amount)
						if err != nil {
							logger.Err("sellerOrderReturnDetailListHandler() => decimal.NewFromString failed, subpackage Invoice.Total invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
								pkgList[i].Subpackages[j].Items[z].Invoice.Total.Amount, pkgList[i].Subpackages[j].OrderId, pkgList[i].Subpackages[j].PId, pkgList[i].Subpackages[j].SId, err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemOrder.Detail.Invoice.Total = uint64(total.IntPart())

						original, err := decimal.NewFromString(pkgList[i].Subpackages[j].Items[z].Invoice.Original.Amount)
						if err != nil {
							logger.Err("sellerOrderReturnDetailListHandler() => decimal.NewFromString failed, subpackage Invoice.Original invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
								pkgList[i].Subpackages[j].Items[z].Invoice.Original.Amount, pkgList[i].Subpackages[j].OrderId, pkgList[i].Subpackages[j].PId, pkgList[i].Subpackages[j].SId, err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemOrder.Detail.Invoice.Original = uint64(original.IntPart())

						special, err := decimal.NewFromString(pkgList[i].Subpackages[j].Items[z].Invoice.Special.Amount)
						if err != nil {
							logger.Err("sellerOrderReturnDetailListHandler() => decimal.NewFromString failed, subpackage Invoice.Special invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
								pkgList[i].Subpackages[j].Items[z].Invoice.Special.Amount, pkgList[i].Subpackages[j].OrderId, pkgList[i].Subpackages[j].PId, pkgList[i].Subpackages[j].SId, err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemOrder.Detail.Invoice.Special = uint64(special.IntPart())

						discount, err := decimal.NewFromString(pkgList[i].Subpackages[j].Items[z].Invoice.Discount.Amount)
						if err != nil {
							logger.Err("sellerOrderReturnDetailListHandler() => decimal.NewFromString failed, subpackage Invoice.Discount invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
								pkgList[i].Subpackages[j].Items[z].Invoice.Discount.Amount, pkgList[i].Subpackages[j].OrderId, pkgList[i].Subpackages[j].PId, pkgList[i].Subpackages[j].SId, err)
							return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
						}
						itemOrder.Detail.Invoice.Discount = uint64(discount.IntPart())

						if pkgList[i].Subpackages[j].Shipments != nil &&
							pkgList[i].Subpackages[j].Shipments.ReturnShipmentDetail != nil {
							if pkgList[i].Subpackages[j].Shipments.ReturnShipmentDetail.RequestedAt != nil {
								itemOrder.Detail.ReturnRequestAt = pkgList[i].Subpackages[j].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
							}
							if pkgList[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippedAt != nil {
								itemOrder.Detail.ReturnShippedAt = pkgList[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippedAt.Format(ISO8601)
							}
						}

						itemDetailList = append(itemDetailList, itemOrder)
					}
				}
			}
		}

		if itemDetailList != nil {
			returnOrderDetail := &pb.SellerReturnOrderDetailList_ReturnOrderDetail{
				OID:       pkgList[i].OrderId,
				Amount:    0,
				RequestAt: pkgList[i].CreatedAt.Format(ISO8601),
				Items:     itemDetailList,
				Address: &pb.SellerReturnOrderDetailList_ReturnOrderDetail_ShipmentAddress{
					FirstName:     pkgList[i].ShippingAddress.FirstName,
					LastName:      pkgList[i].ShippingAddress.LastName,
					Address:       pkgList[i].ShippingAddress.Address,
					Phone:         pkgList[i].ShippingAddress.Phone,
					Mobile:        pkgList[i].ShippingAddress.Mobile,
					Country:       pkgList[i].ShippingAddress.Country,
					City:          pkgList[i].ShippingAddress.City,
					Province:      pkgList[i].ShippingAddress.Province,
					Neighbourhood: pkgList[i].ShippingAddress.Neighbourhood,
					Lat:           "",
					Long:          "",
					ZipCode:       pkgList[i].ShippingAddress.ZipCode,
				},
			}

			subtotal, err := decimal.NewFromString(pkgList[i].Invoice.Subtotal.Amount)
			if err != nil {
				logger.Err("sellerOrderReturnDetailListHandler() => decimal.NewFromString failed, pkgItem.Invoice.Subtotal invalid, subtotal: %s, orderId: %d, pid: %d, error: %s",
					pkgList[i].Invoice.Subtotal.Amount, pkgList[i].OrderId, pkgList[i].PId, err)
				return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

			}
			returnOrderDetail.Amount = uint64(subtotal.IntPart())

			if pkgList[i].ShippingAddress.Location != nil {
				returnOrderDetail.Address.Lat = strconv.Itoa(int(pkgList[i].ShippingAddress.Location.Coordinates[0]))
				returnOrderDetail.Address.Long = strconv.Itoa(int(pkgList[i].ShippingAddress.Location.Coordinates[1]))
			}
			sellerReturnOrderList = append(sellerReturnOrderList, returnOrderDetail)
		} else {
			logger.Err("sellerOrderReturnDetailListHandler() => get item from orderList failed, orderId: %d pid: %d, filterValue: %s, page: %d, perPage: %d", pkgList[i].OrderId, pid, filter, page, perPage)
		}
	}

	sellerReturnOrderDetailList := &pb.SellerReturnOrderDetailList{
		PID:               pid,
		ReturnOrderDetail: sellerReturnOrderList,
	}

	serializedData, err := proto.Marshal(sellerReturnOrderDetailList)
	if err != nil {
		logger.Err("sellerOrderDetailHandler() => could not serialize sellerReturnOrderDetailList, sellerReturnOrderDetailList: %v, error:%s", sellerReturnOrderDetailList, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "SellerReturnOrderDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(totalCount),
			Page:    page,
			PerPage: perPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(sellerReturnOrderDetailList),
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
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShipmentPendingState := server.queryPathStates[ShipmentPendingFilter]
	shipmentPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "packages.deletedAt": nil, queryPathShipmentPendingState.queryPath: queryPathShipmentPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathShipmentDelayedState := server.queryPathStates[ShipmentDelayedFilter]
	shipmentDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShipmentDelayedState.queryPath: queryPathShipmentDelayedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathReturnRequestPendingState := server.queryPathStates[ReturnRequestPendingFilter]
	returnRequestPendingFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnRequestPendingState.queryPath: queryPathReturnRequestPendingState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	approvalPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, approvalPendingFilter)
	if err != nil {
		logger.Err("sellerOrderDashboardReportsHandler() => CountWithFilter for approvalPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	shipmentPendingCount, err := app.Globals.PkgItemRepository.CountWithFilter(ctx, shipmentPendingFilter)
	if err != nil {
		logger.Err("sellerOrderDashboardReportsHandler() => CountWithFilter for shipmentPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	shipmentDelayedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, shipmentDelayedFilter)
	if err != nil {
		logger.Err("sellerOrderDashboardReportsHandler() => CountWithFilter for shipmentDelayedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnRequestPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		logger.Err("sellerOrderDashboardReportsHandler() => CountWithFilter for returnRequestPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderDashboardReports := &pb.SellerOrderDashboardReports{
		SellerId:             userId,
		ApprovalPending:      uint32(approvalPendingCount),
		ShipmentPending:      uint32(shipmentPendingCount),
		ShipmentDelayed:      uint32(shipmentDelayedCount),
		ReturnRequestPending: uint32(returnRequestPendingCount),
	}

	serializedData, err := proto.Marshal(sellerOrderDashboardReports)
	if err != nil {
		logger.Err("sellerOrderDashboardReportsHandler() => could not serialize sellerOrderDashboardReportsHandler, userId: %d, error:%s", userId, err)
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathShippedState.queryPath: queryPathShippedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	shipmentPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, shipmentPendingFilter)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => CountWithFilter for shipmentPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	shipmentDelayedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, shipmentDelayedFilter)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => CountWithFilter for shipmentDelayedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	shippedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, shippedFilter)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => CountWithFilter for shippedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderShipmentReports := &pb.SellerOrderShipmentReports{
		SellerId:        userId,
		ShipmentPending: uint32(shipmentPendingCount),
		ShipmentDelayed: uint32(shipmentDelayedCount),
		Shipped:         uint32(shippedCount),
	}

	serializedData, err := proto.Marshal(sellerOrderShipmentReports)
	if err != nil {
		logger.Err("sellerOrderShipmentReportsHandler() => could not serialize sellerOrderShipmentReports, userId: %d, error:%s", userId, err)
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathReturnDeliveryFailedState.queryPath: queryPathReturnDeliveryFailedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	returnRequestPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnRequestPendingFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnRequestPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShipmentPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnShipmentPendingFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnShipmentPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnShippedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnShippedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnShippedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveredCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveredFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveredFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryFailedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveryFailedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveryFailedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnRequestRejectedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnRequestRejectedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnRequestRejectedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryPendingCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveryPendingFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveryPendingFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	returnDeliveryDelayedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, returnDeliveryDelayedFilter)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => CountWithFilter for returnDeliveryDelayedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
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

	serializedData, err := proto.Marshal(sellerOrderReturnReports)
	if err != nil {
		logger.Err("sellerOrderReturnReportsHandler() => could not serialize sellerOrderReturnReports, userId: %d, error:%s", userId, err)
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
	queryPathDeliveryDelayedState := server.queryPathStates[DeliveryDelayedFilter]
	deliveryPendingAndDelayedFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()},
				bson.M{queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()}}}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, "$or": bson.A{
				bson.M{queryPathDeliveryPendingState.queryPath: queryPathDeliveryPendingState.state.StateName()},
				bson.M{queryPathDeliveryDelayedState.queryPath: queryPathDeliveryDelayedState.state.StateName()}}}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	queryPathDeliveredState := server.queryPathStates[DeliveredFilter]
	deliveredFilter := func() interface{} {
		return []bson.M{
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveredState.queryPath: queryPathDeliveredState.state.StateName()}},
			{"$unwind": "$packages"},
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathDeliveryFailedState.queryPath: queryPathDeliveryFailedState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	deliveryPendingAndDelayedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, deliveryPendingAndDelayedFilter)
	if err != nil {
		logger.Err("sellerOrderDeliverReportsHandler() => CountWithFilter for deliveryPendingAndDelayedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	deliveredCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, deliveredFilter)
	if err != nil {
		logger.Err("sellerOrderDeliverReportsHandler() => CountWithFilter for deliveredFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	deliveryFailedCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, deliveryFailedFilter)
	if err != nil {
		logger.Err("sellerOrderDeliverReportsHandler() => CountWithFilter for deliveryFailedFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderDeliveredReports := &pb.SellerOrderDeliveredReports{
		SellerId:                  userId,
		DeliveryPendingAndDelayed: uint32(deliveryPendingAndDelayedCount),
		Delivered:                 uint32(deliveredCount),
		DeliveryFailed:            uint32(deliveryFailedCount),
	}

	serializedData, err := proto.Marshal(sellerOrderDeliveredReports)
	if err != nil {
		logger.Err("buyerReturnOrderReportsHandler() => could not serialize sellerOrderDeliveredReports, userId: %d, error:%s", userId, err)
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
			{"$unwind": "$packages.subpackages"},
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
			{"$unwind": "$packages.subpackages"},
			{"$match": bson.M{"packages.pid": userId, "deletedAt": nil, queryPathCanceledBySellerState.queryPath: queryPathCanceledBySellerState.state.StateName()}},
			{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"_id": 0, "count": 1}},
		}
	}

	cancelByBuyerCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, cancelByBuyerFilter)
	if err != nil {
		logger.Err("sellerOrderCancelReportsHandler() => CountWithFilter for cancelByBuyerFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	cancelBySellerCount, err := app.Globals.SubPkgRepository.CountWithFilter(ctx, cancelBySellerFilter)
	if err != nil {
		logger.Err("sellerOrderCancelReportsHandler() => CountWithFilter for cancelBySellerFilter failed, userId: %d, error: %s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	sellerOrderCancelReports := &pb.SellerOrderCancelReports{
		SellerId:       userId,
		CancelBySeller: uint32(cancelBySellerCount),
		CancelByBuyer:  uint32(cancelByBuyerCount),
	}

	serializedData, err := proto.Marshal(sellerOrderCancelReports)
	if err != nil {
		logger.Err("sellerOrderCancelReportsHandler() => could not serialize sellerOrderCancelReports, userId: %d, error:%s", userId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")

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