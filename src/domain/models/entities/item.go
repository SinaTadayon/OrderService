package entities

import "time"

type Item struct {
	ItemId 				string 				`bson:"itemId"`
	InventoryId 		string 				`bson:"inventoryId"`
	Title       		string 				`bson:"title"`
	Brand           	string          	`bson:"brand"`
	Warranty        	string          	`bson:"warranty"`
	Categories      	string          	`bson:"categories"`
	Image           	string          	`bson:"image"`
	Returnable      	bool            	`bson:"returnable"`
	Attributes      	Attributes      	`bson:"attributes"`
	CreatedAt      		time.Time			`bson:"createdAt"`
	UpdatedAt      		time.Time			`bson:"updatedAt"`
	DeletedAt       	*time.Time      	`bson:"deletedAt"`
	BuyerInfo       	BuyerInfo       	`bson:"buyerInfo"`
	SellerInfo      	SellerInfo      	`bson:"sellerInfo"`
	PriceInfo       	PriceInfo       	`bson:"priceInfo"`
	ShipmentSpec    	ShipmentSpec    	`bson:"shipmentSpec"`
	ShipmentDetails 	ShipmentDetails 	`bson:"shipmentDetails"`
	OrderStep       	OrderStep       	`bson:"orderStep"`
}

type Attributes struct {
	Quantity        	int32            	`bson:"quantity"`
	Width 				string				`bson:"with"`
	Height				string				`bson:"height"`
	Length				string				`bson:"length"`
	Weight				string				`bson:"weight"`
	Color 				string				`bson:"color"`
	Materials			string				`bson:"materials"`
	Extra				*ExtraAttributes	`bson:"extra"`
}

// TODO will be complete
type ExtraAttributes struct {

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

type OrderStep struct {
	CurrentName   		string				`bson:"currentName"`
	CurrentIndex		int					`bson:"currentIndex"`
	CurrentState		State				`bson:"currentState"`
	CreatedAt 			time.Time			`bson:"createdAt"`
	StepsHistory   		[]StepHistory		`bson:"stepsHistory"`
}

type StepHistory struct {
	Name    			string				`bson:"name"`
	Index				int					`bson:"index"`
	CreatedAt 			time.Time			`bson:"createdAt"`
	StatesHistory		[]State				`bson:"statesHistory"`
}

type State struct {
	Name         		string				`bson:"name"`
	Index        		int					`bson:"index"`
	Action       		Action				`bson:"action"`
	ActionResult 		bool				`bson:"actionResult"`
	Reason       		string				`bson:"reason"`
	CreatedAt    		time.Time			`bson:"createdAt"`
}

/*
 Action sample:
	Name: ApprovedAction
	Type: SellerInfoActor
	Base: ActorAction
	Data: "sample data"
	DispatchedTime: dispatched timestamp
*/
type Action struct {
	Name				string				`bson:"name"`
	Type 				string				`bson:"type"`
	Base 				string				`bson:"base"`
	Data				string				`bson:"data"`
	DispatchedTime		time.Time			`bson:"DispatchedTime"`
}

type SellerInfo struct {
	SellerId 			string 					`bson:"sellerId"`
	Profile				*SellerProfile			`bson:"profile"`
}

type PriceInfo struct {
	Unit             	uint64					`bson:"unit"`
	Total            	uint64					`bson:"total"`
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

