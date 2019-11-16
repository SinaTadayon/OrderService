package entities

import (
	"time"
)

type Item struct {
	ItemId          string            `bson:"itemId"`
	InventoryId     string            `bson:"inventoryId"`
	Title           string            `bson:"title"`
	Brand           string            `bson:"brand"`
	Guaranty        string            `bson:"guaranty"`
	Category        string            `bson:"category"`
	Image           string            `bson:"image"`
	Returnable      bool              `bson:"returnable"`
	Status          string            `bson:"status"`
	Quantity        int32             `bson:"quantity"`
	Attributes      map[string]string `bson:"attributes"`
	CreatedAt       time.Time         `bson:"createdAt"`
	UpdatedAt       time.Time         `bson:"updatedAt"`
	DeletedAt       *time.Time        `bson:"deletedAt"`
	SellerInfo      SellerInfo        `bson:"sellerInfo"`
	Price           Price             `bson:"price"`
	ShipmentSpec    ShipmentSpec      `bson:"shipmentSpec"`
	ShipmentDetails ShipmentDetails   `bson:"shipmentDetails"`
	Progress        Progress          `bson:"progress"`
}

type ShipmentDetails struct {
	SellerShipmentDetail      ShipmentDetail `bson:"sellerShipmentDetail"`
	BuyerReturnShipmentDetail ShipmentDetail `bson:"buyerReturnShipmentDetail"`
}

type ShipmentDetail struct {
	CarrierName    string    `bson:"carrierName"`
	ShippingMethod string    `bson:"shippingMethod"`
	TrackingNumber string    `bson:"trackingNumber"`
	Image          string    `bson:"image"`
	Description    string    `bson:"description"`
	CreatedAt      time.Time `bson:"createdAt"`
}

type Progress struct {
	CurrentStepName  string `bson:"currentStepName"`
	CurrentStepIndex int    `bson:"currentStepIndex"`
	//CurrentState		State				`bson:"currentState"`
	CreatedAt    time.Time     `bson:"createdAt"`
	StepsHistory []StepHistory `bson:"stepsHistory"`
}

type StepHistory struct {
	Name          string    `bson:"name"`
	Index         int       `bson:"index"`
	CreatedAt     time.Time `bson:"createdAt"`
	ActionHistory []Action  `bson:"actionHistory"`
	//StatesHistory		[]StateHistory		`bson:"statesHistory"`
}

type StateHistory struct {
	Name      string    `bson:"name"`
	Index     int       `bson:"index"`
	Type      string    `bson:"type"`
	Action    Action    `bson:"action"`
	Result    bool      `bson:"result"`
	Reason    string    `bson:"reason"`
	CreatedAt time.Time `bson:"createdAt"`
}

type State struct {
	Name           string    `bson:"name"`
	Index          int       `bson:"index"`
	Type           string    `bson:"type"`
	Actions        []Action  `bson:"actions"`
	AcceptedAction Action    `bson:"acceptedAction"`
	Result         bool      `bson:"actionResult"`
	Reason         string    `bson:"reason"`
	CreatedAt      time.Time `bson:"createdAt"`
}

/*
 Action sample:
	Name: ApprovedAction
	Type: SellerInfoActor
	Base: ActorAction
	Data: "sample data"
	Time: dispatched timestamp
*/
type Action struct {
	Name      string                 `bson:"name"`
	Type      string                 `bson:"type"`
	Base      string                 `bson:"base"`
	Data      map[string]interface{} `bson:"data"`
	Result    bool                   `bson:"result"`
	Reason    string                 `bson:"reason"`
	Time      *time.Time             `bson:"time"`
	CreatedAt time.Time              `bson:"createdAt"`
}

type SellerInfo struct {
	SellerId string         `bson:"sellerId"`
	Profile  *SellerProfile `bson:"profile"`
}

type Price struct {
	Unit             uint64  `bson:"unit"`
	Total            uint64  `bson:"total"`
	Original         uint64  `bson:"original"`
	Special          uint64  `bson:"special"`
	Discount         uint64  `bson:"discount"`
	SellerCommission float32 `bson:"sellerCommission"`
	Currency         string  `bson:"currency"`
}

// Time unit hours
type ShipmentSpec struct {
	CarrierName    string `bson:"carrierName"`
	CarrierProduct string `bson:"carrierProduct"`
	CarrierType    string `bson:"carrierType"`
	ShippingCost   uint64 `bson:"shippingCost"`
	VoucherAmount  uint64 `bson:"voucherAmount"`
	Currency       string `bson:"currency"`
	ReactionTime   int32  `bson:"reactionTime"`
	ShippingTime   int32  `bson:"shippingTime"`
	ReturnTime     int32  `bson:"returnTime"`
	Details        string `bson:"Details"`
}
