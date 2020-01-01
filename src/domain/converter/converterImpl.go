package converter

import (
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
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
func (iconv iConverterImpl) Map(in interface{}, out interface{}) (interface{}, error) {

	var ok bool
	var newOrderDto *ordersrv.RequestNewOrder
	newOrderDto, ok = in.(*ordersrv.RequestNewOrder)
	if ok == false {
		return nil, errors.New("mapping from input type not supported")
	}

	_, ok = out.(entities.Order)
	if ok == false {
		return nil, errors.New("mapping to output type not supported")
	}

	return convert(newOrderDto)
}

func convert(newOrderDto *ordersrv.RequestNewOrder) (*entities.Order, error) {

	var order entities.Order

	if newOrderDto.Buyer == nil {
		return nil, errors.New("Buyer of RequestNewOrder invalid")
	}

	if newOrderDto.Packages == nil || len(newOrderDto.Packages) == 0 {
		return nil, errors.New("Packages of RequestNewOrder empty")
	}

	if newOrderDto.Buyer.BuyerId <= 0 {
		return nil, errors.New("BuyerId of NewOrder invalid")
	}

	order.Platform = newOrderDto.Platform

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
	setOrderLocation(newOrderDto.Buyer.ShippingAddress.Lat, newOrderDto.Buyer.ShippingAddress.Long, &order)

	if newOrderDto.Invoice == nil {
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

	if newOrderDto.Invoice.Voucher != nil {
		order.Invoice.Voucher = &entities.Voucher{
			Percent: float64(newOrderDto.Invoice.Voucher.Percent),
			Price:   nil,
			Code:    newOrderDto.Invoice.Voucher.Code,
			Details: nil,
		}

		if newOrderDto.Invoice.Voucher.Price != nil {
			order.Invoice.Voucher.Price = &entities.Money{
				Amount:   newOrderDto.Invoice.Voucher.Price.Amount,
				Currency: newOrderDto.Invoice.Voucher.Price.Currency,
			}
		}

		if newOrderDto.Invoice.Voucher.Details != nil {
			order.Invoice.Voucher.Details = &entities.VoucherDetails{
				Type:             newOrderDto.Invoice.Voucher.Details.Type,
				MaxDiscountValue: newOrderDto.Invoice.Voucher.Details.MaxDiscountValue,
				MinBasketValue:   newOrderDto.Invoice.Voucher.Details.MinBasketValue,
			}

			temp, err := time.Parse(ISO8601, newOrderDto.Invoice.Voucher.Details.StartDate)
			if err != nil {
				return nil, errors.New("Voucher startDate Invalid")
			}
			order.Invoice.Voucher.Details.StartDate = temp

			temp, err = time.Parse(ISO8601, newOrderDto.Invoice.Voucher.Details.EndDate)
			if err != nil {
				return nil, errors.New("Voucher endDate Invalid")
			}
			order.Invoice.Voucher.Details.EndDate = temp
		}
	}

	order.Packages = make([]entities.PackageItem, 0, len(newOrderDto.Packages))
	for _, pkgDto := range newOrderDto.Packages {

		if pkgDto.SellerId <= 0 {
			return nil, errors.New("PId of RequestNewOrder invalid")
		}

		if pkgDto.Invoice == nil {
			return nil, errors.New("Invoice of RequestNewOrder is nil")
		}

		if pkgDto.Shipment == nil {
			return nil, errors.New("Shipment of RequestNewOrder is nil")
		}

		if pkgDto.Items == nil || len(pkgDto.Items) == 0 {
			return nil, errors.New("Items of RequestNewOrder is empty")
		}

		var pkgItem = entities.PackageItem{
			PId:         pkgDto.SellerId,
			OrderId:     0,
			Version:     0,
			SellerInfo:  nil,
			ShopName:    pkgDto.ShopName,
			PayToSeller: nil,
			Subpackages: nil,
			Status:      "",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
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
				PId:   pkgDto.SellerId,
				Items: make([]*entities.Item, 0, len(pkgDto.Items)),
			},
		}
		for _, itemDto := range pkgDto.Items {
			if len(itemDto.InventoryId) == 0 {
				return nil, errors.New("InventoryId of RequestNewOrder invalid")
			}

			if itemDto.Quantity <= 0 {
				return nil, errors.New("Items Quantity of RequestNewOrder invalid")
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
				Attributes:  itemDto.Attributes,
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

					SellerCommission:  itemDto.Invoice.SellerCommission,
					ApplicableVoucher: false,
				},
			}

			if newOrderDto.Invoice.Voucher != nil && (newOrderDto.Invoice.Voucher.Price != nil || newOrderDto.Invoice.Voucher.Percent > 0) {
				item.Invoice.ApplicableVoucher = true
			}

			pkgItem.Subpackages[0].Items = append(pkgItem.Subpackages[0].Items, item)
		}
		order.Packages = append(order.Packages, pkgItem)
	}

	return &order, nil
}

func setOrderLocation(lat, long string, order *entities.Order) {
	var latitude, longitude float64
	var err error
	if len(lat) == 0 || len(long) == 0 {
		return
	}

	if latitude, err = strconv.ParseFloat(lat, 64); err != nil {
		logger.Err("shippingAddress.latitude of RequestNewOrder ")
		return
	}

	if longitude, err = strconv.ParseFloat(long, 64); err != nil {
		logger.Err("shippingAddress.longitude of RequestNewOrder ")
		return
	}

	order.BuyerInfo.ShippingAddress.Location = &entities.Location{}
	order.BuyerInfo.ShippingAddress.Location.Type = "Point"
	order.BuyerInfo.ShippingAddress.Location.Coordinates = []float64{longitude, latitude}
}
