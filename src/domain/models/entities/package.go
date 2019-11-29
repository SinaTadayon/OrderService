package entities

import "time"

type PackageItem struct {
	SellerId     uint64         `bson:"sellerId"`
	OrderId      uint64         `bson:"orderId"`
	Version      uint64         `bson:"version"`
	Invoice      PackageInvoice `bson:"invoice"`
	SellerInfo   *SellerProfile `bson:"sellerInfo"`
	ShipmentSpec ShipmentSpec   `bson:"shipmentSpec"`
	Subpackages  []Subpackage   `bson:"subpackages"`
	Status       string         `bson:"status"`
	CreatedAt    time.Time      `bson:"createdAt"`
	UpdatedAt    time.Time      `bson:"updatedAt"`
	DeletedAt    *time.Time     `bson:"deletedAt"`
}

type PackageInvoice struct {
	Subtotal       uint64 `bson:"subtotal"`
	Discount       uint64 `bson:"discount"`
	ShipmentAmount uint64 `bson:"shipmentAmount"`
}

// Time unit hours
type ShipmentSpec struct {
	CarrierNames   []string `bson:"carrierNames"`
	CarrierProduct string   `bson:"carrierProduct"`
	CarrierType    string   `bson:"carrierType"`
	ShippingCost   uint64   `bson:"shippingCost"`
	VoucherAmount  uint64   `bson:"voucherAmount"`
	Currency       string   `bson:"currency"`
	ReactionTime   int32    `bson:"reactionTime"`
	ShippingTime   int32    `bson:"shippingTime"`
	ReturnTime     int32    `bson:"returnTime"`
	Details        string   `bson:"Details"`
}
