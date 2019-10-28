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
	Voucher 			*Voucher			`bson:"voucher"`
}

type Voucher struct {
	Amount			uint64					`bson:"amount"`
	Code 			string					`bson:"code"`
	Details			*VoucherDetails			`bson:"details"`
}


// TODO will be complete
type VoucherDetails struct {

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