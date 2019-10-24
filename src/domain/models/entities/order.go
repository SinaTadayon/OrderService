package entities

import (
	"github.com/google/uuid"
	"strconv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

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
	ShipmentDetail       ShipmentDetail			`bson:"shipmentDetail"`
	ReturnShipmentDetail ReturnShipmentDetail	`bson:"returnShipmentDetail"`
}

type ShipmentDetail struct {
	ShipmentProvider       	string			`bson:"shipmentProvider"`
	ShipmentTrackingNumber 	string			`bson:"shipmentTrackingNumber"`
	Image                  	string			`bson:"image"`
	Description            	string			`bson:"description"`
	CreatedAt      			time.Time		`bson:"createdAt"`
}

type ReturnShipmentDetail struct {
	ShipmentProvider       	string			`bson:"shipmentProvider"`
	ShipmentTrackingNumber 	string			`bson:"shipmentTrackingNumber"`
	Image                  	string			`bson:"image"`
	Description            	string			`bson:"description"`
	CreatedAt      			time.Time		`bson:"createdAt"`
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

type PaymentRequest struct {
	Amount				int64				`bson:"amount"`
	Currency			string				`bson:"currency"`
	Gateway 			string				`bson:"gateway"`
	CreatedAt   		time.Time			`bson:"createdAt"`
}

type PaymentResponse struct {
	Result 				bool				`bson:"result"`
	Reason				string				`bson:"reason"`
	Description 		string				`bson:"description"`
	CallBackUrl			string				`bson:"callbackUrl"`
	InvoiceId			int64				`bson:"invoiceId"`
	PaymentId			string				`bson:"paymentId"`
	CreatedAt   		time.Time			`bson:"createdAt"`
}

type PaymentResult struct {
	Result 				bool				`bson:"result"`
	Reason				string				`bson:"reason"`
	PaymentId  			string				`bson:"paymentId"`
	InvoiceId 			int64				`bson:"invoiceId"`
	Amount    			int64				`bson:"amount"`
	ReqBody   			string				`bson:"reqBody"`
	ResBody   			string				`bson:"resBody"`
	CreatedAt   		time.Time			`bson:"createdAt"`
}

// TODO get configs of pay to market from siavash
type SystemPayment struct {
	PayToBuyerInfo  []PayToBuyerInfo		`bson:"payToBuyerInfo"`
	PayToSellerInfo []PayToSellerInfo		`bson:"payToSellerInfo"`
	PayToMarket []PayToMarket				`bson:"payToMarket"`
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
	Total    			int64					`bson:"total"`
	Payable  			int64					`bson:"payable"`
	Discount 			int64					`bson:"discount"`
	Currency 			string					`bson:"currency"`
}

type Item struct {
	ProductId       	string 					`bson:"productId"`
	Title           	string 					`bson:"title"`
	Quantity        	int    					`bson:"quantity"`
	Brand           	string 					`bson:"brand"`
	Warranty        	string 					`bson:"warranty"`
	Categories      	string 					`bson:"categories"`
	Image           	string 					`bson:"image"`
	Returnable      	bool   					`bson:"returnable"`
	DeletedAt			*time.Time				`bson:"deletedAt"`
	BuyerInfo       	BuyerInfo  				`bson:"buyerInfo"`
	SellerInfo      	SellerInfo 				`bson:"sellerInfo"`
	PriceInfo       	PriceInfo				`bson:"priceInfo"`
	ShipmentSpecInfo    ShipmentSpecInfo		`bson:"shipmentSpecInfo"`
	ShipmentDetails 	ShipmentDetails			`bson:"shipmentDetails"`
	OrderStep       	OrderStep				`bson:"orderStep"`
}

type BuyerInfo struct {
	FirstName  			string					`bson:"firstName"`
	LastName   			string					`bson:"lastName"`
	Mobile     			string					`bson:"mobile"`
	Email      			string					`bson:"email"`
	NationalId 			string					`bson:"nationalId"`
	IP         			string					`bson:"ip"`
	Finance    			FinanceInfo				`bson:"finance"`
	Address    			AddressInfo				`bson:"address"`
}

type FinanceInfo struct {
	Iban 				string					`bson:"iban"`
	CardNumber			string					`bson:"cardNumber"`
	AccountNumber		string					`bson:"accountNumber"`
	BankName			string					`bson:"backName"`
	Gateway				string					`bson:"gateway"`
}

type AddressInfo struct {
	Address 			string					`bson:"address"`
	Phone   			string					`bson:"phone"`
	Country 			string					`bson:"country"`
	City    			string					`bson:"city"`
	State   			string					`bson:"state"`
	Location			Location				`bson:"location"`
	ZipCode 			string					`bson:"zipCode"`
}

type Location struct {
	Type string    				`bson:"type"`
	Coordinates []float64 		`bson:"coordinates"`
}

type SellerInfo struct {
	Title            	string					`bson:"title"`
	FirstName        	string					`bson:"firstName"`
	LastName         	string					`bson:"lastName"`
	Mobile           	string					`bson:"mobile"`	
	Email            	string					`bson:"email"`
	NationalId       	string					`bson:"nationalId"`
	CompanyName      	string					`bson:"companyName"`
	RegistrationName 	string					`bson:"registrationName"`	
	EconomicCode     	string					`bson:"economicCode"`
	Finance          	FinanceInfo				`bson:"finance"`
	Address          	AddressInfo				`bson:"address"`
}

type PriceInfo struct {
	Unit             	int64					`bson:"unit"`
	Total            	int64					`bson:"total"`
	Payable          	int64					`bson:"payable"`
	Discount         	int64					`bson:"discount"`	
	SellerCommission 	int64					`bson:"sellerCommission"`	
	Currency		 	string					`bson:"currency"`
}

// Time unit hours
type ShipmentSpecInfo struct {
	ProviderName   		string					`bson:"providerName"`
	ReactionTime   		int						`bson:"reactionTime"`
	ShippingTime   		int						`bson:"shippingTime"`
	ReturnTime     		int						`bson:"returnTime"`
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

func GenerateOrderId() string {
	var err error
	var bytes []byte
	var orderId uint32
	bytes, err = uuid.New().MarshalBinary()
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
