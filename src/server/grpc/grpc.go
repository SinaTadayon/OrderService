package grpc_server

import (
	"context"
	"gitlab.faza.io/go-framework/acl"
	"path"
	"runtime/debug"
	"strconv"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"gitlab.faza.io/order-project/order-service/domain/actions"
	scheduler_action "gitlab.faza.io/order-project/order-service/domain/actions/scheduler"
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/domain/states"
	"gitlab.faza.io/order-project/order-service/infrastructure/frame"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"gitlab.faza.io/order-project/order-service/infrastructure/utils"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	pb "gitlab.faza.io/protos/order"
	pg "gitlab.faza.io/protos/payment-gateway"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"net"

	"gitlab.faza.io/go-framework/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	pg.UnimplementedBankResultHookServer
	flowManager          domain.IFlowManager
	address              string
	port                 uint16
	requestFilters       map[RequestName][]FilterValue
	buyerFilterStates    map[FilterValue][]FilterState
	buyerAllStatesMap    map[string][]states.IEnumState
	buyerReturnStatesMap map[string][]states.IEnumState
	sellerFilterStates   map[FilterValue][]FilterState
	sellerStatesMap      map[string][]states.IEnumState
	operatorFilterStates map[FilterValue][]FilterState
	queryPathStates      map[FilterValue]FilterQueryState
	actionStates         map[UserType][]actions.IAction
	reasonConfigs        utils.ReasonConfigs
}

func NewServer(address string, port uint16, flowManager domain.IFlowManager) Server {
	reqFilters := initialRequestFilters()
	actionStateMap := initialActualStateMap()
	queryPathStatesMap := initialQueryStateMap()
	buyerStatesMap := initialBuyerStatesMap()
	buyerAllStatesMapping := initialBuyerAllStatesMapping()
	buyerReturnStatesMapping := initialBuyerReturnStatesMapping()
	sellerFilterStatesMap := initialSellerFilterStatesMap()
	sellerStatesMapping := initialSellerStatesMapping()
	operatorFilterStatesMap := initialOperatorFilterStatesMap()

	rp := utils.InitialReasonConfig()
	return Server{
		flowManager:          flowManager,
		address:              address,
		port:                 port,
		requestFilters:       reqFilters,
		buyerFilterStates:    buyerStatesMap,
		buyerAllStatesMap:    buyerAllStatesMapping,
		buyerReturnStatesMap: buyerReturnStatesMapping,
		sellerFilterStates:   sellerFilterStatesMap,
		sellerStatesMap:      sellerStatesMapping,
		operatorFilterStates: operatorFilterStatesMap,
		queryPathStates:      queryPathStatesMap,
		actionStates:         actionStateMap,
		reasonConfigs:        rp,
	}
}

func (server *Server) RequestHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {

	iFuture := app.Globals.UserService.AuthenticateContextToken(ctx).Get()
	//userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
	if iFuture.Error() != nil {
		app.Globals.Logger.FromContext(ctx).Error("UserService.AuthenticateContextToken failed",
			"fn", "RequestHandler", "error", iFuture.Error().Reason())
		return nil, status.Error(codes.Code(iFuture.Error().Code()), iFuture.Error().Message())
	}

	userAcl := iFuture.Data().(*acl.Acl)
	if uint64(userAcl.User().UserID) != req.Meta.UID {
		app.Globals.Logger.FromContext(ctx).Error("request userId mismatch with token userId", "fn", "RequestHandler",
			"userId", req.Meta.UID, "token", userAcl.User().UserID)
		return nil, status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	}

	if req.Meta.UTP == string(OperatorUser) {
		if !userAcl.UserPerm().Has("order.state.all.view") && RequestType(req.Type) == DataReqType {
			return nil, status.Error(codes.Code(future.Forbidden), "User Not Permitted")
		}

		if !userAcl.UserPerm().Has("order.state.all.action") && RequestType(req.Type) == ActionReqType {
			return nil, status.Error(codes.Code(future.Forbidden), "User Not Permitted")
		}
	}

	if ctx.Value(string(utils.CtxUserID)) == nil {
		ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(req.Meta.UID))
		ctx = context.WithValue(ctx, string(utils.CtxUserACL), userAcl)
	}

	reqType := RequestType(req.Type)
	if reqType == DataReqType {
		return server.requestDataHandler(ctx, req)
	} else {
		return server.requestActionHandler(ctx, req)
	}
}

