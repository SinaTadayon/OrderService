package main

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-ozzo/ozzo-validation/is"

	"github.com/Shopify/sarama"
	validation "github.com/go-ozzo/ozzo-validation"
)

func PaymentPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	var ppr = PaymentPendingRequest{}

	err := json.Unmarshal(message.Value, &ppr)
	if err != nil {
		return nil, err
	}
	err = ppr.validate()
	if err != nil {
		return nil, err
	}

	return message, nil
}

func PaymentPendingAction(message *sarama.ConsumerMessage) error {
	// calculate price and send to payment
	err := PaymentPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func PaymentPendingProduce(topic string, payload []byte) error {
	return nil
}

type PaymentPendingRequest struct {
	orderNumber string
	buyer       Buyer
	amount      Amount
	items       []Item
}
type Amount struct {
	total    float64
	payable  float64
	discount float64
}
type Item struct {
	sku        string
	title      string
	quantity   int32
	brand      string
	warranty   string
	categories string
	seller     ItemSeller
	price      ItemPrice
	shipment   ItemShipment
}
type Buyer struct {
	firstName  string
	lastName   string
	mobile     string
	email      string
	nationalId string
	finance    BuyerFinance
	address    BuyerAddress
}
type BuyerFinance struct {
	iban string
}
type BuyerAddress struct {
	address string
	phone   string
	country string
	city    string
	state   string
	lat     string
	lan     string
	zipCode string
}
type ItemSeller struct {
	title            string
	firstName        string
	lastName         string
	mobile           string
	email            string
	nationalId       string
	companyName      string
	registrationName string
	economicCode     string
	finance          ItemSellerFinance
	address          ItemSellerAddress
}
type ItemSellerFinance struct {
	iban string
}
type ItemSellerAddress struct {
	title   string
	address string
	phone   string
	country string
	city    string
	state   string
	lat     string
	lan     string
	zipCode string
}
type ItemPrice struct {
	unit             float64
	total            float64
	payable          float64
	discount         float64
	sellerCommission float64
}
type ItemShipment struct {
	providerName   string
	reactionTime   int32
	shippingTime   int32
	returnTime     int32
	shipmentDetail string
}

func (ppr *PaymentPendingRequest) validate() error {
	var errValidation []string
	// Validate order number
	errPaymentRequest := validation.ValidateStruct(ppr,
		validation.Field(&ppr.orderNumber, validation.Required, validation.Length(5, 250)),
	)
	if errPaymentRequest != nil {
		errValidation = append(errValidation, errPaymentRequest.Error())
	}

	// Validate Buyer
	errPaymentRequestBuyer := validation.ValidateStruct(&ppr.buyer,
		validation.Field(&ppr.buyer.firstName, validation.Required),
		validation.Field(&ppr.buyer.lastName, validation.Required),
		validation.Field(&ppr.buyer.email, validation.Required, is.Email),
		validation.Field(&ppr.buyer.nationalId, validation.Required, validation.Length(10, 10)),
		validation.Field(&ppr.buyer.mobile, validation.Required),
	)
	if errPaymentRequestBuyer != nil {
		errValidation = append(errValidation, errPaymentRequestBuyer.Error())
	}

	// Validate Buyer finance
	errPaymentRequestBuyerFinance := validation.ValidateStruct(&ppr.buyer.finance,
		validation.Field(&ppr.buyer.finance.iban, validation.Required, validation.Length(26, 26)),
	)
	if errPaymentRequestBuyerFinance != nil {
		errValidation = append(errValidation, errPaymentRequestBuyerFinance.Error())
	}

	// Validate Buyer address
	errPaymentRequestBuyerAddress := validation.ValidateStruct(&ppr.buyer.address,
		validation.Field(&ppr.buyer.address.address, validation.Required),
		validation.Field(&ppr.buyer.address.state, validation.Required),
		validation.Field(&ppr.buyer.address.city, validation.Required),
		validation.Field(&ppr.buyer.address.country, validation.Required),
		validation.Field(&ppr.buyer.address.zipCode, validation.Required),
		validation.Field(&ppr.buyer.address.phone, validation.Required),
	)
	if errPaymentRequestBuyerAddress != nil {
		errValidation = append(errValidation, errPaymentRequestBuyerAddress.Error())
	}

	// Validate amount
	errPaymentRequestAmount := validation.ValidateStruct(&ppr.amount,
		validation.Field(&ppr.amount.total, validation.Required),
		validation.Field(&ppr.amount.discount, validation.Required),
		validation.Field(&ppr.amount.payable, validation.Required),
	)
	if errPaymentRequestAmount != nil {
		errValidation = append(errValidation, errPaymentRequestAmount.Error())
	}

	if len(ppr.items) != 0 {
		for i := range ppr.items {
			// Validate amount
			errPaymentRequestItems := validation.ValidateStruct(&ppr.items[i],
				validation.Field(&ppr.items[i].sku, validation.Required),
				validation.Field(&ppr.items[i].quantity, validation.Required),
				validation.Field(&ppr.items[i].title, validation.Required),
				validation.Field(&ppr.items[i].categories, validation.Required),
				validation.Field(&ppr.items[i].brand, validation.Required),
			)
			if errPaymentRequestItems != nil {
				errValidation = append(errValidation, errPaymentRequestItems.Error())
			}

			errPaymentRequestItemsSeller := validation.ValidateStruct(&ppr.items[i].seller,
				validation.Field(&ppr.items[i].seller.title, validation.Required),
				validation.Field(&ppr.items[i].seller.firstName, validation.Required),
				validation.Field(&ppr.items[i].seller.lastName, validation.Required),
				validation.Field(&ppr.items[i].seller.mobile, validation.Required),
				validation.Field(&ppr.items[i].seller.email, validation.Required),
			)
			if errPaymentRequestItemsSeller != nil {
				errValidation = append(errValidation, errPaymentRequestItemsSeller.Error())
			}

			errPaymentRequestItemsSellerFinance := validation.ValidateStruct(&ppr.items[i].seller.finance,
				validation.Field(&ppr.items[i].seller.finance.iban, validation.Required),
			)
			if errPaymentRequestItemsSellerFinance != nil {
				errValidation = append(errValidation, errPaymentRequestItemsSellerFinance.Error())
			}

			errPaymentRequestItemsSellerAddress := validation.ValidateStruct(&ppr.items[i].seller.address,
				validation.Field(&ppr.items[i].seller.address.title, validation.Required),
				validation.Field(&ppr.items[i].seller.address.address, validation.Required),
				validation.Field(&ppr.items[i].seller.address.phone, validation.Required),
				validation.Field(&ppr.items[i].seller.address.country, validation.Required),
				validation.Field(&ppr.items[i].seller.address.state, validation.Required),
				validation.Field(&ppr.items[i].seller.address.city, validation.Required),
				validation.Field(&ppr.items[i].seller.address.zipCode, validation.Required),
			)
			if errPaymentRequestItemsSellerAddress != nil {
				errValidation = append(errValidation, errPaymentRequestItemsSellerAddress.Error())
			}

			errPaymentRequestItemsPrice := validation.ValidateStruct(&ppr.items[i].price,
				validation.Field(&ppr.items[i].price.unit, validation.Required),
				validation.Field(&ppr.items[i].price.total, validation.Required),
				validation.Field(&ppr.items[i].price.payable, validation.Required),
				validation.Field(&ppr.items[i].price.discount, validation.Required),
				validation.Field(&ppr.items[i].price.sellerCommission, validation.Required),
			)
			if errPaymentRequestItemsPrice != nil {
				errValidation = append(errValidation, errPaymentRequestItemsPrice.Error())
			}

			errPaymentRequestItemsShipment := validation.ValidateStruct(&ppr.items[i].shipment,
				validation.Field(&ppr.items[i].shipment.providerName, validation.Required),
				validation.Field(&ppr.items[i].shipment.reactionTime, validation.Required),
				validation.Field(&ppr.items[i].shipment.shippingTime, validation.Required),
				validation.Field(&ppr.items[i].shipment.returnTime, validation.Required),
				validation.Field(&ppr.items[i].shipment.shipmentDetail, validation.Required),
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
