package entities

import (
	"time"
)

type Item struct {
	ItemId      string            `bson:"itemId"`
	InventoryId string            `bson:"inventoryId"`
	Title       string            `bson:"title"`
	Brand       string            `bson:"brand"`
	Guarantee   string            `bson:"guarantee"`
	Categories  string            `bson:"categories"`
	Image       string            `bson:"image"`
	Returnable      bool              `bson:"returnable"`
	Status          string            `bson:"status"`
	Attributes      map[string]string `bson:"attributes"`
	CreatedAt       time.Time         `bson:"createdAt"`
	UpdatedAt       time.Time         `bson:"updatedAt"`
	DeletedAt       *time.Time        `bson:"deletedAt"`
	SellerInfo      SellerInfo        `bson:"sellerInfo"`
	PriceInfo       PriceInfo         `bson:"priceInfo"`
	ShipmentSpec    ShipmentSpec      `bson:"shipmentSpec"`
	ShipmentDetails ShipmentDetails   `bson:"shipmentDetails"`
	Progress        Progress          `bson:"progress"`
}

type ShipmentDetails struct {
	SellerShipmentDetail       	ShipmentDetail `bson:"sellerShipmentDetail"`
	BuyerReturnShipmentDetail 	ShipmentDetail `bson:"buyerReturnShipmentDetail"`
}

type ShipmentDetail struct {
	CarrierName 	 	string     			`bson:"carrierName"`
	ShippingMethod	 	string 				`bson:"shippingMethod"`
	TrackingNumber   	string    			`bson:"trackingNumber"`
	Image            	string    			`bson:"image"`
	Description      	string    			`bson:"description"`
	CreatedAt        	time.Time 			`bson:"createdAt"`
}

type Progress struct {
	CurrentName   		string				`bson:"currentName"`
	CurrentIndex		int					`bson:"currentIndex"`
	CurrentState		State				`bson:"currentState"`
	ActionHistory		[]Action			`bson:"actionHistory"`
	CreatedAt 			time.Time			`bson:"createdAt"`
	StepsHistory   		[]StepHistory		`bson:"stepsHistory"`
}

type StepHistory struct {
	Name    			string				`bson:"name"`
	Index				int					`bson:"index"`
	CreatedAt 			time.Time			`bson:"createdAt"`
	StatesHistory		[]StateHistory		`bson:"statesHistory"`
}

type StateHistory struct {
	Name         		string				`bson:"name"`
	Index        		int					`bson:"index"`
	Type 				string				`bson:"type"`
	Action      		Action				`bson:"action"`
	Result 				bool				`bson:"result"`
	Reason       		string				`bson:"reason"`
	CreatedAt    		time.Time			`bson:"createdAt"`
}

type State struct {
	Name         		string				`bson:"name"`
	Index        		int					`bson:"index"`
	Type 				string				`bson:"type"`
	Actions				[]Action			`bson:"actions"`
	AcceptedAction      Action				`bson:"acceptedAction"`
	Result 				bool				`bson:"actionResult"`
	Reason       		string				`bson:"reason"`
	CreatedAt    		time.Time			`bson:"createdAt"`
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
	Name				string					`bson:"name"`
	Type 				string					`bson:"type"`
	Base 				string					`bson:"base"`
	Data				map[string]interface{}	`bson:"data"`
	Result 				bool					`bson:"result"`
	Reason       		string					`bson:"reason"`
	Time				*time.Time				`bson:"time"`
	CreatedAt			time.Time				`bson:"createdAt"`
}


type SellerInfo struct {
	SellerId 			string 					`bson:"sellerId"`
	Profile				*SellerProfile			`bson:"profile"`
}

// TODO remove Total
type PriceInfo struct {
	Unit             	uint64					`bson:"unit"`
	Payable          	uint64					`bson:"payable"`
	Discount         	uint64					`bson:"discount"`
	SellerCommission 	uint64					`bson:"sellerCommission"`
	Currency		 	string					`bson:"currency"`
}

// Time unit hours
type ShipmentSpec struct {
	CarrierName			string					`bson:"carrierName"`
	CarrierProduct 		string					`bson:"carrierProduct"`
	CarrierType			string					`bson:"carrierType"`
	ShippingAmount		uint64					`bson:"shippingAmount"`
	VoucherAmount		uint64					`bson:"voucherAmount"`
	Currency			string					`bson:"currency"`
	ReactionTime   		int32					`bson:"reactionTime"`
	ShippingTime   		int32					`bson:"shippingTime"`
	ReturnTime     		int32					`bson:"returnTime"`
	Details 			string					`bson:"Details"`
}