func (server *Server) SchedulerMessageHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {

	app.Globals.Logger.FromContext(ctx).Debug("Received scheduler action request",
		"fn", "SchedulerMessageHandler",
		"request", req)

	if ctx.Value(string(utils.CtxUserID)) == nil {
		ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(0))
	}

	userType := SchedulerUser
	var userAction actions.IAction

	var schedulerActionRequest pb.SchedulerActionRequest
	if err := ptypes.UnmarshalAny(req.Data, &schedulerActionRequest); err != nil {
		app.Globals.Logger.Error("Could not unmarshal schedulerActionRequest from request anything field", "fn", "SchedulerMessageHandler",
			"request", req, "error", err)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	for _, orderReq := range schedulerActionRequest.Orders {
		userActions, ok := server.actionStates[userType]
		if !ok {
			app.Globals.Logger.Error("requested scheduler action not supported", "fn", "SchedulerMessageHandler", "request", req)
			return nil, status.Error(codes.Code(future.BadRequest), "Scheduler Action Invalid")
		}

		for _, action := range userActions {
			if action.ActionEnum().ActionName() == orderReq.ActionState {
				userAction = action
				break
			}
		}

		if userAction == nil {
			app.Globals.Logger.Error("scheduler action invalid", "fn", "SchedulerMessageHandler", "request", req)
			return nil, status.Error(codes.Code(future.BadRequest), "Action Invalid")
		}

		if userAction.ActionEnum() == scheduler_action.PaymentFail {
			event := events.New(events.Action, orderReq.OID, 0, 0,
				orderReq.StateIndex, userAction,
				time.Unix(req.Time.GetSeconds(), int64(req.Time.GetNanos())), nil)

			app.Globals.Logger.Debug("scheduler action event paymentFail",
				"fn", "SchedulerMessageHandler",
				"oid", event.OrderId(),
				"uid", event.UserId(),
				"event", event)

			iFuture := future.Factory().SetCapacity(1).Build()
			iFrame := frame.Factory().SetFuture(iFuture).SetEvent(event).Build()
			server.flowManager.MessageHandler(ctx, iFrame)
			futureData := iFuture.Get()
			if futureData.Error() != nil {
				app.Globals.Logger.FromContext(ctx).Error("flowManager.MessageHandler failed",
					"fn", "SchedulerMessageHandler",
					"event", event,
					"error", futureData.Error().Reason())
			}

		} else {
			for _, pkgReq := range orderReq.Packages {
				subpackages := make([]events.ActionSubpackage, 0, len(pkgReq.Subpackages))
				for _, subPkgReq := range pkgReq.Subpackages {

					subpackage := events.ActionSubpackage{
						SId:   subPkgReq.SID,
						Items: nil,
					}
					subpackage.Items = make([]events.ActionItem, 0, len(subPkgReq.Items))
					for _, item := range subPkgReq.Items {
						actionItem := events.ActionItem{
							InventoryId: item.InventoryId,
							Quantity:    item.Quantity,
						}
						subpackage.Items = append(subpackage.Items, actionItem)
					}
					subpackages = append(subpackages, subpackage)
				}

				actionData := events.ActionData{
					SubPackages:    subpackages,
					Carrier:        "",
					TrackingNumber: "",
				}

				event := events.New(events.Action, orderReq.OID, pkgReq.PID, 0,
					orderReq.StateIndex, userAction,
					time.Unix(req.Time.GetSeconds(), int64(req.Time.GetNanos())), actionData)

				iFuture := future.Factory().SetCapacity(1).Build()
				iFrame := frame.Factory().SetFuture(iFuture).SetEvent(event).Build()

				app.Globals.Logger.Debug("scheduler action event",
					"fn", "SchedulerMessageHandler",
					"oid", event.OrderId(),
					"uid", event.UserId(),
					"event", event)

				server.flowManager.MessageHandler(ctx, iFrame)
				futureData := iFuture.Get()
				if futureData.Error() != nil {
					app.Globals.Logger.Error("flowManager.MessageHandler failed", "fn", "SchedulerMessageHandler", "event", event, "error", futureData.Error().Reason())
				}
			}
		}
	}

	response := &pb.MessageResponse{
		Entity: "ActionResponse",
		Meta:   nil,
		Data:   nil,
	}
	return response, nil
}

