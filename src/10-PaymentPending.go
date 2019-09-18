package main

import (
	"errors"
	"strings"
	"time"

	"github.com/go-ozzo/ozzo-validation/is"

	validation "github.com/go-ozzo/ozzo-validation"
)

type PaymentPendingRequest struct {
	OrderNumber   string
	PaymentDetail PaymentDetail
	SystemPayment SystemPayment
	Status        Status
	Buyer         Buyer
	Amount        Amount
	ShipmentInfo  ShipmentInfo
	Items         []Item
	CreatedAt     time.Time
}
type ShipmentInfo struct {
	ShipmentDetail ShipmentDetail
}
type ShipmentDetail struct {
	ShipmentProvider       string
	ShipmentTrackingNumber string
	Image                  string
}
type Status struct {
	Current   string
	CreatedAt time.Time
	History   []StatusHistory
}
type StatusHistory struct {
	Status    string
	Agent     string
	CreatedAt time.Time
	Reason    string
}
type PaymentDetail struct {
	Status      bool
	Description string
	Request     string
	Response    string
	CreatedAt   time.Time
}
type SystemPayment struct {
	Buyer  []BuyerDetail
	Seller []SellerDetail
	Market []MarketDetail
}
type BuyerDetail struct {
	Status      bool
	Description string
	Request     string
	Response    string
	CreatedAt   time.Time
}
type SellerDetail struct {
	Status      bool
	Description string
	Request     string
	Response    string
	CreatedAt   time.Time
}
type MarketDetail struct {
	Status      bool
	Description string
	Request     string
	Response    string
	CreatedAt   time.Time
}
type Amount struct {
	Total    float64
	Payable  float64
	Discount float64
}
type Item struct {
	Sku        string
	Title      string
	Quantity   int32
	Brand      string
	Warranty   string
	Categories string
	Seller     ItemSeller
	Price      ItemPrice
	Shipment   ItemShipment
}
type Buyer struct {
	FirstName  string
	LastName   string
	Mobile     string
	Email      string
	NationalId string
	Finance    BuyerFinance
	Address    BuyerAddress
	IP         string
}
type BuyerFinance struct {
	Iban string
}
type BuyerAddress struct {
	Address string
	Phone   string
	Country string
	City    string
	State   string
	Lat     string
	Lan     string
	ZipCode string
}
type ItemSeller struct {
	Title            string
	FirstName        string
	LastName         string
	Mobile           string
	Email            string
	NationalId       string
	CompanyName      string
	RegistrationName string
	EconomicCode     string
	Finance          ItemSellerFinance
	Address          ItemSellerAddress
}
type ItemSellerFinance struct {
	Iban string
}
type ItemSellerAddress struct {
	Title   string
	Address string
	Phone   string
	Country string
	City    string
	State   string
	Lat     string
	Lan     string
	ZipCode string
}
type ItemPrice struct {
	Unit             float64
	Total            float64
	Payable          float64
	Discount         float64
	SellerCommission float64
}
type ItemShipment struct {
	ProviderName   string
	ReactionTime   int32
	ShippingTime   int32
	ReturnTime     int32
	ShipmentDetail string
}

