package entities

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"strconv"
	"time"
)

const (
	randomMin int = 100
	randomMax int = 999
)

func init () {
	rand.Seed(time.Now().UnixNano())
}

//type ObjectId struct {
//	ID 	primitive.ObjectID `bson:"_id"`
//}

type Order struct {
	ID           			primitive.ObjectID  `bson:"-"`
	OrderId        			string				`bson:"orderId"`
	PaymentService 			[]PaymentService	`bson:"paymentService"`
	SystemPayment  			SystemPayment		`bson:"systemPayment"`
	BuyerInfo          		BuyerInfo			`bson:"buyerInfo"`
	Amount         			Amount				`bson:"amount"`
	Items          			[]Item				`bson:"items"`
	CreatedAt      			time.Time			`bson:"createdAt"`
	UpdatedAt      			time.Time			`bson:"updatedAt"`
	DeletedAt      			*time.Time			`bson:"deletedAt"`
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

type PaymentService struct {
	PaymentRequest  PaymentRequest  		`bson:"paymentRequest"`
	PaymentResponse PaymentResponse 		`bson:"paymentResponse"`
	PaymentResult   PaymentResult   		`bson:"paymentResult"`
}


// TODO get configs of pay to market from siavash
type SystemPayment struct {
	PayToBuyerInfo  []PayToBuyerInfo		`bson:"payToBuyerInfo"`
	PayToSellerInfo []PayToSellerInfo		`bson:"payToSellerInfo"`
	PayToMarket 	[]PayToMarket			`bson:"payToMarket"`
}

type PayToBuyerInfo struct {
	PaymentRequest  PaymentRequest  		`bson:"paymentRequest"`
	PaymentResponse PaymentResponse 		`bson:"paymentResponse"`
	PaymentResult   PaymentResult   		`bson:"paymentResult"`
}

type PayToSellerInfo struct {
	PaymentRequest  PaymentRequest  		`bson:"paymentRequest"`
	PaymentResponse PaymentResponse 		`bson:"paymentResponse"`
	PaymentResult   PaymentResult   		`bson:"paymentResult"`
}

type PayToMarket struct {
	PaymentRequest  PaymentRequest  		`bson:"paymentRequest"`
	PaymentResponse PaymentResponse 		`bson:"paymentResponse"`
	PaymentResult   PaymentResult   		`bson:"paymentResult"`
}

type Amount struct {
	Total    			uint64				`bson:"total"`
	Payable  			uint64				`bson:"payable"`
	Discount 			uint64				`bson:"discount"`
	ShipmentTotal		uint64				`bson:"shipmentTotal"`
	Currency 			string				`bson:"currency"`
	PaymentMethod		string				`bson:"paymentMethod"`
	PaymentOption		string				`bson:"paymentOption"`
	Voucher 			Voucher				`bson:"voucher"`
}

type Voucher struct {
	Amount			uint64					`bson:"amount"`
	Code 			string					`bson:"code"`
	Details			VoucherDetails			`bson:"details"`
}


// TODO will be complete
type VoucherDetails struct {

}

type Item struct {
	OrderItemId 		string 				`bson:"orderItemId"`
	InventoryId 		string 				`bson:"inventoryId"`
	Title       		string 				`bson:"title"`
	Quantity        	int32            	`bson:"quantity"`
	Brand           	string          	`bson:"brand"`
	Warranty        	string          	`bson:"warranty"`
	Categories      	string          	`bson:"categories"`
	Image           	string          	`bson:"image"`
	Returnable      	bool            	`bson:"returnable"`
	Attributes      	Attributes      	`bson:"attributes"`
	DeletedAt       	*time.Time      	`bson:"deletedAt"`
	BuyerInfo       	BuyerInfo       	`bson:"buyerInfo"`
	SellerInfo      	SellerInfo      	`bson:"sellerInfo"`
	PriceInfo       	PriceInfo       	`bson:"priceInfo"`
	ShipmentSpec    	ShipmentSpec    	`bson:"shipmentSpec"`
	ShipmentDetails 	ShipmentDetails 	`bson:"shipmentDetails"`
	OrderStep       	OrderStep       	`bson:"orderStep"`
}

type Attributes struct {
	Width 				string				`bson:"with"`
	Height				string				`bson:"height"`
	Length				string				`bson:"length"`
	Weight				string				`bson:"weight"`
	Color 				string				`bson:"color"`
	Materials			string				`bson:"materials"`
	Extra				ExtraAttributes		`bson:"extra"`
}

// TODO will be complete
type ExtraAttributes struct {

}

// TODO check with nasser for buyerId
type BuyerInfo struct {
	FirstName  			string					`bson:"firstName"`
	LastName   			string					`bson:"lastName"`
	Mobile     			string					`bson:"mobile"`
	Email      			string					`bson:"email"`
	NationalId 			string					`bson:"nationalId"`
	Gender				string					`bson:"gender"`
	IP         			string					`bson:"ip"`
	FinanceInfo    		FinanceInfo				`bson:"financeInfo"`
	ShippingAddress    	AddressInfo				`bson:"shippingAddress"`
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

func (order Order) IsIdEmpty() bool {
	for _, v := range order.ID {
		if v != 0 {
			return false
		}
	}
	return true
}

// TODO concurrency check
func GenerateOrderId() string {
	var err error
	var bytes []byte
	var orderId uint32
	for {
		bytes, err = uuid.New().MarshalBinary()
		if err == nil {
			orderId = byteToHash(bytes)
			break
		}
	}
	return strconv.FormatUint(uint64(orderId), 10)
}

func byteToHash(bytes []byte) uint32 {
	var h uint32 = 0
	for _, val := range bytes {
		h = 31 * h + uint32(val & 0xff)
	}
	return h
}

func GenerateRandomNumber() uint32 {
	return uint32(rand.Intn(randomMax - randomMin + 1) + randomMin)
}