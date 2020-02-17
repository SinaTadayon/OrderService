package grpc_server

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/models/repository"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	pb "gitlab.faza.io/protos/order"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) operatorOrderListHandler(ctx context.Context, oid uint64, buyerMobile string, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	var orderList []*entities.Order
	var totalCount int64
	var err repository.IRepoError
	if oid > 0 {
		return server.operatorGetOrderByIdHandler(ctx, oid, filter)
	} else if buyerMobile != "" {
		return server.operatorGetOrdersByMobileHandler(ctx, buyerMobile, filter, page, perPage, sortName, direction)
	} else {
		var sortDirect int
		if direction == "ASC" {
			sortDirect = 1
		} else {
			sortDirect = -1
		}

		if filter != "" {
			filters := server.OperatorGeneratePipelineFilter(ctx, filter)
			if sortName != "" {
				orderFilter := func() (interface{}, string, int) {
					return bson.D{{"deletedAt", nil}, {filters[0].(string), filters[1]}},
						sortName, sortDirect
				}
				orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "operatorOrderListHandler", "oid", oid, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
					return nil, status.Error(codes.Code(err.Code()), err.Message())
				}
			} else {
				orderFilter := func() interface{} {
					return bson.D{{"deletedAt", nil}, {filters[0].(string), filters[1]}}
				}
				orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "operatorOrderListHandler", "oid", oid, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
					return nil, status.Error(codes.Code(err.Code()), err.Message())
				}
			}
		} else {
			if sortName != "" {
				orderFilter := func() (interface{}, string, int) {
					return bson.D{{"deletedAt", nil}}, sortName, sortDirect
				}
				orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "operatorOrderListHandler", "oid", oid, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
					return nil, status.Error(codes.Code(err.Code()), err.Message())
				}
			} else {
				orderFilter := func() interface{} {
					return bson.D{{"deletedAt", nil}}
				}
				orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "operatorOrderListHandler", "oid", oid, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
					return nil, status.Error(codes.Code(err.Code()), err.Message())
				}
			}
		}
	}

	//orderList, totalCount, err := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
	//if err != nil {
	//	app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "operatorOrderListHandler", "oid", oid, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
	//	return nil, status.Error(codes.Code(err.Code()), err.Message())
	//}

	if totalCount == 0 || orderList == nil || len(orderList) == 0 {
		app.Globals.Logger.FromContext(ctx).Info("order not found", "fn", "operatorOrderListHandler", "oid", oid, "filter", filter)
		return nil, status.Error(codes.Code(future.NotFound), "Order Not Found")
	}

	operatorOrders := make([]*pb.OperatorOrderList_Order, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		order := &pb.OperatorOrderList_Order{
			OrderId:     orderList[i].OrderId,
			BuyerId:     orderList[i].BuyerInfo.BuyerId,
			PurchasedOn: orderList[i].CreatedAt.Format(ISO8601),
			BasketSize:  0,
			BillTo:      orderList[i].BuyerInfo.FirstName + " " + orderList[i].BuyerInfo.LastName,
			ShipTo:      orderList[i].BuyerInfo.ShippingAddress.FirstName + " " + orderList[i].BuyerInfo.ShippingAddress.LastName,
			Platform:    orderList[i].Platform,
			IP:          orderList[i].BuyerInfo.IP,
			Status:      orderList[i].Status,
			Invoice: &pb.OperatorOrderList_Order_Invoice{
				GrandTotal:     0,
				Subtotal:       0,
				PaymentMethod:  orderList[i].Invoice.PaymentMethod,
				PaymentGateway: orderList[i].Invoice.PaymentGateway,
				Shipment:       0,
			},
		}

		amount, err := decimal.NewFromString(orderList[i].Invoice.GrandTotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, GrandTotal invalid",
				"fn", "operatorOrderListHandler",
				"grandTotal", orderList[i].Invoice.GrandTotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.GrandTotal = uint64(amount.IntPart())

		subtotal, err := decimal.NewFromString(orderList[i].Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
				"fn", "operatorOrderListHandler",
				"subtotal", orderList[i].Invoice.Subtotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.Subtotal = uint64(subtotal.IntPart())

		shipmentTotal, err := decimal.NewFromString(orderList[i].Invoice.ShipmentTotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, shipmentTotal invalid",
				"fn", "operatorOrderListHandler",
				"shipmentTotal", orderList[i].Invoice.ShipmentTotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.Shipment = uint64(shipmentTotal.IntPart())

		if orderList[i].Invoice.Voucher != nil {
			if orderList[i].Invoice.Voucher.Percent > 0 {
				order.Invoice.Voucher = float32(orderList[i].Invoice.Voucher.Percent)
			} else {
				var voucherAmount decimal.Decimal
				if orderList[i].Invoice.Voucher.Price != nil {
					voucherAmount, err = decimal.NewFromString(orderList[i].Invoice.Voucher.Price.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("order.Invoice.Voucher.Price.Amount invalid",
							"fn", "operatorOrderListHandler",
							"price", orderList[i].Invoice.Voucher.Price.Amount,
							"oid", order.OrderId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
				}
				order.Invoice.Voucher = float32(voucherAmount.IntPart())
			}
		}

		if orderList[i].OrderPayment != nil && len(orderList[i].OrderPayment) > 0 {
			if orderList[i].OrderPayment[0].PaymentResult != nil {
				if orderList[i].OrderPayment[0].PaymentResult.Result {
					order.Invoice.PaymentStatus = "success"
				} else {
					order.Invoice.PaymentStatus = "fail"
				}
			} else {
				if orderList[i].Status == string(states.OrderClosedStatus) {
					if orderList[i].OrderPayment[0].PaymentResponse != nil {
						if orderList[i].OrderPayment[0].PaymentResponse.Result {
							order.Invoice.PaymentStatus = "success"
						} else {
							order.Invoice.PaymentStatus = "fail"
						}
					} else {
						order.Invoice.PaymentStatus = "fail"
					}
				} else {
					order.Invoice.PaymentStatus = "pending"
				}
			}
		} else {
			order.Invoice.PaymentStatus = "pending"
		}

		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
					order.BasketSize += orderList[i].Packages[j].Subpackages[z].Items[t].Quantity
				}
			}
		}

		operatorOrders = append(operatorOrders, order)
	}

	operatorOrderList := &pb.OperatorOrderList{
		Orders: operatorOrders,
	}

	serializedData, e := proto.Marshal(operatorOrderList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal operatorOrderListHandler failed",
			"fn", "operatorOrderListHandler", "operatorOrderList", operatorOrderList, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(totalCount),
		Page:    page,
		PerPage: perPage,
	}

	response := &pb.MessageResponse{
		Entity: "OperatorOrderList",
		Meta:   meta,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(operatorOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) operatorOrderDetailHandler(ctx context.Context, oid uint64) (*pb.MessageResponse, error) {

	order, err := app.Globals.OrderRepository.FindById(ctx, oid)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("FindById failed",
			"fn", "operatorOrderDetailHandler",
			"oid", oid, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	orderDetail := &pb.OperatorOrderDetail{
		OrderId:     order.OrderId,
		PurchasedOn: order.CreatedAt.Format(ISO8601),
		IP:          order.BuyerInfo.IP,
		Invoice: &pb.OperatorOrderDetail_Invoice{
			GrandTotal:     0,
			Subtotal:       0,
			PaymentMethod:  order.Invoice.PaymentMethod,
			PaymentGateway: order.Invoice.PaymentGateway,
			ShipmentTotal:  0,
		},
		Billing: &pb.OperatorOrderDetail_BillingInfo{
			BuyerId:    order.BuyerInfo.BuyerId,
			FirstName:  order.BuyerInfo.FirstName,
			LastName:   order.BuyerInfo.LastName,
			Phone:      order.BuyerInfo.Phone,
			Mobile:     order.BuyerInfo.Mobile,
			NationalId: order.BuyerInfo.NationalId,
		},
		ShippingInfo: &pb.OperatorOrderDetail_ShippingInfo{
			FirstName:    order.BuyerInfo.ShippingAddress.FirstName,
			LastName:     order.BuyerInfo.ShippingAddress.LastName,
			Country:      order.BuyerInfo.ShippingAddress.Country,
			City:         order.BuyerInfo.ShippingAddress.City,
			Province:     order.BuyerInfo.ShippingAddress.Province,
			Neighborhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
			Address:      order.BuyerInfo.ShippingAddress.Address,
			ZipCode:      order.BuyerInfo.ShippingAddress.ZipCode,
		},
		Subpackages: nil,
	}

	amount, e := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, GrandTotal invalid",
			"fn", "operatorOrderDetailHandler",
			"grandTotal", order.Invoice.GrandTotal.Amount,
			"oid", order.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.Invoice.GrandTotal = uint64(amount.IntPart())

	subtotal, e := decimal.NewFromString(order.Invoice.Subtotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
			"fn", "operatorOrderDetailHandler",
			"subtotal", order.Invoice.Subtotal.Amount,
			"oid", order.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.Invoice.Subtotal = uint64(subtotal.IntPart())

	shipmentTotal, e := decimal.NewFromString(order.Invoice.ShipmentTotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, shipmentTotal invalid",
			"fn", "operatorOrderDetailHandler",
			"shipmentTotal", order.Invoice.ShipmentTotal.Amount,
			"oid", order.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.Invoice.ShipmentTotal = uint64(shipmentTotal.IntPart())

	if order.Invoice.Voucher != nil {
		if order.Invoice.Voucher.Percent > 0 {
			orderDetail.Invoice.VoucherAmount = float32(order.Invoice.Voucher.Percent)
		} else {
			var voucherAmount decimal.Decimal
			if order.Invoice.Voucher.Price != nil {
				voucherAmount, e = decimal.NewFromString(order.Invoice.Voucher.Price.Amount)
				if e != nil {
					app.Globals.Logger.FromContext(ctx).Error("order.Invoice.Voucher.Price.Amount invalid",
						"fn", "operatorOrderDetailHandler",
						"price", order.Invoice.Voucher.Price.Amount,
						"oid", order.OrderId,
						"error", e)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
			}
			orderDetail.Invoice.VoucherAmount = float32(voucherAmount.IntPart())
		}
	}

	if order.OrderPayment != nil &&
		len(order.OrderPayment) > 0 &&
		order.OrderPayment[0].PaymentResult != nil {
		if order.OrderPayment[0].PaymentResult.Result {
			orderDetail.Invoice.PaymentStatus = "success"
		} else {
			orderDetail.Invoice.PaymentStatus = "fail"
		}
	}

	orderDetail.Subpackages = make([]*pb.OperatorOrderDetail_Subpackage, 0, 32)
	for i := 0; i < len(order.Packages); i++ {
		for j := 0; j < len(order.Packages[i].Subpackages); j++ {
			subpackage := &pb.OperatorOrderDetail_Subpackage{
				SID:                  order.Packages[i].Subpackages[j].SId,
				PID:                  order.Packages[i].Subpackages[j].PId,
				SellerId:             order.Packages[i].Subpackages[j].PId,
				ShopName:             order.Packages[i].ShopName,
				UpdatedAt:            order.Packages[i].Subpackages[j].UpdatedAt.Format(ISO8601),
				States:               nil,
				ShipmentDetail:       nil,
				ReturnShipmentDetail: nil,
				Items:                nil,
				Actions:              nil,
			}

			subpackage.States = make([]*pb.OperatorOrderDetail_Subpackage_StateHistory, 0, len(order.Packages[i].Subpackages[j].Tracking.History))
			for x := 0; x < len(order.Packages[i].Subpackages[j].Tracking.History); x++ {
				state := &pb.OperatorOrderDetail_Subpackage_StateHistory{
					Name:      order.Packages[i].Subpackages[j].Tracking.History[x].Name,
					Index:     int32(order.Packages[i].Subpackages[j].Tracking.History[x].Index),
					UTP:       "",
					CreatedAt: order.Packages[i].Subpackages[j].Tracking.History[x].CreatedAt.Format(ISO8601),
				}

				if order.Packages[i].Subpackages[j].Tracking.History[x].Actions != nil {
					state.UTP = order.Packages[i].Subpackages[j].Tracking.History[x].Actions[len(order.Packages[i].Subpackages[j].Tracking.History[x].Actions)-1].UTP
					//state.CreatedAt = order.Packages[i].Subpackages[j].Tracking.History[x].Actions[len(order.Packages[i].Subpackages[j].Tracking.History[x].Actions)-1].CreatedAt.Format(ISO8601)
				}
				subpackage.States = append(subpackage.States, state)
			}

			if order.Packages[i].Subpackages[j].Shipments != nil && order.Packages[i].Subpackages[j].Shipments.ShipmentDetail != nil {
				subpackage.ShipmentDetail = &pb.OperatorOrderDetail_Subpackage_ShipmentDetail{
					CarrierName:    order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.CourierName,
					ShippingMethod: order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.ShippingMethod,
					TrackingNumber: order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.TrackingNumber,
					Image:          order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.Image,
					Description:    order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.Description,
					CreatedAt:      order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.CreatedAt.Format(ISO8601),
					ShippedAt:      "",
				}
				if order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.ShippedAt != nil {
					subpackage.ShipmentDetail.ShippedAt = order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.ShippedAt.Format(ISO8601)
				}
			}

			if order.Packages[i].Subpackages[j].Shipments != nil && order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail != nil {
				subpackage.ReturnShipmentDetail = &pb.OperatorOrderDetail_Subpackage_ReturnShipmentDetail{
					CarrierName:    order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.CourierName,
					ShippingMethod: order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippingMethod,
					TrackingNumber: order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.TrackingNumber,
					Image:          order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.Image,
					Description:    order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.Description,
					RequestedAt:    "",
					ShippedAt:      "",
					CreatedAt:      order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.CreatedAt.Format(ISO8601),
				}

				if order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.RequestedAt != nil {
					subpackage.ReturnShipmentDetail.RequestedAt = order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.RequestedAt.Format(ISO8601)
				}

				if order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippedAt != nil {
					subpackage.ReturnShipmentDetail.ShippedAt = order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.ShippedAt.Format(ISO8601)
				}
			}

			subpackage.Actions = make([]string, 0, 3)
			for _, action := range server.flowManager.GetState(states.FromString(order.Packages[i].Subpackages[j].Status)).Actions() {
				if action.ActionType() == actions.Operator {
					subpackage.Actions = append(subpackage.Actions, action.ActionEnum().ActionName())
				}
			}

			subpackage.Items = make([]*pb.OperatorOrderDetail_Subpackage_Item, 0, len(order.Packages[i].Subpackages[j].Items))
			for z := 0; z < len(order.Packages[i].Subpackages[j].Items); z++ {
				item := &pb.OperatorOrderDetail_Subpackage_Item{
					InventoryId: order.Packages[i].Subpackages[j].Items[z].InventoryId,
					Brand:       order.Packages[i].Subpackages[j].Items[z].Brand,
					Title:       order.Packages[i].Subpackages[j].Items[z].Title,
					Attributes:  nil,
					Quantity:    order.Packages[i].Subpackages[j].Items[z].Quantity,
					Invoice: &pb.OperatorOrderDetail_Subpackage_Item_Invoice{
						Unit:     0,
						Total:    0,
						Original: 0,
						Special:  0,
						Discount: 0,
						Currency: "IRR",
					},
				}

				if order.Packages[i].Subpackages[j].Items[z].Attributes != nil {
					item.Attributes = make(map[string]*pb.OperatorOrderDetail_Subpackage_Item_Attribute, len(order.Packages[i].Subpackages[j].Items[z].Attributes))
					for attrKey, attribute := range order.Packages[i].Subpackages[j].Items[z].Attributes {
						keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
						for keyTran, value := range attribute.KeyTranslate {
							keyTranslates[keyTran] = value
						}
						valTranslates := make(map[string]string, len(attribute.ValueTranslate))
						for valTran, value := range attribute.ValueTranslate {
							valTranslates[valTran] = value
						}
						item.Attributes[attrKey] = &pb.OperatorOrderDetail_Subpackage_Item_Attribute{
							KeyTranslates:   keyTranslates,
							ValueTranslates: valTranslates,
						}
					}
				}

				unit, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Unit.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Unit invalid",
						"fn", "operatorOrderDetailHandler",
						"unit", order.Packages[i].Subpackages[j].Items[z].Invoice.Unit.Amount,
						"oid", order.OrderId,
						"pid", order.Packages[i].Subpackages[j].PId,
						"sid", order.Packages[i].Subpackages[j].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Unit = uint64(unit.IntPart())

				total, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Total.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Total invalid",
						"fn", "operatorOrderDetailHandler",
						"total", order.Packages[i].Subpackages[j].Items[z].Invoice.Total.Amount,
						"oid", order.OrderId,
						"pid", order.Packages[i].Subpackages[j].PId,
						"sid", order.Packages[i].Subpackages[j].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Total = uint64(total.IntPart())

				original, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Original.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Original invalid",
						"fn", "operatorOrderDetailHandler",
						"original", order.Packages[i].Subpackages[j].Items[z].Invoice.Original.Amount,
						"oid", order.OrderId,
						"pid", order.Packages[i].Subpackages[j].PId,
						"sid", order.Packages[i].Subpackages[j].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Original = uint64(original.IntPart())

				special, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Special.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Special invalid",
						"fn", "operatorOrderDetailHandler",
						"special", order.Packages[i].Subpackages[j].Items[z].Invoice.Special.Amount,
						"oid", order.OrderId,
						"pid", order.Packages[i].Subpackages[j].PId,
						"sid", order.Packages[i].Subpackages[j].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Special = uint64(special.IntPart())

				discount, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Discount.Amount)
				if err != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, subpackage Invoice.Discount invalid",
						"fn", "operatorOrderDetailHandler",
						"discount", order.Packages[i].Subpackages[j].Items[z].Invoice.Discount.Amount,
						"oid", order.OrderId,
						"pid", order.Packages[i].Subpackages[j].PId,
						"sid", order.Packages[i].Subpackages[j].SId,
						"error", err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Discount = uint64(discount.IntPart())

				subpackage.Items = append(subpackage.Items, item)
			}
			orderDetail.Subpackages = append(orderDetail.Subpackages, subpackage)
		}
	}

	serializedData, e := proto.Marshal(orderDetail)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal operatorOrderDetail failed",
			"fn", "operatorOrderDetailHandler",
			"oid", orderDetail.OrderId, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "OperatorOrderDetail",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(orderDetail),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) operatorGetOrderByIdHandler(ctx context.Context, oid uint64, filter FilterValue) (*pb.MessageResponse, error) {

	findOrder, err := app.Globals.OrderRepository.FindById(ctx, oid)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("OrderRepository.FindById",
			"fn", "operatorGetOrderByIdHandler",
			"oid", oid, "filterValue", filter, "error", err)
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	operatorOrders := make([]*pb.OperatorOrderList_Order, 0, 1)
	order := &pb.OperatorOrderList_Order{
		OrderId:     findOrder.OrderId,
		BuyerId:     findOrder.BuyerInfo.BuyerId,
		PurchasedOn: findOrder.CreatedAt.Format(ISO8601),
		BasketSize:  0,
		BillTo:      findOrder.BuyerInfo.FirstName + " " + findOrder.BuyerInfo.LastName,
		ShipTo:      findOrder.BuyerInfo.ShippingAddress.FirstName + " " + findOrder.BuyerInfo.ShippingAddress.LastName,
		Platform:    findOrder.Platform,
		IP:          findOrder.BuyerInfo.IP,
		Status:      findOrder.Status,
		Invoice: &pb.OperatorOrderList_Order_Invoice{
			GrandTotal:     0,
			Subtotal:       0,
			Shipment:       0,
			Voucher:        0,
			PaymentStatus:  "",
			PaymentMethod:  findOrder.Invoice.PaymentMethod,
			PaymentGateway: findOrder.Invoice.PaymentGateway,
		},
	}

	grandTotal, e := decimal.NewFromString(findOrder.Invoice.GrandTotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, GrandTotal invalid",
			"fn", "operatorGetOrderByIdHandler",
			"grandTotal", findOrder.Invoice.GrandTotal.Amount,
			"oid", findOrder.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	order.Invoice.GrandTotal = uint64(grandTotal.IntPart())

	subtotal, e := decimal.NewFromString(findOrder.Invoice.Subtotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
			"fn", "operatorGetOrderByIdHandler",
			"subtotal", findOrder.Invoice.Subtotal.Amount,
			"oid", findOrder.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	order.Invoice.Subtotal = uint64(subtotal.IntPart())

	shipmentTotal, e := decimal.NewFromString(findOrder.Invoice.ShipmentTotal.Amount)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, shipmentTotal invalid",
			"fn", "operatorGetOrderByIdHandler",
			"shipmentTotal", findOrder.Invoice.ShipmentTotal.Amount,
			"oid", findOrder.OrderId,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	order.Invoice.Shipment = uint64(shipmentTotal.IntPart())

	if findOrder.Invoice.Voucher != nil {
		if findOrder.Invoice.Voucher.Percent > 0 {
			order.Invoice.Voucher = float32(findOrder.Invoice.Voucher.Percent)
		} else {
			var voucherAmount decimal.Decimal
			if findOrder.Invoice.Voucher.Price != nil {
				voucherAmount, e = decimal.NewFromString(findOrder.Invoice.Voucher.Price.Amount)
				if e != nil {
					app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, order.Invoice.Voucher.Price.Amount invalid",
						"fn", "operatorGetOrderByIdHandler",
						"price", findOrder.Invoice.Voucher.Price.Amount,
						"oid", order.OrderId,
						"error", e)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
			}
			order.Invoice.Voucher = float32(voucherAmount.IntPart())
		}
	}

	if findOrder.OrderPayment != nil &&
		len(findOrder.OrderPayment) > 0 &&
		findOrder.OrderPayment[0].PaymentResult != nil {
		if findOrder.OrderPayment[0].PaymentResult.Result {
			order.Invoice.PaymentStatus = "success"
		} else {
			order.Invoice.PaymentStatus = "fail"
		}
	} else {
		order.Invoice.PaymentStatus = "pending"
	}

	for j := 0; j < len(findOrder.Packages); j++ {
		for z := 0; z < len(findOrder.Packages[j].Subpackages); z++ {
			for t := 0; t < len(findOrder.Packages[j].Subpackages[z].Items); t++ {
				order.BasketSize += findOrder.Packages[j].Subpackages[z].Items[t].Quantity
			}
		}
	}

	operatorOrders = append(operatorOrders, order)

	operatorOrderList := &pb.OperatorOrderList{
		Orders: operatorOrders,
	}

	serializedData, e := proto.Marshal(operatorOrderList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal operatorGetOrderByIdHandler failed",
			"fn", "operatorGetOrderByIdHandler",
			"operatorOrderList", operatorOrderList, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "OperatorOrderList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(1),
			Page:    1,
			PerPage: 1,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(operatorOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) operatorGetOrdersByMobileHandler(ctx context.Context, buyerMobile string, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {
	var orderList []*entities.Order
	var totalCount int64
	var err repository.IRepoError

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	if filter != "" {
		filters := server.OperatorGeneratePipelineFilter(ctx, filter)
		if sortName != "" {
			orderFilter := func() (interface{}, string, int) {
				return bson.D{{"deletedAt", nil}, {"buyerInfo.mobile", bson.D{{"$regex", buyerMobile}}}, {filters[0].(string), filters[1]}},
					sortName, sortDirect
			}
			orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "operatorGetOrdersByMobileHandler", "buyerMobile", buyerMobile, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
				return nil, status.Error(codes.Code(err.Code()), err.Message())
			}
		} else {
			orderFilter := func() interface{} {
				return bson.D{{"deletedAt", nil}, {"buyerInfo.mobile", bson.D{{"$regex", buyerMobile}}}, {filters[0].(string), filters[1]}}
			}
			orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "operatorGetOrdersByMobileHandler", "buyerMobile", buyerMobile, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
				return nil, status.Error(codes.Code(err.Code()), err.Message())
			}
		}
	} else {
		if sortName != "" {
			orderFilter := func() (interface{}, string, int) {
				return bson.D{{"deletedAt", nil}, {"buyerInfo.mobile", bson.D{{"$regex", buyerMobile}}}}, sortName, sortDirect
			}
			orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPageAndSort failed", "fn", "operatorGetOrdersByMobileHandler", "buyerMobile", buyerMobile, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
				return nil, status.Error(codes.Code(err.Code()), err.Message())
			}
		} else {
			orderFilter := func() interface{} {
				return bson.D{{"deletedAt", nil}, {"buyerInfo.mobile", bson.D{{"$regex", buyerMobile}}}}
			}
			orderList, totalCount, err = app.Globals.OrderRepository.FindByFilterWithPage(ctx, orderFilter, int64(page), int64(perPage))
			if err != nil {
				app.Globals.Logger.FromContext(ctx).Error("FindByFilterWithPage failed", "fn", "operatorGetOrdersByMobileHandler", "buyerMobile", buyerMobile, "filterValue", filter, "page", page, "perPage", perPage, "error", err)
				return nil, status.Error(codes.Code(err.Code()), err.Message())
			}
		}
	}

	if totalCount == 0 || orderList == nil || len(orderList) == 0 {
		app.Globals.Logger.FromContext(ctx).Info("order not found", "fn", "operatorGetOrdersByMobileHandler", "buyerMobile", buyerMobile, "filter", filter)
		return nil, status.Error(codes.Code(future.NotFound), "Order Not Found")
	}

	operatorOrders := make([]*pb.OperatorOrderList_Order, 0, len(orderList))
	for i := 0; i < len(orderList); i++ {
		order := &pb.OperatorOrderList_Order{
			OrderId:     orderList[i].OrderId,
			BuyerId:     orderList[i].BuyerInfo.BuyerId,
			PurchasedOn: orderList[i].CreatedAt.Format(ISO8601),
			BasketSize:  0,
			BillTo:      orderList[i].BuyerInfo.FirstName + " " + orderList[i].BuyerInfo.LastName,
			ShipTo:      orderList[i].BuyerInfo.ShippingAddress.FirstName + " " + orderList[i].BuyerInfo.ShippingAddress.LastName,
			Platform:    orderList[i].Platform,
			IP:          orderList[i].BuyerInfo.IP,
			Status:      orderList[i].Status,
			Invoice: &pb.OperatorOrderList_Order_Invoice{
				GrandTotal:     0,
				Subtotal:       0,
				PaymentMethod:  orderList[i].Invoice.PaymentMethod,
				PaymentGateway: orderList[i].Invoice.PaymentGateway,
				Shipment:       0,
			},
		}

		amount, err := decimal.NewFromString(orderList[i].Invoice.GrandTotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, GrandTotal invalid",
				"fn", "operatorGetOrdersByMobileHandler",
				"grandTotal", orderList[i].Invoice.GrandTotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.GrandTotal = uint64(amount.IntPart())

		subtotal, err := decimal.NewFromString(orderList[i].Invoice.Subtotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, Subtotal invalid",
				"fn", "operatorGetOrdersByMobileHandler",
				"subtotal", orderList[i].Invoice.Subtotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.Subtotal = uint64(subtotal.IntPart())

		shipmentTotal, err := decimal.NewFromString(orderList[i].Invoice.ShipmentTotal.Amount)
		if err != nil {
			app.Globals.Logger.FromContext(ctx).Error("decimal.NewFromString failed, shipmentTotal invalid",
				"fn", "operatorGetOrdersByMobileHandler",
				"shipmentTotal", orderList[i].Invoice.ShipmentTotal.Amount,
				"oid", orderList[i].OrderId,
				"error", err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.Shipment = uint64(shipmentTotal.IntPart())

		if orderList[i].Invoice.Voucher != nil {
			if orderList[i].Invoice.Voucher.Percent > 0 {
				order.Invoice.Voucher = float32(orderList[i].Invoice.Voucher.Percent)
			} else {
				var voucherAmount decimal.Decimal
				if orderList[i].Invoice.Voucher.Price != nil {
					voucherAmount, err = decimal.NewFromString(orderList[i].Invoice.Voucher.Price.Amount)
					if err != nil {
						app.Globals.Logger.FromContext(ctx).Error("order.Invoice.Voucher.Price.Amount invalid",
							"fn", "operatorGetOrdersByMobileHandler",
							"price", orderList[i].Invoice.Voucher.Price.Amount,
							"oid", order.OrderId,
							"error", err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
				}
				order.Invoice.Voucher = float32(voucherAmount.IntPart())
			}
		}

		if orderList[i].OrderPayment != nil && len(orderList[i].OrderPayment) > 0 {
			if orderList[i].OrderPayment[0].PaymentResult != nil {
				if orderList[i].OrderPayment[0].PaymentResult.Result {
					order.Invoice.PaymentStatus = "success"
				} else {
					order.Invoice.PaymentStatus = "fail"
				}
			} else {
				if orderList[i].Status == string(states.OrderClosedStatus) {
					if orderList[i].OrderPayment[0].PaymentResponse != nil {
						if orderList[i].OrderPayment[0].PaymentResponse.Result {
							order.Invoice.PaymentStatus = "success"
						} else {
							order.Invoice.PaymentStatus = "fail"
						}
					} else {
						order.Invoice.PaymentStatus = "fail"
					}
				} else {
					order.Invoice.PaymentStatus = "pending"
				}
			}
		} else {
			order.Invoice.PaymentStatus = "pending"
		}

		for j := 0; j < len(orderList[i].Packages); j++ {
			for z := 0; z < len(orderList[i].Packages[j].Subpackages); z++ {
				for t := 0; t < len(orderList[i].Packages[j].Subpackages[z].Items); t++ {
					order.BasketSize += orderList[i].Packages[j].Subpackages[z].Items[t].Quantity
				}
			}
		}

		operatorOrders = append(operatorOrders, order)
	}

	operatorOrderList := &pb.OperatorOrderList{
		Orders: operatorOrders,
	}

	serializedData, e := proto.Marshal(operatorOrderList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("marshal operatorOrderListHandler failed",
			"fn", "operatorGetOrdersByMobileHandler", "operatorOrderList", operatorOrderList, "error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	meta := &pb.ResponseMetadata{
		Total:   uint32(totalCount),
		Page:    page,
		PerPage: perPage,
	}

	response := &pb.MessageResponse{
		Entity: "OperatorOrderList",
		Meta:   meta,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(operatorOrderList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) OperatorGeneratePipelineFilter(ctx context.Context, filter FilterValue) []interface{} {

	newFilter := make([]interface{}, 2)
	queryPathState := server.queryPathStates[filter]
	newFilter[0] = queryPathState.queryPath
	newFilter[1] = queryPathState.state.StateName()
	return newFilter
}
