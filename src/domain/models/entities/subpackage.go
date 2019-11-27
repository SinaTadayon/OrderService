package entities

import "time"

// subpackage id same as itemId
type Subpackage struct {
	ItemId          uint64          `bson:"itemId"`
	SellerId        uint64          `bson:"sellerId"`
	OrderId         uint64          `bson:"orderId"`
	Version         uint64          `bson:"version"`
	Item            Item            `bson:"item"`
	ShipmentDetails ShipmentDetails `bson:"shipmentDetails"`
	Tracking        Progress        `bson:"tracking"`
	Status          string          `bson:"status"`
	CreatedAt       time.Time       `bson:"createdAt"`
	UpdatedAt       time.Time       `bson:"updatedAt"`
	DeletedAt       *time.Time      `bson:"deletedAt"`
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

type Progress struct {
	StepName     string        `bson:"stepName"`
	StepIndex    int           `bson:"stepIndex"`
	Action       Action        `bson:"action"`
	StepsHistory []StepHistory `bson:"stepsHistory"`
}

type StepHistory struct {
	Name          string    `bson:"name"`
	Index         int       `bson:"index"`
	ActionHistory []Action  `bson:"actionHistory"`
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
}

/*
 Action sample:
	Name: ApprovedAction
	Type: SellerInfoActor
	Data: "sample data"
*/
type Action struct {
	Name      string                 `bson:"name"`
	Type      string                 `bson:"type"`
	Data      map[string]interface{} `bson:"data"`
	Result    string                 `bson:"result"`
	Reason    string                 `bson:"reason"`
	CreatedAt time.Time              `bson:"createdAt"`
}
