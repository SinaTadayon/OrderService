package entities

import "time"

type Package struct {
	PkgId        uint64         `bson:"pkgId"`
	Version      uint64         `bson:"version"`
	Invoice      PackageInvoice `bson:"invoice"`
	SellerInfo   SellerInfo     `bson:"sellerInfo"`
	ShipmentSpec ShipmentSpec   `bson:"shipmentSpec"`
	Subpackage   []SubPackage   `bson:"subpackage"`
	Status       string         `bson:"status"`
	UpdatedAt    time.Time      `bson:"updatedAt"`
	DeletedAt    *time.Time     `bson:"deletedAt"`
}

type SubPackage struct {
	Version         uint64          `bson:"version"`
	Items           []Item          `bson:"items"`
	ShipmentDetails ShipmentDetails `bson:"shipmentDetails"`
	Progress        Progress        `bson:"progress"`
	Status          string          `bson:"status"`
	UpdatedAt       time.Time       `bson:"updatedAt"`
	DeletedAt       *time.Time      `bson:"deletedAt"`
}

type PackageInvoice struct {
	Subtotal       uint64 `bson:"subtotal"`
	Discount       uint64 `bson:"discount"`
	ShipmentAmount uint64 `bson:"shipmentAmount"`
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

type SellerInfo struct {
	SellerId uint64         `bson:"sellerId"`
	Profile  *SellerProfile `bson:"profile"`
}
