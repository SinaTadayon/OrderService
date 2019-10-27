package converters

import (
	"errors"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	ordersrv "gitlab.faza.io/protos/order"
	"strconv"
)

type iConverterImpl struct {

}

func NewConverter() IConverter {
	return &iConverterImpl{}
}

// Get *ordersrv.NewOrderRequest then map to *entities.Order
func (iconv iConverterImpl) Map(in interface{}, out interface{}) (interface{}, error) {

	var ok bool
	var newOrderDto *ordersrv.NewOrderRequest
	newOrderDto ,ok = in.(*ordersrv.NewOrderRequest)
	if ok == false {
		return nil, errors.New("mapping from input type not supported")
	}

	_, ok = out.(entities.Order)
	if ok == false {
		return nil, errors.New("mapping to output type not supported")
	}

	return convert(newOrderDto)
}

func convert(newOrderDto *ordersrv.NewOrderRequest) (*entities.Order, error) {

	var order entities.Order

	if newOrderDto.Buyer == nil {
		return nil, errors.New("buyer of newOrderRequest invalid")
	}

	order.BuyerInfo.FirstName = newOrderDto.Buyer.FirstName
	order.BuyerInfo.LastName = newOrderDto.Buyer.LastName
	order.BuyerInfo.Mobile = newOrderDto.Buyer.Mobile
	order.BuyerInfo.Email = newOrderDto.Buyer.Email
	order.BuyerInfo.NationalId = newOrderDto.Buyer.NationalId
	order.BuyerInfo.Gender = newOrderDto.Buyer.Gender
	order.BuyerInfo.IP = newOrderDto.Buyer.Ip

	if newOrderDto.Buyer.Finance == nil {
		return nil, errors.New("buyer.finance of newOrderRequest invalid")
	}

	order.BuyerInfo.FinanceInfo.Iban = newOrderDto.Buyer.Finance.Iban
	order.BuyerInfo.FinanceInfo.CardNumber = newOrderDto.Buyer.Finance.CardNumber
	order.BuyerInfo.FinanceInfo.AccountNumber = newOrderDto.Buyer.Finance.AccountNumber
	order.BuyerInfo.FinanceInfo.BankName = newOrderDto.Buyer.Finance.BankName

	if newOrderDto.Buyer.ShippingAddress == nil {
		return nil, errors.New("buyer.shippingAddress of newOrderRequest invalid")
	}

	order.BuyerInfo.ShippingAddress.Address = newOrderDto.Buyer.ShippingAddress.Address
	order.BuyerInfo.ShippingAddress.Phone = newOrderDto.Buyer.ShippingAddress.Phone
	order.BuyerInfo.ShippingAddress.Country = newOrderDto.Buyer.ShippingAddress.Country
	order.BuyerInfo.ShippingAddress.City = newOrderDto.Buyer.ShippingAddress.City
	order.BuyerInfo.ShippingAddress.Province = newOrderDto.Buyer.ShippingAddress.Province
	order.BuyerInfo.ShippingAddress.Neighbourhood = newOrderDto.Buyer.ShippingAddress.Neighbourhood
	order.BuyerInfo.ShippingAddress.ZipCode = newOrderDto.Buyer.ShippingAddress.ZipCode
	setOrderLocation(newOrderDto.Buyer.ShippingAddress.Lat, newOrderDto.Buyer.ShippingAddress.Long, &order)

	if newOrderDto.Amount == nil {
		return nil, errors.New("amount of newOrderRequest invalid")
	}

	order.Amount.Total = newOrderDto.Amount.Total
	order.Amount.Payable = newOrderDto.Amount.Payable
	order.Amount.Discount = newOrderDto.Amount.Discount
	order.Amount.ShipmentTotal = newOrderDto.Amount.ShipmentTotal
	order.Amount.Currency = newOrderDto.Amount.Currency
	order.Amount.PaymentMethod = newOrderDto.Amount.PaymentMethod
	order.Amount.PaymentOption = newOrderDto.Amount.PaymentOption

	if newOrderDto.Amount.Voucher != nil {
		order.Amount.Voucher.Amount = newOrderDto.Amount.Voucher.Amount
		order.Amount.Voucher.Code = newOrderDto.Amount.Voucher.Code
		// TODO implement voucher details
	}

	for _, item := range newOrderDto.Items {
		var newItem = entities.Item{}

		if len(item.InventoryId) == 0 {
			return nil, errors.New("inventoryId of newOrderRequest invalid")
		}

		newItem.InventoryId	= item.InventoryId
		newItem.Title = item.Title
		newItem.Quantity = item.Quantity
		newItem.Brand = item.Brand
		newItem.Warranty = item.Warranty
		newItem.Categories = item.Categories
		newItem.Image = item.Image
		newItem.Returnable = item.Returnable
		newItem.SellerInfo.SellerId = item.SellerId

		if item.Attributes != nil {
			newItem.Attributes.Width = item.Attributes.Width
			newItem.Attributes.Height = item.Attributes.Height
			newItem.Attributes.Length = item.Attributes.Length
			newItem.Attributes.Weight = item.Attributes.Weight
			newItem.Attributes.Color = item.Attributes.Color
			newItem.Attributes.Materials = item.Attributes.Materials
			// Todo Implements extra attributes
		}

		if item.Price == nil {
			return nil, errors.New("item price of newOrderRequest invalid")
		}

		newItem.PriceInfo.Unit = item.Price.Unit
		newItem.PriceInfo.Total = item.Price.Total
		newItem.PriceInfo.Payable = item.Price.Payable
		newItem.PriceInfo.Discount = item.Price.Discount
		newItem.PriceInfo.SellerCommission = item.Price.SellerCommission
		newItem.PriceInfo.Currency = item.Price.Currency

		if item.Shipment == nil {
			return nil, errors.New("item shipment of newOrderRequest invalid")
		}

		newItem.ShipmentSpec.CarrierName = item.Shipment.CarrierName
		newItem.ShipmentSpec.CarrierProduct = item.Shipment.CarrierProduct
		newItem.ShipmentSpec.CarrierType = item.Shipment.CarrierType
		newItem.ShipmentSpec.ShippingAmount = item.Shipment.ShippingAmount
		newItem.ShipmentSpec.VoucherAmount = item.Shipment.VoucherAmount
		newItem.ShipmentSpec.Currency = item.Shipment.Currency
		newItem.ShipmentSpec.ReactionTime = item.Shipment.ReactionTime
		newItem.ShipmentSpec.ShippingTime = item.Shipment.ShippingTime
		newItem.ShipmentSpec.ReturnTime = item.Shipment.ReturnTime
		newItem.ShipmentSpec.Details = item.Shipment.Details

		order.Items = append(order.Items, newItem)
	}

	return &order, nil
}

func setOrderLocation(lat, long string , order *entities.Order) {
	var latitude, longitude float64
	var err error
	if len(lat) == 0 || len(long) == 0 {
		return
	}

	if latitude, err = strconv.ParseFloat(lat, 64); err != nil {
		logger.Err("shippingAddress.latitude of newOrderRequest ")
		return
	}

	if longitude, err = strconv.ParseFloat(long, 64); err != nil {
		logger.Err("shippingAddress.longitude of newOrderRequest ")
		return
	}

	order.BuyerInfo.ShippingAddress.Location.Type = "Point"
	order.BuyerInfo.ShippingAddress.Location.Coordinates = []float64{longitude, latitude}
}