func (ppr *PaymentPendingRequest) validate() error {
	var errValidation []string
	// Validate order number
	errPaymentRequest := validation.ValidateStruct(ppr,
		validation.Field(&ppr.OrderNumber, validation.Required, validation.Length(5, 250)),
	)
	if errPaymentRequest != nil {
		errValidation = append(errValidation, errPaymentRequest.Error())
	}

	// Validate Buyer
	errPaymentRequestBuyer := validation.ValidateStruct(&ppr.Buyer,
		validation.Field(&ppr.Buyer.FirstName, validation.Required),
		validation.Field(&ppr.Buyer.LastName, validation.Required),
		validation.Field(&ppr.Buyer.Email, validation.Required, is.Email),
		validation.Field(&ppr.Buyer.NationalId, validation.Required, validation.Length(10, 10)),
		validation.Field(&ppr.Buyer.Mobile, validation.Required),
	)
	if errPaymentRequestBuyer != nil {
		errValidation = append(errValidation, errPaymentRequestBuyer.Error())
	}

	// Validate Buyer finance
	errPaymentRequestBuyerFinance := validation.ValidateStruct(&ppr.Buyer.Finance,
		validation.Field(&ppr.Buyer.Finance.Iban, validation.Required, validation.Length(26, 26)),
	)
	if errPaymentRequestBuyerFinance != nil {
		errValidation = append(errValidation, errPaymentRequestBuyerFinance.Error())
	}

	// Validate Buyer address
	errPaymentRequestBuyerAddress := validation.ValidateStruct(&ppr.Buyer.Address,
		validation.Field(&ppr.Buyer.Address.Address, validation.Required),
		validation.Field(&ppr.Buyer.Address.State, validation.Required),
		validation.Field(&ppr.Buyer.Address.City, validation.Required),
		validation.Field(&ppr.Buyer.Address.Country, validation.Required),
		validation.Field(&ppr.Buyer.Address.ZipCode, validation.Required),
		validation.Field(&ppr.Buyer.Address.Phone, validation.Required),
	)
	if errPaymentRequestBuyerAddress != nil {
		errValidation = append(errValidation, errPaymentRequestBuyerAddress.Error())
	}

	// Validate amount
	errPaymentRequestAmount := validation.ValidateStruct(&ppr.Amount,
		validation.Field(&ppr.Amount.Total, validation.Required),
		validation.Field(&ppr.Amount.Discount, validation.Required),
		validation.Field(&ppr.Amount.Payable, validation.Required),
	)
	if errPaymentRequestAmount != nil {
		errValidation = append(errValidation, errPaymentRequestAmount.Error())
	}

	if len(ppr.Items) != 0 {
		for i := range ppr.Items {
			// Validate amount
			errPaymentRequestItems := validation.ValidateStruct(&ppr.Items[i],
				validation.Field(&ppr.Items[i].Sku, validation.Required),
				validation.Field(&ppr.Items[i].Quantity, validation.Required),
				validation.Field(&ppr.Items[i].Title, validation.Required),
				validation.Field(&ppr.Items[i].Categories, validation.Required),
				validation.Field(&ppr.Items[i].Brand, validation.Required),
			)
			if errPaymentRequestItems != nil {
				errValidation = append(errValidation, errPaymentRequestItems.Error())
			}

			errPaymentRequestItemsSeller := validation.ValidateStruct(&ppr.Items[i].Seller,
				validation.Field(&ppr.Items[i].Seller.Title, validation.Required),
				validation.Field(&ppr.Items[i].Seller.FirstName, validation.Required),
				validation.Field(&ppr.Items[i].Seller.LastName, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Mobile, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Email, validation.Required),
			)
			if errPaymentRequestItemsSeller != nil {
				errValidation = append(errValidation, errPaymentRequestItemsSeller.Error())
			}

			errPaymentRequestItemsSellerFinance := validation.ValidateStruct(&ppr.Items[i].Seller.Finance,
				validation.Field(&ppr.Items[i].Seller.Finance.Iban, validation.Required),
			)
			if errPaymentRequestItemsSellerFinance != nil {
				errValidation = append(errValidation, errPaymentRequestItemsSellerFinance.Error())
			}

			errPaymentRequestItemsSellerAddress := validation.ValidateStruct(&ppr.Items[i].Seller.Address,
				validation.Field(&ppr.Items[i].Seller.Address.Title, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Address.Address, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Address.Phone, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Address.Country, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Address.State, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Address.City, validation.Required),
				validation.Field(&ppr.Items[i].Seller.Address.ZipCode, validation.Required),
			)
			if errPaymentRequestItemsSellerAddress != nil {
				errValidation = append(errValidation, errPaymentRequestItemsSellerAddress.Error())
			}

			errPaymentRequestItemsPrice := validation.ValidateStruct(&ppr.Items[i].Price,
				validation.Field(&ppr.Items[i].Price.Unit, validation.Required),
				validation.Field(&ppr.Items[i].Price.Total, validation.Required),
				validation.Field(&ppr.Items[i].Price.Payable, validation.Required),
				validation.Field(&ppr.Items[i].Price.Discount, validation.Required),
				validation.Field(&ppr.Items[i].Price.SellerCommission, validation.Required),
			)
			if errPaymentRequestItemsPrice != nil {
				errValidation = append(errValidation, errPaymentRequestItemsPrice.Error())
			}

			errPaymentRequestItemsShipment := validation.ValidateStruct(&ppr.Items[i].Shipment,
				validation.Field(&ppr.Items[i].Shipment.ProviderName, validation.Required),
				validation.Field(&ppr.Items[i].Shipment.ReactionTime, validation.Required),
				validation.Field(&ppr.Items[i].Shipment.ShippingTime, validation.Required),
				validation.Field(&ppr.Items[i].Shipment.ReturnTime, validation.Required),
				validation.Field(&ppr.Items[i].Shipment.ShipmentDetail, validation.Required),
			)
			if errPaymentRequestItemsShipment != nil {
				errValidation = append(errValidation, errPaymentRequestItemsShipment.Error())
			}
		}
	}

	res := strings.Join(errValidation, " ")
	// return nil
	if res == "" {
		return nil
	}
	return errors.New(res)
}
