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
	paid     float64
	discount float64
}
type Item struct {
	sku      string
	quantity int32
	seller   ItemSeller
	amount   ItemAmount
	detail   ItemDetail
	shipment ItemShipment
}
type Buyer struct {
	info    BuyerInfo
	finance BuyerFinance
	address BuyerAddress
}
type BuyerInfo struct {
	firstName  string
	lastName   string
	mobile     string
	email      string
	nationalId string
	gender     string
}
type BuyerFinance struct {
	iban       string
	cartNumber string
	bankName   string
}
type BuyerAddress struct {
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
type ItemSeller struct {
	info    ItemSellerInfo
	finance ItemSellerFinance
	address ItemSellerAddress
}
type ItemSellerInfo struct {
	title            string
	firstName        string
	lastName         string
	mobile           string
	email            string
	nationalId       string
	companyName      string
	registrationName string
	economicCode     string
}
type ItemSellerFinance struct {
	iban           string
	cartNumber     string
	bankName       string
	commissionRate float64
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
type ItemAmount struct {
	total            float64
	paid             float64
	discount         float64
	sellerCommission float64
	systemCommission float64
}
type ItemDetail struct {
	description string
	brand       string
	categories  string
}
type ItemShipment struct {
	providerName     string
	reactionTime     string
	shippingTime     string
	returnTime       string
	shipmentFee      float64
	shipmentFeeOwner string
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

	// Validate Buyer Info
	errPaymentRequestBuyerInfo := validation.ValidateStruct(&ppr.buyer.info,
		validation.Field(&ppr.buyer.info.email, validation.Required, is.Email),
		validation.Field(&ppr.buyer.info.nationalId, validation.Required, validation.Length(10, 10)),
		validation.Field(&ppr.buyer.info.mobile, validation.Required),
		validation.Field(&ppr.buyer.info.gender, validation.Required, validation.In("Male", "Female")),
		validation.Field(&ppr.buyer.info.firstName, validation.Required),
		validation.Field(&ppr.buyer.info.lastName, validation.Required),
	)
	if errPaymentRequestBuyerInfo != nil {
		errValidation = append(errValidation, errPaymentRequestBuyerInfo.Error())
	}

	// Validate Buyer finance
	errPaymentRequestBuyerFinance := validation.ValidateStruct(&ppr.buyer.finance,
		validation.Field(&ppr.buyer.finance.iban, validation.Required),
		validation.Field(&ppr.buyer.finance.bankName, validation.Required),
		validation.Field(&ppr.buyer.finance.cartNumber, validation.Required, validation.Length(16, 16)),
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
		validation.Field(&ppr.amount.paid, validation.Required),
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
			)
			if errPaymentRequestItems != nil {
				errValidation = append(errValidation, errPaymentRequestItems.Error())
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
