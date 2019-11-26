package entities

import "time"

// subpackage id same as itemId
type Subpackage struct {
	Id              uint64          `bson:"id"`
	Version         uint64          `bson:"version"`
	Item            Item            `bson:"item"`
	ShipmentDetails ShipmentDetails `bson:"shipmentDetails"`
	Tracking        Progress        `bson:"tracking"`
	Status          string          `bson:"status"`
	CreatedAt       time.Time       `bson:"createdAt"`
	UpdatedAt       time.Time       `bson:"updatedAt"`
	//DeletedAt       *time.Time      `bson:"deletedAt"`
}

type ShipmentDetails struct {
	ShipmentDetail       ShipmentDetail `bson:"ShipmentDetail"`
	ReturnShipmentDetail ShipmentDetail `bson:"ReturnShipmentDetail"`
}

type ShipmentDetail struct {
	CarrierName    string    `bson:"carrierName"`
	ShippingMethod string    `bson:"shippingMethod"`
	TrackingNumber string    `bson:"trackingNumber"`
	Image          string    `bson:"image"`
	Description    string    `bson:"description"`
	ShippedDate    time.Time `bson:"shippedDate"`
	CreatedAt      time.Time `bson:"createdAt"`
}

type Item struct {
	InventoryId string            `bson:"inventoryId"`
	Title       string            `bson:"title"`
	Brand       string            `bson:"brand"`
	Guaranty    string            `bson:"guaranty"`
	Category    string            `bson:"category"`
	Image       string            `bson:"image"`
	Returnable  bool              `bson:"returnable"`
	Quantity    int32             `bson:"quantity"`
	Attributes  map[string]string `bson:"attributes"`
	CreatedAt   time.Time         `bson:"createdAt"`
	DeletedAt   *time.Time        `bson:"deletedAt"`
	Invoice     ItemInvoice       `bson:"invoice"`
}

type ItemInvoice struct {
	Unit              uint64  `bson:"unit"`
	Total             uint64  `bson:"total"`
	Original          uint64  `bson:"original"`
	Special           uint64  `bson:"special"`
	Discount          uint64  `bson:"discount"`
	SellerCommission  float32 `bson:"sellerCommission"`
	Currency          string  `bson:"currency"`
	ApplicableVoucher bool    `bson:"applicableVoucher"`
}