func (server *Server) FinanceOrderItems(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	app.Globals.Logger.FromContext(ctx).Debug("Received financeOrderItem request",
		"fn", "FinanceOrderItems",
		"request", req)

	if ctx.Value(string(utils.CtxUserID)) == nil {
		ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(0))
	}

	var state string
	var filterValue FilterValue
	var sortName string
	var sortDirection int
	var startTimestamp time.Time
	var endTimestamp time.Time

	if req.Meta.Filters != nil {
		filterValue = FilterValue(req.Meta.Filters[0].Value)
		if filterValue != PayToSellerFilter && filterValue != PayToBuyerFilter {
			app.Globals.Logger.Error("filterValue invalid",
				"fn", "FinanceOrderItems",
				"filterValue", filterValue,
				"request", req)
			return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
		}

		if filterValue == PayToSellerFilter {
			state = states.PayToSeller.StateName()
		} else if filterValue == PayToBuyerFilter {
			state = states.PayToBuyer.StateName()
		}
	} else {
		app.Globals.Logger.Error("filters is empty",
			"fn", "FinanceOrderItems",
			"request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	if req.Meta.StartAt != "" {
		temp, err := time.Parse(utils.ISO8601, req.Meta.StartAt)
		if err != nil {
			app.Globals.Logger.Error("StartTimestamp invalid",
				"fn", "FinanceOrderItems",
				"StartTimestamp", req.Meta.StartAt,
				"request", req)
			return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
		}

		startTimestamp = temp
	} else {
		app.Globals.Logger.Error("StartTimestamp is empty",
			"fn", "FinanceOrderItems",
			"request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	if req.Meta.EndAt != "" {
		temp, err := time.Parse(utils.ISO8601, req.Meta.EndAt)
		if err != nil {
			app.Globals.Logger.Error("EndTimestamp invalid",
				"fn", "FinanceOrderItems",
				"endTimestamp", req.Meta.EndAt,
				"request", req)
			return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
		}

		endTimestamp = temp
	} else {
		app.Globals.Logger.Error("EndTimestamp is empty",
			"fn", "FinanceOrderItems",
			"request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	if req.Meta.Sorts != nil {
		sortName = req.Meta.Sorts[0].Name
		if SortDirection(req.Meta.Sorts[0].Direction) == ASC {
			sortDirection = 1
		} else {
			sortDirection = -1
		}
	}

	if req.Meta.Page <= 0 {
		app.Globals.Logger.Error("Page invalid",
			"fn", "FinanceOrderItems",
			"page", req.Meta.Page,
			"request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	if req.Meta.PerPage <= 0 {
		app.Globals.Logger.Error("PerPage invalid",
			"fn", "FinanceOrderItems",
			"perPage", req.Meta.Page,
			"request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	finances, total, err := app.Globals.FinanceReportRepository.FindAllWithPageAndSort(ctx, state, startTimestamp,
		endTimestamp, int64(req.Meta.Page), int64(req.Meta.PerPage), sortName, sortDirection)

	if err != nil {
		app.Globals.Logger.Error("PerPage invalid",
			"fn", "FinanceOrderItems",
			"perPage", req.Meta.PerPage,
			"request", req,
			"error", err.Reason())
		return nil, status.Error(codes.Code(err.Code()), err.Message())
	}

	financeReportList := make([]*pb.FinanceOrderItemDetailList_OrderItemDetail, 0, len(finances))
	for _, finance := range finances {
		if finance.RawSellerShippingNet == nil || finance.RoundupSellerShippingNet == nil {
			app.Globals.Logger.Info("OrderItem doesn't support finance",
				"fn", "FinanceOrderItems",
				"oid", finance.OrderId,
				"pid", finance.PId,
				"sid", finance.SId)
			continue
		}

		financeOrderItem := &pb.FinanceOrderItemDetailList_OrderItemDetail{
			OId:      finance.OrderId,
			SellerId: finance.PId,
			ShipmentAmount: &pb.Money{
				Amount:   finance.ShipmentAmount.Amount,
				Currency: finance.ShipmentAmount.Currency,
			},
			RawShippingNet: &pb.Money{
				Amount:   finance.RawSellerShippingNet.Amount,
				Currency: finance.RawSellerShippingNet.Currency,
			},
			RoundupShippingNet: &pb.Money{
				Amount:   finance.RoundupSellerShippingNet.Amount,
				Currency: finance.RoundupSellerShippingNet.Currency,
			},
			Items:          nil,
			CreatedAt:      finance.CreatedAt.Format(utils.ISO8601),
			UpdatedAt:      finance.UpdatedAt.Format(utils.ISO8601),
			OrderCreatedAt: finance.OrderCreatedAt.Format(utils.ISO8601),
		}

		financeOrderItem.Items = make([]*pb.FinanceOrderItemDetailList_OrderItemDetail_Item, 0, len(finance.Items))

		for _, item := range finance.Items {
			financeItem := &pb.FinanceOrderItemDetailList_OrderItemDetail_Item{
				SId:         finance.SId,
				Sku:         item.SKU,
				InventoryId: item.InventoryId,
				Title:       item.Title,
				Brand:       item.Brand,
				Category:    item.Category,
				Guaranty:    item.Guaranty,
				Image:       item.Image,
				Returnable:  item.Returnable,
				Quantity:    item.Quantity,
				Attributes:  nil,
				Invoice: &pb.FinanceOrderItemDetailList_OrderItemDetail_Item_ItemInvoice{
					Commission: &pb.FinanceOrderItemDetailList_OrderItemDetail_Item_ItemInvoice_ItemCommission{
						ItemCommission:    item.Invoice.Commission.ItemCommission,
						RawUnitPrice:      nil,
						RoundupUnitPrice:  nil,
						RawTotalPrice:     nil,
						RoundupTotalPrice: nil,
					},

					Share: &pb.FinanceOrderItemDetailList_OrderItemDetail_Item_ItemInvoice_ItemShare{
						RawItemNet: &pb.Money{
							Amount:   item.Invoice.Share.RawItemNet.Amount,
							Currency: item.Invoice.Share.RawItemNet.Currency,
						},
						RoundupItemNet: &pb.Money{
							Amount:   item.Invoice.Share.RoundupItemNet.Amount,
							Currency: item.Invoice.Share.RoundupItemNet.Currency,
						},
						RawTotalNet: &pb.Money{
							Amount:   item.Invoice.Share.RawTotalNet.Amount,
							Currency: item.Invoice.Share.RawTotalNet.Currency,
						},
						RoundupTotalNet: &pb.Money{
							Amount:   item.Invoice.Share.RoundupTotalNet.Amount,
							Currency: item.Invoice.Share.RoundupTotalNet.Currency,
						},
						RawUnitSellerShare: &pb.Money{
							Amount:   item.Invoice.Share.RawUnitSellerShare.Amount,
							Currency: item.Invoice.Share.RawUnitSellerShare.Currency,
						},
						RoundupUnitSellerShare: &pb.Money{
							Amount:   item.Invoice.Share.RoundupUnitSellerShare.Amount,
							Currency: item.Invoice.Share.RoundupUnitSellerShare.Currency,
						},
						RawTotalSellerShare: &pb.Money{
							Amount:   item.Invoice.Share.RawTotalSellerShare.Amount,
							Currency: item.Invoice.Share.RawTotalSellerShare.Currency,
						},
						RoundupTotalSellerShare: &pb.Money{
							Amount:   item.Invoice.Share.RoundupTotalSellerShare.Amount,
							Currency: item.Invoice.Share.RoundupTotalSellerShare.Currency,
						},
					},
					SSO: &pb.FinanceOrderItemDetailList_OrderItemDetail_Item_ItemInvoice_ItemSSO{
						Rate:              item.Invoice.SSO.Rate,
						IsObliged:         item.Invoice.SSO.IsObliged,
						RawUnitPrice:      nil,
						RoundupUnitPrice:  nil,
						RawTotalPrice:     nil,
						RoundupTotalPrice: nil,
					},
					VAT: &pb.FinanceOrderItemDetailList_OrderItemDetail_Item_ItemInvoice_ItemVAT{
						Rate:              item.Invoice.VAT.SellerVat.Rate,
						IsObliged:         item.Invoice.VAT.SellerVat.IsObliged,
						RawUnitPrice:      nil,
						RoundupUnitPrice:  nil,
						RawTotalPrice:     nil,
						RoundupTotalPrice: nil,
					},
				},
			}

			if item.Invoice.Commission.RawUnitPrice != nil {
				financeItem.Invoice.Commission.RawUnitPrice = &pb.Money{
					Amount:   item.Invoice.Commission.RawUnitPrice.Amount,
					Currency: item.Invoice.Commission.RawUnitPrice.Currency,
				}
			}

			if item.Invoice.Commission.RoundupUnitPrice != nil {
				financeItem.Invoice.Commission.RoundupUnitPrice = &pb.Money{
					Amount:   item.Invoice.Commission.RoundupUnitPrice.Amount,
					Currency: item.Invoice.Commission.RoundupUnitPrice.Currency,
				}
			}

			if item.Invoice.Commission.RawTotalPrice != nil {
				financeItem.Invoice.Commission.RawTotalPrice = &pb.Money{
					Amount:   item.Invoice.Commission.RawTotalPrice.Amount,
					Currency: item.Invoice.Commission.RawTotalPrice.Currency,
				}
			}

			if item.Invoice.Commission.RoundupTotalPrice != nil {
				financeItem.Invoice.Commission.RoundupTotalPrice = &pb.Money{
					Amount:   item.Invoice.Commission.RoundupUnitPrice.Amount,
					Currency: item.Invoice.Commission.RoundupUnitPrice.Currency,
				}
			}

			if item.Invoice.SSO.RawUnitPrice != nil {
				financeItem.Invoice.SSO.RawUnitPrice = &pb.Money{
					Amount:   item.Invoice.SSO.RawUnitPrice.Amount,
					Currency: item.Invoice.SSO.RawUnitPrice.Currency,
				}
			}

			if item.Invoice.SSO.RoundupUnitPrice != nil {
				financeItem.Invoice.SSO.RoundupUnitPrice = &pb.Money{
					Amount:   item.Invoice.SSO.RoundupUnitPrice.Amount,
					Currency: item.Invoice.SSO.RoundupUnitPrice.Currency,
				}
			}

			if item.Invoice.SSO.RawTotalPrice != nil {
				financeItem.Invoice.SSO.RawTotalPrice = &pb.Money{
					Amount:   item.Invoice.SSO.RawTotalPrice.Amount,
					Currency: item.Invoice.SSO.RawTotalPrice.Currency,
				}
			}

			if item.Invoice.SSO.RoundupTotalPrice != nil {
				financeItem.Invoice.SSO.RoundupTotalPrice = &pb.Money{
					Amount:   item.Invoice.SSO.RoundupTotalPrice.Amount,
					Currency: item.Invoice.SSO.RoundupTotalPrice.Currency,
				}
			}

			if item.Invoice.VAT.SellerVat.RawUnitPrice != nil {
				financeItem.Invoice.VAT.RawUnitPrice = &pb.Money{
					Amount:   item.Invoice.VAT.SellerVat.RawUnitPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RawUnitPrice.Currency,
				}
			}

			if item.Invoice.VAT.SellerVat.RoundupUnitPrice != nil {
				financeItem.Invoice.VAT.RoundupUnitPrice = &pb.Money{
					Amount:   item.Invoice.VAT.SellerVat.RoundupUnitPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RoundupUnitPrice.Currency,
				}
			}

			if item.Invoice.VAT.SellerVat.RawTotalPrice != nil {
				financeItem.Invoice.VAT.RawTotalPrice = &pb.Money{
					Amount:   item.Invoice.VAT.SellerVat.RawTotalPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RawTotalPrice.Currency,
				}
			}

			if item.Invoice.VAT.SellerVat.RoundupTotalPrice != nil {
				financeItem.Invoice.VAT.RoundupTotalPrice = &pb.Money{
					Amount:   item.Invoice.VAT.SellerVat.RoundupTotalPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RoundupTotalPrice.Currency,
				}
			}

			if item.Attributes != nil {
				financeItem.Attributes = make(map[string]*pb.Attribute, len(item.Attributes))
				for attrKey, attribute := range item.Attributes {
					keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
					for keyTran, value := range attribute.KeyTranslate {
						keyTranslates[keyTran] = value
					}

					valTranslates := make(map[string]string, len(attribute.ValueTranslate))
					for valTran, value := range attribute.ValueTranslate {
						valTranslates[valTran] = value
					}

					financeItem.Attributes[attrKey] = &pb.Attribute{
						KeyTrans:   keyTranslates,
						ValueTrans: valTranslates,
					}
				}
			}

			financeOrderItem.Items = append(financeOrderItem.Items, financeItem)
		}

		financeReportList = append(financeReportList, financeOrderItem)
	}

	financeOrderItemDetailList := &pb.FinanceOrderItemDetailList{
		OrderItems: financeReportList,
	}

	serializedData, e := proto.Marshal(financeOrderItemDetailList)
	if e != nil {
		app.Globals.Logger.FromContext(ctx).Error("could not marshal financeOrderItemDetailList",
			"fn", "FinanceOrderItems",
			"request", req,
			"error", e)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "financeOrderItemDetailList",
		Meta: &pb.ResponseMetadata{
			Total:   uint32(total),
			Page:    req.Meta.Page,
			PerPage: req.Meta.PerPage,
		},
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(financeOrderItemDetailList),
			Value:   serializedData,
		},
	}

	return response, nil
}

func (server *Server) requestDataHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	reqName := RequestName(req.Name)
	userType := UserType(req.Meta.UTP)
	//reqADT := RequestADT(req.ADT)

	//var filterType FilterType
	var filterValue FilterValue
	var sortName string
	var sortDirection SortDirection
	var buyerMobile string
	if req.Meta.Filters != nil {
		//filterType = FilterType(req.Meta.Filters[0].UTP)
		filterValue = FilterValue(req.Meta.Filters[0].Value)
	}

	if req.Meta.Sorts != nil {
		sortName = req.Meta.Sorts[0].Name
		sortDirection = SortDirection(req.Meta.Sorts[0].Direction)
	}

	//if reqName == SellerOrderList && filterType != OrderStateFilterType {
	//	logger.Err("requestDataHandler() => request name %s mismatch with %s filter, request: %v", reqName, filterType, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with filter")
	//}

	//if (reqName == SellerReturnOrderList || reqName == BuyerReturnOrderDetailList) && filterType != OrderReturnStateFilter {
	//	logger.Err("requestDataHandler() => request name %s mismatch with %s filterType, request: %v", reqName, filterType, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with filterType")
	//}

	if userType == SellerUser &&
		reqName != SellerOrderList &&
		reqName != SellerOrderDetail &&
		reqName != SellerReturnOrderList &&
		reqName != SellerReturnOrderDetail &&
		reqName != SellerOrderDeliveredReports &&
		reqName != SellerOrderReturnReports &&
		reqName != SellerOrderShipmentReports &&
		reqName != SellerOrderDashboardReports &&
		reqName != SellerOrderCancelReports &&
		reqName != SellerApprovalPendingOrderReports &&
		reqName != SellerAllOrderReports {
		app.Globals.Logger.FromContext(ctx).Error("RequestName with userType mismatch", "fn", "requestDataHandler", "rn", reqName, "utp", userType, "request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "RN UTP Invalid")
	} else if userType == BuyerUser &&
		reqName != BuyerOrderDetailList &&
		reqName != BuyerAllOrderReports &&
		reqName != BuyerReturnOrderReports &&
		reqName != BuyerReturnOrderDetailList {
		app.Globals.Logger.FromContext(ctx).Error("RequestName with userType mismatch", "fn", "requestDataHandler", "rn", reqName, "utp", userType, "request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "RN UTP Invalid")
	} else if userType == OperatorUser &&
		reqName != OperatorOrderList &&
		reqName != OperatorOrderDetail &&
		reqName != OperatorOrderInvoiceDetail {
		app.Globals.Logger.FromContext(ctx).Error("RequestName with userType mismatch", "fn", "requestDataHandler", "rn", reqName, "utp", userType, "request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "RN UTP Invalid")
	}

	//if req.Meta.OID > 0 && reqADT == ListType {
	//	logger.Err("requestDataHandler() => %s orderId mismatch with %s requestADT, request: %v", userType, reqADT, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch Request name with RequestADT")
	//}

	//if req.Meta.OID > 0 && reqName != SellerOrderList && reqName != OperatorOrderList {
	//	logger.Err("requestDataHandler() => %s orderId mismatch with %s requestName, request: %v", userType, reqName, req)
	//	return nil, status.Error(codes.Code(future.BadRequest), "Mismatch OrderId with Request name")
	//}

	if userType == BuyerUser && reqName != BuyerAllOrderReports && reqName != BuyerReturnOrderReports {
		if reqName == BuyerOrderDetailList {
			if filterValue != "" {
				var findFlag = false
				for _, filter := range server.requestFilters[reqName] {
					if filter == filterValue {
						findFlag = true
						break
					}
				}

				if !findFlag && req.Meta.OID <= 0 {
					app.Globals.Logger.FromContext(ctx).Error("RequestName with filter mismatch", "fn", "requestDataHandler", "rn", reqName, "filter", filterValue, "request", req)
					return nil, status.Error(codes.Code(future.BadRequest), "RN Filter Invalid")
				}
			}
		} else {
			var findFlag = false
			for _, filter := range server.requestFilters[reqName] {
				if filter == filterValue {
					findFlag = true
					break
				}
			}

			if !findFlag {
				app.Globals.Logger.FromContext(ctx).Error("RequestName with filter mismatch", "fn", "requestDataHandler", "rn", reqName, "filter", filterValue, "request", req)
				return nil, status.Error(codes.Code(future.BadRequest), "RN Filter Invalid")
			}
		}
	} else if userType == SellerUser &&
		reqName != SellerOrderDashboardReports &&
		reqName != SellerOrderShipmentReports &&
		reqName != SellerOrderDeliveredReports &&
		reqName != SellerOrderReturnReports &&
		reqName != SellerOrderCancelReports &&
		reqName != SellerApprovalPendingOrderReports &&
		reqName != SellerAllOrderReports {
		var findFlag = false
		for _, filter := range server.requestFilters[reqName] {
			if filter == filterValue {
				findFlag = true
				break
			}
		}

		if !findFlag {
			app.Globals.Logger.FromContext(ctx).Error("RequestName with filter mismatch", "fn", "requestDataHandler", "rn", reqName, "filter", filterValue, "request", req)
			return nil, status.Error(codes.Code(future.BadRequest), "RN Filter Invalid")
		}
	}

	if reqName == OperatorOrderDetail && filterValue != "" {
		app.Globals.Logger.FromContext(ctx).Error("RequestName doesn't need any filter", "fn", "requestDataHandler", "rn", reqName, "filter", filterValue, "request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "RN Filter Invalid")
	} else if reqName == OperatorOrderList {
		if req.Meta.Ext != nil {
			buyerMobile = req.Meta.Ext["buyerMobile"]
		}
	}

	//if req.Meta.OID > 0 && reqName == SellerOrderList {
	//	return server.sellerGetOrderByIdHandler(ctx, , req.Meta.PID, filterValue)
	//}

	switch reqName {
	case SellerOrderList:
		return server.sellerOrderListHandler(ctx, req.Meta.OID, req.Meta.PID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case SellerOrderDetail:
		return server.sellerOrderDetailHandler(ctx, req.Meta.PID, req.Meta.OID, filterValue)
	case SellerReturnOrderList:
		return server.sellerReturnOrderListHandler(ctx, req.Meta.PID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case SellerReturnOrderDetail:
		return server.sellerReturnOrderDetailHandler(ctx, req.Meta.PID, req.Meta.OID, filterValue)

	case SellerOrderDashboardReports:
		return server.sellerOrderDashboardReportsHandler(ctx, req.Meta.UID)
	case SellerOrderShipmentReports:
		return server.sellerOrderShipmentReportsHandler(ctx, req.Meta.UID)
	case SellerOrderReturnReports:
		return server.sellerOrderReturnReportsHandler(ctx, req.Meta.UID)
	case SellerOrderDeliveredReports:
		return server.sellerOrderDeliveredReportsHandler(ctx, req.Meta.UID)
	case SellerOrderCancelReports:
		return server.sellerOrderCancelReportsHandler(ctx, req.Meta.UID)
	case SellerAllOrderReports:
		return server.sellerAllOrderReportsHandler(ctx, req.Meta.UID)
	case SellerApprovalPendingOrderReports:
		return server.sellerApprovalPendingOrderReportsHandler(ctx, req.Meta.UID)

	case BuyerOrderDetailList:
		return server.buyerOrderDetailListHandler(ctx, req.Meta.OID, req.Meta.UID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case BuyerAllOrderReports:
		return server.buyerAllOrderReportsHandler(ctx, req.Meta.UID)
	case BuyerReturnOrderReports:
		return server.buyerReturnOrderReportsHandler(ctx, req.Meta.UID)
	case BuyerReturnOrderDetailList:
		return server.buyerReturnOrderDetailListHandler(ctx, req.Meta.UID, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)

	case OperatorOrderList:
		return server.operatorOrderListHandler(ctx, req.Meta.OID, buyerMobile, filterValue, req.Meta.Page, req.Meta.PerPage, sortName, sortDirection)
	case OperatorOrderDetail:
		return server.operatorOrderDetailHandler(ctx, req.Meta.OID)
	case OperatorOrderInvoiceDetail:
		return server.operatorOrderInvoiceDetailHandler(ctx, req.Meta.OID)
	}

	return nil, status.Error(codes.Code(future.BadRequest), "Invalid Request")
}

func (server *Server) requestActionHandler(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	userType := UserType(req.Meta.UTP)
	var userAction actions.IAction

	app.Globals.Logger.FromContext(ctx).Debug("received request action", "fn", "requestActionHandler", "request", req)

	userActions, ok := server.actionStates[userType]
	if !ok {
		app.Globals.Logger.FromContext(ctx).Error("action userType not supported", "fn", "requestActionHandler", "utp", userType, "request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "User Action Invalid")
	}

	for _, action := range userActions {
		if action.ActionEnum().ActionName() == req.Meta.Action.ActionState {
			userAction = action
			break
		}
	}

	if userAction == nil {
		app.Globals.Logger.FromContext(ctx).Error("action invalid", "fn", "requestActionHandler", "action", req.Meta.Action.ActionState, "request", req)
		return nil, status.Error(codes.Code(future.BadRequest), "Action Invalid")
	}

	var reqActionData pb.ActionData
	if err := ptypes.UnmarshalAny(req.Data, &reqActionData); err != nil {
		app.Globals.Logger.FromContext(ctx).Error("Could not unmarshal reqActionData from request field", "fn", "requestActionHandler", "request", req, "error", err)
		return nil, status.Error(codes.Code(future.BadRequest), "Request Invalid")
	}

	subpackages := make([]events.ActionSubpackage, 0, len(reqActionData.Subpackages))
	for _, reqSubpackage := range reqActionData.Subpackages {
		subpackage := events.ActionSubpackage{
			SId: reqSubpackage.SID,
		}
		subpackage.Items = make([]events.ActionItem, 0, len(reqSubpackage.Items))
		for _, item := range reqSubpackage.Items {

			if item.Quantity <= 0 {
				app.Globals.Logger.FromContext(ctx).Error("action quantity invalid", "fn", "requestActionHandler", "action", req.Meta.Action.ActionState, "quantity", item.Quantity, "request", req)
				return nil, status.Error(codes.Code(future.BadRequest), "Action Quantity Invalid")
			}

			actionItem := events.ActionItem{
				InventoryId: item.InventoryId,
				Quantity:    item.Quantity,
			}
			if item.Reasons != nil {
				actionItem.Reasons = make([]entities.Reason, 0, len(item.Reasons))
				for _, reason := range item.Reasons {
					// convert to models.reason
					reas, err := app.Globals.Converter.Map(ctx, reason, entities.Reason{})
					if err != nil {
						return nil, err
					}
					rs := reas.(*entities.Reason)

					actionItem.Reasons = append(actionItem.Reasons, *rs)
				}
			}
			subpackage.Items = append(subpackage.Items, actionItem)
		}
		subpackages = append(subpackages, subpackage)
	}

	actionData := events.ActionData{
		SubPackages:    subpackages,
		Carrier:        reqActionData.Carrier,
		TrackingNumber: reqActionData.TrackingNumber,
	}

	event := events.New(events.Action, req.Meta.OID, req.Meta.PID, req.Meta.UID,
		req.Meta.Action.StateIndex, userAction,
		time.Unix(req.Time.GetSeconds(), int64(req.Time.GetNanos())), actionData)

	iFuture := future.Factory().SetCapacity(1).Build()
	iFrame := frame.Factory().SetFuture(iFuture).SetEvent(event).Build()
	server.flowManager.MessageHandler(ctx, iFrame)
	futureData := iFuture.Get()
	if futureData.Error() != nil {
		return nil, status.Error(codes.Code(futureData.Error().Code()), futureData.Error().Message())
	}

	eventResponse := futureData.Data().(events.ActionResponse)

	actionResponse := &pb.ActionResponse{
		OID:  eventResponse.OrderId,
		SIDs: eventResponse.SIds,
	}

	serializedResponse, err := proto.Marshal(actionResponse)
	if err != nil {
		app.Globals.Logger.FromContext(ctx).Error("could not marshal actionResponse", "fn", "requestActionHandler", "request", req, "response", actionResponse)
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	response := &pb.MessageResponse{
		Entity: "ActionResponse",
		Meta:   nil,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(actionResponse),
			Value:   serializedResponse,
		},
	}

	return response, nil
}

func (server *Server) PaymentGatewayHook(ctx context.Context, req *pg.PaygateHookRequest) (*pg.PaygateHookResponse, error) {

	app.Globals.Logger.FromContext(ctx).Debug("received payment response", "fn", "PaymentGatewayHook",
		"orderId", req.OrderID,
		"PaymentId", req.PaymentId,
		"InvoiceId", req.InvoiceId,
		"result", req.Result)
	futureData := server.flowManager.PaymentGatewayResult(ctx, req).Get()

	if futureData.Error() != nil {
		return nil, status.Error(codes.Code(futureData.Error().Code()), futureData.Error().Message())
	}

	return &pg.PaygateHookResponse{Ok: true}, nil
}

func (server Server) NewOrder(ctx context.Context, req *pb.RequestNewOrder) (*pb.ResponseNewOrder, error) {

	ifuture := app.Globals.UserService.AuthenticateContextToken(ctx).Get()
	//userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
	if ifuture.Error() != nil {
		app.Globals.Logger.FromContext(ctx).Error("UserService.AuthenticateContextToken failed",
			"fn", "RequestHandler", "error", ifuture.Error().Reason())
		return nil, status.Error(codes.Code(ifuture.Error().Code()), ifuture.Error().Message())
	}

	userAcl := ifuture.Data().(*acl.Acl)
	if uint64(userAcl.User().UserID) != req.Buyer.BuyerId {
		app.Globals.Logger.FromContext(ctx).Error("request userId with token userId mismatch", "fn", "NewOrder", "uid", req.Buyer.BuyerId, "token", userAcl.User().UserID)
		return nil, status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	}

	if ctx.Value(string(utils.CtxUserID)) == nil {
		if userAcl != nil {
			ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(userAcl.User().UserID))
		} else {
			ctx = context.WithValue(ctx, string(utils.CtxUserID), uint64(0))
		}
	}

	if ctx.Value(string(utils.CtxUserACL)) == nil {
		if userAcl != nil {
			ctx = context.WithValue(ctx, string(utils.CtxUserACL), userAcl)
		}
	}

	iFuture := future.Factory().SetCapacity(1).Build()
	iFrame := frame.Factory().SetDefaultHeader(frame.HeaderNewOrder, req).SetFuture(iFuture).Build()
	server.flowManager.MessageHandler(ctx, iFrame)
	futureData := iFuture.Get()

	//if futureData.Error() != nil {
	//	futureErr := futureData.Error()
	//	return nil, status.Error(codes.Code(futureErr.Code()), futureErr.Message())
	//}

	var responseNewOrder pb.ResponseNewOrder

	if ipgResponse, ok := futureData.Data().(entities.PaymentIPGResponse); ok {
		responseNewOrder = pb.ResponseNewOrder{
			Action: pb.ResponseNewOrder_Redirect,
			Response: &pb.ResponseNewOrder_Ipg{
				Ipg: &pb.IPGResponse{
					CallbackUrl: ipgResponse.CallBackUrl,
				},
			},
		}

	} else if mpgResponse, ok := futureData.Data().(entities.PaymentMPGResponse); ok {
		responseNewOrder = pb.ResponseNewOrder{
			Action: pb.ResponseNewOrder_MPG,
			Response: &pb.ResponseNewOrder_Mpg{
				Mpg: &pb.MPGResponse{
					HostRequest:     mpgResponse.HostRequest,
					HostRequestSign: mpgResponse.HostRequestSign,
				},
			},
		}
	} else {
		app.Globals.Logger.FromContext(ctx).Error("NewOrder received data of futureData invalid", "fn", "NewOrder", "data", futureData.Data())
		return nil, status.Error(codes.Code(future.InternalError), "Unknown Error")
	}

	return &responseNewOrder, nil
}

func (server Server) ReasonsList(ctx context.Context, in *pb.ReasonsListRequest) (list *pb.ReasonsListResponse, err error) {
	ls := make([]*pb.ReasonDetail, 0)
	for _, reason := range server.reasonConfigs {
		i := &pb.ReasonDetail{
			Key:            reason.Key,
			Translation:    reason.Translation,
			HasDescription: reason.HasDescription,
			Cancel:         reason.Cancel,
			Return:         reason.Return,
			IsActive:       reason.IsActive,
		}
		switch reason.Responsible {
		case utils.ReasonResponsibleBuyer:
			i.Responsible = pb.ReasonDetail_BUYER
		case utils.ReasonResponsibleSeller:
			i.Responsible = pb.ReasonDetail_SELLER
		case utils.ReasonResponsibleNone:
			i.Responsible = pb.ReasonDetail_NONE
		default:
			i.Responsible = pb.ReasonDetail_NONE
		}
		ls = append(ls, i)
	}
	list = &pb.ReasonsListResponse{
		Reasons: ls,
	}
	return
}

func (server Server) ReportOrderItems(req *pb.RequestReportOrderItems, srv pb.OrderService_ReportOrderItemsServer) error {

	iFuture := app.Globals.UserService.AuthenticateContextToken(srv.Context()).Get()
	//userAcl, err := app.Globals.UserService.AuthenticateContextToken(ctx)
	if iFuture.Error() != nil {
		app.Globals.Logger.FromContext(srv.Context()).Error("UserService.AuthenticateContextToken failed",
			"fn", "ReportOrderItems", "error", iFuture.Error().Reason())
		return status.Error(codes.Code(future.Forbidden), "User Not Authorized")
	}

	userAcl := iFuture.Data().(*acl.Acl)
	if userAcl.User().UserID <= 0 {
		app.Globals.Logger.Error("Token userId not authorized",
			"fn", "ReportOrderItems",
			"userId", userAcl.User().UserID)
		return status.Error(codes.Code(future.Forbidden), "User token not authorized")
	}

	if !userAcl.UserPerm().Has("order.state.all.view") || !userAcl.UserPerm().Has("order.state.all.action") {
		return status.Error(codes.Code(future.Forbidden), "User Not Permitted")
	}

	iFuture = server.flowManager.ReportOrderItems(srv.Context(), req, srv).Get()

	if iFuture.Error() != nil {
		return status.Error(codes.Code(iFuture.Error().Code()), iFuture.Error().Message())
	}

	return nil
}

func (server Server) VerifyUserSuccessOrder(ctx context.Context, req *pb.VerifyUserOrderRequest) (*pb.VerifyUserOrderResponse, error) {
	futureData := server.flowManager.VerifyUserSuccessOrder(ctx, req.UserId).Get()

	if futureData.Error() != nil {
		return nil, status.Error(codes.Code(futureData.Error().Code()), futureData.Error().Message())
	}

	app.Globals.Logger.FromContext(ctx).Debug("VerifyUserSuccessOrder received",
		"fn", "VerifyUserSuccessOrder",
		"uid", req.UserId,
		"IsSuccessOrder", futureData.Data().(bool))

	return &pb.VerifyUserOrderResponse{
		UserId:         req.UserId,
		IsSuccessOrder: futureData.Data().(bool),
	}, nil
}

func (server Server) Start() {
	port := strconv.Itoa(int(server.port))
	lis, err := net.Listen("tcp", server.address+":"+port)
	if err != nil {
		app.Globals.Logger.Error("Failed to listen to TCP on port", "fn", "Start", "port", port, "error", err)
	}
	app.Globals.Logger.Info("GRPC server started", "fn", "Start", "address", server.address, "port", port)

	customFunc := func(p interface{}) (err error) {
		app.Globals.Logger.Error("rpc panic recovered", "fn", "Start",
			"panic", p, "stacktrace", string(debug.Stack()))
		return grpc.Errorf(codes.Unknown, "panic triggered: %v", p)
	}

	//zapLogger, _ := zap.NewProduction()
	//stackDisableOpt := zap.AddStacktrace(stackTraceDisabler{})
	//noStackLogger := app.Globals.ZapLogger.WithOptions(stackDisableOpt)

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}

	uIntOpt := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_prometheus.UnaryServerInterceptor,
		grpc_recovery.UnaryServerInterceptor(opts...),
		myUnaryLogger(app.Globals.Logger),
		//grpc_zap.UnaryServerInterceptor(zapLogger),
	))

	sIntOpt := grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		grpc_prometheus.StreamServerInterceptor,
		grpc_recovery.StreamServerInterceptor(opts...),
		//grpc_zap.StreamServerInterceptor(app.Globals.ZapLogger),
	))

	// enable grpc prometheus interceptors to log timing info for grpc APIs
	grpc_prometheus.EnableHandlingTimeHistogram()

	//Start GRPC server and register the server
	grpcServer := grpc.NewServer(uIntOpt, sIntOpt)
	pb.RegisterOrderServiceServer(grpcServer, &server)
	pg.RegisterBankResultHookServer(grpcServer, &server)
	if err := grpcServer.Serve(lis); err != nil {
		app.Globals.Logger.Error("GRPC server start field", "fn", "Start", "error", err.Error())
		panic("GRPC server start field")
	}
}

func (server Server) StartTest() {
	port := strconv.Itoa(int(server.port))
	lis, err := net.Listen("tcp", server.address+":"+port)
	if err != nil {
		applog.GLog.Logger.Error("Failed to listen to TCP",
			"port", port,
			"error", err.Error())
	}
	applog.GLog.Logger.Debug("app started", "address", server.address, "port", port)

	// Start GRPC server and register the server
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, &server)
	pg.RegisterBankResultHookServer(grpcServer, &server)
	if err := grpcServer.Serve(lis); err != nil {
		panic("GRPC server start field")
	}
}

func myUnaryLogger(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		resp, err = handler(ctx, req)
		dur := time.Since(startTime)
		lg := log.FromContext(ctx)
		lg = lg.With(
			zap.Duration("took_sec", dur),
			zap.String("grpc.Method", path.Base(info.FullMethod)),
			zap.String("grpc.Service", path.Dir(info.FullMethod)[1:]),
			zap.String("grpc.Code", grpc.Code(err).String()),
		)
		lg.Debug("finished unary call")
		return
	}
}
