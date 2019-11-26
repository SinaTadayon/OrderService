package entities

import "time"

type PackageItem struct {
	Id           uint64         `bson:"id"`
	Version      uint64         `bson:"version"`
	Invoice      PackageInvoice `bson:"invoice"`
	SellerInfo   SellerInfo     `bson:"sellerInfo"`
	ShipmentSpec ShipmentSpec   `bson:"shipmentSpec"`
	Subpackages  []Subpackage   `bson:"subpackages"`
	Status       string         `bson:"status"`
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
