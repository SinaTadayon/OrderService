package converter

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	ordersrv "gitlab.faza.io/protos/order"
	"strconv"
	"time"
)

const (
	// ISO8601 standard time format
	ISO8601 = "2006-01-02T15:04:05-0700"
)

type iConverterImpl struct {
}

func NewConverter() IConverter {
	return &iConverterImpl{}
}

// Get *ordersrv.RequestNewOrder then map to *entities.Order
func (iconv iConverterImpl) Map(ctx context.Context, in interface{}, out interface{}) (interface{}, error) {

	var ok bool
	var newOrderDto *ordersrv.RequestNewOrder
	newOrderDto, ok = in.(*ordersrv.RequestNewOrder)
	if ok == false {
		applog.GLog.Logger.FromContext(ctx).Error("mapping from input type not supported",
			"fn", "Map",
			"in", in)
		return nil, errors.New("mapping from input type not supported")
	}

	_, ok = out.(entities.Order)
	if ok == false {
		applog.GLog.Logger.FromContext(ctx).Error("mapping to output type not supported",
			"fn", "Map",
			"in", in)
		return nil, errors.New("mapping to output type not supported")
	}

	return convert(ctx, newOrderDto)
}

func convert(ctx context.Context, newOrderDto *ordersrv.RequestNewOrder) (*entities.Order, error) {

	var order entities.Order
	timestamp := time.Now().UTC()

	if newOrderDto.Buyer == nil {
		applog.GLog.Logger.FromContext(ctx).Error("Buyer of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("Buyer of RequestNewOrder invalid")
	}

	if newOrderDto.Packages == nil || len(newOrderDto.Packages) == 0 {
		applog.GLog.Logger.FromContext(ctx).Error("Packages of RequestNewOrder empty",
			"fn", "convert")
		return nil, errors.New("Packages of RequestNewOrder empty")
	}

	if newOrderDto.Buyer.BuyerId <= 0 {
		applog.GLog.Logger.FromContext(ctx).Error("BuyerId of NewOrder invalid",
			"fn", "convert")
		return nil, errors.New("BuyerId of NewOrder invalid")
	}

	if newOrderDto.Invoice == nil {
		applog.GLog.Logger.FromContext(ctx).Error("Invoice of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("Invoice of RequestNewOrder invalid")
	}

	if newOrderDto.Invoice.Discount == nil {
		applog.GLog.Logger.FromContext(ctx).Error("Invoice.Discount of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("Invoice.Discount of RequestNewOrder invalid")
	}

	if newOrderDto.Invoice.GrandTotal == nil {
		applog.GLog.Logger.FromContext(ctx).Error("Invoice.GrandTotal of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("Invoice.GrandTotal of RequestNewOrder invalid")
	}

	if newOrderDto.Invoice.Subtotal == nil {
		applog.GLog.Logger.FromContext(ctx).Error("Invoice.Subtotal of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("Invoice.Subtotal of RequestNewOrder invalid")
	}

	if newOrderDto.Invoice.ShipmentTotal == nil {
		applog.GLog.Logger.FromContext(ctx).Error("Invoice.ShipmentTotal of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("Invoice.ShipmentTotal of RequestNewOrder invalid")
	}

	if newOrderDto.Invoice.Vat == nil {
		applog.GLog.Logger.FromContext(ctx).Error("Invoice.Vat of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("Invoice.Vat of RequestNewOrder invalid")
	}

	order.Platform = newOrderDto.Platform
	order.DocVersion = entities.DocumentVersion
	order.CreatedAt = timestamp
	order.UpdatedAt = timestamp

	order.BuyerInfo.BuyerId = newOrderDto.Buyer.BuyerId
	order.BuyerInfo.FirstName = newOrderDto.Buyer.FirstName
	order.BuyerInfo.LastName = newOrderDto.Buyer.LastName
	order.BuyerInfo.Mobile = newOrderDto.Buyer.Mobile
	order.BuyerInfo.Phone = newOrderDto.Buyer.Phone
	order.BuyerInfo.Email = newOrderDto.Buyer.Email
	order.BuyerInfo.NationalId = newOrderDto.Buyer.NationalId
	order.BuyerInfo.Gender = newOrderDto.Buyer.Gender
	order.BuyerInfo.IP = newOrderDto.Buyer.Ip

	if newOrderDto.Buyer.Finance != nil {
		//return nil, errors.New("buyer.finance of RequestNewOrder invalid")
		order.BuyerInfo.FinanceInfo.Iban = newOrderDto.Buyer.Finance.Iban
		order.BuyerInfo.FinanceInfo.CardNumber = newOrderDto.Buyer.Finance.CardNumber
		order.BuyerInfo.FinanceInfo.AccountNumber = newOrderDto.Buyer.Finance.AccountNumber
		order.BuyerInfo.FinanceInfo.BankName = newOrderDto.Buyer.Finance.BankName
	}

	if newOrderDto.Buyer.ShippingAddress == nil {
		applog.GLog.Logger.FromContext(ctx).Error("buyer.shippingAddress of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("buyer.shippingAddress of RequestNewOrder invalid")
	}

	order.BuyerInfo.ShippingAddress.FirstName = newOrderDto.Buyer.ShippingAddress.FirstName
	order.BuyerInfo.ShippingAddress.LastName = newOrderDto.Buyer.ShippingAddress.LastName
	order.BuyerInfo.ShippingAddress.Address = newOrderDto.Buyer.ShippingAddress.Address
	order.BuyerInfo.ShippingAddress.Mobile = newOrderDto.Buyer.ShippingAddress.Mobile
	order.BuyerInfo.ShippingAddress.Phone = newOrderDto.Buyer.ShippingAddress.Phone
	order.BuyerInfo.ShippingAddress.Country = newOrderDto.Buyer.ShippingAddress.Country
	order.BuyerInfo.ShippingAddress.City = newOrderDto.Buyer.ShippingAddress.City
	order.BuyerInfo.ShippingAddress.Province = newOrderDto.Buyer.ShippingAddress.Province
	order.BuyerInfo.ShippingAddress.Neighbourhood = newOrderDto.Buyer.ShippingAddress.Neighbourhood
	order.BuyerInfo.ShippingAddress.ZipCode = newOrderDto.Buyer.ShippingAddress.ZipCode
	err := setOrderLocation(newOrderDto.Buyer.ShippingAddress.Lat, newOrderDto.Buyer.ShippingAddress.Long, &order)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("Lat/Long ShippingAddress invalid",
			"fn", "convert")
		return nil, errors.New("Lat/Long ShippingAddress Invalid")
	}

	if newOrderDto.Invoice == nil {
		applog.GLog.Logger.FromContext(ctx).Error("invoice of RequestNewOrder is nil",
			"fn", "convert")
		return nil, errors.New("invoice of RequestNewOrder invalid")
	}

	order.Invoice.GrandTotal = entities.Money{
		Amount:   newOrderDto.Invoice.GrandTotal.Amount,
		Currency: newOrderDto.Invoice.GrandTotal.Currency,
	}

	order.Invoice.Subtotal = entities.Money{
		Amount:   newOrderDto.Invoice.Subtotal.Amount,
		Currency: newOrderDto.Invoice.Subtotal.Currency,
	}

	order.Invoice.Discount = entities.Money{
		Amount:   newOrderDto.Invoice.Discount.Amount,
		Currency: newOrderDto.Invoice.Discount.Currency,
	}

	order.Invoice.ShipmentTotal = entities.Money{
		Amount:   newOrderDto.Invoice.ShipmentTotal.Amount,
		Currency: newOrderDto.Invoice.ShipmentTotal.Currency,
	}

	order.Invoice.PaymentMethod = newOrderDto.Invoice.PaymentMethod
	order.Invoice.PaymentGateway = newOrderDto.Invoice.PaymentGateway
	order.Invoice.PaymentOption = nil

	//order.Invoice.Share =
	//order.Invoice.Commission =
	//order.Invoice.Voucher =
	//order.Invoice.CartRule =
	//order.Invoice.SSO =
	//order.Invoice.TAX =
	order.Invoice.VAT = &entities.VAT{
		Rate:         newOrderDto.Invoice.Vat.Value,
		RawTotal:     nil,
		RoundupTotal: nil,
		CreatedAt:    &timestamp,
		UpdatedAt:    &timestamp,
		Extended:     nil,
	}

	if newOrderDto.Invoice.Voucher != nil {
		order.Invoice.Voucher = &entities.Voucher{
			Percent:      float64(newOrderDto.Invoice.Voucher.Percent),
			AppliedPrice: nil,
			Price:        nil,
			Code:         newOrderDto.Invoice.Voucher.Code,
			Details:      nil,
			Settlement:   "",
			SettlementAt: nil,
			Reserved:     "",
			ReservedAt:   nil,
			Extended:     nil,
		}

		if newOrderDto.Invoice.Voucher.RawAppliedPrice != nil {
			order.Invoice.Voucher.AppliedPrice = &entities.Money{
				Amount:   newOrderDto.Invoice.Voucher.RawAppliedPrice.Amount,
				Currency: newOrderDto.Invoice.Voucher.RawAppliedPrice.Currency,
			}
		}

		if newOrderDto.Invoice.Voucher.RoundupAppliedPrice != nil {
			order.Invoice.Voucher.RoundupAppliedPrice = &entities.Money{
				Amount:   newOrderDto.Invoice.Voucher.RoundupAppliedPrice.Amount,
				Currency: newOrderDto.Invoice.Voucher.RoundupAppliedPrice.Currency,
			}
		}

		if newOrderDto.Invoice.Voucher.Price != nil {
			order.Invoice.Voucher.Price = &entities.Money{
				Amount:   newOrderDto.Invoice.Voucher.Price.Amount,
				Currency: newOrderDto.Invoice.Voucher.Price.Currency,
			}
		}

		if newOrderDto.Invoice.Voucher.Details != nil {
			order.Invoice.Voucher.Details = &entities.VoucherDetails{
				Title:            newOrderDto.Invoice.Voucher.Details.Title,
				Prefix:           newOrderDto.Invoice.Voucher.Details.Prefix,
				UseLimit:         newOrderDto.Invoice.Voucher.Details.UseLimit,
				Count:            newOrderDto.Invoice.Voucher.Details.Count,
				Length:           newOrderDto.Invoice.Voucher.Details.Length,
				Categories:       newOrderDto.Invoice.Voucher.Details.Info.Categories,
				Products:         newOrderDto.Invoice.Voucher.Details.Info.Products,
				Users:            newOrderDto.Invoice.Voucher.Details.Info.Users,
				Sellers:          newOrderDto.Invoice.Voucher.Details.Info.Sellers,
				IsFirstPurchase:  newOrderDto.Invoice.Voucher.Details.IsFirstPurchase,
				Type:             newOrderDto.Invoice.Voucher.Details.Type,
				MaxDiscountValue: newOrderDto.Invoice.Voucher.Details.MaxDiscountValue,
				MinBasketValue:   newOrderDto.Invoice.Voucher.Details.MinBasketValue,
				VoucherType:      newOrderDto.Invoice.Voucher.Details.VoucherType.String(),
				VoucherSponsor:   newOrderDto.Invoice.Voucher.Details.VoucherSponsor.String(),
				Extended:         nil,
			}

			temp, err := time.Parse(ISO8601, newOrderDto.Invoice.Voucher.Details.StartDate)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("Voucher startDate of RequestNewOrder Invalid",
					"fn", "convert",
					"startDate", newOrderDto.Invoice.Voucher.Details.StartDate)
				return nil, errors.New("Voucher startDate Invalid")
			}
			order.Invoice.Voucher.Details.StartDate = temp

			temp, err = time.Parse(ISO8601, newOrderDto.Invoice.Voucher.Details.EndDate)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("Voucher EndDate of RequestNewOrder Invalid",
					"fn", "convert",
					"EndDate", newOrderDto.Invoice.Voucher.Details.EndDate)
				return nil, errors.New("Voucher endDate Invalid")
			}
			order.Invoice.Voucher.Details.EndDate = temp
		} else {
			applog.GLog.Logger.FromContext(ctx).Error("voucher detail of RequestNewOrder is nil",
				"fn", "convert")
			return nil, errors.New("voucher detail is nil")
		}
	}

	order.Packages = make([]*entities.PackageItem, 0, len(newOrderDto.Packages))
	for _, pkgDto := range newOrderDto.Packages {

		if pkgDto.SellerId <= 0 {
			applog.GLog.Logger.FromContext(ctx).Error("SellerId of RequestNewOrder invalid",
				"fn", "convert")
			return nil, errors.New("SellerId of RequestNewOrder invalid")
		}

		if pkgDto.Invoice == nil {
			applog.GLog.Logger.FromContext(ctx).Error("Invoice of RequestNewOrder is nil",
				"fn", "convert")
			return nil, errors.New("Invoice of RequestNewOrder is nil")
		}

		if pkgDto.Shipment == nil {
			applog.GLog.Logger.FromContext(ctx).Error("Package Shipment of RequestNewOrder is nil",
				"fn", "convert")
			return nil, errors.New("Package Shipment of RequestNewOrder is nil")
		}

		if pkgDto.Items == nil || len(pkgDto.Items) == 0 {
			applog.GLog.Logger.FromContext(ctx).Error("Package Items of RequestNewOrder is empty",
				"fn", "convert")
			return nil, errors.New("Package Items of RequestNewOrder is empty")
		}

		if pkgDto.Invoice.ShipmentPrice == nil {
			applog.GLog.Logger.FromContext(ctx).Error("ShipmentPrice of Package Invoice is nil",
				"fn", "convert")
			return nil, errors.New("ShipmentPrice of Package Invoice is nil")
		}

		if pkgDto.Invoice.Discount == nil {
			applog.GLog.Logger.FromContext(ctx).Error("Discount of Package Invoice is nil",
				"fn", "convert")
			return nil, errors.New("Discount of Package Invoice is nil")
		}

		if pkgDto.Invoice.Subtotal == nil {
			applog.GLog.Logger.FromContext(ctx).Error("Subtotal of Package Invoice is nil",
				"fn", "convert")
			return nil, errors.New("Subtotal of Package Invoice is nil")
		}

		if pkgDto.Invoice.Sso == nil {
			applog.GLog.Logger.FromContext(ctx).Error("SSO of Package Invoice is nil",
				"fn", "convert")
			return nil, errors.New("SSO of Package Invoice is nil")
		}

		var pkgItem = &entities.PackageItem{
			PId:         pkgDto.SellerId,
			OrderId:     0,
			Version:     0,
			SellerInfo:  nil,
			ShopName:    pkgDto.ShopName,
			PayToSeller: nil,
			Subpackages: nil,
			Status:      "",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
			DeletedAt:   nil,
			Extended:    nil,

			ShippingAddress: entities.AddressInfo{
				FirstName:     order.BuyerInfo.ShippingAddress.FirstName,
				LastName:      order.BuyerInfo.ShippingAddress.LastName,
				Address:       order.BuyerInfo.ShippingAddress.Address,
				Phone:         order.BuyerInfo.ShippingAddress.Phone,
				Mobile:        order.BuyerInfo.ShippingAddress.Mobile,
				Country:       order.BuyerInfo.ShippingAddress.Country,
				City:          order.BuyerInfo.ShippingAddress.City,
				Province:      order.BuyerInfo.ShippingAddress.Province,
				Neighbourhood: order.BuyerInfo.ShippingAddress.Neighbourhood,
				Location:      order.BuyerInfo.ShippingAddress.Location,
				ZipCode:       order.BuyerInfo.ShippingAddress.ZipCode,
				Extended:      order.BuyerInfo.ShippingAddress.Extended,
			},

			Invoice: entities.PackageInvoice{
				Subtotal: entities.Money{
					Amount:   pkgDto.Invoice.Subtotal.Amount,
					Currency: pkgDto.Invoice.Subtotal.Currency,
				},
				Discount: entities.Money{
					Amount:   pkgDto.Invoice.Discount.Amount,
					Currency: pkgDto.Invoice.Discount.Currency,
				},
				ShipmentAmount: entities.Money{
					Amount:   pkgDto.Invoice.ShipmentPrice.Amount,
					Currency: pkgDto.Invoice.ShipmentPrice.Currency,
				},
				Share:      nil,
				Commission: nil,
				Voucher:    nil,
				CartRule:   nil,
				SSO: &entities.PackageSSO{
					Rate:         pkgDto.Invoice.Sso.Value,
					IsObliged:    pkgDto.Invoice.Sso.IsObliged,
					RawTotal:     nil,
					RoundupTotal: nil,
					CreatedAt:    &timestamp,
					UpdatedAt:    &timestamp,
					Extended:     nil,
				},
				VAT:      nil,
				TAX:      nil,
				Extended: nil,
			},

			ShipmentSpec: entities.ShipmentSpec{
				CarrierNames:   pkgDto.Shipment.CarrierNames,
				CarrierProduct: pkgDto.Shipment.CarrierProduct,
				CarrierType:    pkgDto.Shipment.CarrierType,
				ShippingCost:   nil,
				ReactionTime:   pkgDto.Shipment.ReactionTime,
				ShippingTime:   pkgDto.Shipment.ReturnTime,
				ReturnTime:     pkgDto.Shipment.ReturnTime,
				Details:        pkgDto.Shipment.Details,
			},
		}

		if pkgDto.Shipment.ShippingCost != nil {
			pkgItem.ShipmentSpec.ShippingCost = &entities.Money{
				Amount:   pkgDto.Shipment.ShippingCost.Amount,
				Currency: pkgDto.Shipment.ShippingCost.Currency,
			}
		}

		pkgItem.Subpackages = []*entities.Subpackage{
			{
				PId:       pkgDto.SellerId,
				CreatedAt: timestamp,
				UpdatedAt: timestamp,
				Items:     make([]*entities.Item, 0, len(pkgDto.Items)),
			},
		}
		for _, itemDto := range pkgDto.Items {
			if len(itemDto.InventoryId) == 0 {
				applog.GLog.Logger.FromContext(ctx).Error("InventoryId of RequestNewOrder invalid",
					"fn", "convert",
					"inventoryId", itemDto.InventoryId)
				return nil, errors.New("InventoryId of RequestNewOrder invalid")
			}

			if itemDto.Quantity <= 0 {
				applog.GLog.Logger.FromContext(ctx).Error("Items Quantity of RequestNewOrder invalid",
					"fn", "convert",
					"Quantity", itemDto.Quantity)
				return nil, errors.New("Items Quantity of RequestNewOrder invalid")
			}

			if itemDto.Invoice == nil {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice of RequestNewOrder is nil",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice of RequestNewOrder invalid")
			}

			if itemDto.Invoice.Unit == nil {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.Unit of RequestNewOrder is nil",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice.Unit of RequestNewOrder invalid")
			}

			if itemDto.Invoice.Discount == nil {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.Discount of RequestNewOrder is nil",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice.Discount of RequestNewOrder invalid")
			}

			if itemDto.Invoice.Special == nil {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.Special of RequestNewOrder is nil",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice.Special of RequestNewOrder invalid")
			}

			if itemDto.Invoice.Original == nil {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.Original of RequestNewOrder is nil",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice.Original of RequestNewOrder invalid")
			}

			if itemDto.Invoice.Total == nil {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.Total of RequestNewOrder is nil",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice.Total of RequestNewOrder invalid")
			}

			if itemDto.Invoice.Vat == nil {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.Vat of RequestNewOrder is nil",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice.Vat of RequestNewOrder invalid")
			}

			if itemDto.Invoice.ItemCommission < 0 {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.ItemCommission of RequestNewOrder is negative",
					"fn", "convert")
				return nil, errors.New("itemDto.Invoice.ItemCommission of RequestNewOrder invalid")
			}

			if itemDto.Invoice.Unit.Amount != itemDto.Invoice.Special.Amount &&
				itemDto.Invoice.Unit.Amount != itemDto.Invoice.Original.Amount {
				applog.GLog.Logger.FromContext(ctx).Error("itemDto.Invoice.Unit of RequestNewOrder doesn't equal to special or original",
					"fn", "convert",
					"unit", itemDto.Invoice.Unit,
					"special", itemDto.Invoice.Special,
					"original", itemDto.Invoice.Original)
				return nil, errors.New("itemDto.Invoice.Unit of RequestNewOrder invalid")
			}

			var item = &entities.Item{
				SKU:         itemDto.Sku,
				InventoryId: itemDto.InventoryId,
				Title:       itemDto.Title,
				Brand:       itemDto.Brand,
				Guaranty:    itemDto.Guaranty,
				Category:    itemDto.Category,
				Image:       itemDto.Image,
				Returnable:  itemDto.Returnable,
				Quantity:    itemDto.Quantity,
				Attributes:  nil,
				Invoice: entities.ItemInvoice{
					Unit: entities.Money{
						Amount:   itemDto.Invoice.Unit.Amount,
						Currency: itemDto.Invoice.Unit.Currency,
					},
					Total: entities.Money{
						Amount:   itemDto.Invoice.Total.Amount,
						Currency: itemDto.Invoice.Total.Currency,
					},
					Original: entities.Money{
						Amount:   itemDto.Invoice.Original.Amount,
						Currency: itemDto.Invoice.Original.Currency,
					},
					Special: entities.Money{
						Amount:   itemDto.Invoice.Special.Amount,
						Currency: itemDto.Invoice.Special.Currency,
					},
					Discount: entities.Money{
						Amount:   itemDto.Invoice.Discount.Amount,
						Currency: itemDto.Invoice.Discount.Currency,
					},
					SellerCommission: 0,
					Commission: &entities.ItemCommission{
						ItemCommission:    itemDto.Invoice.ItemCommission,
						RawUnitPrice:      nil,
						RoundupUnitPrice:  nil,
						RawTotalPrice:     nil,
						RoundupTotalPrice: nil,
						CreatedAt:         &timestamp,
						UpdatedAt:         &timestamp,
						Extended:          nil,
					},
					Share: nil,
					//SellerCommission:  itemDto.Invoice.SellerCommission,
					ApplicableVoucher: false,
					Voucher:           nil,
					CartRule:          nil,
					SSO:               nil,
					VAT: &entities.ItemVAT{
						SellerVat: &entities.SellerVAT{
							Rate:              itemDto.Invoice.Vat.Value,
							IsObliged:         itemDto.Invoice.Vat.IsObliged,
							RawUnitPrice:      nil,
							RoundupUnitPrice:  nil,
							RawTotalPrice:     nil,
							RoundupTotalPrice: nil,
							CreatedAt:         &timestamp,
							UpdatedAt:         &timestamp,
							Extended:          nil,
						},
						BusinessVat: &entities.BusinessVAT{
							Rate:              newOrderDto.Invoice.Vat.Value,
							RawUnitPrice:      nil,
							RoundupUnitPrice:  nil,
							RawTotalPrice:     nil,
							RoundupTotalPrice: nil,
							CreatedAt:         &timestamp,
							UpdatedAt:         &timestamp,
							Extended:          nil,
						},
						Extended: nil,
					},
					TAX:      nil,
					Extended: nil,
				},
			}

			if newOrderDto.Invoice.Voucher != nil && (newOrderDto.Invoice.Voucher.RoundupAppliedPrice != nil) {
				item.Invoice.ApplicableVoucher = true
			}

			if itemDto.Attributes != nil {
				item.Attributes = make(map[string]*entities.Attribute, len(itemDto.Attributes))
				for attrKey, attribute := range itemDto.Attributes {
					keyTranslates := make(map[string]string, len(attribute.KeyTrans))
					for keyTran, value := range attribute.KeyTrans {
						keyTranslates[keyTran] = value
					}
					valTranslates := make(map[string]string, len(attribute.ValueTrans))
					for valTran, value := range attribute.ValueTrans {
						valTranslates[valTran] = value
					}
					item.Attributes[attrKey] = &entities.Attribute{
						KeyTranslate:   keyTranslates,
						ValueTranslate: valTranslates,
					}
				}
			}

			pkgItem.Subpackages[0].Items = append(pkgItem.Subpackages[0].Items, item)
		}
		order.Packages = append(order.Packages, pkgItem)
	}

	return &order, nil
}

func setOrderLocation(lat, long string, order *entities.Order) error {
	var latitude, longitude float64
	var err error
	if len(lat) == 0 || len(long) == 0 {
		return nil
	}

	if latitude, err = strconv.ParseFloat(lat, 64); err != nil {
		return err
	}

	if longitude, err = strconv.ParseFloat(long, 64); err != nil {
		return err
	}

	order.BuyerInfo.ShippingAddress.Location = &entities.Location{}
	order.BuyerInfo.ShippingAddress.Location.Type = "Point"
	order.BuyerInfo.ShippingAddress.Location.Coordinates = []float64{longitude, latitude}
	return nil
}
