package grpc_server

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	pb "gitlab.faza.io/protos/order"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) operatorOrderListHandler(ctx context.Context, oid uint64, filter FilterValue, page, perPage uint32,
	sortName string, direction SortDirection) (*pb.MessageResponse, error) {

	var sortDirect int
	if direction == "ASC" {
		sortDirect = 1
	} else {
		sortDirect = -1
	}

	var orderFilter func() (interface{}, string, int)
	if oid > 0 {
		return server.operatorGetOrderByIdHandler(ctx, oid, filter)
	} else {
		if filter != "" {
			filters := server.OperatorGeneratePipelineFilter(ctx, filter)
			orderFilter = func() (interface{}, string, int) {
				return bson.D{{"deletedAt", nil}, {filters[0].(string), filters[1]}},
					sortName, sortDirect
			}
		} else {
			orderFilter = func() (interface{}, string, int) {
				return bson.D{{"deletedAt", nil}}, sortName, sortDirect
			}
		}
	}

	orderList, totalCount, err := app.Globals.OrderRepository.FindByFilterWithPageAndSort(ctx, orderFilter, int64(page), int64(perPage))
	if err != nil {
		logger.Err("operatorOrderListHandler() => CountWithFilter failed,  oid: %d, filterValue: %s, page: %d, perPage: %d, error: %s", oid, filter, page, perPage, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	if totalCount == 0 || orderList == nil || len(orderList) == 0 {
		logger.Err("operatorOrderListHandler() => order not found, orderId: %d, filter:%s", oid, filter)
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
			logger.Err("operatorOrderListHandler() => decimal.NewFromString failed, GrandTotal invalid, grandTotal: %s, orderId: %d, error:%s",
				orderList[i].Invoice.GrandTotal.Amount, orderList[i].OrderId, err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.GrandTotal = uint64(amount.IntPart())

		subtotal, err := decimal.NewFromString(orderList[i].Invoice.Subtotal.Amount)
		if err != nil {
			logger.Err("operatorOrderListHandler() => decimal.NewFromString failed, Subtotal invalid, subtotal: %s, orderId: %d, error:%s",
				orderList[i].Invoice.Subtotal.Amount, orderList[i].OrderId, err)
			return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
		}
		order.Invoice.Subtotal = uint64(subtotal.IntPart())

		shipmentTotal, err := decimal.NewFromString(orderList[i].Invoice.ShipmentTotal.Amount)
		if err != nil {
			logger.Err("operatorOrderListHandler() => decimal.NewFromString failed, shipmentTotal invalid, shipmentTotal: %s, orderId: %d, error:%s",
				orderList[i].Invoice.ShipmentTotal.Amount, orderList[i].OrderId, err)
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
						logger.Err("operatorOrderListHandler() => order.Invoice.Voucher.Price.Amount invalid, price: %s, orderId: %d, error: %s", orderList[i].Invoice.Voucher.Price.Amount, order.OrderId, err)
						return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
					}
				}
				order.Invoice.Voucher = float32(voucherAmount.IntPart())
			}
		}

		if orderList[i].OrderPayment != nil &&
			len(orderList[i].OrderPayment) > 0 &&
			orderList[i].OrderPayment[0].PaymentResult != nil {
			if orderList[i].OrderPayment[0].PaymentResult.Result {
				order.Invoice.PaymentStatus = "success"
			} else {
				order.Invoice.PaymentStatus = "fail"
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

	serializedData, err := proto.Marshal(operatorOrderList)
	if err != nil {
		logger.Err("operatorOrderListHandler() => could not serialize operatorOrderListHandler, operatorOrderList: %v, error:%s", operatorOrderList, err)
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
		logger.Err("operatorOrderDetailHandler() => FindById failed, oid: %d, error: %s", oid, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
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

	amount, err := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
	if err != nil {
		logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, GrandTotal invalid, grandTotal: %s, orderId: %d, error:%s",
			order.Invoice.GrandTotal.Amount, order.OrderId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.Invoice.GrandTotal = uint64(amount.IntPart())

	subtotal, err := decimal.NewFromString(order.Invoice.Subtotal.Amount)
	if err != nil {
		logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, Subtotal invalid, subtotal: %s, orderId: %d, error:%s",
			order.Invoice.Subtotal.Amount, order.OrderId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.Invoice.Subtotal = uint64(subtotal.IntPart())

	shipmentTotal, err := decimal.NewFromString(order.Invoice.ShipmentTotal.Amount)
	if err != nil {
		logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, shipmentTotal invalid, shipmentTotal: %s, orderId: %d, error:%s",
			order.Invoice.ShipmentTotal.Amount, order.OrderId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	orderDetail.Invoice.ShipmentTotal = uint64(shipmentTotal.IntPart())

	if order.Invoice.Voucher != nil {
		if order.Invoice.Voucher.Percent > 0 {
			orderDetail.Invoice.VoucherAmount = float32(order.Invoice.Voucher.Percent)
		} else {
			var voucherAmount decimal.Decimal
			if order.Invoice.Voucher.Price != nil {
				voucherAmount, err = decimal.NewFromString(order.Invoice.Voucher.Price.Amount)
				if err != nil {
					logger.Err("operatorOrderDetailHandler() => order.Invoice.Voucher.Price.Amount invalid, price: %s, orderId: %d, error: %s", order.Invoice.Voucher.Price.Amount, order.OrderId, err)
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
					CarrierName:    order.Packages[i].Subpackages[j].Shipments.ShipmentDetail.CarrierName,
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
					CarrierName:    order.Packages[i].Subpackages[j].Shipments.ReturnShipmentDetail.CarrierName,
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
					Attributes:  order.Packages[i].Subpackages[j].Items[z].Attributes,
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

				unit, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Unit.Amount)
				if err != nil {
					logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Unit invalid, unit: %s, orderId: %d, pid: %d, sid: %d, error: %s",
						order.Packages[i].Subpackages[j].Items[z].Invoice.Unit.Amount, order.OrderId, order.Packages[i].Subpackages[j].PId, order.Packages[i].Subpackages[j].SId, err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Unit = uint64(unit.IntPart())

				total, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Total.Amount)
				if err != nil {
					logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Total invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
						order.Packages[i].Subpackages[j].Items[z].Invoice.Total.Amount, order.OrderId, order.Packages[i].Subpackages[j].PId, order.Packages[i].Subpackages[j].SId, err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Total = uint64(total.IntPart())

				original, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Original.Amount)
				if err != nil {
					logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Original invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
						order.Packages[i].Subpackages[j].Items[z].Invoice.Original.Amount, order.OrderId, order.Packages[i].Subpackages[j].PId, order.Packages[i].Subpackages[j].SId, err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Original = uint64(original.IntPart())

				special, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Special.Amount)
				if err != nil {
					logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Special invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
						order.Packages[i].Subpackages[j].Items[z].Invoice.Special.Amount, order.OrderId, order.Packages[i].Subpackages[j].PId, order.Packages[i].Subpackages[j].SId, err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Special = uint64(special.IntPart())

				discount, err := decimal.NewFromString(order.Packages[i].Subpackages[j].Items[z].Invoice.Discount.Amount)
				if err != nil {
					logger.Err("operatorOrderDetailHandler() => decimal.NewFromString failed, subpackage Invoice.Discount invalid, total: %s, orderId: %d, pid: %d, sid: %d, error: %s",
						order.Packages[i].Subpackages[j].Items[z].Invoice.Discount.Amount, order.OrderId, order.Packages[i].Subpackages[j].PId, order.Packages[i].Subpackages[j].SId, err)
					return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
				}
				item.Invoice.Discount = uint64(discount.IntPart())

				subpackage.Items = append(subpackage.Items, item)
			}
			orderDetail.Subpackages = append(orderDetail.Subpackages, subpackage)
		}
	}

	serializedData, err := proto.Marshal(orderDetail)
	if err != nil {
		logger.Err("operatorOrderDetailHandler() => could not serialize operatorOrderDetail, orderId: %d, error:%s", orderDetail.OrderId, err)
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
		logger.Err("operatorGetOrderByIdHandler() => OrderRepository.FindById,  oid: %d, filterValue: %s, error: %s", oid, filter, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
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

	grandTotal, err := decimal.NewFromString(findOrder.Invoice.GrandTotal.Amount)
	if err != nil {
		logger.Err("operatorGetOrderByIdHandler() => decimal.NewFromString failed, GrandTotal invalid, grandTotal: %s, orderId: %d, error:%s",
			findOrder.Invoice.GrandTotal.Amount, findOrder.OrderId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	order.Invoice.GrandTotal = uint64(grandTotal.IntPart())

	subtotal, err := decimal.NewFromString(findOrder.Invoice.Subtotal.Amount)
	if err != nil {
		logger.Err("operatorGetOrderByIdHandler() => decimal.NewFromString failed, Subtotal invalid, subtotal: %s, orderId: %d, error:%s",
			findOrder.Invoice.Subtotal.Amount, findOrder.OrderId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	order.Invoice.Subtotal = uint64(subtotal.IntPart())

	shipmentTotal, err := decimal.NewFromString(findOrder.Invoice.ShipmentTotal.Amount)
	if err != nil {
		logger.Err("operatorGetOrderByIdHandler() => decimal.NewFromString failed, shipmentTotal invalid, shipmentTotal: %s, orderId: %d, error:%s",
			findOrder.Invoice.ShipmentTotal.Amount, findOrder.OrderId, err)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}
	order.Invoice.Shipment = uint64(shipmentTotal.IntPart())

	if findOrder.Invoice.Voucher != nil {
		if findOrder.Invoice.Voucher.Percent > 0 {
			order.Invoice.Voucher = float32(findOrder.Invoice.Voucher.Percent)
		} else {
			var voucherAmount decimal.Decimal
			if findOrder.Invoice.Voucher.Price != nil {
				voucherAmount, err = decimal.NewFromString(findOrder.Invoice.Voucher.Price.Amount)
				if err != nil {
					logger.Err("operatorGetOrderByIdHandler() => decimal.NewFromString failed, order.Invoice.Voucher.Price.Amount invalid, price: %s, orderId: %d, error: %s",
						findOrder.Invoice.Voucher.Price.Amount, order.OrderId, err)
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

	serializedData, err := proto.Marshal(operatorOrderList)
	if err != nil {
		logger.Err("operatorGetOrderByIdHandler() => could not serialize operatorGetOrderByIdHandler, operatorOrderList: %v, error:%s", operatorOrderList, err)
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

func (server *Server) OperatorGeneratePipelineFilter(ctx context.Context, filter FilterValue) []interface{} {

	newFilter := make([]interface{}, 2)
	queryPathState := server.queryPathStates[filter]
	newFilter[0] = queryPathState.queryPath
	newFilter[1] = queryPathState.state.StateName()
	return newFilter
}