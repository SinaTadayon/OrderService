package main

import "github.com/Shopify/sarama"

func PaymentPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}

func PaymentPendingAction(message *sarama.ConsumerMessage) error {

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
