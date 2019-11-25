package converter

import (
	"github.com/pkg/errors"
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

// Get *ordersrv.RequestNewOrder then map to *entities.Order
func (iconv iConverterImpl) Map(in interface{}, out interface{}) (interface{}, error) {

	var ok bool
	var newOrderDto ordersrv.RequestNewOrder
	newOrderDto, ok = in.(ordersrv.RequestNewOrder)
	if ok == false {
		return nil, errors.New("mapping from input type not supported")
	}

	_, ok = out.(entities.Order)
	if ok == false {
		return nil, errors.New("mapping to output type not supported")
	}

	return convert(&newOrderDto)
}

func convert(newOrderDto *ordersrv.RequestNewOrder) (*entities.Order, error) {

	var order entities.Order

	if newOrderDto.Buyer == nil {
		return nil, errors.New("buyer of RequestNewOrder invalid")
	}

	if newOrderDto.Items == nil || len(newOrderDto.Items) == 0 {
		return nil, errors.New("items of RequestNewOrder empty")
	}

	order.BuyerInfo.BuyerId = newOrderDto.Buyer.BuyerId
	order.BuyerInfo.FirstName = newOrderDto.Buyer.FirstName
	order.BuyerInfo.LastName = newOrderDto.Buyer.LastName
	order.BuyerInfo.Mobile = newOrderDto.Buyer.Mobile
	order.BuyerInfo.Phone = newOrderDto.Buyer.Phone
	order.BuyerInfo.Email = newOrderDto.Buyer.Email
	order.BuyerInfo.NationalId = newOrderDto.Buyer.NationalId
	order.BuyerInfo.Gender = newOrderDto.Buyer.Gender
	order.BuyerInfo.IP = newOrderDto.Buyer.Ip

	if newOrderDto.Buyer.Finance == nil {
		return nil, errors.New("buyer.finance of RequestNewOrder invalid")
	}

	order.BuyerInfo.FinanceInfo.Iban = newOrderDto.Buyer.Finance.Iban
	order.BuyerInfo.FinanceInfo.CardNumber = newOrderDto.Buyer.Finance.CardNumber
	order.BuyerInfo.FinanceInfo.AccountNumber = newOrderDto.Buyer.Finance.AccountNumber
	order.BuyerInfo.FinanceInfo.BankName = newOrderDto.Buyer.Finance.BankName

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

	if newOrderDto.Amount == nil {
		return nil, errors.New("amount of RequestNewOrder invalid")
	}

	order.Amount.Total = newOrderDto.Amount.Total
	order.Amount.Subtotal = newOrderDto.Amount.Subtotal
	order.Amount.Discount = newOrderDto.Amount.Discount
	order.Amount.ShipmentTotal = newOrderDto.Amount.ShipmentTotal
	order.Amount.Currency = newOrderDto.Amount.Currency
	order.Amount.PaymentMethod = newOrderDto.Amount.PaymentMethod
	order.Amount.PaymentOption = newOrderDto.Amount.PaymentOption

	if newOrderDto.Amount.Voucher != nil {
		order.Amount.Voucher = &entities.Voucher{
			Amount: newOrderDto.Amount.Voucher.Amount,
			Code:   newOrderDto.Amount.Voucher.Code,
		}
		// TODO implement voucher details
	}

	order.Items = make([]entities.Item, 0, len(newOrderDto.Items))

	for _, item := range newOrderDto.Items {
		if len(item.InventoryId) == 0 {
			return nil, errors.New("inventoryId of RequestNewOrder invalid")
		}

		if item.Quantity <= 0 {
			return nil, errors.New("item Count of RequestNewOrder invalid")
		}

		for i := 0; i < int(item.Quantity); i++ {
			var newItem = entities.Item{}
			newItem.SellerInfo.SellerId = item.SellerId
			newItem.InventoryId = item.InventoryId
			newItem.Title = item.Title
			newItem.Brand = item.Brand
			newItem.Guaranty = item.Guaranty
			newItem.Category = item.Category
			newItem.Image = item.Image
			newItem.Quantity = item.Quantity
			newItem.Returnable = item.Returnable

			newItem.Attributes = item.Attributes

			if item.Price == nil {
				return nil, errors.New("item price of RequestNewOrder invalid")
			}

			newItem.Price.Unit = item.Price.Unit
			newItem.Price.Total = item.Price.Total
			newItem.Price.Discount = item.Price.Discount
			newItem.Price.Original = item.Price.Original
			newItem.Price.Special = item.Price.Special
			newItem.Price.SellerCommission = item.Price.SellerCommission
			newItem.Price.Currency = item.Price.Currency

			if item.Shipment == nil {
				return nil, errors.New("item shipment of RequestNewOrder invalid")
			}

			newItem.ShipmentSpec.CarrierName = item.Shipment.CarrierName
			newItem.ShipmentSpec.CarrierProduct = item.Shipment.CarrierProduct
			newItem.ShipmentSpec.CarrierType = item.Shipment.CarrierType
			newItem.ShipmentSpec.ShippingCost = item.Shipment.ShippingCost
			newItem.ShipmentSpec.VoucherAmount = item.Shipment.VoucherAmount
			newItem.ShipmentSpec.Currency = item.Shipment.Currency
			newItem.ShipmentSpec.ReactionTime = item.Shipment.ReactionTime
			newItem.ShipmentSpec.ShippingTime = item.Shipment.ShippingTime
			newItem.ShipmentSpec.ReturnTime = item.Shipment.ReturnTime
			newItem.ShipmentSpec.Details = item.Shipment.Details

			order.Items = append(order.Items, newItem)
		}
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
